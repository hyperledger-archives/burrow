package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

func Transfer(srcPort, srcChannel string, amount sdk.Coins, sender, receiver []byte, source bool) {
	types.NewMsgTransfer(srcPort, srcChannel, amount, sender, receiver, source)
}
