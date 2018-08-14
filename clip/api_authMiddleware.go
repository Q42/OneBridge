package clip

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/context"
)

type key int

const authUser key = 0

// Middleware function, which will be called for each request
func (bridge *Bridge) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := urlPart(r.RequestURI, 2)
		for idx, u := range data.Self.Users {
			if u.ID == username || username == "0" {
				// We found the token in our map
				context.Set(r, authUser, username)
				now := time.Now()
				data.Self.Users[idx].LastUseDate = now.Format("2006-01-02T15:04:05")
				// Pass down the request to the next middleware (or final handler)
				next.ServeHTTP(w, r)
				return
			}
		}

		// If we hit this, there are no users yet
		if username == "0" {
			current := time.Now()
			user := BridgeUser{
				Type:       "hue",
				ID:         strings.ToLower(randomHexString(16)),
				DeviceType: "OneBridge#FirstUser",
				CreateDate: current.Format("2006-01-02T15:04:05"),
			}
			data.Self.Users = append(data.Self.Users, user)
			context.Set(r, authUser, user.ID)
			next.ServeHTTP(w, r)
		}

		httpError(r)(w, "Forbidden", http.StatusForbidden)
	})
}
