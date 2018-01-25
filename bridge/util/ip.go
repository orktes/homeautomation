package util

import "net"

func GetIPAddress() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	// handle err
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}

		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if !ip.IsLoopback() && ip.To4() != nil {
				return ip.String()
			}

		}
	}

	return ""
}
