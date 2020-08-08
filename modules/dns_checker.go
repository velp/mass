package modules

type DNSChecker struct {}

func NewDNSChecker() ModuleInterface {
	return &DNSChecker{}
}

func (d *DNSChecker) Run(goroutines int) {}

func (d *DNSChecker) Stop(wait bool) {}
