package rpctransactor

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/execution/pbtransactor"
	"github.com/hyperledger/burrow/txs"
	"golang.org/x/net/context"
)

type transactorServer struct {
	transactor *execution.Transactor
	accounts   *execution.Accounts
	txCodec    txs.Codec
	reader     state.Reader
}

func NewTransactorServer(transactor *execution.Transactor, accounts *execution.Accounts, reader state.Reader,
	txCodec txs.Codec) pbtransactor.TransactorServer {
	return &transactorServer{
		transactor: transactor,
		accounts:   accounts,
		reader:     reader,
		txCodec:    txCodec,
	}
}

func (ts *transactorServer) BroadcastTx(ctx context.Context, param *pbtransactor.TxParam) (*pbtransactor.TxReceipt, error) {
	receipt, err := ts.transactor.BroadcastTxRaw(param.GetTx())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactorServer) Call(ctx context.Context, param *pbtransactor.CallParam) (*pbtransactor.CallResult, error) {
	fromAddress, err := crypto.AddressFromBytes(param.GetFrom())
	if err != nil {
		return nil, err
	}
	address, err := crypto.AddressFromBytes(param.GetAddress())
	if err != nil {
		return nil, err
	}
	call, err := ts.transactor.Call(ts.reader, fromAddress, address, param.GetData())
	return &pbtransactor.CallResult{
		Return:  call.Return,
		GasUsed: call.GasUsed,
	}, nil
}

func (ts *transactorServer) CallCode(ctx context.Context, param *pbtransactor.CallCodeParam) (*pbtransactor.CallResult, error) {
	fromAddress, err := crypto.AddressFromBytes(param.GetFrom())
	if err != nil {
		return nil, err
	}
	call, err := ts.transactor.CallCode(ts.reader, fromAddress, param.GetCode(), param.GetData())
	return &pbtransactor.CallResult{
		Return:  call.Return,
		GasUsed: call.GasUsed,
	}, nil
}

func (ts *transactorServer) Transact(ctx context.Context, param *pbtransactor.TransactParam) (*pbtransactor.TxReceipt, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	address, err := crypto.MaybeAddressFromBytes(param.GetAddress())
	if err != nil {
		return nil, err
	}
	receipt, err := ts.transactor.Transact(inputAccount, address, param.GetData(), param.GetGasLimit(), param.GetValue(),
		param.GetFee())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactorServer) TransactAndHold(ctx context.Context, param *pbtransactor.TransactParam) (*pbevents.EventDataCall, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	address, err := crypto.MaybeAddressFromBytes(param.GetAddress())
	if err != nil {
		return nil, err
	}
	edt, err := ts.transactor.TransactAndHold(ctx, inputAccount, address, param.GetData(), param.GetGasLimit(),
		param.GetValue(), param.GetFee())
	if err != nil {
		return nil, err
	}
	return pbevents.GetEventDataCall(edt), nil
}

func (ts *transactorServer) Send(ctx context.Context, param *pbtransactor.SendParam) (*pbtransactor.TxReceipt, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	toAddress, err := crypto.AddressFromBytes(param.GetToAddress())
	if err != nil {
		return nil, err
	}
	receipt, err := ts.transactor.Send(inputAccount, toAddress, param.GetAmount())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactorServer) SendAndHold(ctx context.Context, param *pbtransactor.SendParam) (*pbtransactor.TxReceipt, error) {
	inputAccount, err := ts.inputAccount(param.GetInputAccount())
	if err != nil {
		return nil, err
	}
	toAddress, err := crypto.AddressFromBytes(param.GetToAddress())
	if err != nil {
		return nil, err
	}
	receipt, err := ts.transactor.SendAndHold(ctx, inputAccount, toAddress, param.GetAmount())
	if err != nil {
		return nil, err
	}
	return txReceipt(receipt), nil
}

func (ts *transactorServer) SignTx(ctx context.Context, param *pbtransactor.SignTxParam) (*pbtransactor.SignedTx, error) {
	txEnv, err := ts.txCodec.DecodeTx(param.GetTx())
	if err != nil {
		return nil, err
	}
	signers, err := signersFromPrivateAccounts(param.GetPrivateAccounts())
	if err != nil {
		return nil, err
	}
	txEnvSigned, err := ts.transactor.SignTx(txEnv, signers)
	if err != nil {
		return nil, err
	}
	bs, err := ts.txCodec.EncodeTx(txEnvSigned)
	if err != nil {
		return nil, err
	}
	return &pbtransactor.SignedTx{
		Tx: bs,
	}, nil
}

func (ts *transactorServer) inputAccount(inAcc *pbtransactor.InputAccount) (*execution.SequentialSigningAccount, error) {
	return ts.accounts.GetSequentialSigningAccount(inAcc.GetAddress(), inAcc.GetPrivateKey())
}

func txReceipt(receipt *txs.Receipt) *pbtransactor.TxReceipt {
	return &pbtransactor.TxReceipt{
		ContractAddress: receipt.ContractAddress.Bytes(),
		CreatesContract: receipt.CreatesContract,
		TxHash:          receipt.TxHash,
	}
}

func signersFromPrivateAccounts(privateAccounts []*pbtransactor.PrivateAccount) ([]acm.AddressableSigner, error) {
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

func privateAccount(privateAccount *pbtransactor.PrivateAccount) (acm.PrivateAccount, error) {
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
