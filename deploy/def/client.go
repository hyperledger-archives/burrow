package def

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"reflect"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis/spec"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Client struct {
	MempoolSigning bool
	// Memoised clients and info
	chainID               string
	transactClient        rpctransact.TransactClient
	queryClient           rpcquery.QueryClient
	executionEventsClient rpcevents.ExecutionEventsClient
	keyClient             keys.KeyClient
}

// Connect GRPC clients using ChainURL
func (c *Client) Dial(chainAddress, keysClientAddress string) error {
	conn, err := grpc.Dial(chainAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.transactClient = rpctransact.NewTransactClient(conn)
	c.queryClient = rpcquery.NewQueryClient(conn)
	c.executionEventsClient = rpcevents.NewExecutionEventsClient(conn)
	if keysClientAddress == "" {
		logrus.Info("Using mempool signing since no keyClient set, pass --keys to sign locally or elsewhere")
		c.MempoolSigning = true
	} else {
		logrus.Info("Using keys server at: %s", keysClientAddress)
		c.keyClient, err = keys.NewRemoteKeyClient(keysClientAddress, logging.NewNoopLogger())
	}

	if err != nil {
		return err
	}
	stat, err := c.Status()
	if err != nil {
		return err
	}
	c.chainID = stat.ChainID
	return nil
}

func (c *Client) Transact() rpctransact.TransactClient {
	return c.transactClient
}

func (c *Client) Query() rpcquery.QueryClient {
	return c.queryClient
}

func (c *Client) Events() rpcevents.ExecutionEventsClient {
	return c.executionEventsClient
}

func (c *Client) Status() (*rpc.ResultStatus, error) {
	return c.queryClient.Status(context.Background(), &rpcquery.StatusParam{})
}

func (c *Client) GetAccount(address crypto.Address) (*acm.ConcreteAccount, error) {
	return c.queryClient.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: address})
}

func (c *Client) GetName(name string) (*names.Entry, error) {
	return c.queryClient.GetName(context.Background(), &rpcquery.GetNameParam{Name: name})
}

func (c *Client) GetValidatorSet() (*rpcquery.ValidatorSet, error) {
	return c.queryClient.GetValidatorSet(context.Background(), &rpcquery.GetValidatorSetParam{IncludeHistory: true})
}

func (c *Client) SignAndBroadcast(tx payload.Payload) (*exec.TxExecution, error) {
	txEnv, err := c.SignTx(tx)
	if err != nil {
		return nil, err
	}
	return c.BroadcastEnvelope(txEnv)
}

func (c *Client) SignTx(tx payload.Payload) (*txs.Envelope, error) {
	txEnv := txs.Enclose(c.chainID, tx)
	if c.MempoolSigning {
		logrus.Info("Using mempool signing")
		return txEnv, nil
	}
	var err error
	inputs := tx.GetInputs()
	signers := make([]acm.AddressableSigner, len(inputs))
	for i, input := range inputs {
		signers[i], err = keys.AddressableSigner(c.keyClient, input.Address)
		if err != nil {
			return nil, err
		}
	}
	err = txEnv.Sign(signers...)
	if err != nil {
		return nil, err
	}
	return txEnv, nil
}

// Creates a keypair using attached keys service
func (c *Client) CreateKey(keyName, curveTypeString string) (crypto.PublicKey, error) {
	if c.keyClient == nil {
		return crypto.PublicKey{}, fmt.Errorf("could not create key pair since no keys service is attached, " +
			"pass --keys flag")
	}
	var err error
	curveType := crypto.CurveTypeEd25519
	if curveTypeString != "" {
		curveType, err = crypto.CurveTypeFromString(curveTypeString)
		if err != nil {
			return crypto.PublicKey{}, err
		}
	}
	address, err := c.keyClient.Generate(keyName, curveType)
	if err != nil {
		return crypto.PublicKey{}, err
	}
	return c.keyClient.PublicKey(address)
}

// Broadcast payload for remote signing
func (c *Client) Broadcast(tx payload.Payload) (*exec.TxExecution, error) {
	return c.transactClient.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{Payload: tx.Any()})
}

// Broadcast envelope - can be locally signed or remote signing will be attempted
func (c *Client) BroadcastEnvelope(txEnv *txs.Envelope) (*exec.TxExecution, error) {
	return c.transactClient.BroadcastTxSync(context.Background(), &rpctransact.TxEnvelopeParam{Envelope: txEnv})
}

func (c *Client) ParseUint64(amount string) (uint64, error) {
	if amount == "" {
		return 0, nil
	}
	return strconv.ParseUint(amount, 10, 64)
}

// Simulated call

