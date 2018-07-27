// https://github.com/probonopd/ESP8266HueEmulator
package main

import (
	"context"
	"fmt"
	"github.com/namsral/flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hermanbanken/huemmux/clip"
	"github.com/hermanbanken/huemmux/hue"
)

var details *hue.AdvertiseDetails

func init() {
	details = &hue.AdvertiseDetails{}
	flag.StringVar(&details.FriendlyName, "name", "OneBridge", "bridge friendly name")
	flag.StringVar(&details.LocalIP, "ip", hue.Localip(), "which IP to bind the server to")
	flag.UintVar(&details.LocalHttpPort, "port", 80, "port to bind to")
	details.ApiVersion = "1.23.0"
	details.SwVersion = "20180109"
	details.DatastoreVersion = 72

	var macs, _ = hue.GetMacAddr() // "00:17:88:ff:ff:ff" //
	details.Mac = macs[0]
	details.BridgeID = strings.ToUpper(hue.ConvertMacToBridgeID(macs[0]))
	details.Uuid = "2f402f80-da50-11e1-9b23-" + strings.ToLower(details.BridgeID)
}

func main() {
	flag.Parse()

	hue.Advertise(*details)

	h := mux.NewRouter()
	h.StrictSlash(true)
	h.HandleFunc("/description.xml", sendDescription)
	clip.Register(h.PathPrefix("/api").Subrouter(), details)
	static := http.FileServer(http.Dir("static"))
	h.PathPrefix("/static").Handler(http.StripPrefix("/static", static))
	h.PathPrefix("/debug").Handler(http.StripPrefix("/debug", static))
	h.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { http.Redirect(w, req, "/static/", 302) })

	hub := clip.NewHub()
	h.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) { clip.ServeWs(hub, w, req) })
	go hub.Run()

	loggedRouter := handlers.LoggingHandler(os.Stdout, h)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", details.LocalHttpPort),
		Handler: loggedRouter,
	}

	go func() {
		log.Printf("OneBridge running on http://%s:%v", details.LocalIP, details.LocalHttpPort)
		log.Fatal(srv.ListenAndServe())
		proxy := goproxy.NewProxyHttpServer()
		proxy.Verbose = true
	}()

	<-stop

	log.Printf("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
