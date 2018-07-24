package hue

import (
	"strings"
)

func ConvertMacToBridgeID(mac string) string {
	var bridgeID = mac
	bridgeID = strings.Replace(bridgeID, ":", "", -1)
	return bridgeID[0:6] + "FFFE" + bridgeID[6:]
}