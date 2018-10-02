package clip

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"onebridge/hue"
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
	if resourceType == "scenes" {
		item["lights"] = convertIdsInArray(bridgeIdx, item["lights"])
		appData, hasAppdata := item["appdata"].(map[string]interface{})
		if hasAppdata && appData["data"] != nil {
			appData["data"] = convertIdsAppData(bridgeIdx, appData["data"])
		}
	}
	if resourceType == "resourcelinks" {
		links, ok := item["links"].([]interface{})
		if ok {
			for idx, url := range links {
				links[idx] = convertIdsInURL(bridgeIdx, url)
			}
		}
	}
	if resourceType == "rules" {
		for key, value := range item {
			list, isList := value.([]interface{})
			if isList && (key == "actions" || key == "conditions") {
				for _, part := range list {
					object, isObject := part.(map[string]interface{})
					if isObject && object["address"] != nil {
						object["address"] = convertIdsInURL(bridgeIdx, object["address"])
					}
				}
			}
		}
	}

	return item
}

var regexID = regexp.MustCompile(`^[0-9]+$`)

func resourceList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var resourceType = vars["resourcetype"]

	// if resourceType == "resourcelinks" || resourceType == "rules" || resourceType == "scenes" {
	// 	w.Write([]byte("{}"))
	// 	return
	// }

	errors := make([]error, 0)

	type tuple struct {
		key   string
		value interface{}
	}
	var allChannel = make(chan tuple, len(data.Delegates))
	var wg sync.WaitGroup

	// TODO idiomatic way to structure this:
	// forEachBridge(fetch("http://"))
	// 	.then(postProcess)
	// 	.merge(byThisGoRoutine)

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

		wg.Add(len(target))
		for key, item := range target {
			mappedKey := key
			if regexID.Match([]byte(key)) {
				mappedKey = resourceIDFromBridge(key, idx)
			}
			allChannel <- tuple{
				mappedKey,
				postProcess(idx, item.(map[string]interface{}), resourceType),
			}
		}
		wg.Done()
	})
	wg.Add(count)

	for _, err := range errors {
		if strings.HasSuffix(err.Error(), "connect: no route to host") {
			// Bridge might be offline or moved to other IP: trigger scan
			go rescan()
		}
		fmt.Printf("NonFatalError: %v\n", err)
	}

	var all = make(map[string]interface{})
	go func() {
		for item := range allChannel {
			all[item.key] = item.value
			wg.Done()
		}
	}()

	wg.Wait()
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

func resourceSingle(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var resourceType = vars["resourcetype"]
		var resourceID = vars["resourceid"]

		// Special group which lists all lights/sensors in the home
		if resourceType == "groups" && resourceID == "0" {
			w.WriteHeader(http.StatusOK)
			bytes, _ := json.Marshal(getGroupZero(details))
			w.Write(bytes)
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

var pathComponentTypeID = regexp.MustCompile(`[a-z]+/[0-9]+`)

func convertIdsInURL(bridgeIdx int, any interface{}) interface{} {
	str, ok := any.(string)
	if ok {
		result := pathComponentTypeID.ReplaceAllStringFunc(str, func(match string) string {
			split := strings.Index(match, "/")
			if split < 0 {
				return match
			}
			resourceType := match[0 : split+1]
			resourceID := resourceIDFromBridge(match[split+1:], bridgeIdx)
			return fmt.Sprintf("%s%s", resourceType, resourceID)
		})
		return result
	}
	fmt.Print("nok")
	return any
}

var appDataRoomID = regexp.MustCompile(`_r[0-9]+`)

func convertIdsAppData(bridgeIdx int, any interface{}) interface{} {
	str, ok := any.(string)
	if ok {
		result := appDataRoomID.ReplaceAllFunc([]byte(str), func(match []byte) []byte {
			resourceID := resourceIDFromBridge(string(match[2:]), bridgeIdx)
			return append([]byte("_r"), []byte(resourceID)...)
		})
		return string(result)
	}
	fmt.Print("nok")
	return any
}

func fetchInternal(url string, details *hue.AdvertiseDetails) map[string]interface{} {
	res, err := netClient.Get(fmt.Sprintf("http://%s:%d%s", details.Network.IP, details.HTTPPort, url))
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	var target map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&target)
	if err != nil {
		log.Printf("Unmarshal incorrect: %s\n", err)
	}
	return target
}

func getGroupZero(details *hue.AdvertiseDetails) interface{} {
	type GroupZero struct {
		Name    string   `json:"name"`
		Lights  []string `json:"lights"`
		Sensors []string `json:"sensors"`
		Type    string   `json:"type"`
		State   struct {
			AllOn bool `json:"all_on"`
			AnyOn bool `json:"any_on"`
		}
		Recycle bool                   `json:"recycle"`
		Action  map[string]interface{} `json:"action"`
	}
	type tuple struct {
		idx    int
		result GroupZero
		error  error
	}

	var wg sync.WaitGroup
	all := make(chan tuple, len(data.Delegates))
	count := forEachBridge(func(bridge *Bridge, idx int) {
		res, err := netClient.Get(fmt.Sprintf("http://%s/api/%s/groups/0", bridge.IP, bridge.Users[0].ID))
		if err != nil {
			all <- tuple{idx, GroupZero{}, err}
			return
		}
		defer res.Body.Close()
		var target GroupZero
		json.NewDecoder(res.Body).Decode(&target)

		all <- tuple{idx, target, nil}
	})
	wg.Add(count)

	var group GroupZero
	group.State.AllOn = true
	group.State.AnyOn = false

	go func() {
		for item := range all {
			if item.error != nil {
				if strings.HasSuffix(item.error.Error(), "connect: no route to host") {
					// Bridge might be offline or moved to other IP: trigger scan
					go rescan()
				}
				fmt.Printf("NonFatalError: %v\n", item.error)
				wg.Done()
				continue
			}

			group.Name = item.result.Name
			group.Recycle = item.result.Recycle
			group.Type = item.result.Type
			group.Action = item.result.Action
			for _, ID := range item.result.Lights {
				group.Lights = append(group.Lights, resourceIDFromBridge(ID, item.idx))
			}
			for _, ID := range item.result.Sensors {
				group.Sensors = append(group.Sensors, resourceIDFromBridge(ID, item.idx))
			}
			group.State.AllOn = item.result.State.AllOn && group.State.AllOn
			group.State.AnyOn = item.result.State.AnyOn || group.State.AnyOn
			wg.Done()
		}
	}()

	wg.Wait()
	return group
}
