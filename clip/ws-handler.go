package clip

import (
	"encoding/json"
)

type Message struct {
	Type string
}

func HandleWebsocket(c *Client, message []byte) {
	var t *Message
	var err = json.Unmarshal(message, &t)
	if err != nil {
		panic(err)
	}

	if t.Type == "request" {
		request(c, message)
	} else if t.Type == "link" {
		link(c, message)
	} else {
		// Default
		c.hub.broadcast <- message
	}
}

type req struct {
	Paths []string
}

func request(c *Client, message []byte) {
	var t *req
	json.Unmarshal(message, &t)

	for _, path := range t.Paths {
		if path == "bridges" {
			wsReplyBridges(c)
		}
	}

}
