package clip

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/q42/onebridge/hue"

	"github.com/gorilla/mux"
)

// Register clip routes
func Register(r *mux.Router, details *hue.AdvertiseDetails) {
	r.HandleFunc("/nouser/config", serveConfigNoAuth(details)).Methods("GET")
	r.HandleFunc("/", linkNewUser(details)).Methods("POST")
	r.HandleFunc("/nupnp", nupnp).Methods("GET")

	authed := r.PathPrefix("/").Subrouter()
	authed.NotFoundHandler = notFound

	// sse? http://localhost:8083/v2_pre1/api
	authed.Use(data.Self.authMiddleware)
	authed.HandleFunc("/{username}/bridges", apiGetDelegates).Methods("GET")
	authed.HandleFunc("/{username}/bridges", apiAddDelegate(details)).Methods("POST")
	authed.HandleFunc("/{username}/bridges/{id}", apiUpdateDelegate(details)).Methods("PUT")
	authed.HandleFunc("/{username}", serveRoot(details)).Methods("GET")
	authed.HandleFunc("/{username}/config", serveConfig(details)).Methods("GET")
	authed.HandleFunc("/{username}/config", putConfig).Methods("PUT")
	authed.HandleFunc("/{username}/{resourcetype}", resourceList).Methods("GET")
	authed.HandleFunc("/{username}/{resourcetype}/new", resourceNew).Methods("GET")
	authed.HandleFunc("/{username}/{resourcetype}/{resourceid}", resourceSingle(details)).Methods("GET")
	authed.HandleFunc("/{username}/{resourcetype}/{resourceid}", resourceUpdateBody).Methods("PUT")
	authed.HandleFunc("/{username}/{resourcetype}/{resourceid}/{field}", resourceUpdate).Methods("PUT")
}

var notFound = &clipCatchAll{}

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
