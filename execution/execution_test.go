// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package execution

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"runtime/debug"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	. "github.com/hyperledger/burrow/binary"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/permission/snatives"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tmthrgd/go-hex"
)

var (
	dbBackend           = dbm.MemDBBackend
	dbDir               = ""
	permissionsContract = evm.SNativeContracts()["Permissions"]
)

/*
Permission Tests:

- SendTx:
x	- 1 input, no perm, call perm, create perm
x	- 1 input, perm
x	- 2 inputs, one with perm one without

- CallTx, CALL
x	- 1 input, no perm, send perm, create perm
x	- 1 input, perm
x	- contract runs call but doesn't have call perm
x	- contract runs call and has call perm
x	- contract runs call (with perm), runs contract that runs call (without perm)
x	- contract runs call (with perm), runs contract that runs call (with perm)

- CallTx for Create, CREATE
x	- 1 input, no perm, send perm, call perm
x 	- 1 input, perm
x	- contract runs create but doesn't have create perm
x	- contract runs create but has perm
x	- contract runs call with empty address (has call and create perm)

- NameTx
	- no perm, send perm, call perm
	- with perm

- BondTx
x	- 1 input, no perm
x	- 1 input, perm
x	- 1 bonder with perm, input without send or bond
x	- 1 bonder with perm, input with send
x	- 1 bonder with perm, input with bond
x	- 2 inputs, one with perm one without

- SendTx for new account
x 	- 1 input, 1 unknown ouput, input with send, not create  (fail)
x 	- 1 input, 1 unknown ouput, input with send and create (pass)
x 	- 2 inputs, 1 unknown ouput, both inputs with send, one with create, one without (fail)
x 	- 2 inputs, 1 known output, 1 unknown ouput, one input with create, one without (fail)
x 	- 2 inputs, 1 unknown ouput, both inputs with send, both inputs with create (pass )
x 	- 2 inputs, 1 known output, 1 unknown ouput, both inputs with create, (pass)


- CALL for new account
x	- unknown output, without create (fail)
x	- unknown output, with create (pass)


- SNative (CallTx, CALL):
	- for each of CallTx, Call
x		- call each snative without permission, fails
x		- call each snative with permission, pass
	- list:
x		- base: has,set,unset
x		- globals: set
x 		- roles: has, add, rm


*/

// keys
var users = makeUsers(10)
var logger = logging.NewNoopLogger()
var deterministicGenesis = genesis.NewDeterministicGenesis(34059836243380576)
var testGenesisDoc, testPrivAccounts, _ = deterministicGenesis.
	GenesisDoc(3, true, 1000, 1, true, 1000)
var testChainID = testGenesisDoc.ChainID()

type testExecutor struct {
	*executor
	blockchain *bcm.Blockchain
	event.Emitter
}

func makeUsers(n int) []acm.AddressableSigner {
	users := make([]acm.AddressableSigner, n)
	for i := 0; i < n; i++ {
		secret := "mysecret" + strconv.Itoa(i)
		users[i] = acm.GeneratePrivateAccountFromSecret(secret)
	}
	return users
}
func newBlockchain(genesisDoc *genesis.GenesisDoc) *bcm.Blockchain {
	testDB := dbm.NewDB("test", dbBackend, ".")
	bc, _ := bcm.LoadOrNewBlockchain(testDB, testGenesisDoc, logger)

	return bc
}

func makeExecutor(state *State) *testExecutor {
	emitter := event.NewEmitter(logger)
	blockchain := newBlockchain(testGenesisDoc)
	return &testExecutor{
		executor:   newExecutor("makeExecutorCache", true, state, blockchain.Tip, emitter, logger),
		blockchain: blockchain,
		Emitter:    emitter,
	}
}

func newBaseGenDoc(globalPerm, accountPerm ptypes.AccountPermissions) genesis.GenesisDoc {
	var genAccounts []genesis.Account
	for _, user := range users[:5] {
		accountPermCopy := accountPerm // Create new instance for custom overridability.
		genAccounts = append(genAccounts, genesis.Account{
			BasicAccount: genesis.BasicAccount{
				Address: user.Address(),
				Amount:  1000000,
			},
			Permissions: accountPermCopy,
		})
	}

	return genesis.GenesisDoc{
		GenesisTime:       time.Now(),
		ChainName:         testGenesisDoc.ChainName,
		GlobalPermissions: globalPerm,
		Accounts:          genAccounts,
		Validators: []genesis.Validator{
			{
				BasicAccount: genesis.BasicAccount{
					PublicKey: users[0].PublicKey(),
					Amount:    10,
				},
				UnbondTo: []genesis.BasicAccount{
					{
						Address: users[0].Address(),
					},
				},
			},
		},
	}
}

//func getAccount(state state.AccountGetter, address acm.Address) acm.MutableAccount {
//	acc, _ := state.GetMutableAccount(state, address)
//	return acc
//}

func TestSendFails(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(ptypes.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(ptypes.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(ptypes.CreateContract, true)
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	exe := makeExecutor(st)

	//-------------------
	// send txs

	// simple send tx should fail
	tx := payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	signAndExecute(t, true, exe, testChainID, tx, users[0])

	// simple send tx with call perm should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[2].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[4].Address(), 5)

	signAndExecute(t, true, exe, testChainID, tx, users[2])

	// simple send tx with create perm should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[3].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[4].Address(), 5)
	signAndExecute(t, true, exe, testChainID, tx, users[3])

	// simple send tx to unknown account without create_account perm should fail
	acc := getAccount(exe.stateCache, users[3].Address())
	acc.MutablePermissions().Base.Set(ptypes.Send, true)
	exe.stateCache.UpdateAccount(acc)
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[3].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].Address(), 5)
	signAndExecute(t, true, exe, testChainID, tx, users[3])
}

func TestName(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Send, true)
	genDoc.Accounts[1].Permissions.Base.Set(ptypes.Name, true)
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//-------------------
	// name txs

	// simple name tx without perm should fail
	tx, err := payload.NewNameTx(st, users[0].PublicKey(), "somename", "somedata", 10000, 100)
	if err != nil {
		t.Fatal(err)
	}
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[0])

	// simple name tx with perm should pass
	tx, err = payload.NewNameTx(st, users[1].PublicKey(), "somename", "somedata", 10000, 100)
	if err != nil {
		t.Fatal(err)
	}
	signAndExecute(t, false, batchCommitter, testChainID, tx, users[1])
}

