package clip

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// NUPNP contains discovery information
type NUPNP struct {
	ID                string `json:"id"`
	InternalIPAddress string `json:"internalipaddress"`
	Name              string `json:"name"`
	Mac               string `json:"mac"`
}

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func delegateByMac(mac string) *Bridge {
	for idx, bridge := range data.Delegates {
		if bridge.Mac == mac {
			return &data.Delegates[idx]
		}
	}
	return nil
}

// Scans for bridges and updates IPs if necessary
func rescan() {
	for _, bridge := range nupnpg() {
		if delegate := delegateByMac(bridge.Mac); delegate != nil {
			delegate.IP = bridge.InternalIPAddress
		}
	}
}

func wsReplyBridges(c *Client) {
	type discoveryResult struct {
		Name   string
		Bridge Bridge
		Linked bool
	}
	var bridges = make([]discoveryResult, 0)
	for _, bridge := range nupnpg() {
		var linked = false
		if delegate := delegateByMac(bridge.Mac); delegate != nil {
			linked = true
			delegate.IP = bridge.InternalIPAddress
		}
		bridges = append(bridges, discoveryResult{
			Bridge: Bridge{
				ID:  bridge.ID,
				IP:  bridge.InternalIPAddress,
				Mac: bridge.Mac,
			},
			Name:   bridge.Name,
			Linked: linked,
		})
	}
	b, _ := json.Marshal(bridges)
	c.send <- b
}

func nupnp(w http.ResponseWriter, r *http.Request) {
	res, _ := netClient.Get("https://www.meethue.com/api/nupnp")
	io.Copy(w, res.Body)
}

func nupnpg() []NUPNP {
	res, _ := netClient.Get("https://www.meethue.com/api/nupnp")
	var target []NUPNP
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&target)

	for idx, bridge := range target {
		configRes, _ := netClient.Get(fmt.Sprintf("http://%s/api/nouser/config", bridge.InternalIPAddress))
		var configTarget configShort
		defer configRes.Body.Close()
		json.NewDecoder(configRes.Body).Decode(&configTarget)
		target[idx].Name = *configTarget.Name
		target[idx].Mac = *configTarget.Mac
	}

	return target
}

func link(c *Client, message []byte) {
	type linkreq struct {
		BridgeID  *string
		BridgeIP  *string
		BridgeMac *string
	}
	var t *linkreq
	json.Unmarshal(message, &t)
	if t.BridgeID == nil || t.BridgeIP == nil || t.BridgeMac == nil {
		return
	}

	quit := make(chan bool)
	ticker := time.NewTicker(time.Millisecond * 500)
	go func() {
		for {
			select {
			case <-quit:
				c.send <- []byte(fmt.Sprintf(`{ "type": "status", "status": "polling-stopped", "bridgeid": "%s" }`, *t.BridgeID))
				ticker.Stop()
				return
			case <-ticker.C:
				response, _ := netClient.Post(fmt.Sprintf("http://%s/api", *t.BridgeIP), "application/json", strings.NewReader(`{ "devicetype": "onebridge#1", "generateclientkey": true }`))
				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					log.Printf("Error: %v", err)
				}
				var arbitraryJSON []map[string]interface{}
				json.Unmarshal(body, &arbitraryJSON)
				if username := get(arbitraryJSON[0], []string{"success", "username"}); username != nil {
					ticker.Stop()
					var success = arbitraryJSON[0]["success"].(map[string]interface{})
					var clientKey = success["clientkey"].(string)
					var username = success["username"].(string)
					var bridge = Bridge{
						IP:    *t.BridgeIP,
						ID:    *t.BridgeID,
						Mac:   *t.BridgeMac,
						Users: []BridgeUser{BridgeUser{ID: username, ClientKey: &clientKey, Type: "hue"}},
					}
					addDelegate(bridge)
					c.send <- []byte(fmt.Sprintf(`{ "username": "%s", "id": "%s" }`, success["username"], *t.BridgeID))
					ticker.Stop()
				} else {
					c.send <- []byte(fmt.Sprintf(`{ "type": "status", "status": "polling", "bridgeid": "%s" }`, *t.BridgeID))
				}
			}
		}
	}()

	time.Sleep(time.Second * 10)
	quit <- true
}

func get(js map[string]interface{}, fields []string) interface{} {
	for idx, field := range fields {
		if val, ok := js[field]; ok {
			if idx < len(fields)-1 {
				js = val.(map[string]interface{})
			} else {
				return val
			}
		} else {
			return nil
		}
	}
	return js
}
