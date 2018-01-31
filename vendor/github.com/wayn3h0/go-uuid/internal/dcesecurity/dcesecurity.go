package dcesecurity

import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/wayn3h0/go-uuid/internal/layout"
	"github.com/wayn3h0/go-uuid/internal/timebased"
	"github.com/wayn3h0/go-uuid/internal/version"
)

// Generate returns a new DCE security uuid.
func New(domain Domain) ([]byte, error) {
	uuid, err := timebased.New()
	if err != nil {
		return nil, err
	}

	switch domain {
	case User:
		uid := os.Getuid()
		binary.BigEndian.PutUint32(uuid[0:], uint32(uid)) // network byte order
	case Group:
		gid := os.Getgid()
		binary.BigEndian.PutUint32(uuid[0:], uint32(gid)) // network byte order
	default:
		return nil, errors.New("uuid: domain is invalid")
	}

	// set version(v2)
	version.Set(uuid, version.DCESecurity)
	// set layout(RFC4122)
	layout.Set(uuid, layout.RFC4122)

	return uuid, nil
}
