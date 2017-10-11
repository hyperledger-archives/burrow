package tm

import (
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/core"
	"github.com/hyperledger/burrow/consensus/tendermint/query"
)

func serv() error{


	txCodec := txs.NewGoWireCodec()
	transactor := execution.NewTransactor(blockchain, state, eventEmitter,
		BroadcastTxAsyncFunc(validatorNode, txCodec))

	nameReg := execution.NewNameReg(state, blockchain)


	service := core.NewService(
		state,
		eventEmitter,
		nameReg,
		blockchain,
		transactor,
		query.NewNodeView(validatorNode, txCodec),
		logger,
	)
	_, err := StartServer(service, "/websocket", ":46657", eventSwitch, NewLogger(logger))
	if err != nil {
		return err
	}
	return nil
}
