package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
)

func DetermineNetworkConfig(dstIP net.IP) (*net.Interface, net.IP, net.HardwareAddr, error) {
	router, err := routing.New()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while creating routing object: %s", err)
	}

	iface, _, srcIP, err := router.Route(dstIP)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error routing to ip %s: %s", dstIP, err)
	}

	dstMAC, err := getDstMACAddress(iface, srcIP, dstIP)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error destination MAC address for %s: %s", dstIP, err)
	}

	return iface, srcIP, dstMAC, nil
}

func getDstMACAddress(iface *net.Interface, srcIP net.IP, dstIP net.IP) (net.HardwareAddr, error) {
	// Open up a pcap handle for packet reads/writes.
	handle, err := pcap.OpenLive(iface.Name, 65536, false, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	// Timeout context for operation (1 minute)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	// Start up a goroutine to catch ARP replies.
	result := make(chan net.HardwareAddr)
	go fetchARPReply(ctx, handle, iface, dstIP, result)
	// Start up ARP request sending
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			// Waiting timeout
			return nil, ctx.Err()
		case hwAddr := <-result:
			return hwAddr, nil
		case <-ticker.C:
			// Send ARP request packets out to the handle
			if err := sendARPRequest(handle, iface, srcIP, dstIP); err != nil {
				return nil, fmt.Errorf("error writing packets on %v: %v", iface.Name, err)
			}
		}
	}
}

// fetchARPReply fetches incoming ARP responses and tries find response from target host to get MAC address
func fetchARPReply(ctx context.Context, handle *pcap.Handle, iface *net.Interface, dstIP net.IP, result chan<- net.HardwareAddr) {
	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()
	for {
		var packet gopacket.Packet
		select {
		case <-ctx.Done():
			// Waiting timeout
			return
		case packet = <-in:
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}
			arp := arpLayer.(*layers.ARP)
			if arp.Operation != layers.ARPReply || bytes.Equal([]byte(iface.HardwareAddr), arp.SourceHwAddress) {
				// This is a packet we sent.
				continue
			}
			if bytes.Equal(arp.SourceProtAddress, []byte(dstIP)) {
				log.Printf("IP %v is at %v", net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress))
				result <- net.HardwareAddr(arp.SourceHwAddress)
				return
			}
		}
	}
}

// sendARPRequest sends an ARP request for the IP address to the pcap handle
func sendARPRequest(handle *pcap.Handle, iface *net.Interface, srcIP net.IP, dstIP net.IP) error {
	// Set up all the layers' fields we can.
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: srcIP,
		DstProtAddress:    dstIP,
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}
	// Set up buffer and options for serialization.
	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	// Send one packet for every address.
	gopacket.SerializeLayers(buffer, opts, &eth, &arp)
	log.Printf("Send ARP request who-has %s tell %s", dstIP, srcIP)
	if err := handle.WritePacketData(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}
