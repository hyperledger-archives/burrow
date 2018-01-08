package uuid

import (
	"bytes"
	"fmt"

	"github.com/wayn3h0/go-uuid/internal/dcesecurity"
	"github.com/wayn3h0/go-uuid/internal/layout"
	"github.com/wayn3h0/go-uuid/internal/namebased/md5"
	"github.com/wayn3h0/go-uuid/internal/namebased/sha1"
	"github.com/wayn3h0/go-uuid/internal/random"
	"github.com/wayn3h0/go-uuid/internal/timebased"
	"github.com/wayn3h0/go-uuid/internal/version"
)

var (
	Nil = UUID{} // Nil UUID
)

// NewTimeBased returns a new time based UUID (version 1).
func NewTimeBased() (UUID, error) {
	u, err := timebased.New()
	if err != nil {
		return Nil, err
	}

	uuid := UUID{}
	copy(uuid[:], u)

	return uuid, nil
}

// NewV1 same as NewTimeBased.
func NewV1() (UUID, error) {
	return NewTimeBased()
}

// NewDCESecurity returns a new DCE security UUID (version 2).
func NewDCESecurity(domain Domain) (UUID, error) {
	u, err := dcesecurity.New(dcesecurity.Domain(domain))
	if err != nil {
		return Nil, err
	}

	uuid := UUID{}
	copy(uuid[:], u)

	return uuid, nil
}

// NewV2 same as NewDCESecurity.
func NewV2(domain Domain) (UUID, error) {
	return NewDCESecurity(domain)
}

// NewNameBasedMD5 returns a new name based UUID with MD5 hash (version 3).
func NewNameBasedMD5(namespace, name string) (UUID, error) {
	u, err := md5.New(namespace, name)
	if err != nil {
		return Nil, err
	}

	uuid := UUID{}
	copy(uuid[:], u)

	return uuid, nil
}

// NewV3 same as NewNameBasedMD5.
func NewV3(namespace, name string) (UUID, error) {
	return NewNameBasedMD5(namespace, name)
}

// NewRandom returns a new random UUID (version 4).
func NewRandom() (UUID, error) {
	u, err := random.New()
	if err != nil {
		return Nil, err
	}

	uuid := UUID{}
	copy(uuid[:], u)

	return uuid, nil
}

// NewV4 same as NewRandom.
func NewV4() (UUID, error) {
	return NewRandom()
}

// New same as NewRandom.
func New() (UUID, error) {
	return NewRandom()
}

// NewNameBasedSHA1 returns a new name based UUID with SHA1 hash (version 5).
func NewNameBasedSHA1(namespace, name string) (UUID, error) {
	u, err := sha1.New(namespace, name)
	if err != nil {
		return Nil, err
	}

	uuid := UUID{}
	copy(uuid[:], u)

	return uuid, nil
}

// NewV5 same as NewNameBasedSHA1.
func NewV5(namespace, name string) (UUID, error) {
	return NewNameBasedSHA1(namespace, name)
}

// UUID respresents an UUID type compliant with specification in RFC 4122.
type UUID [16]byte

// Layout returns layout of UUID.
func (this UUID) Layout() Layout {
	return Layout(layout.Get(this[:]))
}

// Version returns version of UUID.
func (this UUID) Version() Version {
	return Version(version.Get(this[:]))
}

// Equal returns true if current uuid equal to passed uuid.
func (this UUID) Equal(another UUID) bool {
	return bytes.EqualFold(this[:], another[:])
}

// Format returns the formatted string of UUID.
func (this UUID) Format(style Style) string {
	switch style {
	case StyleWithoutDash:
		return fmt.Sprintf("%x", this[:])
	//case StyleStandard:
	default:
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", this[:4], this[4:6], this[6:8], this[8:10], this[10:])
	}
}

// String returns the string of UUID with standard style(xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx | 8-4-4-4-12).
func (this UUID) String() string {
	return this.Format(StyleStandard)
}
