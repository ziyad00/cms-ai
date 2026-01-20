package api

import (
	"crypto/rand"
	"encoding/hex"
)

func newID(prefix string) string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return prefix + "-unknown"
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}
