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

	sensor := convertIdsInURL(17, "/sensors/36/state")
	sensorExpected := fmt.Sprintf("/sensors/%s/state", resourceIDFromBridge("36", 17))
	if sensor != sensorExpected {
		t.Errorf("Should convert sensor state url, expected: '%s', actual: '%s'", sensorExpected, sensor)
	}
}

func TestApiGroupPutOk(t *testing.T) {
	payload := `{ "type": "Other" }`
	// group 20920 = bridge 18, group 17
	req, err := http.NewRequest("PUT", "/api/0/groups/20920", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `[{ "success": { "/groups/20920/class": "TV" }}]`
	matched, err := regexp.MatchString(expected, rr.Body.String())
	if !matched || err != nil {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

}
