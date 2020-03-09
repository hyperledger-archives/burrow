package contexts

import (
	"strings"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestNameContext(t *testing.T) {
	accountState := acmstate.NewMemoryState()

	privKey := newPrivKey(t)
	account := newAccountFromPrivKey(privKey)

	db := dbm.NewMemDB()
	genesisDoc, _, _ := genesis.NewDeterministicGenesis(3450976).GenesisDoc(23, 10)
	blockchain := bcm.NewBlockchain(db, genesisDoc)
	state := state.NewState(db)

	ctx := &NameContext{
		State:      accountState,
		Logger:     logging.NewNoopLogger(),
		Blockchain: blockchain,
		NameReg:    names.NewCache(state),
	}

	callTx := &payload.CallTx{}
	err := ctx.Execute(execFromTx(callTx), callTx)
	require.Error(t, err, "should not continue with incorrect payload")

	nameTx := &payload.NameTx{
		Input: &payload.TxInput{
			Address: account.Address,
		},
	}

	err = ctx.Execute(execFromTx(nameTx), nameTx)
	require.Error(t, err, "account should not exist")

	accountState.Accounts[account.Address] = account
	nameTx.Name = "foobar"

	err = ctx.Execute(execFromTx(nameTx), nameTx)
	require.Error(t, err, "insufficient amount")

	costPerBlock := names.NameCostPerBlock(names.NameBaseCost(ctx.tx.Name, ctx.tx.Data))
	nameTx.Input.Amount = names.MinNameRegistrationPeriod * costPerBlock

	err = ctx.Execute(execFromTx(nameTx), nameTx)
	require.NoError(t, err, "should successfully set namereg")
}

func TestValidateStrings(t *testing.T) {
	nameTx := &payload.NameTx{}
	err := validateStrings(nameTx)
	require.Error(t, err, "should fail on empty name")

	nameTx.Name = strings.Repeat("A", names.MaxNameLength+1)
	err = validateStrings(nameTx)
	require.Error(t, err, "should fail because name is too long")

	nameTx.Name = "foo"

	nameTx.Data = strings.Repeat("A", names.MaxDataLength+1)
	err = validateStrings(nameTx)
	require.Error(t, err, "should fail because data is too long")

	nameTx.Data = "bar"
	err = validateStrings(nameTx)
	require.NoError(t, err, "name reg entry should be valid")
}

func newPrivKey(t *testing.T) crypto.PrivateKey {
	privKey, err := crypto.GeneratePrivateKey(nil, crypto.CurveTypeEd25519)
	require.NoError(t, err)
	return privKey
}

func newAccountFromPrivKey(privKey crypto.PrivateKey) *acm.Account {
	pubKey := privKey.GetPublicKey()
	address := pubKey.GetAddress()

	return &acm.Account{
		Address:   address,
		PublicKey: pubKey,
		Balance:   1337,
	}
}

func execFromTx(payl payload.Payload) *exec.TxExecution {
	return &exec.TxExecution{
		Envelope: &txs.Envelope{
			Tx: &txs.Tx{
				Payload: payl,
			},
		},
	}
}
