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

package abi

// Ethereum defines types and calling conventions for the ABI
// (application binary interface) here: https://github.com/ethereum/wiki/wiki/Ethereum-Contract-ABI
// We make a start of representing them here

type Type string

type Arg struct {
	Name string
	Type Type
}

type Return struct {
	Name string
	Type Type
}

const (
	// We don't need to be exhaustive here, just make what we used strongly typed
	Address Type = "address"
	Int     Type = "int"
	Uint64  Type = "uint64"
	Bytes32 Type = "bytes32"
	String  Type = "string"
	Bool    Type = "bool"
)
