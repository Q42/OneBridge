package hue

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/koron/go-ssdp"
)

// AdvertiseDetails contains everything that is needed to advertise as a Hue bridge.
type AdvertiseDetails struct {
	APIVersion       string
	BridgeID         string
	UUID             string
	Network          NetworkInfo
	HTTPPort         uint
	FriendlyName     string
	SwVersion        string
	DatastoreVersion int
}

// Advertise starts the advertising as a Hue bridge.
func Advertise(details AdvertiseDetails) {
	// start advertising it!
	_, err2 := advertiseSSDP(details)
	if err2 != nil {
		log.Printf("Cannot advertise over ssdp: %s", err2.Error())
	}
}

func advertiseSSDP(details AdvertiseDetails) (*ssdp.Advertiser, error) {
	serverSignature := fmt.Sprintf("Linux/3.14.0 UPnP/1.0 IpBridge/%s\r\nhue-bridgeid: %s", details.APIVersion, details.BridgeID)
	adv, err := ssdp.Advertise(
		"urn:schemas-upnp-org:device:basic:1",
		"uuid:"+details.UUID+"",
		fmt.Sprintf("http://%s:%v/description.xml", details.Network.IP, details.HTTPPort),
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

// DescriptionXML returns similar XML to which the Hue bridge exposes at /description.xml for discovery.
func DescriptionXML(details AdvertiseDetails) string {
	deviceXML := `<root xmlns="urn:schemas-upnp-org:device-1-0">
	<specVersion>
		<major>1</major>
		<minor>0</minor>
	</specVersion>
	<URLBase>$BaseURL</URLBase>
	<device>
		<deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
		<friendlyName>$friendlyName ($ip)</friendlyName>
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

	deviceXML = strings.Replace(deviceXML, "$BaseURL", fmt.Sprintf("http://%s:%v/", details.Network.IP, 80), -1)
	deviceXML = strings.Replace(deviceXML, "$bridgeID", details.BridgeID, -1)
	deviceXML = strings.Replace(deviceXML, "$uuid", details.UUID, -1)
	deviceXML = strings.Replace(deviceXML, "$ip", details.Network.IP, -1)
	deviceXML = strings.Replace(deviceXML, "$friendlyName", details.FriendlyName, -1)

	return deviceXML
}