func TestCallFails(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(ptypes.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(ptypes.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(ptypes.CreateContract, true)
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//-------------------
	// call txs

	address4 := users[4].Address()
	// simple call tx should fail
	tx, _ := payload.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &address4, nil, 100, 100, 100)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[0])

	// simple call tx with send permission should fail
	tx, _ = payload.NewCallTx(batchCommitter.stateCache, users[1].PublicKey(), &address4, nil, 100, 100, 100)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[1])

	// simple call tx with create permission should fail
	tx, _ = payload.NewCallTx(batchCommitter.stateCache, users[3].PublicKey(), &address4, nil, 100, 100, 100)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[3])

	//-------------------
	// create txs

	// simple call create tx should fail
	tx, _ = payload.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), nil, nil, 100, 100, 100)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[0])

	// simple call create tx with send perm should fail
	tx, _ = payload.NewCallTx(batchCommitter.stateCache, users[1].PublicKey(), nil, nil, 100, 100, 100)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[1])

	// simple call create tx with call perm should fail
	tx, _ = payload.NewCallTx(batchCommitter.stateCache, users[2].PublicKey(), nil, nil, 100, 100, 100)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[2])
}

func TestSendPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Send, true) // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	// A single input, having the permission, should succeed
	tx := payload.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	signAndExecute(t, false, batchCommitter, testChainID, tx, users[0])

	// Two inputs, one with permission, one without, should fail
	tx = payload.NewSendTx()
	require.NoError(t, tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5))
	require.NoError(t, tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5))
	require.NoError(t, tx.AddOutput(users[2].Address(), 10))
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[:2]...)
}

func TestCallPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Call, true) // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	exe := makeExecutor(st)

	//------------------------------
	// call to simple contract
	fmt.Println("\n##### SIMPLE CONTRACT")

	// create simple contract
	simpleContractAddr := crypto.NewContractAddress(users[0].Address(), 100)
	simpleAcc := acm.ConcreteAccount{
		Address:     simpleContractAddr,
		Balance:     0,
		Code:        []byte{0x60},
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	st.writeState.UpdateAccount(simpleAcc)

	// A single input, having the permission, should succeed
	tx, _ := payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &simpleContractAddr, nil, 100, 100, 100)
	signAndExecute(t, false, exe, testChainID, tx, users[0])

	//----------------------------------------------------------
	// call to contract that calls simple contract - without perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (FAIL)")

	// create contract that calls the simple contract
	contractCode := callContractCode(simpleContractAddr)
	caller1ContractAddr := crypto.NewContractAddress(users[0].Address(), 101)
	caller1Acc := acm.ConcreteAccount{
		Address:     caller1ContractAddr,
		Balance:     10000,
		Code:        contractCode,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	exe.stateCache.UpdateAccount(caller1Acc)

	// A single input, having the permission, but the contract doesn't have permission
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr) //
	require.Error(t, err)

	//----------------------------------------------------------
	// call to contract that calls simple contract - with perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (PASS)")

	// A single input, having the permission, and the contract has permission
	caller1Acc.MutablePermissions().Base.Set(ptypes.Call, true)
	exe.stateCache.UpdateAccount(caller1Acc)
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr) //
	require.NoError(t, err)

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// caller1Contract does not have call perms, but caller2Contract does.
	fmt.Println("\n##### CALL TO CONTRACT CALLING SIMPLE CONTRACT (FAIL)")

	contractCode2 := callContractCode(caller1ContractAddr)
	caller2ContractAddr := crypto.NewContractAddress(users[0].Address(), 102)
	caller2Acc := acm.ConcreteAccount{
		Address:     caller2ContractAddr,
		Balance:     1000,
		Code:        contractCode2,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	caller1Acc.MutablePermissions().Base.Set(ptypes.Call, false)
	caller2Acc.MutablePermissions().Base.Set(ptypes.Call, true)
	exe.stateCache.UpdateAccount(caller1Acc)
	exe.stateCache.UpdateAccount(caller2Acc)

	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr) //
	require.Error(t, err)

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// both caller1 and caller2 have permission
	fmt.Println("\n##### CALL TO CONTRACT CALLING SIMPLE CONTRACT (PASS)")

	caller1Acc.MutablePermissions().Base.Set(ptypes.Call, true)
	exe.stateCache.UpdateAccount(caller1Acc)

	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr) //
	require.NoError(t, err)
}

func TestCreatePermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.CreateContract, true) // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Call, true)           // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	exe := makeExecutor(st)

	//------------------------------
	// create a simple contract
	fmt.Println("\n##### CREATE SIMPLE CONTRACT")

	contractCode := []byte{0x60}
	createCode := wrapContractForCreate(contractCode)

	// A single input, having the permission, should succeed
	tx, _ := payload.NewCallTx(exe.stateCache, users[0].PublicKey(), nil, createCode, 100, 100, 100)
	signAndExecute(t, false, exe, testChainID, tx, users[0])

	// ensure the contract is there
	contractAddr := crypto.NewContractAddress(tx.Input.Address, tx.Input.Sequence)
	contractAcc := getAccount(exe.stateCache, contractAddr)
	if contractAcc == nil {
		t.Fatalf("failed to create contract %s", contractAddr)
	}
	if !bytes.Equal(contractAcc.Code(), contractCode) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), contractCode)
	}

	//------------------------------
	// create contract that uses the CREATE op
	fmt.Println("\n##### CREATE FACTORY")

	contractCode = []byte{0x60}
	createCode = wrapContractForCreate(contractCode)
	factoryCode := createContractCode()
	createFactoryCode := wrapContractForCreate(factoryCode)

	// A single input, having the permission, should succeed
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), nil, createFactoryCode, 100, 100, 100)
	signAndExecute(t, false, exe, testChainID, tx, users[0])

	// ensure the contract is there
	contractAddr = crypto.NewContractAddress(tx.Input.Address, tx.Input.Sequence)
	contractAcc = getAccount(exe.stateCache, contractAddr)
	if contractAcc == nil {
		t.Fatalf("failed to create contract %s", contractAddr)
	}
	if !bytes.Equal(contractAcc.Code(), factoryCode) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.Code(), factoryCode)
	}

	//------------------------------
	// call the contract (should FAIL)
	fmt.Println("\n###### CALL THE FACTORY (FAIL)")

	// A single input, having the permission, should succeed
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &contractAddr, createCode, 100, 100, 100)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, contractAddr) //
	require.Error(t, err)

	//------------------------------
	// call the contract (should PASS)
	fmt.Println("\n###### CALL THE FACTORY (PASS)")

	contractAcc.MutablePermissions().Base.Set(ptypes.CreateContract, true)
	exe.stateCache.UpdateAccount(contractAcc)

	// A single input, having the permission, should succeed
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &contractAddr, createCode, 100, 100, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, contractAddr) //
	require.NoError(t, err)

	//--------------------------------
	fmt.Println("\n##### CALL to empty address")
	code := callContractCode(crypto.Address{})

	contractAddr = crypto.NewContractAddress(users[0].Address(), 110)
	contractAcc = acm.ConcreteAccount{
		Address:     contractAddr,
		Balance:     1000,
		Code:        code,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	contractAcc.MutablePermissions().Base.Set(ptypes.Call, true)
	contractAcc.MutablePermissions().Base.Set(ptypes.CreateContract, true)
	exe.stateCache.UpdateAccount(contractAcc)

	// this should call the 0 address but not create ...
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &contractAddr, createCode, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, crypto.Address{}) //
	require.NoError(t, err)
	zeroAcc := getAccount(exe.stateCache, crypto.Address{})
	if len(zeroAcc.Code()) != 0 {
		t.Fatal("the zero account was given code from a CALL!")
	}
}

func TestCreateAccountPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Send, true)          // give the 0 account permission
	genDoc.Accounts[1].Permissions.Base.Set(ptypes.Send, true)          // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.CreateAccount, true) // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// SendTx to unknown account

	// A single input, having the permission, should succeed
	tx := payload.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].Address(), 5)
	signAndExecute(t, false, batchCommitter, testChainID, tx, users[0])

	// Two inputs, both with send, one with create, one without, should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 10)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[:2]...)

	// Two inputs, both with send, one with create, one without, two ouputs (one known, one unknown) should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 4)
	tx.AddOutput(users[4].Address(), 6)
	signAndExecute(t, true, batchCommitter, testChainID, tx, users[:2]...)

	// Two inputs, both with send, both with create, should pass
	acc := getAccount(batchCommitter.stateCache, users[1].Address())
	acc.MutablePermissions().Base.Set(ptypes.CreateAccount, true)
	batchCommitter.stateCache.UpdateAccount(acc)
	tx = payload.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 10)
	signAndExecute(t, false, batchCommitter, testChainID, tx, users[:2]...)

	// Two inputs, both with send, both with create, two outputs (one known, one unknown) should pass
	tx = payload.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 7)
	tx.AddOutput(users[4].Address(), 3)
	signAndExecute(t, false, batchCommitter, testChainID, tx, users[:2]...)

	//----------------------------------------------------------
	// CALL to unknown account

	acc = getAccount(batchCommitter.stateCache, users[0].Address())
	acc.MutablePermissions().Base.Set(ptypes.Call, true)
	batchCommitter.stateCache.UpdateAccount(acc)

	// call to contract that calls unknown account - without create_account perm
	// create contract that calls the simple contract
	contractCode := callContractCode(users[9].Address())
	caller1ContractAddr := crypto.NewContractAddress(users[4].Address(), 101)
	caller1Acc := acm.ConcreteAccount{
		Address:     caller1ContractAddr,
		Balance:     0,
		Code:        contractCode,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	batchCommitter.stateCache.UpdateAccount(caller1Acc)

	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ := payload.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCallEnv := txs.Enclose(testChainID, txCall)
	txCallEnv.Sign(users[0])

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, batchCommitter, txCallEnv, caller1ContractAddr) //
	require.Error(t, err)

	// NOTE: for a contract to be able to CreateAccount, it must be able to call
	// NOTE: for a users to be able to CreateAccount, it must be able to send!
	caller1Acc.MutablePermissions().Base.Set(ptypes.CreateAccount, true)
	caller1Acc.MutablePermissions().Base.Set(ptypes.Call, true)
	batchCommitter.stateCache.UpdateAccount(caller1Acc)
	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ = payload.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCallEnv = txs.Enclose(testChainID, txCall)
	txCallEnv.Sign(users[0])

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, batchCommitter, txCallEnv, caller1ContractAddr) //
	require.NoError(t, err)

}

// holla at my boy
var DougAddress crypto.Address

func init() {
	copy(DougAddress[:], ([]byte)("THISISDOUG"))
}

func TestSNativeCALL(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(ptypes.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	exe := makeExecutor(st)

	//----------------------------------------------------------
	// Test CALL to SNative contracts

	// make the main contract once
	doug := acm.ConcreteAccount{
		Address:     DougAddress,
		Balance:     0,
		Code:        nil,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()

	doug.MutablePermissions().Base.Set(ptypes.Call, true)
	//doug.Permissions.Base.Set(permission.HasBase, true)
	exe.stateCache.UpdateAccount(doug)

	fmt.Println("\n#### HasBase")
	// HasBase
	snativeAddress, pF, data := snativePermTestInputCALL("hasBase", users[3], ptypes.Bond, false)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### SetBase")
	// SetBase
	snativeAddress, pF, data = snativePermTestInputCALL("setBase", users[3], ptypes.Bond, false)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], ptypes.Bond, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})
	snativeAddress, pF, data = snativePermTestInputCALL("setBase", users[3], ptypes.CreateContract, true)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], ptypes.CreateContract, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeAddress, pF, data = snativePermTestInputCALL("unsetBase", users[3], ptypes.CreateContract, false)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], ptypes.CreateContract, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeAddress, pF, data = snativePermTestInputCALL("setGlobal", users[3], ptypes.CreateContract, true)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], ptypes.CreateContract, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### HasRole")
	// HasRole
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "bumble")
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### AddRole")
	// AddRole
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "chuck")
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})
	snativeAddress, pF, data = snativeRoleTestInputCALL("addRole", users[3], "chuck")
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "chuck")
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### RemoveRole")
	// RemoveRole
	snativeAddress, pF, data = snativeRoleTestInputCALL("removeRole", users[3], "chuck")
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "chuck")
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})
}

