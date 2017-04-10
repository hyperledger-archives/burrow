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

package manager

import (
	burrowmint "github.com/monax/burrow/manager/burrow-mint"
)

//------------------------------------------------------------------------------
// Helper functions

func AssertValidApplicationManagerModule(name, minorVersionString string) bool {
	switch name {
	case "burrowmint":
		return minorVersionString == burrowmint.GetBurrowMintVersion().GetMinorVersionString()
	case "geth":
		// TODO: [ben] implement Geth 1.4 as an application manager
		return false
	}
	return false
}
