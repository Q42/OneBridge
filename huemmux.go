// https://github.com/probonopd/ESP8266HueEmulator
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/koron/go-ssdp"
)

var localIP = localip()
var httpPort = 80
var httpsPort = 443

var apiVersion = "1.23.0"
var swVersion = "20180109"
var friendlyName = "Huemmux"
var mac = "ac:bc:32:b6:f9:a7" // "00:17:88:ff:ff:ff" //
var bridgeID = strings.ToUpper(convertMacToBridgeID(mac))
var uuid = "2f402f80-da50-11e1-9b23-" + strings.ToLower(bridgeID)
var datastoreVersion = 72

func convertMacToBridgeID(mac string) string {
	var bridgeID = mac
	bridgeID = strings.Replace(bridgeID, ":", "", -1)
	return bridgeID[0:6] + "FFFE" + bridgeID[6:]
}

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	// start advertising it!
	log.Print("Advertising huemmux service on network")
	signature := fmt.Sprintf("Linux/3.14.0 UPnP/1.0 IpBridge/%s\r\nhue-bridgeid: %s", apiVersion, bridgeID)
	_, err2 := advertiseSSDP(signature, uuid)
	if err2 != nil {
		log.Printf("Cannot advertise over ssdp: %s", err2.Error())
	}

	h := http.NewServeMux()
	h.HandleFunc("/description.xml", logHandler(sendDescription))
	h.HandleFunc("/api/nouser/config", logHandler(clipNoUserConfig))
	h.HandleFunc("/api/", logHandler(clipHandler))
	h.HandleFunc("/", logHandler(func(w http.ResponseWriter, req *http.Request) {
		// The "/" pattern matches everything, so we need to check
		// that we're at the root here.
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Welcome to the home page!")
	}))

	log.Printf("Started Huemmux HTTP service on port http://%s:%v and https://%s:%v", localIP, httpPort, localIP, httpsPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", httpPort), h))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", 8080), h))
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", httpsPort), "server.crt", "server.key", h))
	// openssl genrsa -out server.key 2048
	// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

	log.Fatal(http.ListenAndServe(":8080", proxy))
}

func localip() string {
	conn, _ := net.Dial("udp", "1.2.3.4:80") // handle err...
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func advertiseSSDP(serverSignature string, deviceUUID string) (*ssdp.Advertiser, error) {
	log.Printf("Advertising as " + serverSignature + " (" + deviceUUID + ")")

	adv, err := ssdp.Advertise(
		"urn:schemas-upnp-org:device:basic:1",
		"uuid:"+deviceUUID+"",
		fmt.Sprintf("http://%s:%v/description.xml", localIP, httpPort),
		serverSignature,
		100)

	if err != nil {
		return nil, err
	}

	go sendAlive(adv)

	return adv, nil
}

func sendAlive(advertiser *ssdp.Advertiser) {
	aliveTick := time.Tick(15 * time.Second)

	for {
		select {
		case <-aliveTick:
			if err := advertiser.Alive(); err != nil {
				log.Fatal(err.Error())
			}
		}
	}
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

// LinkRequest
type LinkRequest struct {
	Devicetype string //
}

func clipHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/" && r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		log.Println(string(body))
		var link LinkRequest
		err = json.Unmarshal(body, &link)
		if err != nil {
			panic(err)
		}
		log.Println(link.Devicetype)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"success":{"username": "83b7780291a6ceffbe0bd049104df", "clientkey": "557D78B63DE0D099F7B8AB507C8383E3" }}]`))
		log.Print("Writing username")
	} else {
		log.Print("Writing config")
		clipFullConfig(w, r)
	}
}

func sendDescription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	log.Printf("creating device xml")
	deviceXML := `<root xmlns="urn:schemas-upnp-org:device-1-0">
	<specVersion>
		<major>1</major>
		<minor>0</minor>
	</specVersion>
	<URLBase>$BaseURL</URLBase>
	<device>
		<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
		<friendlyName>$friendlyName ($localIP)</friendlyName>
		<manufacturer>Royal Philips Electronics</manufacturer>
		<manufacturerURL>http://www.philips.com</manufacturerURL>
		<modelDescription>Philips hue Personal Wireless Lighting</modelDescription>
		<modelName>Philips hue bridge 2015</modelName>
		<modelNumber>BSB002</modelNumber>
		<modelURL>http://www.meethue.com</modelURL>
		<serialNumber>$bridgeID</serialNumber>
		<UDN>uuid:$uuid</UDN>
		<presentationURL>index.html</presentationURL>
		<iconList>
			<icon>
				<mimetype>image/png</mimetype>
				<height>48</height>
				<width>48</width>
				<depth>24</depth>
				<url>hue_logo_0.png</url>
			</icon>
		</iconList>
	</device>
