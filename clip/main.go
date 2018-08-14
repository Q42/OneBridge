package clip

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
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

const authUser key = 0

var notFound = &clipCatchAll{}

// Register clip routes
func Register(r *mux.Router, details *hue.AdvertiseDetails) {
	r.HandleFunc("/nouser/config", serveConfigNoAuth(details)).Methods("GET")
	r.HandleFunc("/", linkNewUser(details)).Methods("POST")
	r.HandleFunc("/nupnp", nupnp).Methods("GET")

	authed := r.PathPrefix("/").Subrouter()
	authed.NotFoundHandler = notFound

	authed.Use(data.Self.authMiddleware)
	authed.HandleFunc("/{username}/bridges", getDelegates).Methods("GET")
	authed.HandleFunc("/{username}/bridges", addDelegate(details)).Methods("POST")
	authed.HandleFunc("/{username}", serveRoot(details)).Methods("GET")
	authed.HandleFunc("/{username}/config", serveConfig(details)).Methods("GET")
	authed.HandleFunc("/{username}/config", putConfig).Methods("PUT")
	authed.HandleFunc("/{username}/{resourcetype}", resourceList).Methods("GET")
	authed.HandleFunc("/{username}/{resourcetype}/new", resourceNew).Methods("GET")
	authed.HandleFunc("/{username}/{resourcetype}/{resourceid}", resourceSingle).Methods("GET")
}

type clipCatchAll struct{}

func (h *clipCatchAll) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Unhandled CLIP: %s %s", r.Method, r.RequestURI)
	if r.Body != nil {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Printf(" `%s`", string(body))
	}
	fmt.Print("\n")
	// no pattern matched; send 404 response
	http.NotFound(w, r)
	return
}

// Middleware function, which will be called for each request
func (bridge *Bridge) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := urlPart(r.RequestURI, 2)
		for idx, u := range data.Self.Users {
			if u.ID == username || username == "0" {
				// We found the token in our map
				context.Set(r, authUser, username)
				now := time.Now()
				data.Self.Users[idx].LastUseDate = now.Format("2006-01-02T15:04:05")
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
				return
			}
		}

		// If we hit this, there are no users yet
		if username == "0" {
			current := time.Now()
			user := BridgeUser{
				Type:       "hue",
				ID:         strings.ToLower(randomHexString(16)),
				DeviceType: "OneBridge#FirstUser",
				CreateDate: current.Format("2006-01-02T15:04:05"),
			}
			data.Self.Users = append(data.Self.Users, user)
			context.Set(r, authUser, user.ID)
			next.ServeHTTP(w, r)
		}

		httpError(r)(w, "Forbidden", http.StatusForbidden)
	})
}

