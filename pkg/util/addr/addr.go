/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addr

import (
	"net"
)

func addrToIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPAddr:
		return v.IP
	case *net.IPNet:
		return v.IP
	default:
		return nil
	}
}

func localIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var ipAddrs []string

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // ignore error
		}

		for _, addr := range addrs {
			if ip := addrToIP(addr); ip != nil {
				ipAddrs = append(ipAddrs, ip.String())
			}
		}
	}

	return ipAddrs
}

func IPs() []string {
	return localIPs()
}
