package utils

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type PortRange struct {
	collection []int
	current    int
	total      int
}

func (r *PortRange) Next() int {
	r.current++
	if r.current == r.total {
		r.current = 0
	}
	return r.collection[r.current]
}

func ParsePortRange(portRangeStr string) (portRange PortRange, err error) {
	portRange = PortRange{
		current: 0,
		total:   0,
	}
	parts := strings.Split(portRangeStr, "-")
	if len(parts) != 2 {
		err = fmt.Errorf("cannot parse port range %s, must be in format <port_start>-<port_end>", portRangeStr)
		return
	}
	p0, err := strconv.Atoi(parts[0])
	if err != nil {
		err = fmt.Errorf("cannot parse start port, must be a number from range 0-65536")
		return
	}
	p1, err := strconv.Atoi(parts[1])
	if err != nil {
		err = fmt.Errorf("cannot parse end port, must be a number from range 0-65536")
		return
	}
	if p0 > p1 {
		err = fmt.Errorf("start port must be less than end port")
		return
	}
	// Generate list of ports
	ports := make([]int, p1-p0+1)
	for i := range ports {
		ports[i] = p0 + i
	}
	// Shuffle ports
	r := rand.New(rand.NewSource(time.Now().Unix()))
	portRange.collection = make([]int, len(ports))
	perm := r.Perm(len(ports))
	for i, randIndex := range perm {
		portRange.collection[i] = ports[randIndex]
	}
	portRange.total = len(ports)
	return
}

type IPv4Range struct {
	start   net.IP
	end     net.IP
	current net.IP
}

func (r *IPv4Range) Next() net.IP {
	res := r.current
	for i := 0; i < 4; i++ {
		if r.current[15-i] >= r.end[15-i] {
			r.current[15-i] = r.start[15-i]
		} else {
			r.current[15-i]++
			break
		}
	}
	return res
}

func ParseIPv4Range(ipv4Range string) (resRange IPv4Range, err error) {
	var i0, i1 int
	resRange = IPv4Range{
		current: net.ParseIP("0.0.0.0"),
		start:   net.ParseIP("0.0.0.0"),
		end:     net.ParseIP("0.0.0.0"),
	}

	ip := strings.Split(ipv4Range, ".")
	if len(ip) != 4 {
		err = fmt.Errorf("cannot parse IPv4 address range (.): %s", ipv4Range)
		return
	}
	for i := 0; i < 4; i++ {
		if strings.Contains(ip[i], "-") {
			s := strings.Split(ip[i], "-")
			if len(s) != 2 {
				err = fmt.Errorf("cannot parse IPv4 address range (-): %s", ipv4Range)
			}
			i0, err = strconv.Atoi(s[0])
			if err != nil {
				return
			}
			a0 := byte(i0)
			i1, err = strconv.Atoi(s[1])
			if err != nil {
				return
			}
			a1 := byte(i1)
			if a0 < a1 {
				resRange.start[12+i] = a0
				resRange.end[12+i] = a1
			} else {
				resRange.start[12+i] = a1
				resRange.end[12+i] = a0
			}
		} else {
			i0, err = strconv.Atoi(ip[i])
			if err != nil {
				return
			}
			a0 := byte(i0)
			resRange.start[12+i] = a0
			resRange.end[12+i] = a0
		}
	}
	return
}
