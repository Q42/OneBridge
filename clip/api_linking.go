package clip

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"onebridge/hue"
	"strings"
	"time"
)

func linkNewUser(details *hue.AdvertiseDetails) func(w http.ResponseWriter, r *http.Request) {
	type linkRequest struct {
		Devicetype string
	}
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
