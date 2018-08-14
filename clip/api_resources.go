package clip

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

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
