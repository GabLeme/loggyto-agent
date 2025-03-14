package utils

import (
	"net"
	"os"
	"runtime"
)

func GetHostMetadata() map[string]string {
	hostname, _ := os.Hostname()
	ip := getLocalIP()

	return map[string]string{
		"host_name":    hostname,
		"machine_ip":   ip,
		"os":           runtime.GOOS,
		"architecture": runtime.GOARCH,
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}
