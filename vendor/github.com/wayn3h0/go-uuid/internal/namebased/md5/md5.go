package md5

import (
	"crypto/md5"

	"github.com/wayn3h0/go-uuid/internal/layout"
	"github.com/wayn3h0/go-uuid/internal/version"
)

// New returns a new name-based uses SHA-1 hashing uuid.
func New(namespace, name string) ([]byte, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(namespace))
	if err != nil {
		return nil, err
	}
	_, err = hash.Write([]byte(name))
	if err != nil {
		return nil, err
	}

	sum := hash.Sum(nil)

	uuid := make([]byte, 16)
	copy(uuid, sum)

	// set version(v3)
	version.Set(uuid, version.NameBasedMD5)
	// set layout(RFC4122)
	layout.Set(uuid, layout.RFC4122)

	return uuid, nil
}
