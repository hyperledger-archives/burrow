// Copyright 2019 Monax Industries Limited
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

package project

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullVersion(t *testing.T) {
	assert.Equal(t, History.CurrentVersion().String(), FullVersion())
	commit = "0e90ed60"
	date = "2018-11-07T14:29:28Z"
	assert.Equal(t, fmt.Sprintf("%v+commit.%v+%v", History.CurrentVersion().String(), commit, date), FullVersion())
}
