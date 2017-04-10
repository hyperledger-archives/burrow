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

package state

import (
	"bytes"
	"testing"

	"github.com/tendermint/go-wire"
)

func TestStateToFromVMAccount(t *testing.T) {
	acmAcc1, _ := RandAccount(true, 456)
	vmAcc := toVMAccount(acmAcc1)
	acmAcc2 := toStateAccount(vmAcc)

	acmAcc1Bytes := wire.BinaryBytes(acmAcc1)
	acmAcc2Bytes := wire.BinaryBytes(acmAcc2)
	if !bytes.Equal(acmAcc1Bytes, acmAcc2Bytes) {
		t.Errorf("Unexpected account wire bytes\n%X vs\n%X",
			acmAcc1Bytes, acmAcc2Bytes)
	}

}
