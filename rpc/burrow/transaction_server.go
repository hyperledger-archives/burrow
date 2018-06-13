package burrow

import (
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"golang.org/x/net/context"
)

type transactionServer struct {
	service *rpc.Service
}

func NewTransactionServer(service *rpc.Service) TransactionServer {
	return &transactionServer{service}
}

func (ts *transactionServer) BroadcastTx(ctx context.Context, param *TxParam) (*TxReceipt, error) {
	receipt, err := ts.service.Transactor().BroadcastTxRaw(param.Tx)
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactionServer) Call(context.Context, *CallParam) (*CallResult, error) {
	panic("implement me")
}

func (ts *transactionServer) CallCode(context.Context, *CallCodeParam) (*CallResult, error) {
	panic("implement me")
}

func (ts *transactionServer) Transact(ctx context.Context, param *TransactParam) (*TxReceipt, error) {
	inputAccount, err := ts.service.SigningAccount(param.InputAccount.Address, param.InputAccount.PrivateKey)
	if err != nil {
		return nil, err
	}
	address, err := crypto.MaybeAddressFromBytes(param.Address)
	if err != nil {
		return nil, err
	}
	receipt, err := ts.service.Transactor().Transact(inputAccount, address, param.Data, param.GasLimit, param.Fee)
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactionServer) TransactAndHold(ctx context.Context, param *TransactParam) (*EventDataCall, error) {
	inputAccount, err := ts.service.SigningAccount(param.InputAccount.Address, param.InputAccount.PrivateKey)
	if err != nil {
		return nil, err
	}
	address, err := crypto.MaybeAddressFromBytes(param.Address)
	if err != nil {
		return nil, err
	}
	edt, err := ts.service.Transactor().TransactAndHold(inputAccount, address, param.Data, param.GasLimit, param.Fee)
	if err != nil {
		return nil, err
	}
	return eventDataCall(edt), nil
}

func (ts *transactionServer) Send(context.Context, *SendParam) (*TxReceipt, error) {
	panic("implement me")
}

func (ts *transactionServer) SendAndHold(context.Context, *SendParam) (*TxReceipt, error) {
	panic("implement me")
}

func (ts *transactionServer) SignTx(context.Context, *SignTxParam) (*SignedTx, error) {
	panic("implement me")
}

func eventDataCall(edt *events.EventDataCall) *EventDataCall {
	return &EventDataCall{
		Origin:     edt.Origin.Bytes(),
		TxHash:     edt.TxHash,
		CallData:   callData(edt.CallData),
		StackDepth: int64(edt.StackDepth),
		Return:     edt.Return,
		Exception:  edt.Exception.Error(),
	}
}
func callData(cd *events.CallData) *CallData {
	return &CallData{
		Caller: cd.Caller.Bytes(),
		Callee: cd.Callee.Bytes(),
		Data:   cd.Data,
		Gas:    cd.Gas,
	}
}

func txReceipt(receipt *txs.Receipt) *TxReceipt {
	return &TxReceipt{
		ContractAddress: receipt.ContractAddress.Bytes(),
		CreatesContract: receipt.CreatesContract,
		TxHash:          receipt.TxHash,
	}
}
