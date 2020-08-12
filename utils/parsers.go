package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func ParsePortRange(portRangeStr string) (RandomUint32Range, error) {
	parts := strings.Split(portRangeStr, "-")
	if len(parts) != 2 {
		return RandomUint32Range{}, fmt.Errorf("cannot parse port range %s, must be in format <port_start>-<port_end>", portRangeStr)
	}
	p0, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return RandomUint32Range{}, fmt.Errorf("cannot parse start port, must be a number from range 0-65535")
	}
	p1, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return RandomUint32Range{}, fmt.Errorf("cannot parse end port, must be a number from range 0-65535")
	}
	if p0 > p1 {
		return RandomUint32Range{}, fmt.Errorf("start port must be less than end port")
	}
	return NewRandomUint32Range(uint32(p0), uint32(p1)), nil
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