</root>`

	deviceXML = strings.Replace(deviceXML, "$BaseURL", fmt.Sprintf("http://%s:%v/", localIP, 80), -1)
	deviceXML = strings.Replace(deviceXML, "$bridgeID", bridgeID, -1)
	deviceXML = strings.Replace(deviceXML, "$uuid", uuid, -1)
	deviceXML = strings.Replace(deviceXML, "$localIP", localIP, -1)
	deviceXML = strings.Replace(deviceXML, "$friendlyName", friendlyName, -1)

	w.Write([]byte(deviceXML))
}

func clipNoUserConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "nginx")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Connection", "close")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Mon, 1 Aug 2011 09:00:00 GMT")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	config := `{
	"name": "$friendlyName",
	"datastoreversion": "$datastoreVersion",
	"swversion": "$swVersion",
	"apiversion": "$apiVersion",
	"mac": "$mac",
	"bridgeid": "$bridgeID",
	"factorynew": false,
	"replacesbridgeid": null,
	"modelid": "BSB002",
	"starterkitid": ""
}`

	config = strings.Replace(config, "\n", "", -1)
	config = strings.Replace(config, "\t", "", -1)
	config = strings.Replace(config, "$friendlyName", friendlyName, -1)
	config = strings.Replace(config, "$bridgeID", bridgeID, -1)
	config = strings.Replace(config, "$apiVersion", apiVersion, -1)
	config = strings.Replace(config, "$swVersion", swVersion, -1)
	config = strings.Replace(config, "$datastoreVersion", fmt.Sprintf("%v", datastoreVersion), -1)
	config = strings.Replace(config, "$mac", mac, -1)

	w.Write([]byte(config))
}

func clipFullConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "nginx")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Connection", "close")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Mon, 1 Aug 2011 09:00:00 GMT")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	config := `{ "lights": [], "scenes": [], "sensors": [], "config": { 
		"name": "$friendlyName",
		"zigbeechannel": 15,
		"mac": "$mac",
		"dhcp": true,
		"ipaddress": "$localIP",
		"netmask": "255.255.255.0",
		"gateway": "192.168.178.1",
		"proxyaddress": "none",
		"proxyport": 0,
		"UTC": "$utcTime",
		"localtime": "$localTime",
		"timezone": "Europe/Amsterdam",
		"whitelist": {
				"83b7780291a6ceffbe0bd049104df": {
						"last use date": "2018-07-17T07:21:38",
						"create date": "2018-07-08T08:55:10",
						"name": "my_hue_app#iphone"
				}
		},
		"swversion": "$swVersion",
		"apiversion": "$apiVersion",
		"swupdate": {
				"updatestate": 0,
				"url": "",
				"text": "",
				"notify": false
		},
		"linkbutton": true,
		"portalservices": true,
		"portalconnection": "connected",
		"portalstate": {
				"signedon": true,
				"incoming": false,
				"outgoing": true,
				"communication": "disconnected"
		}
}}`

	l := time.Now()
	t := time.Now().Add(time.Hour * -2)
	config = strings.Replace(config, "\n", "", -1)
	config = strings.Replace(config, "\t", "", -1)
	config = strings.Replace(config, "$friendlyName", friendlyName, -1)
	config = strings.Replace(config, "$localIP", localIP, -1)
	config = strings.Replace(config, "$mac", mac, -1)
	config = strings.Replace(config, "$bridgeID", bridgeID, -1)
	config = strings.Replace(config, "$apiVersion", apiVersion, -1)
	config = strings.Replace(config, "$swVersion", swVersion, -1)
	config = strings.Replace(config, "$datastoreVersion", fmt.Sprintf("%v", datastoreVersion), -1)
	config = strings.Replace(config, "$utcTime", fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()), -1)
	config = strings.Replace(config, "$localTime", fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", l.Year(), l.Month(), l.Day(), l.Hour(), l.Minute(), l.Second()), -1)

	w.Write([]byte(config))
}

func randomClientKey() []byte {
	token := make([]byte, 16)
	rand.Read(token)
	return token
}