func TestSNativeTx(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(ptypes.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(ptypes.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// Test SNativeTx

	fmt.Println("\n#### SetBase")
	// SetBase
	snativeArgs := snativePermTestInputTx("setBase", users[3], ptypes.Bond, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, ptypes.SetBase, snativeArgs)
	acc := getAccount(batchCommitter.stateCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(ptypes.Bond); v {
		t.Fatal("expected permission to be set false")
	}
	snativeArgs = snativePermTestInputTx("setBase", users[3], ptypes.CreateContract, true)
	testSNativeTxExpectPass(t, batchCommitter, ptypes.SetBase, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(ptypes.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeArgs = snativePermTestInputTx("unsetBase", users[3], ptypes.CreateContract, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, ptypes.UnsetBase, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(ptypes.CreateContract); v {
		t.Fatal("expected permission to be set false")
	}

	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeArgs = snativePermTestInputTx("setGlobal", users[3], ptypes.CreateContract, true)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, ptypes.SetGlobal, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, acm.GlobalPermissionsAddress)
	if v, _ := acc.MutablePermissions().Base.Get(ptypes.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### AddRole")
	// AddRole
	snativeArgs = snativeRoleTestInputTx("addRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, ptypes.AddRole, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v := acc.Permissions().HasRole("chuck"); !v {
		t.Fatal("expected role to be added")
	}

	fmt.Println("\n#### RemoveRole")
	// RemoveRole
	snativeArgs = snativeRoleTestInputTx("removeRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, ptypes.RemoveRole, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v := acc.Permissions().HasRole("chuck"); v {
		t.Fatal("expected role to be removed")
	}
}

func TestTxSequence(t *testing.T) {
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true, 1000)
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PublicKey()
	acc1 := getAccount(state, privAccounts[1].Address())

	// Test a variety of sequence numbers for the tx.
	// The tx should only pass when i == 1.
	for i := uint64(0); i < 3; i++ {
		sequence := acc0.Sequence() + i
		tx := payload.NewSendTx()
		tx.AddInputWithSequence(acc0PubKey, 1, sequence)
		tx.AddOutput(acc1.Address(), 1)
		txEnv := txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))
		stateCopy := state.Copy(dbm.NewMemDB())
		err := execTxWithState(stateCopy, txEnv)
		if i == 1 {
			// Sequence is good.
			if err != nil {
				t.Errorf("Expected good sequence to pass: %v", err)
			}
			// Check acc.Sequence().
			newAcc0 := getAccount(stateCopy, acc0.Address())
			if newAcc0.Sequence() != sequence {
				t.Errorf("Expected account sequence to change to %v, got %v",
					sequence, newAcc0.Sequence())
			}
		} else {
			// Sequence is bad.
			if err == nil {
				t.Errorf("Expected bad sequence to fail")
			}
			// Check acc.Sequence(). (shouldn't have changed)
			newAcc0 := getAccount(stateCopy, acc0.Address())
			if newAcc0.Sequence() != acc0.Sequence() {
				t.Errorf("Expected account sequence to not change from %v, got %v",
					acc0.Sequence(), newAcc0.Sequence())
			}
		}
	}
}

func TestNameTxs(t *testing.T) {
	st, err := MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	require.NoError(t, err)
	st.writeState.Save()

	names.MinNameRegistrationPeriod = 5
	blockchain := newBlockchain(testGenesisDoc)
	startingBlock := blockchain.LastBlockHeight()

	// try some bad names. these should all fail
	nameStrings := []string{"", "\n", "123#$%", "\x00", string([]byte{20, 40, 60, 80}),
		"baffledbythespectacleinallofthisyouseeehesaidwithouteyessurprised", "no spaces please"}
	data := "something about all this just doesn't feel right."
	fee := uint64(1000)
	numDesiredBlocks := uint64(5)
	for _, name := range nameStrings {
		amt := fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*
			names.NameBaseCost(name, data)
		tx, _ := payload.NewNameTx(st, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
		txEnv := txs.Enclose(testChainID, tx)
		txEnv.Sign(testPrivAccounts[0])

		if err := execTxWithState(st, txEnv); err == nil {
			t.Fatalf("Expected invalid name error from %s", name)
		}
	}

	// try some bad data. these should all fail
	name := "hold_it_chum"
	datas := []string{"cold&warm", "!@#$%^&*()", "<<<>>>>", "because why would you ever need a ~ or a & or even a % in a json file? make your case and we'll talk"}
	for _, data := range datas {
		amt := fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*
			names.NameBaseCost(name, data)
		tx, _ := payload.NewNameTx(st, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
		txEnv := txs.Enclose(testChainID, tx)
		txEnv.Sign(testPrivAccounts[0])

		if err := execTxWithState(st, txEnv); err == nil {
			t.Fatalf("Expected invalid data error from %s", data)
		}
	}

	validateEntry := func(t *testing.T, entry *names.Entry, name, data string, addr crypto.Address, expires uint64) {
		if entry == nil {
			t.Fatalf("Could not find name %s", name)
		}
		if entry.Owner != addr {
			t.Fatalf("Wrong owner. Got %s expected %s", entry.Owner, addr)
		}
		if data != entry.Data {
			t.Fatalf("Wrong data. Got %s expected %s", entry.Data, data)
		}
		if name != entry.Name {
			t.Fatalf("Wrong name. Got %s expected %s", entry.Name, name)
		}
		if expires != entry.Expires {
			t.Fatalf("Wrong expiry. Got %d, expected %d: %s", entry.Expires, expires, debug.Stack())
		}
	}

	// try a good one, check data, owner, expiry
	name = "@looking_good/karaoke_bar.broadband"
	data = "on this side of neptune there are 1234567890 people: first is OMNIVORE+-3. Or is it. Ok this is pretty restrictive. No exclamations :(. Faces tho :')"
	amt := fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*names.NameBaseCost(name, data)
	tx, _ := payload.NewNameTx(st, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[0]))
	if err := execTxWithState(st, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err := st.GetNameEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks)

	// fail to update it as non-owner, in same block
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[1]))
	if err := execTxWithState(st, txEnv); err == nil {
		t.Fatal("Expected error")
	}

	// update it as owner, just to increase expiry, in same block
	// NOTE: we have to resend the data or it will clear it (is this what we want?)
	tx, _ = payload.NewNameTx(st, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[0]))
	if err := execTxWithStateNewBlock(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks*2)

	// update it as owner, just to increase expiry, in next block
	tx, _ = payload.NewNameTx(st, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[0]))
	if err := execTxWithStateNewBlock(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks*3)

	// fail to update it as non-owner
	// Fast forward
	for blockchain.Tip.LastBlockHeight() < entry.Expires-1 {
		commitNewBlock(st, blockchain)
	}
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[1]))
	if err := execTxWithStateAndBlockchain(st, blockchain, txEnv); err == nil {
		t.Fatal("Expected error")
	}
	commitNewBlock(st, blockchain)

	// once expires, non-owner succeeds
	startingBlock = blockchain.LastBlockHeight()
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[1]))
	if err := execTxWithStateAndBlockchain(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[1].Address(), startingBlock+numDesiredBlocks)

	// update it as new owner, with new data (longer), but keep the expiry!
	data = "In the beginning there was no thing, not even the beginning. It hadn't been here, no there, nor for that matter anywhere, not especially because it had not to even exist, let alone to not. Nothing especially odd about that."
	oldCredit := amt - fee
	numDesiredBlocks = 10
	amt = fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*names.NameBaseCost(name, data) - oldCredit
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[1]))
	if err := execTxWithStateAndBlockchain(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[1].Address(), startingBlock+numDesiredBlocks)

	// test removal
	amt = fee
	data = ""
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[1]))
	if err := execTxWithStateNewBlock(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	if entry != nil {
		t.Fatal("Expected removed entry to be nil")
	}

	// create entry by key0,
	// test removal by key1 after expiry
	startingBlock = blockchain.LastBlockHeight()
	name = "looking_good/karaoke_bar"
	data = "some data"
	amt = fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*names.NameBaseCost(name, data)
	tx, _ = payload.NewNameTx(st, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[0]))
	if err := execTxWithStateAndBlockchain(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks)
	// Fast forward
	for blockchain.Tip.LastBlockHeight() < entry.Expires {
		commitNewBlock(st, blockchain)
	}

	amt = fee
	data = ""
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(testPrivAccounts[1]))
	if err := execTxWithStateNewBlock(st, blockchain, txEnv); err != nil {
		t.Fatal(err)
	}
	entry, err = st.GetNameEntry(name)
	require.NoError(t, err)
	if entry != nil {
		t.Fatal("Expected removed entry to be nil")
	}
}

