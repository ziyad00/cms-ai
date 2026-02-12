package api

import (
	"github.com/google/uuid"
)

func newID(prefix string) string {
	return uuid.New().String()
}
