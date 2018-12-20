package clip

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/q42/onebridge/hue"

	"github.com/gorilla/mux"
)

var details = hue.AdvertiseDetails{}

var router = func() *mux.Router {
	h := mux.NewRouter()
	h.StrictSlash(true)
	subrouter := h.PathPrefix("/api").Subrouter()
	Register(subrouter, &details)
	return h
}()

func TestApiRootPostWithoutBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/", nil)
	req.Header.Add("User-Agent", "Hue")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the response body is what we expect.
	expected := `.*error.*`
	matched, err := regexp.MatchString(expected, rr.Body.String())
	if !matched || err != nil {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestApiRootPostOk(t *testing.T) {
	payload := fmt.Sprintf(`{ "devicetype": "%s" }`, "Golang#Testing")

	req, err := http.NewRequest("POST", "/api/", strings.NewReader(payload))
	req.Header.Add("User-Agent", "Hue")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `[{"success":{"username":".*"}}]`
	matched, err := regexp.MatchString(expected, rr.Body.String())
	if !matched || err != nil {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	if len(data.Self.Users) != 1 {
		t.Error("Should add a User")
	}
	if data.Self.Users[0].DeviceType != "Golang#Testing" {
		t.Error("Should register the POSTed deviceType")
	}

}
