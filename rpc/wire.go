package rpc

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-crypto"
)

var cdc = amino.NewCodec()

func init() {
	//types.RegisterEvidences(AminoCodec)
	crypto.RegisterAmino(cdc)
	//core_types.RegisterAmino(AminoCodec)
}
