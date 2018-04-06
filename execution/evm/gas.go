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

package evm

const (
	GasSha3          uint64 = 1
	GasGetAccount    uint64 = 1
	GasStorageUpdate uint64 = 1
	GasCreateAccount uint64 = 1

	GasBaseOp  uint64 = 0 // TODO: make this 1
	GasStackOp uint64 = 1

	GasEcRecover     uint64 = 1
	GasSha256Word    uint64 = 1
	GasSha256Base    uint64 = 1
	GasRipemd160Word uint64 = 1
	GasRipemd160Base uint64 = 1
	GasIdentityWord  uint64 = 1
	GasIdentityBase  uint64 = 1
)
