package ibc

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func NewConnectionOpenInit(connectionID, clientID, counterpartyConnectionID, counterpartyClientID string, counterpartyPrefix commitment.PrefixI) {
	types.NewMsgConnectionOpenInit(connectionID, clientID, counterpartyConnectionID, counterpartyClientID, counterpartyPrefix, []byte{})
}

// func NewConnectionOpenTry() {
// 	types.NewMsgConnectionOpenTry()
// }

// func NewConnectionOpenAck() {
// 	types.NewMsgConnectionOpenAck()
// }

// func NewConnectionOpenConfirm() {
// 	types.NewMsgConnectionOpenConfirm()
// }
