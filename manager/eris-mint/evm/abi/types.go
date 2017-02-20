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
