package dns

import (
	"errors"
	"net"
)

func HostnameToIp(hostname string) (net.IP, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, nil
		}
	}
	return nil, errors.New("no IPv4 address found")
}
