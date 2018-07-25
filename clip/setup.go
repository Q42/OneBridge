package clip

import (
	"io"
	"net/http"
	"time"
)

type NUPNP struct {
	IP                string
	InternalIPAddress string
}

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func nupnp(w http.ResponseWriter, r *http.Request) {
	res, _ := netClient.Get("https://www.meethue.com/api/nupnp")
	io.Copy(w, res.Body)

	// target []NUPNP
	// defer r.Body.Close()
	// return json.NewDecoder(r.Body).Decode(target)
}
