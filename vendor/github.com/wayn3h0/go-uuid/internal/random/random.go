package random

import (
	"crypto/rand"

	"github.com/wayn3h0/go-uuid/internal/layout"
	"github.com/wayn3h0/go-uuid/internal/version"
)

// New returns a new randomly uuid.
func New() ([]byte, error) {
	uuid := make([]byte, 16)
	n, err := rand.Read(uuid[:])
	if err != nil {
		return nil, err
	}
	if n != len(uuid) {
		return nil, err
	}

	// set version(v4)
	version.Set(uuid, version.Random)
	// set layout(RFC4122)
	layout.Set(uuid, layout.RFC4122)

	return uuid, nil
}
