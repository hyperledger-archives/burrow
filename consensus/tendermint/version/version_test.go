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

package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinorVersionTendermintEqual(t *testing.T) {
	// assert explicitly on major and minor version number
	assert.Equal(t, tendermintVersionMajorConst, tendermintVersionMajor,
		fmt.Sprintf("Major version number for Tendermint consensus is not %v as expected: %v",
			tendermintVersionMajorConst, tendermintVersionMajor))
	assert.Equal(t, tendermintVersionMinorConst, tendermintVersionMinor,
		fmt.Sprintf("Minor version number for Tendermint consensus is not %v as expected: %v",
			tendermintVersionMinorConst, tendermintVersionMinor))
	// assert patch number can not regress
	if tendermintVersionPatchConst > tendermintVersionPatch {
		t.Errorf("Patch version has regressed for Tendermint consensus: expected minimally %v, got %v",
			tendermintVersionPatchConst, tendermintVersionPatch)
		t.Fail()
	}
}

func TestSemanticVersioningTendermint(t *testing.T) {
	// assert that reading the semantic version from Tendermint vendored source
	// succeeds without error; at runtime initialisation, on error we default
	// to hard-coded semantic version
	if _, err := getTendermintMajorVersionFromSource(); err != nil {
		t.Errorf("Failed to read Major version from Tendermint source code: %s", err)
		t.Fail()
	}
	if _, err := getTendermintMinorVersionFromSource(); err != nil {
		t.Errorf("Failed to read Minor version from Tendermint source code: %s", err)
		t.Fail()
	}
	if _, err := getTendermintPatchVersionFromSource(); err != nil {
		t.Errorf("Failed to read Patch version from Tendermint source code: %s", err)
		t.Fail()
	}
}