// Test creating a contract from futher down the call stack
/*
contract Factory {
   address a;
   function create() returns (address){
       a = new PreFactory();
       return a;
   }
}

contract PreFactory{
   address a;
   function create(Factory c) returns (address) {
   	a = c.create();
   	return a;
   }
}
*/

// run-time byte code for each of the above
var preFactoryCode, _ = hex.DecodeString("60606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
var factoryCode, _ = hex.DecodeString("60606040526000357C010000000000000000000000000000000000000000000000000000000090048063EFC81A8C146037576035565B005B60426004805050606E565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B6000604051610153806100E0833901809050604051809103906000F0600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905060DD565B90566060604052610141806100126000396000F360606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
var createData, _ = hex.DecodeString("9ed93318")

func TestCreates(t *testing.T) {
	//evm.SetDebug(true)
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())

	newAcc1 := getAccount(state, acc1.Address())
	newAcc1.SetCode(preFactoryCode)
	newAcc2 := getAccount(state, acc2.Address())
	newAcc2.SetCode(factoryCode)

	state.writeState.UpdateAccount(newAcc1)
	state.writeState.UpdateAccount(newAcc2)

	createData = append(createData, acc2.Address().Word256().Bytes()...)

	// call the pre-factory, triggering the factory to run a create
	tx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address(),
			Amount:   1,
			Sequence: acc0.Sequence() + 1,
		},
		Address:  addressPtr(acc1),
		GasLimit: 10000,
		Data:     createData,
	}

	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(privAccounts[0]))
	err := execTxWithState(state, txEnv)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc1 = getAccount(state, acc1.Address())
	firstCreatedAddress, err := state.GetStorage(acc1.Address(), LeftPadWord256(nil))
	require.NoError(t, err)

	acc0 = getAccount(state, acc0.Address())
	// call the pre-factory, triggering the factory to run a create
	tx = &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address(),
			Amount:   1,
			Sequence: acc0.Sequence() + 1,
		},
		Address:  addressPtr(acc1),
		GasLimit: 100000,
		Data:     createData,
	}

	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(privAccounts[0]))
	err = execTxWithState(state, txEnv)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc1 = getAccount(state, acc1.Address())
	secondCreatedAddress, err := state.GetStorage(acc1.Address(), LeftPadWord256(nil))
	require.NoError(t, err)

	if firstCreatedAddress == secondCreatedAddress {
		t.Errorf("Multiple contracts created with the same address!")
	}
}

/*
contract Caller {
   function send(address x){
       x.send(msg.value);
   }
}
*/
var callerCode, _ = hex.DecodeString("60606040526000357c0100000000000000000000000000000000000000000000000000000000900480633e58c58c146037576035565b005b604b6004808035906020019091905050604d565b005b8073ffffffffffffffffffffffffffffffffffffffff16600034604051809050600060405180830381858888f19350505050505b5056")
var sendData, _ = hex.DecodeString("3e58c58c")

func TestContractSend(t *testing.T) {
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())

	newAcc1 := getAccount(state, acc1.Address())
	newAcc1.SetCode(callerCode)
	state.writeState.UpdateAccount(newAcc1)

	sendData = append(sendData, acc2.Address().Word256().Bytes()...)
	sendAmt := uint64(10)
	acc2Balance := acc2.Balance()

	// call the contract, triggering the send
	tx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address(),
			Amount:   sendAmt,
			Sequence: acc0.Sequence() + 1,
		},
		Address:  addressPtr(acc1),
		GasLimit: 1000,
		Data:     sendData,
	}

	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(privAccounts[0]))
	err := execTxWithState(state, txEnv)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc2 = getAccount(state, acc2.Address())
	if acc2.Balance() != sendAmt+acc2Balance {
		t.Errorf("Value transfer from contract failed! Got %d, expected %d", acc2.Balance(), sendAmt+acc2Balance)
	}
}

func TestMerklePanic(t *testing.T) {
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true,
		1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc1 := getAccount(state, privAccounts[1].Address())

	state.writeState.Save()
	// SendTx.
	{
		stateSendTx := state.Copy(dbm.NewMemDB())
		tx := &payload.SendTx{
			Inputs: []*payload.TxInput{
				{
					Address:  acc0.Address(),
					Amount:   1,
					Sequence: acc0.Sequence() + 1,
				},
			},
			Outputs: []*payload.TxOutput{
				{
					Address: acc1.Address(),
					Amount:  1,
				},
			},
		}

		txEnv := txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))
		err := execTxWithState(stateSendTx, txEnv)
		if err != nil {
			t.Errorf("Got error in executing send transaction, %v", err)
		}
		// uncomment for panic fun!
		//stateSendTx.Save()
	}

	// CallTx. Just runs through it and checks the transfer. See vm, rpc tests for more
	{
		stateCallTx := state.Copy(dbm.NewMemDB())
		newAcc1 := getAccount(stateCallTx, acc1.Address())
		newAcc1.SetCode([]byte{0x60})
		stateCallTx.writeState.UpdateAccount(newAcc1)
		tx := &payload.CallTx{
			Input: &payload.TxInput{
				Address:  acc0.Address(),
				Amount:   1,
				Sequence: acc0.Sequence() + 1,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		txEnv := txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))
		err := execTxWithState(stateCallTx, txEnv)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
	}
	state.writeState.Save()
	trygetacc0 := getAccount(state, privAccounts[0].Address())
	fmt.Println(trygetacc0.Address())
}

