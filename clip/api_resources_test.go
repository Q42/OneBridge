package clip

import (
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
