package modules

import (
	"log"
	"net"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/velp/mass/utils"
)

type TCPsynSender struct {
	wg               *sync.WaitGroup
	closer           chan struct{}
	netDevice        string
	srcIPRange       utils.IPv4Range
	srcPortRange     utils.RandomUint32Range
	ipHdrIDGenerator utils.RandomUint32Range
	tcpSeqGenerator  utils.RandomUint32Range
	dstIP            net.IP
	srcMAC           net.HardwareAddr
	dstMAC           net.HardwareAddr
}

func NewTCPsynSender(iface *net.Interface, srcIPRange utils.IPv4Range, srcPortRange utils.RandomUint32Range, dstMAC net.HardwareAddr, dstIP net.IP) ModuleInterface {
	return &TCPsynSender{
		wg:               new(sync.WaitGroup),
		netDevice:        iface.Name,
		srcIPRange:       srcIPRange,
		srcPortRange:     srcPortRange,
		ipHdrIDGenerator: utils.NewRandomUint32Range(0, 65535),
		tcpSeqGenerator:  utils.NewRandomUint32Range(1010825923, 1010891458),
		dstIP:            dstIP,
		srcMAC:           iface.HardwareAddr,
		dstMAC:           dstMAC,
	}
}

func (d *TCPsynSender) Run(goroutines int) {
	d.closer = make(chan struct{})
	for i := 0; i < goroutines; i++ {
		d.wg.Add(1)
		go d.send(i)
	}
}

func (d *TCPsynSender) Stop(wait bool) {
	close(d.closer)
	if wait {
		d.wg.Wait()
	}
}

func (d *TCPsynSender) send(idx int) {
	defer d.wg.Done()
	log.Printf("TCP SYN sender #%v started", idx)
	defer log.Printf("TCP SYN sender #%v stoped", idx)
	// Generators
	srcPortGenerator := d.srcPortRange
	srcIPGenerator := d.srcIPRange
	ipHdrIDGenerator := d.ipHdrIDGenerator
	tcpSeqGenerator := d.tcpSeqGenerator
	// Prepare static parts of packet and options
	serializeOpts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	ethernetLayer := &layers.Ethernet{
		SrcMAC:       d.srcMAC,
		DstMAC:       d.dstMAC,
		EthernetType: 0x800,
	}
	ipv4Layer := &layers.IPv4{
		Version:  4,
		TTL:      64,
		Id:       uint16(ipHdrIDGenerator.Next()),
		Protocol: layers.IPProtocolTCP,
		DstIP:    d.dstIP,
	}
	tcpLayer := &layers.TCP{
		DstPort: layers.TCPPort(53),
		SYN:     true,
		Seq:     tcpSeqGenerator.Next(),
		Window:  1024,
	}
	// Open device
	handle, err := pcap.OpenLive(d.netDevice, 65536, false, pcap.BlockForever)
	if err != nil {
		log.Printf("opening device %s failed: %s", d.netDevice, err)
		return
	}
	defer handle.Close()
	for {
		select {
		case <-d.closer:
			return
		default:
			// Change IPv4 layer
			ipv4Layer.SrcIP = srcIPGenerator.Next()
			// Change TCP layer
			tcpLayer.SrcPort = layers.TCPPort(srcPortGenerator.Next())
			tcpLayer.SetNetworkLayerForChecksum(ipv4Layer)
			// Serialize
			buffer := gopacket.NewSerializeBuffer()
			if err := gopacket.SerializeLayers(buffer, serializeOpts, ethernetLayer, ipv4Layer, tcpLayer); err != nil {
				log.Printf("packet preparation failed: %s", err)
			}
			// Send
			err = handle.WritePacketData(buffer.Bytes())
			if err != nil {
				log.Printf("sending packet failed: %s", err)
			}
		}
	}
}
