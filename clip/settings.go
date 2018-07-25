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

type NUPNP struct {
	Id                string `json:"id"`
	InternalIPAddress string `json:"internalipaddress"`
}

var netClient = &http.Client{
	Timeout: time.Second * 10,
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
	return target
}

func link(c *Client, message []byte) {
	type linkreq struct {
		Id                string
		Internalipaddress string
	}
	var t *linkreq
	json.Unmarshal(message, &t)

	quit := make(chan bool)
	ticker := time.NewTicker(time.Millisecond * 500)
	go func() {
		for {
			select {
			case <-quit:
				return
			case <-ticker.C:
				response, _ := netClient.Post(fmt.Sprintf("http://%s/api", t.Internalipaddress), "application/json", strings.NewReader(`{ "devicetype": "onebridge#1" }`))
				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					log.Printf("Error: %v", err)
				}
				var arbitrary_json []map[string]interface{}
				json.Unmarshal(body, &arbitrary_json)
				if username := get(arbitrary_json[0], []string{"success", "username"}); username != nil {
					ticker.Stop()
					var success = arbitrary_json[0]["success"].(map[string]interface{})
					c.send <- []byte(fmt.Sprintf(`{ "username": "%s", "id": "%s" }`, success["username"], t.Id))
				} else {
					c.send <- []byte(`{ "type": "status", "status": "polling" }`)
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
