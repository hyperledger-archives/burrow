package payload

import (
	"fmt"
)

/*
Payload (Transaction) is an atomic operation on the ledger state.

Account Txs:
 - SendTx         Send coins to address
 - CallTx         Send a msg to a contract that runs in the vm
 - NameTx	  Store some value under a name in the global namereg

Validation Txs:
 - BondTx         New validator posts a bond
 - UnbondTx       Validator leaves

Admin Txs:
 - PermsTx
*/

type Type uint32

// Types of Payload implementations
const (
	TypeUnknown = Type(0x00)
	// Account transactions
	TypeSend = Type(0x01)
	TypeCall = Type(0x02)
	TypeName = Type(0x03)

	// Validation transactions
	TypeBond   = Type(0x11)
	TypeUnbond = Type(0x12)

	// Admin transactions
	TypePermissions = Type(0x21)
	TypeGovernance  = Type(0x22)
)

var nameFromType = map[Type]string{
	TypeUnknown:     "UnknownTx",
	TypeSend:        "SendTx",
	TypeCall:        "CallTx",
	TypeName:        "NameTx",
	TypeBond:        "BondTx",
	TypeUnbond:      "UnbondTx",
	TypePermissions: "PermsTx",
	TypeGovernance:  "GovTx",
}

var typeFromName = make(map[string]Type)

func init() {
	for t, n := range nameFromType {
		typeFromName[n] = t
	}
}

func TxTypeFromString(name string) Type {
	return typeFromName[name]
}

func (typ Type) String() string {
	name, ok := nameFromType[typ]
	if ok {
		return name
	}
	return "UnknownTx"
}

func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
}

func (typ *Type) UnmarshalText(data []byte) error {
	*typ = TxTypeFromString(string(data))
	return nil
}

// Protobuf support
func (typ Type) Marshal() ([]byte, error) {
	return typ.MarshalText()
}

func (typ *Type) Unmarshal(data []byte) error {
	return typ.UnmarshalText(data)
}

type Payload interface {
	String() string
	GetInputs() []*TxInput
	Type() Type
	Any() *Any
	// The serialised size in bytes
	Size() int
}

func New(txType Type) (Payload, error) {
	switch txType {
	case TypeSend:
		return &SendTx{}, nil
	case TypeCall:
		return &CallTx{}, nil
	case TypeName:
		return &NameTx{}, nil
	case TypeBond:
		return &BondTx{}, nil
	case TypeUnbond:
		return &UnbondTx{}, nil
	case TypePermissions:
		return &PermsTx{}, nil
	case TypeGovernance:
		return &GovTx{}, nil
	}
	return nil, fmt.Errorf("unknown payload type: %d", txType)
}
