// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package tendermint

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
