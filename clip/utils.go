package clip

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

func urlPart(path string, index int) string {
	j := strings.Index(path, "/")
	for index > 0 {
		index = index - 1
		if j >= 0 {
			path = path[(j + 1):]
		} else {
			path = ""
		}
		j = strings.Index(path, "/")
	}
	if j < 0 {
		return path
	}
	return path[:j]
}

func randomClientKey() []byte {
	token := make([]byte, 16)
	rand.Read(token)
	return token
}

func randomHexString(len int) string {
	src := make([]byte, 16)
	rand.Read(src)
	dst := make([]byte, hex.EncodedLen(len))
	hex.Encode(dst, src)
	return fmt.Sprintf("%s", dst)
}
