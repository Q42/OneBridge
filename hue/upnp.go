package hue

import (
	"fmt"
	"time"
	"log"
	"strings"
	"github.com/koron/go-ssdp"
)

type AdvertiseDetails struct {
	ApiVersion string
	BridgeID string
	Uuid string
	LocalIP string
	LocalHttpPort uint
	FriendlyName string
	SwVersion string
	DatastoreVersion int
	Mac string
}

func Advertise(details AdvertiseDetails) {
	// start advertising it!
	_, err2 := advertiseSSDP(details)
	if err2 != nil {
		log.Printf("Cannot advertise over ssdp: %s", err2.Error())
	}
}

func advertiseSSDP(details AdvertiseDetails) (*ssdp.Advertiser, error) {
	serverSignature := fmt.Sprintf("Linux/3.14.0 UPnP/1.0 IpBridge/%s\r\nhue-bridgeid: %s", details.ApiVersion, details.BridgeID)
	adv, err := ssdp.Advertise(
		"urn:schemas-upnp-org:device:basic:1",
		"uuid:"+details.Uuid+"",
		fmt.Sprintf("http://%s:%v/description.xml", details.LocalIP, details.LocalHttpPort),
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

func DescriptionXML(details AdvertiseDetails) string {
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

	deviceXML = strings.Replace(deviceXML, "$BaseURL", fmt.Sprintf("http://%s:%v/", details.LocalIP, 80), -1)
	deviceXML = strings.Replace(deviceXML, "$bridgeID", details.BridgeID, -1)
	deviceXML = strings.Replace(deviceXML, "$uuid", details.Uuid, -1)
	deviceXML = strings.Replace(deviceXML, "$localIP", details.LocalIP, -1)
	deviceXML = strings.Replace(deviceXML, "$friendlyName", details.FriendlyName, -1)

	return deviceXML
}