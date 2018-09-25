package clip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

func resourceIDFromBridge(id string, bridgeIdx int) string {
	uid, _ := strconv.ParseInt(id, 10, 32)
	return fmt.Sprintf("%s9%s", strconv.FormatInt(int64(bridgeIdx+1), 9), strconv.FormatInt(uid, 9))
}

var nineBaseID, _ = regexp.Compile("[0-8]+9[0-8]+")

func resourceIDToBridge(id string) (int, string) {
	if !nineBaseID.MatchString(id) {
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
		stream, hasStream := item["stream"].(map[string]interface{})
		if hasStream && stream["proxynode"] != nil {
			stream["proxynode"] = convertIdsInURL(bridgeIdx, stream["proxynode"])
		}
	}
	return item
}

func resourceList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resourceType = vars["resourcetype"]

	if resourceType == "resourcelinks" || resourceType == "rules" || resourceType == "scenes" {
		w.Write([]byte("{}"))
		return
	}

	errors := make([]error, 0)
	var all = make(map[string]interface{})
	var wg sync.WaitGroup

	count := forEachBridge(func(bridge *Bridge, idx int) {
		res, err := netClient.Get(fmt.Sprintf("http://%s/api/%s/%s", bridge.IP, bridge.Users[0].ID, resourceType))
		if err != nil {
			errors = append(errors, err)
			wg.Done()
			return
		}
		defer res.Body.Close()
		var target map[string]interface{}
		json.NewDecoder(res.Body).Decode(&target)
		for key, item := range target {
			all[resourceIDFromBridge(key, idx)] = postProcess(idx, item.(map[string]interface{}), resourceType)
		}
		wg.Done()
	})

	wg.Add(count)
	wg.Wait()
	for _, err := range errors {
		if strings.HasSuffix(err.Error(), "connect: no route to host") {
			// Bridge might be offline or moved to other IP: trigger scan
			go rescan()
		}
		fmt.Printf("NonFatalError: %v\n", err)
	}
	err := json.NewEncoder(w).Encode(all)
	if err != nil {
		fmt.Printf("FatalError: %v\n", err)
		httpError(r)(w, err.Error(), 500)
	}
}

// Updating property like "group/1" with { "type": "Other" }
func resourceUpdateBody(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resourceType = vars["resourcetype"]
	var resourceID = vars["resourceid"]

	bix, rid := resourceIDToBridge(resourceID)
	if bix >= len(data.Delegates) {
		httpError(r)(w, "Bridge not found", 404)
		return
	}
	bridge := data.Delegates[bix]
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/%s/%s/%s", bridge.IP, bridge.Users[0].ID, resourceType, rid), r.Body)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	res, err := netClient.Do(req)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.Copy(w, res.Body)
}

// Updating property like "group/0/action"
func resourceUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resourceType = vars["resourcetype"]
	var resourceID = vars["resourceid"]
	var field = vars["field"]

	// Handle special "Home" group 0
	if resourceType == "groups" && resourceID == "0" {
		var wg sync.WaitGroup
		var result []byte
		count := forEachBridge(func(bridge *Bridge, idx int) {
			req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s/api/%s/%s/%s/%s", bridge.IP, bridge.Users[0].ID, resourceType, resourceID, field), r.Body)
			res, err := netClient.Do(req)
			if err != nil {
				fmt.Printf("NonFatalError: %v\n", err)
				wg.Done()
				return
			}
			defer res.Body.Close()
			result, _ = ioutil.ReadAll(res.Body)
			wg.Done()
		})
		wg.Add(count)
		wg.Wait()
		if len(result) == 0 {
			httpError(r)(w, "No results available", 500)
			fmt.Printf("FatalError: none of the delegates replied\n")
			return
		}
		w.Write(result)
		return
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

	res, err := netClient.Do(req)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.Copy(w, res.Body)
}

func resourceNew(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Type: %v\n", vars["resourcetype"])
	fmt.Fprintf(w, `{ "lastscan": "none" }`)
}

func resourceSingle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resourceType = vars["resourcetype"]
	var resourceID = vars["resourceid"]

	if resourceType == "resourcelinks" || resourceType == "rules" || resourceType == "scenes" {
		httpError(r)(w, "Not implemented", 404)
		return
	}

	// Special group which lists all lights/sensors in the home
	if resourceType == "groups" && resourceID == "0" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{}") // TODO fetch from all, and merge
		return
	}

	// Getting single resource from delegate bridge
	bix, rid := resourceIDToBridge(resourceID)
	if bix >= len(data.Delegates) {
		httpError(r)(w, "Bridge not found", 404)
		return
	}
	bridge := data.Delegates[bix]
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/api/%s/%s/%s", bridge.IP, bridge.Users[0].ID, resourceType, rid), nil)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	res, err := netClient.Do(req)
	if err != nil {
		httpError(r)(w, err.Error(), 500)
		return
	}

	defer res.Body.Close()
	var target map[string]interface{}
	json.NewDecoder(res.Body).Decode(&target)
	postProcess(bix, target, resourceType)
	w.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(target)
	w.Write(bytes)
}

func forEachBridge(fn func(bridge *Bridge, idx int)) int {
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

var pathCompmonentTypeID = regexp.MustCompile(`[a-z]+/[0-9]+`)

func convertIdsInURL(bridgeIdx int, any interface{}) interface{} {
	str, ok := any.(string)
	if ok {
		result := pathCompmonentTypeID.ReplaceAllFunc([]byte(str), func(match []byte) []byte {
			split := bytes.Index(match, []byte("/"))
			if split < 0 {
				return match
			}
			resourceType := match[0 : split+1]
			resourceID := resourceIDFromBridge(string(match[split+1:]), bridgeIdx)
			return append(resourceType, []byte(resourceID)...)
		})
		return string(result)
	}
	fmt.Print("nok")
	return any
}
