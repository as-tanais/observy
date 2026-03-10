package agent

import (
	"log"
	"net"
	"net/http"
)

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var localIP string

func init() {
	localIP = getLocalIP()
	if localIP != "" {
		log.Printf("Agent local IP detected: %s", localIP)
	}
}

func addRealIPHeader(req *http.Request) {
	if localIP != "" {
		req.Header.Set("X-Real-IP", localIP)
	}
}
