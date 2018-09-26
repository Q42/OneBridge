package clip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"onebridge/hue"
	"time"

	"github.com/gorilla/context"
)

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

		// TODO merge all
		root.Lights = fetchInternal("/api/0/lights", details)
		root.Groups = fetchInternal("/api/0/groups", details)
		root.Scenes = fetchInternal("/api/0/scenes", details)
		root.Schedules = fetchInternal("/api/0/schedules", details)
		root.Rules = fetchInternal("/api/0/rules", details)
		root.Sensors = fetchInternal("/api/0/sensors", details)
		root.ResourceLinks = fetchInternal("/api/0/resourcelinks", details)

		bytes, _ := json.Marshal(root)
		writeStandardHeaders(w)
		w.Write(bytes)
	}
}

func applyConfig(details *hue.AdvertiseDetails, config interface{}, username *string) {
	if config, ok := config.(*configShort); ok {
		config.BridgeID = &details.BridgeID
		config.FactoryNew = false
		config.APIVersion = &details.APIVersion
		config.DatastoreVersion = fmt.Sprintf("%d", details.DatastoreVersion)
		config.Mac = &details.Network.Mac
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
