package util

import (
	"regexp"
)

var hexRe = regexp.MustCompile(`^[0-9a-fA-F]*$`)
var addrRe = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
var pubRe = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
var privRe = regexp.MustCompile(`^[0-9a-fA-F]{128}$`)

// Is the candidate a proper hex string.
func IsHex(str string) bool {
	return hexRe.MatchString(str)
}

// Is the candidate a hash (32 bytes, same as public keys)
func IsHash(str string) bool {
	return pubRe.MatchString(str)
}

// Is the candidate a proper public address string (20 bytes, hex).
func IsAddress(str string) bool {
	return addrRe.MatchString(str)
}

// Is the candidate a public key string (32 bytes). This is not a good name.
func IsPubKey(str string) bool {
	return pubRe.MatchString(str)
}

// Is the candidate a private key string (64 bytes, hex). This is not a good name.
func IsPrivKey(str string) bool {
	return privRe.MatchString(str)
}