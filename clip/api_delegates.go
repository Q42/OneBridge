package clip

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/q42/onebridge/hue"

	"github.com/gorilla/mux"
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

func apiUpdateDelegate(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var bridge Bridge
		_, err := parseBody(&bridge, w, r)
		if err != nil {
			return
		}

		vars := mux.Vars(r)
		delegateIdx, err := strconv.ParseInt(vars["id"], 10, 32)
		if err != nil {
			httpError(r)(w, err.Error(), 404)
			return
		}
		if delegateIdx >= int64(len(data.Delegates)) {
			httpError(r)(w, fmt.Sprintf("id out of bounds (%d, %d)", delegateIdx, len(data.Delegates)), 404)
			return
		}

		log.Printf("Updated delegate bridge %v\n", bridge)
		data.Delegates[delegateIdx] = bridge

		writeStandardHeaders(w)
		w.Write([]byte(fmt.Sprintf(`[{"success":true}]`)))
	}
}

func addDelegate(bridge Bridge) {
	data.Delegates = append(data.Delegates, bridge)
}
