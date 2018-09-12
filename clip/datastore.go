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
	ClientKey   *string
}

// OneBridgeData Datastore root of OneBridge
type OneBridgeData struct {
	Self      Bridge
	Delegates []Bridge
}

var data = OneBridgeData{
	Self:      Bridge{Users: make([]BridgeUser, 0)}, // defaults, instead of `null`
	Delegates: make([]Bridge, 0),                    // defaults, instead of `null`
}

// SetupDatastore will read from file & then write any changes to the data to disk
func SetupDatastore(dir string) {
	file := fmt.Sprintf("%s/onebridge.data.json", dir)
	fmt.Printf("Reading %s...", file)
	if err := Load(file, &data); err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	fmt.Println("done")

	ticker := time.NewTicker(time.Second * 10)
	go func() {
		for range ticker.C {
			Save(file, data)
		}
	}()
}

// ManualSave use this when gracefully exiting
func ManualSave(dir string) {
	file := fmt.Sprintf("%s/onebridge.data.json", dir)
	fmt.Printf("Writing %s...", file)
	Save(file, data)
	fmt.Println("done")
}

// RefreshDetails updates the network details of the bridge in CLIP
func RefreshDetails(bridge Bridge) {
	data.Self.ID = bridge.ID
	data.Self.IP = bridge.IP
	data.Self.Mac = bridge.Mac
}
