package clip

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Bridge contains basic information to identify a physical bridge
type Bridge struct {
	ID    string
	Mac   string
	IP    string
	Users []BridgeUser
}

// BridgeUser contains the "whitelist" entries
type BridgeUser struct {
	Type        string // hue, google, etc.
	ID          string
	DeviceType  string
	LastUseDate string
	CreateDate  string
}

// OneBridgeData Datastore root of OneBridge
type OneBridgeData struct {
	Self      Bridge
	Delegates []Bridge
}

var data OneBridgeData

// SetupDatastore will read from file & then write any changes to the data to disk
func SetupDatastore() {
	fmt.Print("Reading onebridge.data.json... ")
	if err := Load("./onebridge.data.json", &data); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	fmt.Println("done")

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for range ticker.C {
			Save("./onebridge.data.json", data)
		}
	}()
}

// ManualSave use this when gracefully exiting
func ManualSave() {
	fmt.Print("Writing onebridge.data.json... ")
	Save("./onebridge.data.json", data)
	fmt.Println("done")
}
