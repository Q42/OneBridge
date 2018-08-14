package hue

import (
	"fmt"
	"log"
	"net"
)

// GetNetworkInfo traverses all interfaces and returns the network info of the most likely used interface.
func GetNetworkInfo() (*NetworkInfo, error) {
	usedIP := getUsedIP()
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, ifa := range ifas {
		var info = NetworkInfo{
			Mac: ifa.HardwareAddr.String(),
		}

		if info.Mac == "" {
			continue
		}

		addrs, _ := ifa.Addrs()
		for _, addr := range addrs {
			cidrIP, cidrNet, err := net.ParseCIDR(addr.String())
			if cidrIP.String() != usedIP || err != nil {
				continue
			}
			info.IP = usedIP
			info.Netmask = quadString(cidrNet.Mask)
			info.Gateway = firstAddr(*cidrNet)
			return &info, nil
		}
	}
	return nil, nil
}

// NetworkInfo contains network details
type NetworkInfo struct {
	IP      string
	Mac     string
	Netmask string
	Gateway string
}

func getUsedIP() string {
	conn, _ := net.Dial("udp", "1.2.3.4:80") // handle err...
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
	is, err := GetNetworkInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(is)
}

func quadString(m []byte) string {
	if len(m) == 4 {
		return fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
	}
	return ""
}

func firstAddr(ipNet net.IPNet) string {
	l := ipNet.IP
	prefix := l[:len(l)-1]
	final := l[len(l)-1]
	oneUp := append(prefix, final+1)
	return quadString(oneUp)
}
