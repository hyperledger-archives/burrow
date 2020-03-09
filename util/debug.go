package util

import "fmt"

// If we ever want to ship some specific debug we could make this a string var and set it through -ldflags "-X ..."

const debug = true
//const debug = false

// Using this in place of Printf statements makes it easier to find any errant debug statements and the switch means
// we can turn them off at minimal runtime cost.
func Debugf(format string, args ...interface{}) {
	if debug {
		fmt.Printf(format+"\n", args...)
	}
}
