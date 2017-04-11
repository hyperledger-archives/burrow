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

package logging

import (
	"github.com/go-kit/kit/log/term"
	"github.com/hyperledger/burrow/logging/structure"
)

func Colors(keyvals ...interface{}) term.FgBgColor {
	for i := 0; i < len(keyvals)-1; i += 2 {
		if keyvals[i] != structure.LevelKey {
			continue
		}
		switch keyvals[i+1] {
		case "debug":
			return term.FgBgColor{Fg: term.DarkGray}
		case "info":
			return term.FgBgColor{Fg: term.Gray}
		case "warn":
			return term.FgBgColor{Fg: term.Yellow}
		case "error":
			return term.FgBgColor{Fg: term.Red}
		case "crit":
			return term.FgBgColor{Fg: term.Gray, Bg: term.DarkRed}
		default:
			return term.FgBgColor{}
		}
	}
	return term.FgBgColor{}
}
