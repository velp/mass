package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/velp/mass/modules"
	"github.com/velp/mass/utils"
)

var (
	srcIPRangeArg   string
	srcPortRangeArg string
	goroutinesArg   int
	dstIPArg        string
	dnsDomainArg    string
	moduleArg       string
)

func init() {
	flag.StringVar(&srcIPRangeArg, "src-ip-range", "", "Sets source IP range for spoofing in format 192.168.10-40.1-255. (default: ip address from interface)")
	flag.StringVar(&srcPortRangeArg, "src-port-range", "30000-65536", "Sets source port range.")
	flag.IntVar(&goroutinesArg, "goroutines", 10, "Number of goroutines to generate traffic.")
	flag.StringVar(&dstIPArg, "dst-ip", "", "Target IP address.")
	flag.StringVar(&dnsDomainArg, "dns-domain", "example.com", "Domain which will be used in DNS A query. Masked part (*) will be randomized.")
	flag.StringVar(&moduleArg, "module", "dns-flooder", "Module to run tests. Supported modules: dns-flooder, dns-checker.")
}

func main() {
	flag.Parse()
	if dstIPArg == "" {
		log.Fatalf("argument -dst-ip is required")
	}
	var dstIP net.IP
	if dstIP = net.ParseIP(dstIPArg).To4(); dstIP == nil {
		log.Fatalf("error as non-ip target %s is passed", dstIP)
	}
	if dnsDomainArg == "" {
		log.Fatalf("argument -dns-domain can't be empty")
	}
	iface, ifaceSrcIP, dstMAC, err := utils.DetermineNetworkConfig(dstIP)
	if err != nil {
		log.Fatalf("cannot determine network data: %s", err)
	}
	srcIPRangeStr := ifaceSrcIP.String()
	if srcIPRangeArg != "" {
		srcIPRangeStr = srcIPRangeArg
	}
	srcIPRange, err := utils.ParseIPv4Range(srcIPRangeStr)
	if err != nil {
		log.Fatal(err)
	}
	srcPortRange, err := utils.ParsePortRange(srcPortRangeArg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Network data:\n\tSource: %s (%s, real IP %s) from %s\n\tDestination: %s (%s)", srcIPRangeStr, iface.HardwareAddr, ifaceSrcIP, iface.Name, dstIP, dstMAC.String())

	var mod modules.ModuleInterface
	switch moduleArg {
	case "dns-flooder":
		mod = modules.NewDNSFlooder(iface, srcIPRange, srcPortRange, dstMAC, dstIP, dnsDomainArg)
	case "dns-checker":
		mod = modules.NewDNSChecker(dstIP, dnsDomainArg)
	default:
		log.Fatalf("unsupported module %s", moduleArg)
	}
	mod.Run(goroutinesArg)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	log.Printf("Signal received attempting")

	mod.Stop(true)
}
