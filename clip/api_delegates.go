package clip

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"onebridge/hue"
)

func apiGetDelegates(w http.ResponseWriter, r *http.Request) {
	bytes, _ := json.Marshal(data.Delegates)
	writeStandardHeaders(w)
	w.Write(bytes)
}

func apiAddDelegate(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var bridge Bridge
		_, err := parseBody(&bridge, w, r)
		if err != nil {
			return
		}

		log.Printf("Added delegate bridge %v\n", bridge)
		data.Delegates = append(data.Delegates, bridge)

		writeStandardHeaders(w)
		w.Write([]byte(fmt.Sprintf(`[{"success":true}]`)))
	}
}

func addDelegate(bridge Bridge) {
	data.Delegates = append(data.Delegates, bridge)
}
