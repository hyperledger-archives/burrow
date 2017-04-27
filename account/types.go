package account

import (
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/tendermint/go-crypto"
)

// TODO: [ben] Account and PrivateAccount need to become a pure interface
// and then move the implementation to the manager types.
// Eg, Geth has its accounts, different from BurrowMint

// Account resides in the application state, and is mutated by transactions
// on the blockchain.
// Serialized by wire.[read|write]Reflect
type Account struct {
	Address     []byte        `json:"address"`
	PubKey      crypto.PubKey `json:"pub_key"`
	Sequence    int           `json:"sequence"`
	Balance     int64         `json:"balance"`
	Code        []byte        `json:"code"`         // VM code
	StorageRoot []byte        `json:"storage_root"` // VM storage merkle root.

	Permissions ptypes.AccountPermissions `json:"permissions"`
}

type PrivAccount struct {
	Address []byte         `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}
