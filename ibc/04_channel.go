package ibc

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func NewChannelOpenInit(portID, channelID, version string, channelOrder exported.Order, connectionHops []string, counterpartyPortID, counterpartyChannelID string) {
	types.NewMsgChannelOpenInit(portID, channelID, version, channelOrder, connectionHops, counterpartyPortID, counterpartyChannelID, []byte{})
}

// func NewChannelOpenTry() {
// 	types.NewMsgChannelOpenTry()
// }

// func NewChannelOpenAck() {
// 	types.NewMsgChannelOpenAck()
// }

// func NewChannelOpenConfirm() {
// 	types.NewMsgChannelOpenConfirm()
// }

// func NewChannelCloseInit() {
// 	types.NewMsgChannelCloseInit(portID, channelID, cliCtx.GetFromAddress())
// }

// func NewChannelCloseConfirm() {
// 	types.NewMsgChannelCloseConfirm()
// }