type QueryArg struct {
	Input   string
	Address string
	Data    string
}

func (c *Client) QueryContract(arg *QueryArg) (*exec.TxExecution, error) {
	logArg("Query contract", arg)
	tx, err := c.Call(&CallArg{
		Input:   arg.Input,
		Address: arg.Address,
		Data:    arg.Data,
	})
	if err != nil {
		return nil, err
	}
	return c.transactClient.CallTxSim(context.Background(), tx)
}

// Transaction types

type GovArg struct {
	Input       string
	Native      string
	Power       string
	Sequence    string
	Permissions []string
	Roles       []string
	Address     string
	PublicKey   string
}

func (c *Client) UpdateAccount(arg *GovArg) (*payload.GovTx, error) {
	logArg("GovTx", arg)
	input, err := c.TxInput(arg.Input, arg.Native, arg.Sequence)
	if err != nil {
		return nil, err
	}
	update := &spec.TemplateAccount{
		Permissions: arg.Permissions,
		Roles:       arg.Permissions,
	}
	if arg.Address != "" {
		address, err := crypto.AddressFromHexString(arg.Address)
		if err != nil {
			return nil, fmt.Errorf("could not parse UpdateAccoount Address: %v", err)
		}
		update.Address = &address
	}
	if arg.PublicKey != "" {
		publicKey, err := publicKeyFromString(arg.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("could not parse UpdateAccount PublicKey: %v", err)
		}
		update.PublicKey = &publicKey
		// Update arg for variable usage
		arg.Address = publicKey.Address().String()
	}
	if update.PublicKey == nil {
		// Attempt to get public key from connected key client
		if update.Address != nil {
			// Try key client
			if c.keyClient != nil {
				publicKey, err := c.keyClient.PublicKey(*update.Address)
				if err != nil {
					logrus.Info("Could not retrieve public key for %v from keys server", *update.Address)
				} else {
					update.PublicKey = &publicKey
				}
			}
			// We can still proceed with just address set
		} else {
			return nil, fmt.Errorf("neither target address or public key were provided to govern account")
		}
	}
	_, err = permission.PermFlagFromStringList(arg.Permissions)
	if err != nil {
		return nil, fmt.Errorf("could not parse UpdateAccoutn permissions: %v", err)
	}

	if arg.Native != "" {
		native, err := c.ParseUint64(arg.Native)
		if err != nil {
			return nil, fmt.Errorf("could not parse native token amount: %v", err)
		}
		update.Amounts = update.Balances().Native(native)
	}
	if arg.Power != "" {
		power, err := c.ParseUint64(arg.Power)
		if err != nil {
			return nil, fmt.Errorf("could not parse native token amount: %v", err)
		}
		update.Amounts = update.Balances().Power(power)
	}
	tx := &payload.GovTx{
		Inputs:         []*payload.TxInput{input},
		AccountUpdates: []*spec.TemplateAccount{update},
	}
	return tx, nil
}

func publicKeyFromString(publicKey string) (crypto.PublicKey, error) {
	bs, err := hex.DecodeString(publicKey)
	if err != nil {
		return crypto.PublicKey{}, fmt.Errorf("could not parse public key string %s as hex: %v", publicKey, err)
	}
	switch len(bs) {
	case crypto.PublicKeyLength(crypto.CurveTypeEd25519):
		return crypto.PublicKeyFromBytes(bs, crypto.CurveTypeEd25519)
	case crypto.PublicKeyLength(crypto.CurveTypeSecp256k1):
		return crypto.PublicKeyFromBytes(bs, crypto.CurveTypeSecp256k1)
	default:
		return crypto.PublicKey{}, fmt.Errorf("public key string %s has byte length %d which is not the size of either "+
			"ed25519 or compressed secp256k1 keys so cannot construct public key", publicKey, len(bs))
	}
}

type CallArg struct {
	Input    string
	Amount   string
	Sequence string
	Address  string
	Fee      string
	Gas      string
	Data     string
}

func (c *Client) Call(arg *CallArg) (*payload.CallTx, error) {
	logArg("CallTx", arg)
	input, err := c.TxInput(arg.Input, arg.Amount, arg.Sequence)
	if err != nil {
		return nil, err
	}
	var contractAddress *crypto.Address
	if arg.Address != "" {
		address, err := crypto.AddressFromHexString(arg.Address)
		if err != nil {
			return nil, err
		}
		contractAddress = &address
	}
	fee, err := c.ParseUint64(arg.Fee)
	if err != nil {
		return nil, err
	}
	gas, err := c.ParseUint64(arg.Gas)
	if err != nil {
		return nil, err
	}
	code, err := hex.DecodeString(arg.Data)
	if err != nil {
		return nil, err
	}
	tx := &payload.CallTx{
		Input:    input,
		Address:  contractAddress,
		Data:     code,
		Fee:      fee,
		GasLimit: gas,
	}
	return tx, nil
}