// TODO: test overflows.
// TODO: test for unbonding validators.
func TestTxs(t *testing.T) {
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc1 := getAccount(state, privAccounts[1].Address())

	// SendTx.
	{
		stateSendTx := state.Copy(dbm.NewMemDB())
		tx := &payload.SendTx{
			Inputs: []*payload.TxInput{
				{
					Address:  acc0.Address(),
					Amount:   1,
					Sequence: acc0.Sequence() + 1,
				},
			},
			Outputs: []*payload.TxOutput{
				{
					Address: acc1.Address(),
					Amount:  1,
				},
			},
		}

		txEnv := txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))
		err := execTxWithState(stateSendTx, txEnv)
		if err != nil {
			t.Errorf("Got error in executing send transaction, %v", err)
		}
		newAcc0 := getAccount(stateSendTx, acc0.Address())
		if acc0.Balance()-1 != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-1, newAcc0.Balance())
		}
		newAcc1 := getAccount(stateSendTx, acc1.Address())
		if acc1.Balance()+1 != newAcc1.Balance() {
			t.Errorf("Unexpected newAcc1 balance. Expected %v, got %v",
				acc1.Balance()+1, newAcc1.Balance())
		}
	}

	// CallTx. Just runs through it and checks the transfer. See vm, rpc tests for more
	{
		stateCallTx := state.Copy(dbm.NewMemDB())
		newAcc1 := getAccount(stateCallTx, acc1.Address())
		newAcc1.SetCode([]byte{0x60})
		stateCallTx.writeState.UpdateAccount(newAcc1)
		tx := &payload.CallTx{
			Input: &payload.TxInput{
				Address:  acc0.Address(),
				Amount:   1,
				Sequence: acc0.Sequence() + 1,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		txEnv := txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))
		err := execTxWithState(stateCallTx, txEnv)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
		newAcc0 := getAccount(stateCallTx, acc0.Address())
		if acc0.Balance()-1 != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-1, newAcc0.Balance())
		}
		newAcc1 = getAccount(stateCallTx, acc1.Address())
		if acc1.Balance()+1 != newAcc1.Balance() {
			t.Errorf("Unexpected newAcc1 balance. Expected %v, got %v",
				acc1.Balance()+1, newAcc1.Balance())
		}
	}
	trygetacc0 := getAccount(state, privAccounts[0].Address())
	fmt.Println(trygetacc0.Address())

	// NameTx.
	{
		entryName := "satoshi"
		entryData := `
A  purely   peer-to-peer   version   of   electronic   cash   would   allow   online
payments  to  be  sent   directly  from  one  party  to  another  without   going  through  a
financial institution.   Digital signatures provide part of the solution, but the main
benefits are lost if a trusted third party is still required to prevent double-spending.
We propose a solution to the double-spending problem using a peer-to-peer network.
The   network   timestamps   transactions  by  hashing   them   into   an   ongoing   chain   of
hash-based proof-of-work, forming a record that cannot be changed without redoing
the proof-of-work.   The longest chain not only serves as proof of the sequence of
events witnessed, but proof that it came from the largest pool of CPU power.   As
long as a majority of CPU power is controlled by nodes that are not cooperating to
attack the network, they'll generate the longest chain and outpace attackers.   The
network itself requires minimal structure.   Messages are broadcast on a best effort
basis,   and   nodes   can   leave  and   rejoin   the  network   at  will,  accepting   the   longest
proof-of-work chain as proof of what happened while they were gone `
		entryAmount := uint64(10000)

		stateNameTx := state
		tx := &payload.NameTx{
			Input: &payload.TxInput{
				Address:  acc0.Address(),
				Amount:   entryAmount,
				Sequence: acc0.Sequence() + 1,
			},
			Name: entryName,
			Data: entryData,
		}

		txEnv := txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))

		err := execTxWithState(stateNameTx, txEnv)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
		newAcc0 := getAccount(stateNameTx, acc0.Address())
		if acc0.Balance()-entryAmount != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-entryAmount, newAcc0.Balance())
		}
		entry, err := stateNameTx.GetNameEntry(entryName)
		require.NoError(t, err)
		if entry == nil {
			t.Errorf("Expected an entry but got nil")
		}
		if entry.Data != entryData {
			t.Errorf("Wrong data stored")
		}

		// test a bad string
		tx.Data = string([]byte{0, 1, 2, 3, 127, 128, 129, 200, 251})
		tx.Input.Sequence += 1
		txEnv = txs.Enclose(testChainID, tx)
		require.NoError(t, txEnv.Sign(privAccounts[0]))
		err = execTxWithState(stateNameTx, txEnv)
		if _, ok := err.(payload.ErrTxInvalidString); !ok {
			t.Errorf("Expected invalid string error. Got: %s", err.Error())
		}
	}

	// BondTx. TODO
	/*
		{
			state := state.Copy()
			tx := &payload.BondTx{
				PublicKey: acc0PubKey.(acm.PublicKeyEd25519),
				Inputs: []*payload.TxInput{
					&payload.TxInput{
						Address:  acc0.Address(),
						Amount:   1,
						Sequence: acc0.Sequence() + 1,
						PublicKey:   acc0PubKey,
					},
				},
				UnbondTo: []*payload.TxOutput{
					&payload.TxOutput{
						Address: acc0.Address(),
						Amount:  1,
					},
				},
			}
			tx.Signature = privAccounts[0] acm.ChainSign(testChainID, tx).(crypto.SignatureEd25519)
			tx.Inputs[0].Signature = privAccounts[0] acm.ChainSign(testChainID, tx)
			err := execTxWithState(state, tx)
			if err != nil {
				t.Errorf("Got error in executing bond transaction, %v", err)
			}
			newAcc0 := getAccount(state, acc0.Address())
			if newAcc0.Balance() != acc0.Balance()-1 {
				t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
					acc0.Balance()-1, newAcc0.Balance())
			}
			_, acc0Val := state.BondedValidators.GetByAddress(acc0.Address())
			if acc0Val == nil {
				t.Errorf("acc0Val not present")
			}
			if acc0Val.BondHeight != blockchain.LastBlockHeight()+1 {
				t.Errorf("Unexpected bond height. Expected %v, got %v",
					blockchain.LastBlockHeight(), acc0Val.BondHeight)
			}
			if acc0Val.Power != 1 {
				t.Errorf("Unexpected voting power. Expected %v, got %v",
					acc0Val.Power, acc0.Balance())
			}
			if acc0Val.Accum != 0 {
				t.Errorf("Unexpected accum. Expected 0, got %v",
					acc0Val.Accum)
			}
		} */

	// TODO UnbondTx.

}

