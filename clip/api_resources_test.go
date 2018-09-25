package clip

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestIdConversion(t *testing.T) {
	oneBridgeID := resourceIDFromBridge("1", 0)
	if oneBridgeID != "191" {
		t.Errorf("Should convert bridgeId + resourceId to base 9: %s", oneBridgeID)
	}

	bridgeIdx, resourceID := resourceIDToBridge(oneBridgeID)
	if bridgeIdx != 0 {
		t.Errorf("Should convert bridgeIdx back")
	}
	if resourceID != "1" {
		t.Errorf("Should convert resourceID back")
	}
}

func TestLargeIdConversion(t *testing.T) {
	oneBridgeID := resourceIDFromBridge("18", 17)
	if oneBridgeID != "20920" {
		t.Errorf("Should convert bridgeId + resourceId to base 9: %s", oneBridgeID)
	}

	bridgeIdx, resourceID := resourceIDToBridge(oneBridgeID)
	if bridgeIdx != 17 {
		t.Errorf("Should convert bridgeIdx back")
	}
	if resourceID != "18" {
		t.Errorf("Should convert resourceID back")
	}
}

func TestUrlConversion(t *testing.T) {
	light56 := convertIdsInURL(17, "/lights/56")
	light56Expected := fmt.Sprintf("/lights/%s", resourceIDFromBridge("56", 17))
	if light56 != light56Expected {
		t.Errorf("Should convert light url, expected: '%s', actual: '%s'", light56Expected, light56)
	}

	bridge := convertIdsInURL(17, "/bridge")
	bridgeExpected := "/bridge"
	if bridge != bridgeExpected {
		t.Errorf("Should convert bridge url, expected: '%s', actual: '%s'", bridgeExpected, bridge)
	}
}
