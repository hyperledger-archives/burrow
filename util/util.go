// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