func TestSelfDestruct(t *testing.T) {
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true, 1000)

	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PublicKey()
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())
	sendingAmount, refundedBalance, oldBalance := uint64(1), acc1.Balance(), acc2.Balance()

	newAcc1 := getAccount(state, acc1.Address())

	// store 0x1 at 0x1, push an address, then self-destruct:)
	contractCode := []byte{0x60, 0x01, 0x60, 0x01, 0x55, 0x73}
	contractCode = append(contractCode, acc2.Address().Bytes()...)
	contractCode = append(contractCode, 0xff)
	newAcc1.SetCode(contractCode)
	state.writeState.UpdateAccount(newAcc1)

	// send call tx with no data, cause self-destruct
	tx := payload.NewCallTxWithSequence(acc0PubKey, addressPtr(acc1), nil, sendingAmount, 1000, 0, acc0.Sequence()+1)

	// we use cache instead of execTxWithState so we can run the tx twice
	exe := NewBatchCommitter(state, newBlockchain(testGenesisDoc).Tip, event.NewNoOpPublisher(), logger)
	signAndExecute(t, false, exe, testChainID, tx, privAccounts[0])

	// if we do it again, we won't get an error, but the self-destruct
	// shouldn't happen twice and the caller should lose fee
	tx.Input.Sequence += 1
	signAndExecute(t, false, exe, testChainID, tx, privAccounts[0])

	// commit the block
	exe.Commit()

	// acc2 should receive the sent funds and the contracts balance
	newAcc2 := getAccount(state, acc2.Address())
	newBalance := sendingAmount + refundedBalance + oldBalance
	if newAcc2.Balance() != newBalance {
		t.Errorf("Unexpected newAcc2 balance. Expected %v, got %v",
			newAcc2.Balance(), newBalance)
	}
	newAcc1 = getAccount(state, acc1.Address())
	if newAcc1 != nil {
		t.Errorf("Expected account to be removed")
	}
}

func signAndExecute(t *testing.T, shouldFail bool, exe BatchExecutor, chainID string, tx payload.Payload,
	signers ...acm.AddressableSigner) *txs.Envelope {

	env := txs.Enclose(chainID, tx)
	require.NoError(t, env.Sign(signers...), "Could not sign tx in call: %s", debug.Stack())
	if shouldFail {
		require.Error(t, exe.Execute(env), "Tx should fail in call: %s", debug.Stack())
	} else {
		require.NoError(t, exe.Execute(env), "Could not execute tx in call: %s", debug.Stack())
	}
	return env
}

func execTxWithStateAndBlockchain(state *State, blockchain *bcm.Blockchain, txEnv *txs.Envelope) error {
	exe := newExecutor("execTxWithStateAndBlockchainCache", true, state, blockchain.Tip,
		event.NewNoOpPublisher(), logger)
	if err := exe.Execute(txEnv); err != nil {
		return err
	} else {
		exe.Commit()
		commitNewBlock(state, blockchain)
		return nil
	}
}

func execTxWithState(state *State, txEnv *txs.Envelope) error {
	return execTxWithStateAndBlockchain(state, newBlockchain(testGenesisDoc), txEnv)
}

func commitNewBlock(state *State, blockchain *bcm.Blockchain) {
	blockchain.CommitBlock(blockchain.LastBlockTime().Add(time.Second), sha3.Sha3(blockchain.LastBlockHash()),
		state.writeState.Hash())
}

func execTxWithStateNewBlock(state *State, blockchain *bcm.Blockchain, txEnv *txs.Envelope) error {
	if err := execTxWithStateAndBlockchain(state, blockchain, txEnv); err != nil {
		return err
	}
	commitNewBlock(state, blockchain)
	return nil
}

func makeGenesisState(numAccounts int, randBalance bool, minBalance uint64, numValidators int, randBonded bool,
	minBonded int64) (*State, []acm.AddressableSigner) {
	testGenesisDoc, privAccounts, _ := deterministicGenesis.GenesisDoc(numAccounts, randBalance, minBalance,
		numValidators, randBonded, minBonded)
	s0, err := MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	if err != nil {
		panic(fmt.Errorf("could not make genesis state: %v", err))
	}
	s0.writeState.Save()
	return s0, privAccounts
}

func getAccount(accountGetter state.AccountGetter, address crypto.Address) acm.MutableAccount {
	acc, _ := state.GetMutableAccount(accountGetter, address)
	return acc
}

func addressPtr(account acm.Account) *crypto.Address {
	if account == nil {
		return nil
	}
	accountAddresss := account.Address()
	return &accountAddresss
}

//-------------------------------------------------------------------------------------
// helpers

var ExceptionTimeOut = errors.NewCodedError(errors.ErrorCodeGeneric, "timed out waiting for event")

// run ExecTx and wait for the Call event on given addr
// returns the msg data and an error/exception
func execTxWaitAccountCall(t *testing.T, exe *testExecutor, txEnv *txs.Envelope,
	address crypto.Address) (*events.EventDataCall, error) {

	ch := make(chan *events.EventDataCall)
	ctx := context.Background()
	const subscriber = "execTxWaitEvent"
	events.SubscribeAccountCall(ctx, exe, subscriber, address, txEnv.Tx.Hash(), -1, ch)
	defer exe.UnsubscribeAll(ctx, subscriber)
	err := exe.Execute(txEnv)
	if err != nil {
		return nil, err
	}
	exe.Commit()
	exe.blockchain.CommitBlock(time.Time{}, nil, nil)
	ticker := time.NewTicker(5 * time.Second)

	select {
	case eventDataCall := <-ch:
		fmt.Println("MSG: ", eventDataCall)
		return eventDataCall, eventDataCall.Exception.AsError()

	case <-ticker.C:
		return nil, ExceptionTimeOut
	}

}

// give a contract perms for an snative, call it, it calls the snative, but shouldn't have permission
func testSNativeCALLExpectFail(t *testing.T, exe *testExecutor, doug acm.MutableAccount,
	snativeAddress crypto.Address, data []byte) {
	testSNativeCALL(t, false, exe, doug, 0, snativeAddress, data, nil)
}

// give a contract perms for an snative, call it, it calls the snative, ensure the check funciton (f) succeeds
func testSNativeCALLExpectPass(t *testing.T, exe *testExecutor, doug acm.MutableAccount, snativePerm ptypes.PermFlag,
	snativeAddress crypto.Address, data []byte, f func([]byte) error) {
	testSNativeCALL(t, true, exe, doug, snativePerm, snativeAddress, data, f)
}

