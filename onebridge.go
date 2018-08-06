package main

// https://github.com/probonopd/ESP8266HueEmulator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"onebridge/clip"
	"onebridge/hue"

	"github.com/elazarl/goproxy"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
)

var details *hue.AdvertiseDetails

func init() {
	details = &hue.AdvertiseDetails{}
	flag.StringVar(&details.FriendlyName, "name", "OneBridge", "bridge friendly name")
	flag.StringVar(&details.LocalIP, "ip", hue.Localip(), "which IP to bind the server to")
	flag.UintVar(&details.LocalHTTPPort, "port", 80, "port to bind to")
	flag.StringVar(&details.APIVersion, "apiversion", "1.23.0", "bridge api version")
	flag.StringVar(&details.SwVersion, "swversion", "20180109", "bridge software version")
	flag.IntVar(&details.DatastoreVersion, "datastoreversion", 72, "bridge datastore version")

	var macs, _ = hue.GetMacAddr() // "00:17:88:ff:ff:ff" //
	details.Mac = macs[0]
	details.BridgeID = strings.ToUpper(hue.ConvertMacToBridgeID(macs[0]))
	details.UUID = "2f402f80-da50-11e1-9b23-" + strings.ToLower(details.BridgeID)
}

func main() {
	flag.Parse()

	h := mux.NewRouter()
	h.StrictSlash(true)

	// Serve traditional CLIP API
	hue.Advertise(*details)
	h.HandleFunc("/description.xml", sendDescription)
	clip.Register(h.PathPrefix("/api").Subrouter(), details)

	// Serve WebSocket server for ws OneBridge API
	hub := clip.NewHub()
	h.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) { clip.ServeWs(hub, w, req) })
	go hub.Run()

	// Serve traditional debug tools
	debug := http.FileServer(http.Dir("debug"))
	h.PathPrefix("/debug").Handler(http.StripPrefix("/debug", debug))

	// Serving client React app
	docs := http.FileServer(http.Dir("docs"))
	h.PathPrefix("/OneBridge").Handler(http.StripPrefix("/OneBridge", docs))

	h.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { http.Redirect(w, req, "/OneBridge/", 302) })

	loggedRouter := handlers.LoggingHandler(os.Stdout, h)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", details.LocalHTTPPort),
		Handler: loggedRouter,
	}

	go clip.SetupDatastore()
	clip.RefreshDetails(clip.Bridge{ID: details.BridgeID, Mac: details.Mac, IP: details.LocalIP})

	go func() {
		log.Printf("OneBridge running on http://%s:%v", details.LocalIP, details.LocalHTTPPort)
		log.Fatal(srv.ListenAndServe())
		proxy := goproxy.NewProxyHttpServer()
		proxy.Verbose = true
	}()

	<-stop

	log.Printf("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clip.ManualSave()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	// log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%v", details.LocalIP, httpsPort), "server.crt", "server.key", h))
	// openssl genrsa -out server.key 2048
	// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
}

func sendDescription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(hue.DescriptionXML(*details)))
}
