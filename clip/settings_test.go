package clip

import (
	"encoding/json"
	"testing"
)

func TestTimeConsuming(t *testing.T) {
	var target []NUPNP
	err := json.Unmarshal([]byte(`[{"id": "001788fffefffff","internalipaddress": "192.168.1.1"}]`), &target)
	if target == nil {
		t.Errorf("Unmarshal incorrect: %s", err)
	}
}
