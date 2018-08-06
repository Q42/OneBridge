package clip

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"onebridge/hue"

	"github.com/gorilla/mux"
)

type linkRequest struct {
	Devicetype string //
}

// Register clip routes
func Register(r *mux.Router, details *hue.AdvertiseDetails) {
	r.HandleFunc("/nouser/config", noUserConfig(details))
	r.HandleFunc("/", linkNewUser(details)).Methods("POST")
	r.HandleFunc("/", fullConfig(details)).Methods("GET")
	r.HandleFunc("/nupnp", nupnp)
	r.Handle("/", headerLoggerHandler(http.NotFoundHandler()))
}

func linkNewUser(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`[{"error":{"type":5,"address":"/","description":"invalid/missing parameters in body"}}]`))
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Something bad happened!"))
			log.Println(err)
			return
		}

		var link linkRequest
		err = json.Unmarshal(body, &link)

		if err != nil {
			log.Println(string(body))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid or unparsable JSON."))
			log.Println(err)
			return
		}

		current := time.Now()
		user := BridgeUser{
			Type:       "hue",
			ID:         strings.ToLower(randomHexString(16)),
			DeviceType: link.Devicetype,
			CreateDate: current.Format("2006-01-02T15:04:05"),
		}

		log.Printf("DeviceType %s => User.ID %s \n", user.DeviceType, user.ID)
		data.Self.Users = append(data.Self.Users, user)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`[{"success":{"username": "%s" }}]`, user.ID)))
	}
}

func noUserConfig(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "Mon, 1 Aug 2011 09:00:00 GMT")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		config := `{
		"name": "$friendlyName",
		"datastoreversion": "$datastoreVersion",
		"swversion": "$swVersion",
		"apiversion": "$apiVersion",
		"mac": "$mac",
		"bridgeid": "$bridgeID",
		"factorynew": false,
		"replacesbridgeid": null,
		"modelid": "BSB002",
		"starterkitid": ""
	}`

		config = strings.Replace(config, "\n", "", -1)
		config = strings.Replace(config, "\t", "", -1)
		config = strings.Replace(config, "$friendlyName", details.FriendlyName, -1)
		config = strings.Replace(config, "$bridgeID", details.BridgeID, -1)
		config = strings.Replace(config, "$apiVersion", details.APIVersion, -1)
		config = strings.Replace(config, "$swVersion", details.SwVersion, -1)
		config = strings.Replace(config, "$datastoreVersion", fmt.Sprintf("%v", details.DatastoreVersion), -1)
		config = strings.Replace(config, "$mac", details.Mac, -1)

		w.Write([]byte(config))
	}
}

func fullConfig(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "Mon, 1 Aug 2011 09:00:00 GMT")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		config := `{ "lights": [], "scenes": [], "sensors": [], "config": { 
		"name": "$friendlyName",
		"zigbeechannel": 15,
		"mac": "$mac",
		"dhcp": true,
		"ipaddress": "$localIP",
		"netmask": "255.255.255.0",
		"gateway": "192.168.178.1",
		"proxyaddress": "none",
		"proxyport": 0,
		"UTC": "$utcTime",
		"localtime": "$localTime",
		"timezone": "Europe/Amsterdam",
		"whitelist": {
				"83b7780291a6ceffbe0bd049104df": {
						"last use date": "2018-07-17T07:21:38",
						"create date": "2018-07-08T08:55:10",
						"name": "my_hue_app#iphone"
				}
		},
		"swversion": "$swVersion",
		"apiversion": "$apiVersion",
		"swupdate": {
				"updatestate": 0,
				"url": "",
				"text": "",
				"notify": false
		},
		"linkbutton": true,
		"portalservices": true,
		"portalconnection": "connected",
		"portalstate": {
				"signedon": true,
				"incoming": false,
				"outgoing": true,
				"communication": "disconnected"
		}
}}`

		l := time.Now()
		t := time.Now().Add(time.Hour * -2)
		config = strings.Replace(config, "\n", "", -1)
		config = strings.Replace(config, "\t", "", -1)
		config = strings.Replace(config, "$friendlyName", details.FriendlyName, -1)
		config = strings.Replace(config, "$localIP", details.LocalIP, -1)
		config = strings.Replace(config, "$mac", details.Mac, -1)
		config = strings.Replace(config, "$bridgeID", details.BridgeID, -1)
		config = strings.Replace(config, "$apiVersion", details.APIVersion, -1)
		config = strings.Replace(config, "$swVersion", details.SwVersion, -1)
		config = strings.Replace(config, "$datastoreVersion", fmt.Sprintf("%v", details.DatastoreVersion), -1)
		config = strings.Replace(config, "$utcTime", fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()), -1)
		config = strings.Replace(config, "$localTime", fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", l.Year(), l.Month(), l.Day(), l.Hour(), l.Minute(), l.Second()), -1)

		w.Write([]byte(config))
	}
}

func randomClientKey() []byte {
	token := make([]byte, 16)
	rand.Read(token)
	return token
}

func randomHexString(len int) string {
	src := make([]byte, 16)
	rand.Read(src)
	dst := make([]byte, hex.EncodedLen(len))
	hex.Encode(dst, src)
	return fmt.Sprintf("%s", dst)
}

func headerLoggerHandler(deleg http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("URL: %s, %s, %s", req.URL.Path, req.Method, req.URL.String())
		for name, headers := range req.Header {
			name = strings.ToLower(name)
			for _, h := range headers {
				log.Printf("%s: %s", name, h)
			}
		}
		deleg.ServeHTTP(w, req)
		log.Println()
		return
	})
}
