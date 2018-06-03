package tm

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/rpc/core/types"
)

var AminoCodec = amino.NewCodec()

func init() {
	//types.RegisterEvidences(AminoCodec)
	//crypto.RegisterAmino(AminoCodec)
	core_types.RegisterAmino(AminoCodec)
}
