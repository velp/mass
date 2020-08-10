package modules

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type DNSChecker struct {
	wg        *sync.WaitGroup
	closer    chan struct{}
	domain    string
	dstIP     net.IP
	totalReq  uint64
	failedReq uint64
}

func NewDNSChecker(dstIP net.IP, domain string) ModuleInterface {
	return &DNSChecker{
		wg:     new(sync.WaitGroup),
		domain: domain,
		dstIP:  dstIP,
	}
}

func (d *DNSChecker) Run(goroutines int) {
	d.closer = make(chan struct{})
	for i := 0; i < goroutines; i++ {
		d.wg.Add(1)
		go d.send(i)
	}
}

func (d *DNSChecker) Stop(wait bool) {
	close(d.closer)
	if wait {
		d.wg.Wait()
	}
	log.Printf("Total requests: %d failed: %d success: %d", d.totalReq, d.failedReq, d.totalReq-d.failedReq)
}

func (d *DNSChecker) send(idx int) {
	defer d.wg.Done()
	log.Printf("DNS checker #%v started", idx)
	defer log.Printf("DNS checker #%v stoped", idx)
	ticker := time.NewTicker(1 * time.Second)
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dealer := net.Dialer{
				Timeout: 3 * time.Second,
			}
			return dealer.DialContext(ctx, "udp", d.dstIP.String()+":53")
		},
	}
	var startTime time.Time
	for {
		select {
		case <-d.closer:
			return
		case <-ticker.C:
			atomic.AddUint64(&d.totalReq, 1)
			startTime = time.Now()
			ips, err := resolver.LookupIPAddr(context.Background(), d.domain)
			if err != nil {
				log.Printf("could not get IP addresses for domain %s: %s (request took: %s)", d.domain, err, time.Since(startTime))
				atomic.AddUint64(&d.failedReq, 1)
				continue
			}
			if len(ips) == 0 {
				log.Printf("domain %s doesn't have IP address (request took: %s)", d.domain, time.Since(startTime))
				atomic.AddUint64(&d.failedReq, 1)
				continue
			}
			ipsStr := ""
			for _, ip := range ips {
				ipsStr += ip.String() + " "
			}
			log.Printf("ip addresses: %s (request took: %s)", ipsStr, time.Since(startTime))
		}
	}
}
