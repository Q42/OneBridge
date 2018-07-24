package hue

import (
    "fmt"
    "log"
    "net"
)

func GetMacAddr() ([]string, error) {
    ifas, err := net.Interfaces()
    if err != nil {
        return nil, err
    }
    var as []string
    for _, ifa := range ifas {
        a := ifa.HardwareAddr.String()
        if a != "" {
            as = append(as, a)
        }
    }
    return as, nil
}

func Localip() string {
	conn, _ := net.Dial("udp", "1.2.3.4:80") // handle err...
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {
    as, err := GetMacAddr()
    if err != nil {
        log.Fatal(err)
    }
    for _, a := range as {
        fmt.Println(a)
    }
}
