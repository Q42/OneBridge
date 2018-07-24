// https://github.com/probonopd/ESP8266HueEmulator
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"flag"

	"github.com/hermanbanken/huemmux/hue"
	"github.com/hermanbanken/huemmux/clip"
	"github.com/elazarl/goproxy"
)

var macs, _ = hue.GetMacAddr() // "00:17:88:ff:ff:ff" //
var bridgeID = strings.ToUpper(hue.ConvertMacToBridgeID(macs[0]))
var uuid = "2f402f80-da50-11e1-9b23-" + strings.ToLower(bridgeID)
var datastoreVersion = 72
var details *hue.AdvertiseDetails

func init() {
	details = &hue.AdvertiseDetails{}
	flag.StringVar(&details.FriendlyName, "name", "OneBridge", "bridge friendly name")
	flag.StringVar(&details.LocalIP, "ip", hue.Localip(), "which IP to bind the server to")
	flag.UintVar(&details.LocalHttpPort, "port", 80, "port to bind to")
	details.ApiVersion = "1.23.0"
	details.SwVersion = "20180109"
	details.datastoreVersion = 72
	details.Mac = macs[0]
}

func main() {
	flag.Parse()

	hue.Advertise(*details)

	h := http.NewServeMux()
	h.HandleFunc("/description.xml", logHandler(sendDescription))
	h.HandleFunc("/api/", logHandler(clip.Handler(details)))
	h.HandleFunc("/", logHandler(func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	}))

	log.Printf("OneBridge running on http://%s:%v", details.LocalIP, details.LocalHttpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%v", details.LocalIP, details.LocalHttpPort), h))
	// log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%v", details.LocalIP, httpsPort), "server.crt", "server.key", h))
	// openssl genrsa -out server.key 2048
	// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
}

func logHandler(deleg func(w http.ResponseWriter, req *http.Request)) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("URL: %s, %s, %s", req.URL.Path, req.Method, req.URL.String())
		for name, headers := range req.Header {
			name = strings.ToLower(name)
			for _, h := range headers {
				log.Printf("%s: %s", name, h)
			}
		}
		deleg(w, req)
		log.Println()
		return
	}
}

func sendDescription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(hue.DescriptionXML(*details)))
}

func randomClientKey() []byte {
	token := make([]byte, 16)
	rand.Read(token)
	return token
}
