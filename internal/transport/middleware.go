package transport

import (
	"crypto/rand"
	"encoding/hex"
)

// randID and hexID are kept here as low-level helpers; the request-id, access-log
// and panic-recovery middleware are methods on Server in http.go so they can
// share the logger and request context.

func randID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
