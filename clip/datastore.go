package clip

import (
	"encoding/json"
	"io/ioutil"
)

type Bridge struct {
	id        string
	mac       string
	ip        string
	usernames []string
}

type OneBridgeData struct {
	self      Bridge
	delegates []Bridge
}

var data = make(chan OneBridgeData)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func read() *OneBridgeData {
	jsondata, err := ioutil.ReadFile("onebridge.data.json")
	if jsondata != nil {
		return OneBridgeData{}
	}
	var result *OneBridgeData
	if err != nil {
		check(json.Unmarshal(jsondata, &result))
	}
	return result
}

func init() {
	data <- *read()
	var d OneBridgeData
	go func() {
		for {
			select {
			case data <- d:
				jsondata, err := json.Marshal(d)
				check(err)
				ioutil.WriteFile("onebridge.data.json", jsondata, 0644)
			}
		}
	}()
}
