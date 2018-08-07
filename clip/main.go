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

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

type linkRequest struct {
	Devicetype string //
}

type key int

const AuthUser key = 0

// Register clip routes
func Register(r *mux.Router, details *hue.AdvertiseDetails) {
	r.HandleFunc("/nouser/config", noUserConfig(details))
	r.HandleFunc("/", linkNewUser(details)).Methods("POST")
	r.HandleFunc("/", fullConfig(details)).Methods("GET")
	r.HandleFunc("/nupnp", nupnp)

	authed := r.PathPrefix("/").Subrouter()
	authed.Use(data.Self.authMiddleware)
	authed.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(`Clip Username %s`, context.Get(r, AuthUser)))
		w.WriteHeader(http.StatusOK)
	})
	r.Handle("/", headerLoggerHandler(http.NotFoundHandler()))
}

// Middleware function, which will be called for each request
func (bridge *Bridge) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := urlPart(r.RequestURI, 2)
		for _, u := range data.Self.Users {
			if u.ID == username {
				// We found the token in our map
				context.Set(r, AuthUser, username)
				log.Printf("Authenticated user %s\n", username)
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
				return
			}
		}

		// Write an error and stop the handler chain
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

func urlPart(path string, index int) string {
	j := strings.Index(path, "/")
	for index > 0 {
		index = index - 1
		if j >= 0 {
			path = path[(j + 1):]
		} else {
			path = ""
		}
		j = strings.Index(path, "/")
	}
	if j < 0 {
		return path
	}
	return path[:j]
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
		writeStandardHeaders(w)
		w.Write([]byte(fmt.Sprintf(`[{"success":{"username": "%s" }}]`, user.ID)))
	}
}

func noUserConfig(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		writeStandardHeaders(w)

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
		writeStandardHeaders(w)

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
		"whitelist": $whitelist,
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
		localTime := l.Format("2006-01-02T15:04:05")
		utcTime := t.Format("2006-01-02T15:04:05")

		config = strings.Replace(config, "\n", "", -1)
		config = strings.Replace(config, "\t", "", -1)
		config = strings.Replace(config, "$friendlyName", details.FriendlyName, -1)
		config = strings.Replace(config, "$localIP", details.LocalIP, -1)
		config = strings.Replace(config, "$mac", details.Mac, -1)
		config = strings.Replace(config, "$bridgeID", details.BridgeID, -1)
		config = strings.Replace(config, "$apiVersion", details.APIVersion, -1)
		config = strings.Replace(config, "$swVersion", details.SwVersion, -1)
		config = strings.Replace(config, "$whitelist", string(getWhitelist()), -1)
		config = strings.Replace(config, "$datastoreVersion", fmt.Sprintf("%v", details.DatastoreVersion), -1)
		config = strings.Replace(config, "$utcTime", utcTime, -1)
		config = strings.Replace(config, "$localTime", localTime, -1)

		w.Write([]byte(config))
	}
}

func getWhitelist() []byte {
	type whitelistEntry struct {
		LastUseDate string `json:"last use date"`
		CreateDate  string `json:"create date"`
		Name        string `json:"name"`
	}

	datas := make(map[string]whitelistEntry)
	for _, u := range data.Self.Users {
		datas[u.ID] = whitelistEntry{LastUseDate: u.LastUseDate, CreateDate: u.CreateDate, Name: u.DeviceType}
	}
	jsonData, _ := json.Marshal(datas)
	return jsonData
}

func writeStandardHeaders(w http.ResponseWriter) {
	w.Header().Set("Server", "nginx")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Mon, 1 Aug 2011 09:00:00 GMT")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
