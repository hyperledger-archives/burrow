package burrow

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"golang.org/x/net/context"
)

type transactionServer struct {
	service *rpc.Service
	txCodec txs.Codec
	reader  state.Reader
}

func NewTransactionServer(service *rpc.Service, reader state.Reader, txCodec txs.Codec) TransactionServer {
	return &transactionServer{
		service: service,
		reader:  reader,
		txCodec: txCodec,
	}
}

func (ts *transactionServer) BroadcastTx(ctx context.Context, param *TxParam) (*TxReceipt, error) {
	receipt, err := ts.service.Transactor().BroadcastTxRaw(param.GetTx())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactionServer) Call(ctx context.Context, param *CallParam) (*CallResult, error) {
	fromAddress, err := crypto.AddressFromBytes(param.GetFrom())
	if err != nil {
		return nil, err
	}
	address, err := crypto.AddressFromBytes(param.GetAddress())
	if err != nil {
		return nil, err
	}
	call, err := ts.service.Transactor().Call(ts.reader, fromAddress, address, param.GetData())
	return &CallResult{
		Return:  call.Return,
		GasUsed: call.GasUsed,
	}, nil
}

func (ts *transactionServer) CallCode(ctx context.Context, param *CallCodeParam) (*CallResult, error) {
	fromAddress, err := crypto.AddressFromBytes(param.GetFrom())
	if err != nil {
		return nil, err
	}
	call, err := ts.service.Transactor().CallCode(ts.reader, fromAddress, param.GetCode(), param.GetData())
	return &CallResult{
		Return:  call.Return,
		GasUsed: call.GasUsed,
	}, nil
}

func (ts *transactionServer) Transact(ctx context.Context, param *TransactParam) (*TxReceipt, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	address, err := crypto.MaybeAddressFromBytes(param.GetAddress())
	if err != nil {
		return nil, err
	}
	receipt, err := ts.service.Transactor().Transact(inputAccount, address, param.GetData(), param.GetGasLimit(),
		param.GetFee())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactionServer) TransactAndHold(ctx context.Context, param *TransactParam) (*EventDataCall, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	address, err := crypto.MaybeAddressFromBytes(param.GetAddress())
	if err != nil {
		return nil, err
	}
	edt, err := ts.service.Transactor().TransactAndHold(ctx, inputAccount, address, param.GetData(),
		param.GetGasLimit(), param.GetFee())
	if err != nil {
		return nil, err
	}
	return eventDataCall(edt), nil
}

func (ts *transactionServer) Send(ctx context.Context, param *SendParam) (*TxReceipt, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	toAddress, err := crypto.AddressFromBytes(param.GetToAddress())
	if err != nil {
		return nil, err
	}
	receipt, err := ts.service.Transactor().Send(inputAccount, toAddress, param.GetAmount())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactionServer) SendAndHold(ctx context.Context, param *SendParam) (*TxReceipt, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	toAddress, err := crypto.AddressFromBytes(param.GetToAddress())
	if err != nil {
		return nil, err
	}
	receipt, err := ts.service.Transactor().SendAndHold(ctx, inputAccount, toAddress, param.GetAmount())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactionServer) SignTx(ctx context.Context, param *SignTxParam) (*SignedTx, error) {
	txEnv, err := ts.txCodec.DecodeTx(param.GetTx())
	if err != nil {
		return nil, err
	}
	signers, err := signersFromPrivateAccounts(param.GetPrivateAccounts())
	if err != nil {
		return nil, err
	}
	txEnvSigned, err := ts.service.Transactor().SignTx(txEnv, signers)
	if err != nil {
		return nil, err
	}
	bs, err := ts.txCodec.EncodeTx(txEnvSigned)
	if err != nil {
		return nil, err
	}
	return &SignedTx{
		Tx: bs,
	}, nil
}

func (ts *transactionServer) inputAccount(inAcc *InputAccount) (*execution.SequentialSigningAccount, error) {
	return ts.service.SigningAccount(inAcc.GetAddress(), inAcc.GetPrivateKey())
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

func signersFromPrivateAccounts(privateAccounts []*PrivateAccount) ([]acm.AddressableSigner, error) {
	signers := make([]acm.AddressableSigner, len(privateAccounts))
	var err error
	for i, pa := range privateAccounts {
		signers[i], err = privateAccount(pa)
		if err != nil {
			return nil, err
		}
	}
	return signers, nil
}

func privateAccount(privateAccount *PrivateAccount) (acm.PrivateAccount, error) {
	privateKey, err := crypto.PrivateKeyFromRawBytes(privateAccount.PrivateKey, crypto.CurveTypeEd25519)
	if err != nil {
		return nil, err
	}
	publicKey := privateKey.GetPublicKey()
	return acm.ConcretePrivateAccount{
		Address:    publicKey.Address(),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}.PrivateAccount(), nil
}
