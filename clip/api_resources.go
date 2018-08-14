package clip

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func resourceIDFromBridge(id string, bridgeIdx int) string {
	uid, _ := strconv.ParseInt(id, 10, 32)
	return fmt.Sprintf("%s9%s", strconv.FormatInt(int64(bridgeIdx+1), 9), strconv.FormatInt(uid, 9))
}

var nineBaseId, _ = regexp.Compile("[0-8]+9[0-8]+")

func resourceIDToBridge(id string) (int, string) {
	if !nineBaseId.MatchString(id) {
		fmt.Printf("Could not convert '%s' to tuple of bridge id and resource id.\n", id)
		return 0, id
	}
	split := strings.Index(id, "9")
	bid := id[:split]
	rid := id[split+1:]
	intBid, _ := strconv.ParseInt(bid, 9, 32)
	intRid, _ := strconv.ParseInt(rid, 9, 32)
	return int(intBid - 1), strconv.FormatInt(intRid, 10)
}

func postProcess(bridgeIdx int, item map[string]interface{}, resourceType string) interface{} {
	if resourceType == "groups" {
		item["lights"] = convertIdsInArray(bridgeIdx, item["lights"])
		if item["sensors"] != nil {
			item["sensors"] = convertIdsInArray(bridgeIdx, item["sensors"])
		}
		if item["locations"] != nil {
			item["locations"] = convertIdsInMap(bridgeIdx, item["locations"])
		}
	}
	return item
}

func resourceList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	var resourceType = vars["resourcetype"]

	if resourceType == "resourcelinks" || resourceType == "rules" || resourceType == "scenes" {
		w.Write([]byte("{}"))
		return
	}

	resources := make(chan int)
	errors := make(chan error)
	var all = make(map[string]interface{})

	count := forEachBridge(func(bridge *Bridge, idx int) {
		res, err := netClient.Get(fmt.Sprintf("http://%s/api/%s/%s", bridge.IP, bridge.Users[0].ID, resourceType))
		if err != nil {
			errors <- err
			return
		}
		defer res.Body.Close()
		var target map[string]interface{}
		json.NewDecoder(res.Body).Decode(&target)
		for key, item := range target {
			all[resourceIDFromBridge(key, idx)] = postProcess(idx, item.(map[string]interface{}), resourceType)
		}
		resources <- 1
	})

	for range resources {
		count = count - 1
		if count == 0 {
			json.NewEncoder(w).Encode(all)
			return
		}
	}
}

func resourceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resourceType = vars["resourcetype"]
	var resourceID = vars["resourceid"]
	var field = vars["field"]

	if resourceType == "groups" && resourceID == "0" {
		fmt.Println("TODO FORALL bridges do action")
	}

	bix, rid := resourceIDToBridge(resourceID)
	if bix >= len(data.Delegates) {
		httpError(r)(w, "Bridge not found", 404)
		return
	}
	bridge := data.Delegates[bix]
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/%s/%s/%s/%s", bridge.IP, bridge.Users[0].ID, resourceType, rid, field), r.Body)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, err := netClient.Do(req)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	io.Copy(w, res.Body)
}

func resourceNew(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Type: %v\n", vars["resourcetype"])
	fmt.Fprintf(w, `{ "lastscan": null }`)
}

func resourceSingle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Type: %v id: %v\n", vars["resourcetype"], vars["resourceid"])
	fmt.Fprintf(w, "{}")
}

func forEachBridge(fn func(bridge *Bridge, idx int)) int {
	// TODO parallelize
	for idx := range data.Delegates {
		go fn(&data.Delegates[idx], idx)
	}
	return len(data.Delegates)
}

func convertIdsInArray(bridgeIdx int, any interface{}) interface{} {
	if any == nil {
		return nil
	}
	switch reflect.TypeOf(any).Kind() {
	case reflect.Slice:
		list := reflect.ValueOf(any)
		vsm := reflect.MakeSlice(reflect.TypeOf(any), list.Len(), list.Cap())
		for i := 0; i < list.Len(); i++ {
			converted := resourceIDFromBridge(list.Index(i).Interface().(string), bridgeIdx)
			vsm.Index(i).Set(reflect.ValueOf(converted))
		}
		return vsm.Interface()
	}
	return any
}

func convertIdsInMap(bridgeIdx int, any interface{}) interface{} {
	if any == nil {
		return nil
	}
	switch reflect.TypeOf(any).Kind() {
	case reflect.Map:
		list := reflect.ValueOf(any)
		vsm := reflect.MakeMapWithSize(reflect.TypeOf(any), list.Len())
		keys := list.MapKeys()
		for i := 0; i < len(keys); i++ {
			value := list.MapIndex(keys[i])
			mappedKey := resourceIDFromBridge(keys[i].Interface().(string), bridgeIdx)
			vsm.SetMapIndex(reflect.ValueOf(mappedKey), value)
		}
		return vsm.Interface()
	}
	return any
}
