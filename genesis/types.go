package genesis

import (
	"fmt"
	"os"
	"time"

	ptypes "github.com/eris-ltd/eris-db/permission/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

//------------------------------------------------------------
// we store the GenesisDoc in the db under this key

var GenDocKey = []byte("GenDocKey")

//------------------------------------------------------------
// core types for a genesis definition

type BasicAccount struct {
	Address []byte `json:"address"`
	Amount  int64  `json:"amount"`
}

type GenesisAccount struct {
	Address     []byte                     `json:"address"`
	Amount      int64                      `json:"amount"`
	Name        string                     `json:"name"`
	Permissions *ptypes.AccountPermissions `json:"permissions"`
}

type GenesisValidator struct {
	PubKey   crypto.PubKey  `json:"pub_key"`
	Amount   int64          `json:"amount"`
	Name     string         `json:"name"`
	UnbondTo []BasicAccount `json:"unbond_to"`
}

// GenesisPrivateValidator marshals the state of the private
// validator for the purpose of Genesis creation; and hence
// is defined in genesis and not under consensus, where
// PrivateValidator (currently inherited from Tendermint) is.
type GenesisPrivateValidator struct {
	Address    []byte         `json:"address"`
	PubKey     crypto.PubKey  `json:"pub_key"`
	PrivKey    crypto.PrivKey `json:"priv_key"`
	LastHeight int64          `json:"last_height"`
	LastRound  int64          `json:"last_round"`
	LastStep   int64          `json:"last_step"`
}

type GenesisParams struct {
	GlobalPermissions *ptypes.AccountPermissions `json:"global_permissions"`
}

//------------------------------------------------------------
// GenesisDoc is stored in the state database

type GenesisDoc struct {
	GenesisTime time.Time          `json:"genesis_time"`
	ChainID     string             `json:"chain_id"`
	Params      *GenesisParams     `json:"params"`
	Accounts    []GenesisAccount   `json:"accounts"`
	Validators  []GenesisValidator `json:"validators"`
}

//------------------------------------------------------------
// Make genesis state from file

func GenesisDocFromJSON(jsonBlob []byte) (genState *GenesisDoc) {
	var err error
	wire.ReadJSONPtr(&genState, jsonBlob, &err)
	if err != nil {
		fmt.Printf("Couldn't read GenesisDoc: %v", err)
		// TODO: on error return error, not exit
		os.Exit(1)
	}
	return
}