type SendArg struct {
	Input    string
	Amount   string
	Sequence string
	Output   string
}

func (c *Client) Send(arg *SendArg) (*payload.SendTx, error) {
	logArg("SendTx", arg)
	input, err := c.TxInput(arg.Input, arg.Amount, arg.Sequence)
	if err != nil {
		return nil, err
	}
	outputAddress, err := crypto.AddressFromHexString(arg.Output)
	if err != nil {
		return nil, err
	}
	tx := &payload.SendTx{
		Inputs: []*payload.TxInput{input},
		Outputs: []*payload.TxOutput{{
			Address: outputAddress,
			Amount:  input.Amount,
		}},
	}
	return tx, nil
}

type NameArg struct {
	Input    string
	Amount   string
	Sequence string
	Name     string
	Data     string
	Fee      string
}

func (c *Client) Name(arg *NameArg) (*payload.NameTx, error) {
	logArg("NameTx", arg)
	input, err := c.TxInput(arg.Input, arg.Amount, arg.Sequence)
	if err != nil {
		return nil, err
	}
	fee, err := c.ParseUint64(arg.Fee)
	if err != nil {
		return nil, err
	}
	tx := &payload.NameTx{
		Input: input,
		Name:  arg.Name,
		Data:  arg.Data,
		Fee:   fee,
	}
	return tx, nil
}

type PermArg struct {
	Input      string
	Sequence   string
	Action     string
	Target     string
	Permission string
	Value      string
	Role       string
}

func (c *Client) Permissions(arg *PermArg) (*payload.PermsTx, error) {
	logArg("PermsTx", arg)
	input, err := c.TxInput(arg.Input, "", arg.Sequence)
	if err != nil {
		return nil, err
	}
	action, err := permission.PermStringToFlag(arg.Action)
	if err != nil {
		return nil, err
	}
	permArgs := permission.PermArgs{
		Action: action,
	}
	if arg.Target != "" {
		target, err := crypto.AddressFromHexString(arg.Target)
		if err != nil {
			return nil, err
		}
		permArgs.Target = &target
	}
	if arg.Value != "" {
		var value bool
		switch arg.Value {
		case "true":
			value = true
		case "false":
			value = false
		default:
			return nil, fmt.Errorf("did not recognise value %s as boolean, use 'true' or 'false'", arg.Value)
		}
		permArgs.Value = &value
	}
	if arg.Permission != "" {
		perm, err := permission.PermStringToFlag(arg.Permission)
		if err != nil {
			return nil, err
		}
		permArgs.Permission = &perm
	}

	if arg.Role != "" {
		permArgs.Role = &arg.Role
	}

	tx := &payload.PermsTx{
		Input:    input,
		PermArgs: permArgs,
	}
	return tx, nil
}

func (c *Client) TxInput(inputString, amountString, sequenceString string) (*payload.TxInput, error) {
	var err error
	var inputAddress crypto.Address
	if inputString != "" {
		inputAddress, err = crypto.AddressFromHexString(inputString)
		if err != nil {
			return nil, fmt.Errorf("could not parse input address from %s: %v", inputString, err)
		}
	}
	var amount uint64
	if amountString != "" {
		amount, err = c.ParseUint64(amountString)
	}
	var sequence uint64
	sequence, err = c.GetSequence(sequenceString, inputAddress)
	if err != nil {
		return nil, err
	}
	return &payload.TxInput{
		Address:  inputAddress,
		Amount:   amount,
		Sequence: sequence,
	}, nil
}

func (c *Client) GetSequence(sequence string, inputAddress crypto.Address) (uint64, error) {
	if sequence == "" {
		if c.MempoolSigning {
			// Perform mempool signing
			return 0, nil
		}
		// Get from chain
		acc, err := c.queryClient.GetAccount(context.Background(), &rpcquery.GetAccountParam{Address: inputAddress})
		if err != nil {
			return 0, err
		}
		return acc.Sequence + 1, nil
	}
	return c.ParseUint64(sequence)
}

func argMap(value interface{}) map[string]interface{} {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	rt := rv.Type()
	fields := make(map[string]interface{})
	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).Kind() == reflect.String {
			fields[rt.Field(i).Name] = rv.Field(i).String()
		}
	}
	return fields
}

func logArg(message string, value interface{}) {
	logrus.WithFields(argMap(value)).Info(message)
}
