// Copyright 2015-2017 Monax Industries Limited.
// This file is part of the Monax platform (Monax)

// Monax is free software: you can use, redistribute it and/or modify
// it only under the terms of the GNU General Public License, version
// 3, as published by the Free Software Foundation.

// Monax is distributed WITHOUT ANY WARRANTY pursuant to
// the terms of the Gnu General Public Licence, version 3, including
// (but not limited to) Clause 15 thereof. See the text of the
// GNU General Public License, version 3 for full terms.

// You should have received a copy of the GNU General Public License,
// version 3, with Monax.  If not, see <http://www.gnu.org/licenses/>.

package sanity

import (
	"fmt"
)

//--------------------------------------------------------------------------------------------------
// panic wrappers
// NOTE: [ben] Fail fast and fail hard, these are wrappers that point to code that needs
// to be addressed, but simplify finding them in the code;

// A panic resulting from a sanity check means there is a programmer error
// and some gaurantee is not satisfied.
func PanicSanity(v interface{}) {
	panic(fmt.Sprintf("Paniced on a Sanity Check: %v", v))
}

// A panic here means something has gone horribly wrong, in the form of data corruption or
// failure of the operating system. In a correct/healthy system, these should never fire.
// If they do, it's indicative of a much more serious problem.
func PanicCrisis(v interface{}) {
	panic(fmt.Sprintf("Paniced on a Crisis: %v", v))
}

// Indicates a failure of consensus. Someone was malicious or something has
// gone horribly wrong. These should really boot us into an "emergency-recover" mode
func PanicConsensus(v interface{}) {
	panic(fmt.Sprintf("Paniced on a Consensus Failure: %v", v))
}

// For those times when we're not sure if we should panic
func PanicQ(v interface{}) {
	panic(fmt.Sprintf("Paniced questionably: %v", v))
}