func testSNativeCALL(t *testing.T, expectPass bool, exe *testExecutor, doug acm.MutableAccount,
	snativePerm ptypes.PermFlag, snativeAddress crypto.Address, data []byte, f func([]byte) error) {
	if expectPass {
		doug.MutablePermissions().Base.Set(snativePerm, true)
	}

	doug.SetCode(callContractCode(snativeAddress))
	dougAddress := doug.Address()

	exe.stateCache.UpdateAccount(doug)
	tx, _ := payload.NewCallTx(exe.stateCache, users[0].PublicKey(), &dougAddress, data, 100, 10000, 100)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	t.Logf("subscribing to %v", events.EventStringAccountCall(snativeAddress))
	ev, err := execTxWaitAccountCall(t, exe, txEnv, snativeAddress)
	if err == ExceptionTimeOut {
		t.Fatal("Timed out waiting for event")
	}
	if expectPass {
		require.NoError(t, err)
		ret := ev.Return
		if err := f(ret); err != nil {
			t.Fatal(err)
		}
	} else {
		require.Error(t, err)
	}
}

func testSNativeTxExpectFail(t *testing.T, batchCommitter *testExecutor, snativeArgs snatives.PermArgs) {
	testSNativeTx(t, false, batchCommitter, 0, snativeArgs)
}

func testSNativeTxExpectPass(t *testing.T, batchCommitter *testExecutor, perm ptypes.PermFlag,
	snativeArgs snatives.PermArgs) {
	testSNativeTx(t, true, batchCommitter, perm, snativeArgs)
}

func testSNativeTx(t *testing.T, expectPass bool, batchCommitter *testExecutor, perm ptypes.PermFlag,
	snativeArgs snatives.PermArgs) {
	if expectPass {
		acc := getAccount(batchCommitter.stateCache, users[0].Address())
		acc.MutablePermissions().Base.Set(perm, true)
		batchCommitter.stateCache.UpdateAccount(acc)
	}
	tx, _ := payload.NewPermissionsTx(batchCommitter.stateCache, users[0].PublicKey(), snativeArgs)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	err := batchCommitter.Execute(txEnv)
	if expectPass {
		if err != nil {
			t.Fatal("Unexpected exception", err)
		}
	} else {
		if err == nil {
			t.Fatal("Expected exception")
		}
	}
}

func boolToWord256(v bool) Word256 {
	var vint byte
	if v {
		vint = 0x1
	} else {
		vint = 0x0
	}
	return LeftPadWord256([]byte{vint})
}

func permNameToFuncID(name string) []byte {
	function, err := permissionsContract.FunctionByName(name)
	if err != nil {
		panic("didn't find snative function signature!")
	}
	id := function.ID()
	return id[:]
}

func snativePermTestInputCALL(name string, user acm.AddressableSigner, perm ptypes.PermFlag,
	val bool) (addr crypto.Address, pF ptypes.PermFlag, data []byte) {
	addr = permissionsContract.Address()
	switch name {
	case "hasBase", "unsetBase":
		data = user.Address().Word256().Bytes()
		data = append(data, Uint64ToWord256(uint64(perm)).Bytes()...)
	case "setBase":
		data = user.Address().Word256().Bytes()
		data = append(data, Uint64ToWord256(uint64(perm)).Bytes()...)
		data = append(data, boolToWord256(val).Bytes()...)
	case "setGlobal":
		data = Uint64ToWord256(uint64(perm)).Bytes()
		data = append(data, boolToWord256(val).Bytes()...)
	}
	data = append(permNameToFuncID(name), data...)
	var err error
	if pF, err = ptypes.PermStringToFlag(name); err != nil {
		panic(fmt.Sprintf("failed to convert perm string (%s) to flag", name))
	}
	return
}

func snativePermTestInputTx(name string, user acm.AddressableSigner, perm ptypes.PermFlag,
	val bool) (snativeArgs snatives.PermArgs) {

	switch name {
	case "hasBase":
		snativeArgs = snatives.HasBaseArgs(user.Address(), perm)
	case "unsetBase":
		snativeArgs = snatives.UnsetBaseArgs(user.Address(), perm)
	case "setBase":
		snativeArgs = snatives.SetBaseArgs(user.Address(), perm, val)
	case "setGlobal":
		snativeArgs = snatives.SetGlobalArgs(perm, val)
	}
	return
}

func snativeRoleTestInputCALL(name string, user acm.AddressableSigner,
	role string) (addr crypto.Address, pF ptypes.PermFlag, data []byte) {
	addr = permissionsContract.Address()
	data = user.Address().Word256().Bytes()
	data = append(data, RightPadBytes([]byte(role), 32)...)
	data = append(permNameToFuncID(name), data...)

	var err error
	if pF, err = ptypes.PermStringToFlag(name); err != nil {
		panic(fmt.Sprintf("failed to convert perm string (%s) to flag", name))
	}
	return
}

func snativeRoleTestInputTx(name string, user acm.AddressableSigner, role string) (snativeArgs snatives.PermArgs) {
	switch name {
	case "hasRole":
		snativeArgs = snatives.HasRoleArgs(user.Address(), role)
	case "addRole":
		snativeArgs = snatives.AddRoleArgs(user.Address(), role)
	case "removeRole":
		snativeArgs = snatives.RemoveRoleArgs(user.Address(), role)
	}
	return
}

// convenience function for contract that calls a given address
func callContractCode(contractAddr crypto.Address) []byte {
	// calldatacopy into mem and use as input to call
	memOff, inputOff := byte(0x0), byte(0x0)
	value := byte(0x1)
	inOff := byte(0x0)
	retOff, retSize := byte(0x0), byte(0x20)

	// this is the code we want to run (call a contract and return)
	return bc.MustSplice(CALLDATASIZE, PUSH1, inputOff, PUSH1, memOff,
		CALLDATACOPY, PUSH1, retSize, PUSH1, retOff, CALLDATASIZE, PUSH1, inOff,
		PUSH1, value, PUSH20, contractAddr,
		// Zeno loves us - call with half of the available gas each time we CALL
		PUSH1, 2, GAS, DIV, CALL,
		PUSH1, 32, PUSH1, 0, RETURN)
}

// convenience function for contract that is a factory for the code that comes as call data
func createContractCode() []byte {
	// TODO: gas ...

	// calldatacopy the calldatasize
	memOff, inputOff := byte(0x0), byte(0x0)
	contractCode := []byte{0x60, memOff, 0x60, inputOff, 0x36, 0x37}

	// create
	value := byte(0x1)
	contractCode = append(contractCode, []byte{0x60, value, 0x36, 0x60, memOff, 0xf0}...)
	return contractCode
}

// wrap a contract in create code
func wrapContractForCreate(contractCode []byte) []byte {
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	code := append([]byte{0x7f}, RightPadWord256(contractCode).Bytes()...)
	// store it in memory
	code = append(code, []byte{0x60, 0x0, 0x52}...)
	// return whats in memory
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	// return init code, contract code, expected return
	return code
}