// Just like http.Error, but which formats http errors for Hue (200 + status message)
func httpError(r *http.Request) func(w http.ResponseWriter, status string, statusCode int) {
	ua := r.Header.Get("User-Agent")
	isHue := strings.Contains(ua, "Hue") || strings.Contains(ua, "hue")
	isBrowser := strings.Contains(ua, "Mozilla")

	if isBrowser {
		return http.Error
	}

	if !isHue {
		fmt.Printf("Unknown UA: %s\n", ua)
	}

	return func(w http.ResponseWriter, status string, statusCode int) {
		// Use HTTP200-Hue-error instead of HTTP error
		writeStandardHeaders(w)
		w.WriteHeader(http.StatusOK)
		if statusCode == http.StatusForbidden {
			w.Write([]byte(`[{"error":{"type":1,"address":"/","description":"unauthorized user"}}]`))
		} else if statusCode == http.StatusBadRequest {
			w.Write([]byte(`[{"error":{"type":5,"address":"/","description":"invalid/missing parameters in body"}}]`))
		} else if statusCode == http.StatusMethodNotAllowed {
			w.Write([]byte(`[{"error":{"type":4,"address":"/","description":"method, GET, not available for resource, /"}}]`))
		} else {
			// TODO learn more error.type's and implement those separately
			w.Write([]byte(fmt.Sprintf(`[{"error":{"type":0,"address":"/","description":"%s"}}]`, status)))
		}
	}
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

func parseBody(result interface{}, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Body == nil {
		httpError(r)(w, "Invalid/missing parameters in body", http.StatusBadRequest)
		return nil, errors.New("JSON body missing")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(r)(w, "Something bad happened!", http.StatusInternalServerError)
		log.Println(err)
		return nil, err
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		log.Println(string(body))
		httpError(r)(w, "Invalid or unparsable JSON.", http.StatusBadRequest)
		log.Println(err)
		return body, err
	}
	return body, nil
}

func linkNewUser(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var link linkRequest
		_, err := parseBody(&link, w, r)
		if err != nil {
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

func putConfig(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	body, err := parseBody(&data, w, r)
	if err != nil {
		return
	}

	if data["timezone"] != nil {
		writeStandardHeaders(w)
		w.Write([]byte(fmt.Sprintf(`[{"success":{"/config/timezone":"%s"}}]`, data["timezone"])))
	} else {
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		notFound.ServeHTTP(w, r)
	}
}

func getDelegates(w http.ResponseWriter, r *http.Request) {
	bytes, _ := json.Marshal(data.Delegates)
	writeStandardHeaders(w)
	w.Write(bytes)
}

func addDelegate(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var bridge Bridge
		_, err := parseBody(&bridge, w, r)
		if err != nil {
			return
		}

		log.Printf("Added delegate bridge %s\n", bridge)
		data.Delegates = append(data.Delegates, bridge)

		writeStandardHeaders(w)
		w.Write([]byte(fmt.Sprintf(`[{"success":true}]`)))
	}
}

func resourceList(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Printf("Type: %v\n", vars["resourcetype"])
	fmt.Fprintf(w, "[]")
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

func emptyArray(w http.ResponseWriter, r *http.Request) {
	writeStandardHeaders(w)
	w.Write([]byte("[]"))
}

func applyConfig(details *hue.AdvertiseDetails, config interface{}, username *string) {
	if config, ok := config.(*configShort); ok {
		config.BridgeID = &details.BridgeID
		config.FactoryNew = false
		config.APIVersion = &details.APIVersion
		config.DatastoreVersion = fmt.Sprintf("%d", details.DatastoreVersion)
		config.Mac = &details.Mac
		config.Name = &details.FriendlyName
		config.ReplacesBridgeID = nil
		config.StarterKitID = ""
		config.SwVersion = &details.SwVersion
	}

	if config, ok := config.(*configLong); ok {
		config.ZigbeeChannel = 15
		config.LinkButton = true
		config.PortalServices = true

		config.IPAddress = details.Network.IP
		config.Dhcp = true
		config.Netmask = details.Network.Netmask
		config.Gateway = details.Network.Gateway
		config.ProxyAddress = "none"
		config.ProxyPort = 0

		l := time.Now()
		t := time.Now().UTC()
		config.LocalTime = l.Format("2006-01-02T15:04:05")
		config.UTC = t.Format("2006-01-02T15:04:05")
		zoneName, _ := time.Now().Zone()
		config.TimeZone = zoneName

		config.Whitelist = make(map[string]whitelistEntry)
		for _, u := range data.Self.Users {
			if username == nil || u.ID == *username {
				config.Whitelist[u.ID] = whitelistEntry{LastUseDate: u.LastUseDate, CreateDate: u.CreateDate, Name: u.DeviceType}
			}
		}
	}
}

func serveConfigNoAuth(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var config = configShort{}
		applyConfig(details, &config, nil)
		bytes, _ := json.Marshal(config)
		writeStandardHeaders(w)
		w.Write(bytes)
	}
}

func serveConfig(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authuser := context.Get(r, authUser)
		var username = authuser.(string)
		if username == "" {
			httpError(r)(w, "Forbidden", http.StatusForbidden)
			return
		}
		var config = configLong{}
		applyConfig(details, &config, nil)
		bytes, _ := json.Marshal(config)
		writeStandardHeaders(w)
		w.Write(bytes)
	}
}

func serveRoot(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authuser := context.Get(r, authUser)
		var username = authuser.(string)

		var config = configLong{}
		applyConfig(details, &config, &username)

		var root = configFull{Config: config}
		bytes, _ := json.Marshal(root)
		writeStandardHeaders(w)
		w.Write(bytes)
	}
}

func getWhitelist(only *string) []byte {
	datas := make(map[string]whitelistEntry)
	for _, u := range data.Self.Users {
		if only == nil || u.ID == *only {
			datas[u.ID] = whitelistEntry{LastUseDate: u.LastUseDate, CreateDate: u.CreateDate, Name: u.DeviceType}
		}
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
