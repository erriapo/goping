// Copyright 2018 Gavin Chun Jin. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package core provides utilities
package core

import (
	"net"
	"strings"
)

func is4(address string) bool {
	return strings.Count(address, ":") < 2
}

func is6(address string) bool {
	return !is4(address)
}

func ipType(ip string) string {
	if is4(ip) {
		return "ipv4"
	} else {
		return "ipv6"
	}
}

// ScanInterfaces returns a map from the interface name to an IPv4 address
func ScanInterfaces() (map[string]net.IP, error) {
	retval := make(map[string]net.IP)

	interfaces, err := net.Interfaces()
	if err != nil {
		return retval, err
	}

	var addresses []net.Addr
	for _, k := range interfaces {
		// TODO skip if Flag is not up
		addresses, err = k.Addrs()
		if err != nil {
			continue
		}
		if len(addresses) == 0 {
			// Assume Network is unreachable
			continue
		}
		var cidr net.IP
		for _, h := range addresses {
			cidr, _, err = net.ParseCIDR(h.String())
			if err != nil {
				continue
			}
			// We choose the first IPv4 address
			if ipType(h.String()) == "ipv4" {
				retval[k.Name] = cidr
			}
			//fmt.Printf("\t\t\t\t%v - %v\n", cidr, ipType(h.String()))
		}
	}
	return retval, nil
}
