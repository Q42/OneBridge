package hue

import (
	"strings"
)

// ConvertMacToBridgeID will adapt a mac address to be a Hue BridgeID,
// just like how Hue converts mac addresses to bridge ids too.
func ConvertMacToBridgeID(mac string) string {
	var bridgeID = mac
	bridgeID = strings.Replace(bridgeID, ":", "", -1)
	return bridgeID[0:6] + "FFFE" + bridgeID[6:]
}
