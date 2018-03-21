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

	acm "github.com/hyperledger/burrow/account"
	. "github.com/hyperledger/burrow/binary"
	bcm "github.com/hyperledger/burrow/blockchain"
	"github.com/hyperledger/burrow/event"
	exe_events "github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/evm"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	evm_events "github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tmthrgd/go-hex"
)

var (
	dbBackend           = dbm.MemDBBackendStr
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
var logger = loggers.NewNoopInfoTraceLogger()
var deterministicGenesis = genesis.NewDeterministicGenesis(34059836243380576)
var testGenesisDoc, testPrivAccounts, _ = deterministicGenesis.
	GenesisDoc(3, true, 1000, 1, true, 1000)
var testChainID = testGenesisDoc.ChainID()

func makeUsers(n int) []acm.PrivateAccount {
	users := make([]acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		secret := "mysecret" + strconv.Itoa(i)
		users[i] = acm.GeneratePrivateAccountFromSecret(secret)
	}
	return users
}

func makeExecutor(state *State) *executor {
	return newExecutor(true, state, testChainID, bcm.NewBlockchain(nil, testGenesisDoc), event.NewEmitter(logger),
		logger)
}

func newBaseGenDoc(globalPerm, accountPerm ptypes.AccountPermissions) genesis.GenesisDoc {
	genAccounts := []genesis.Account{}
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

//func getAccount(state acm.Getter, address acm.Address) acm.MutableAccount {
//	acc, _ := acm.GetMutableAccount(state, address)
//	return acc
//}

func TestSendFails(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(permission.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(permission.CreateContract, true)
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//-------------------
	// send txs

	// simple send tx should fail
	tx := txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[0])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple send tx with call perm should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[2].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[4].Address(), 5)
	tx.SignInput(testChainID, 0, users[2])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple send tx with create perm should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[3].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[4].Address(), 5)
	tx.SignInput(testChainID, 0, users[3])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple send tx to unknown account without create_account perm should fail
	acc := getAccount(batchCommitter.stateCache, users[3].Address())
	acc.MutablePermissions().Base.Set(permission.Send, true)
	batchCommitter.stateCache.UpdateAccount(acc)
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[3].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].Address(), 5)
	tx.SignInput(testChainID, 0, users[3])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}
}

func TestName(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Name, true)
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//-------------------
	// name txs

	// simple name tx without perm should fail
	tx, err := txs.NewNameTx(st, users[0].PublicKey(), "somename", "somedata", 10000, 100)
	if err != nil {
		t.Fatal(err)
	}
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple name tx with perm should pass
	tx, err = txs.NewNameTx(st, users[1].PublicKey(), "somename", "somedata", 10000, 100)
	if err != nil {
		t.Fatal(err)
	}
	tx.Sign(testChainID, users[1])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal(err)
	}
}

func TestCallFails(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(permission.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(permission.CreateContract, true)
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//-------------------
	// call txs

	address4 := users[4].Address()
	// simple call tx should fail
	tx, _ := txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &address4, nil, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call tx with send permission should fail
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[1].PublicKey(), &address4, nil, 100, 100, 100)
	tx.Sign(testChainID, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call tx with create permission should fail
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[3].PublicKey(), &address4, nil, 100, 100, 100)
	tx.Sign(testChainID, users[3])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	//-------------------
	// create txs

	// simple call create tx should fail
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), nil, nil, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call create tx with send perm should fail
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[1].PublicKey(), nil, nil, 100, 100, 100)
	tx.Sign(testChainID, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call create tx with call perm should fail
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[2].PublicKey(), nil, nil, 100, 100, 100)
	tx.Sign(testChainID, users[2])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}
}

func TestSendPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true) // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	// A single input, having the permission, should succeed
	tx := txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}

	// Two inputs, one with permission, one without, should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[2].Address(), 10)
	tx.SignInput(testChainID, 0, users[0])
	tx.SignInput(testChainID, 1, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}
}

func TestCallPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//------------------------------
	// call to simple contract
	fmt.Println("\n##### SIMPLE CONTRACT")

	// create simple contract
	simpleContractAddr := acm.NewContractAddress(users[0].Address(), 100)
	simpleAcc := acm.ConcreteAccount{
		Address:     simpleContractAddr,
		Balance:     0,
		Code:        []byte{0x60},
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	st.UpdateAccount(simpleAcc)

	// A single input, having the permission, should succeed
	tx, _ := txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &simpleContractAddr, nil, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}

	//----------------------------------------------------------
	// call to contract that calls simple contract - without perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (FAIL)")

	// create contract that calls the simple contract
	contractCode := callContractCode(simpleContractAddr)
	caller1ContractAddr := acm.NewContractAddress(users[0].Address(), 101)
	caller1Acc := acm.ConcreteAccount{
		Address:     caller1ContractAddr,
		Balance:     10000,
		Code:        contractCode,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	batchCommitter.stateCache.UpdateAccount(caller1Acc)

	// A single input, having the permission, but the contract doesn't have permission
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception := execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(caller1ContractAddr)) //
	if exception == "" {
		t.Fatal("Expected exception")
	}

	//----------------------------------------------------------
	// call to contract that calls simple contract - with perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (PASS)")

	// A single input, having the permission, and the contract has permission
	caller1Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.stateCache.UpdateAccount(caller1Acc)
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(caller1ContractAddr)) //
	if exception != "" {
		t.Fatal("Unexpected exception:", exception)
	}

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// caller1Contract does not have call perms, but caller2Contract does.
	fmt.Println("\n##### CALL TO CONTRACT CALLING SIMPLE CONTRACT (FAIL)")

	contractCode2 := callContractCode(caller1ContractAddr)
	caller2ContractAddr := acm.NewContractAddress(users[0].Address(), 102)
	caller2Acc := acm.ConcreteAccount{
		Address:     caller2ContractAddr,
		Balance:     1000,
		Code:        contractCode2,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	caller1Acc.MutablePermissions().Base.Set(permission.Call, false)
	caller2Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.stateCache.UpdateAccount(caller1Acc)
	batchCommitter.stateCache.UpdateAccount(caller2Acc)

	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(caller1ContractAddr)) //
	if exception == "" {
		t.Fatal("Expected exception")
	}

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// both caller1 and caller2 have permission
	fmt.Println("\n##### CALL TO CONTRACT CALLING SIMPLE CONTRACT (PASS)")

	caller1Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.stateCache.UpdateAccount(caller1Acc)

	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(caller1ContractAddr)) //
	if exception != "" {
		t.Fatal("Unexpected exception", exception)
	}
}

func TestCreatePermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.CreateContract, true) // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true)           // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//------------------------------
	// create a simple contract
	fmt.Println("\n##### CREATE SIMPLE CONTRACT")

	contractCode := []byte{0x60}
	createCode := wrapContractForCreate(contractCode)

	// A single input, having the permission, should succeed
	tx, _ := txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), nil, createCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}
	// ensure the contract is there
	contractAddr := acm.NewContractAddress(tx.Input.Address, tx.Input.Sequence)
	contractAcc := getAccount(batchCommitter.stateCache, contractAddr)
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
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), nil, createFactoryCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}
	// ensure the contract is there
	contractAddr = acm.NewContractAddress(tx.Input.Address, tx.Input.Sequence)
	contractAcc = getAccount(batchCommitter.stateCache, contractAddr)
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
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &contractAddr, createCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	// we need to subscribe to the Call event to detect the exception
	_, exception := execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(contractAddr)) //
	if exception == "" {
		t.Fatal("expected exception")
	}

	//------------------------------
	// call the contract (should PASS)
	fmt.Println("\n###### CALL THE FACTORY (PASS)")

	contractAcc.MutablePermissions().Base.Set(permission.CreateContract, true)
	batchCommitter.stateCache.UpdateAccount(contractAcc)

	// A single input, having the permission, should succeed
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &contractAddr, createCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(contractAddr)) //
	if exception != "" {
		t.Fatal("unexpected exception", exception)
	}

	//--------------------------------
	fmt.Println("\n##### CALL to empty address")
	code := callContractCode(acm.Address{})

	contractAddr = acm.NewContractAddress(users[0].Address(), 110)
	contractAcc = acm.ConcreteAccount{
		Address:     contractAddr,
		Balance:     1000,
		Code:        code,
		Sequence:    0,
		StorageRoot: Zero256.Bytes(),
		Permissions: permission.ZeroAccountPermissions,
	}.MutableAccount()
	contractAcc.MutablePermissions().Base.Set(permission.Call, true)
	contractAcc.MutablePermissions().Base.Set(permission.CreateContract, true)
	batchCommitter.stateCache.UpdateAccount(contractAcc)

	// this should call the 0 address but not create ...
	tx, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &contractAddr, createCode, 100, 10000, 100)
	tx.Sign(testChainID, users[0])
	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(acm.Address{})) //
	if exception != "" {
		t.Fatal("unexpected exception", exception)
	}
	zeroAcc := getAccount(batchCommitter.stateCache, acm.Address{})
	if len(zeroAcc.Code()) != 0 {
		t.Fatal("the zero account was given code from a CALL!")
	}
}

func TestCreateAccountPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)          // give the 0 account permission
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)          // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.CreateAccount, true) // give the 0 account permission
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// SendTx to unknown account

	// A single input, having the permission, should succeed
	tx := txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].Address(), 5)
	tx.SignInput(testChainID, 0, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}

	// Two inputs, both with send, one with create, one without, should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 10)
	tx.SignInput(testChainID, 0, users[0])
	tx.SignInput(testChainID, 1, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// Two inputs, both with send, one with create, one without, two ouputs (one known, one unknown) should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 4)
	tx.AddOutput(users[4].Address(), 6)
	tx.SignInput(testChainID, 0, users[0])
	tx.SignInput(testChainID, 1, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// Two inputs, both with send, both with create, should pass
	acc := getAccount(batchCommitter.stateCache, users[1].Address())
	acc.MutablePermissions().Base.Set(permission.CreateAccount, true)
	batchCommitter.stateCache.UpdateAccount(acc)
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 10)
	tx.SignInput(testChainID, 0, users[0])
	tx.SignInput(testChainID, 1, users[1])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Unexpected error", err)
	}

	// Two inputs, both with send, both with create, two outputs (one known, one unknown) should pass
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.stateCache, users[0].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.stateCache, users[1].PublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].Address(), 7)
	tx.AddOutput(users[4].Address(), 3)
	tx.SignInput(testChainID, 0, users[0])
	tx.SignInput(testChainID, 1, users[1])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Unexpected error", err)
	}

	//----------------------------------------------------------
	// CALL to unknown account

	acc = getAccount(batchCommitter.stateCache, users[0].Address())
	acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.stateCache.UpdateAccount(acc)

	// call to contract that calls unknown account - without create_account perm
	// create contract that calls the simple contract
	contractCode := callContractCode(users[9].Address())
	caller1ContractAddr := acm.NewContractAddress(users[4].Address(), 101)
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
	txCall, _ := txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCall.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception := execTxWaitEvent(t, batchCommitter, txCall, evm_events.EventStringAccountCall(caller1ContractAddr)) //
	if exception == "" {
		t.Fatal("Expected exception")
	}

	// NOTE: for a contract to be able to CreateAccount, it must be able to call
	// NOTE: for a users to be able to CreateAccount, it must be able to send!
	caller1Acc.MutablePermissions().Base.Set(permission.CreateAccount, true)
	caller1Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.stateCache.UpdateAccount(caller1Acc)
	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ = txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCall.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, txCall, evm_events.EventStringAccountCall(caller1ContractAddr)) //
	if exception != "" {
		t.Fatal("Unexpected exception", exception)
	}

}

// holla at my boy
var DougAddress acm.Address

func init() {
	copy(DougAddress[:], ([]byte)("THISISDOUG"))
}

func TestSNativeCALL(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(permission.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

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

	doug.MutablePermissions().Base.Set(permission.Call, true)
	//doug.Permissions.Base.Set(permission.HasBase, true)
	batchCommitter.stateCache.UpdateAccount(doug)

	fmt.Println("\n#### HasBase")
	// HasBase
	snativeAddress, pF, data := snativePermTestInputCALL("hasBase", users[3], permission.Bond, false)
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### SetBase")
	// SetBase
	snativeAddress, pF, data = snativePermTestInputCALL("setBase", users[3], permission.Bond, false)
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.Bond, false)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})
	snativeAddress, pF, data = snativePermTestInputCALL("setBase", users[3], permission.CreateContract, true)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeAddress, pF, data = snativePermTestInputCALL("unsetBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeAddress, pF, data = snativePermTestInputCALL("setGlobal", users[3], permission.CreateContract, true)
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### HasRole")
	// HasRole
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "bumble")
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### AddRole")
	// AddRole
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "chuck")
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})
	snativeAddress, pF, data = snativeRoleTestInputCALL("addRole", users[3], "chuck")
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "chuck")
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Println("\n#### RemoveRole")
	// RemoveRole
	snativeAddress, pF, data = snativeRoleTestInputCALL("removeRole", users[3], "chuck")
	testSNativeCALLExpectFail(t, batchCommitter, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativeRoleTestInputCALL("hasRole", users[3], "chuck")
	testSNativeCALLExpectPass(t, batchCommitter, doug, pF, snativeAddress, data, func(ret []byte) error {
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
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(permission.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st, err := MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// Test SNativeTx

	fmt.Println("\n#### SetBase")
	// SetBase
	snativeArgs := snativePermTestInputTx("setBase", users[3], permission.Bond, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetBase, snativeArgs)
	acc := getAccount(batchCommitter.stateCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(permission.Bond); v {
		t.Fatal("expected permission to be set false")
	}
	snativeArgs = snativePermTestInputTx("setBase", users[3], permission.CreateContract, true)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetBase, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(permission.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeArgs = snativePermTestInputTx("unsetBase", users[3], permission.CreateContract, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.UnsetBase, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(permission.CreateContract); v {
		t.Fatal("expected permission to be set false")
	}

	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeArgs = snativePermTestInputTx("setGlobal", users[3], permission.CreateContract, true)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetGlobal, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, permission.GlobalPermissionsAddress)
	if v, _ := acc.MutablePermissions().Base.Get(permission.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### AddRole")
	// AddRole
	snativeArgs = snativeRoleTestInputTx("addRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.AddRole, snativeArgs)
	acc = getAccount(batchCommitter.stateCache, users[3].Address())
	if v := acc.Permissions().HasRole("chuck"); !v {
		t.Fatal("expected role to be added")
	}

	fmt.Println("\n#### RemoveRole")
	// RemoveRole
	snativeArgs = snativeRoleTestInputTx("removeRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.RemoveRole, snativeArgs)
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
		tx := txs.NewSendTx()
		tx.AddInputWithSequence(acc0PubKey, 1, sequence)
		tx.AddOutput(acc1.Address(), 1)
		tx.Inputs[0].Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		stateCopy := state.Copy(dbm.NewMemDB())
		err := execTxWithState(stateCopy, tx)
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
	state, err := MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	require.NoError(t, err)
	state.Save()

	txs.MinNameRegistrationPeriod = 5
	blockchain := bcm.NewBlockchain(nil, testGenesisDoc)
	startingBlock := blockchain.LastBlockHeight()

	// try some bad names. these should all fail
	names := []string{"", "\n", "123#$%", "\x00", string([]byte{20, 40, 60, 80}),
		"baffledbythespectacleinallofthisyouseeehesaidwithouteyessurprised", "no spaces please"}
	data := "something about all this just doesn't feel right."
	fee := uint64(1000)
	numDesiredBlocks := uint64(5)
	for _, name := range names {
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*
			txs.NameBaseCost(name, data)
		tx, _ := txs.NewNameTx(state, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
		tx.Sign(testChainID, testPrivAccounts[0])

		if err := execTxWithState(state, tx); err == nil {
			t.Fatalf("Expected invalid name error from %s", name)
		}
	}

	// try some bad data. these should all fail
	name := "hold_it_chum"
	datas := []string{"cold&warm", "!@#$%^&*()", "<<<>>>>", "because why would you ever need a ~ or a & or even a % in a json file? make your case and we'll talk"}
	for _, data := range datas {
		amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*
			txs.NameBaseCost(name, data)
		tx, _ := txs.NewNameTx(state, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
		tx.Sign(testChainID, testPrivAccounts[0])

		if err := execTxWithState(state, tx); err == nil {
			t.Fatalf("Expected invalid data error from %s", data)
		}
	}

	validateEntry := func(t *testing.T, entry *NameRegEntry, name, data string, addr acm.Address, expires uint64) {

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
			t.Fatalf("Wrong expiry. Got %d, expected %d", entry.Expires, expires)
		}
	}

	// try a good one, check data, owner, expiry
	name = "@looking_good/karaoke_bar.broadband"
	data = "on this side of neptune there are 1234567890 people: first is OMNIVORE+-3. Or is it. Ok this is pretty restrictive. No exclamations :(. Faces tho :')"
	amt := fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)
	tx, _ := txs.NewNameTx(state, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithState(state, tx); err != nil {
		t.Fatal(err)
	}
	entry, err := state.GetNameRegEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks)

	// fail to update it as non-owner, in same block
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithState(state, tx); err == nil {
		t.Fatal("Expected error")
	}

	// update it as owner, just to increase expiry, in same block
	// NOTE: we have to resend the data or it will clear it (is this what we want?)
	tx, _ = txs.NewNameTx(state, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks*2)

	// update it as owner, just to increase expiry, in next block
	tx, _ = txs.NewNameTx(state, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), startingBlock+numDesiredBlocks*3)

	// fail to update it as non-owner
	// Fast forward
	for blockchain.Tip().LastBlockHeight() < entry.Expires-1 {
		commitNewBlock(state, blockchain)
	}
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err == nil {
		t.Fatal("Expected error")
	}
	commitNewBlock(state, blockchain)

	// once expires, non-owner succeeds
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[1].Address(), blockchain.LastBlockHeight()+numDesiredBlocks)

	// update it as new owner, with new data (longer), but keep the expiry!
	data = "In the beginning there was no thing, not even the beginning. It hadn't been here, no there, nor for that matter anywhere, not especially because it had not to even exist, let alone to not. Nothing especially odd about that."
	oldCredit := amt - fee
	numDesiredBlocks = 10
	amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data) - oldCredit
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[1].Address(), blockchain.LastBlockHeight()+numDesiredBlocks)

	// test removal
	amt = fee
	data = ""
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
	require.NoError(t, err)
	if entry != nil {
		t.Fatal("Expected removed entry to be nil")
	}

	// create entry by key0,
	// test removal by key1 after expiry
	name = "looking_good/karaoke_bar"
	data = "some data"
	amt = fee + numDesiredBlocks*txs.NameByteCostMultiplier*txs.NameBlockCostMultiplier*txs.NameBaseCost(name, data)
	tx, _ = txs.NewNameTx(state, testPrivAccounts[0].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[0])
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].Address(), blockchain.LastBlockHeight()+numDesiredBlocks)
	// Fast forward
	for blockchain.Tip().LastBlockHeight() < entry.Expires {
		commitNewBlock(state, blockchain)
	}

	amt = fee
	data = ""
	tx, _ = txs.NewNameTx(state, testPrivAccounts[1].PublicKey(), name, data, amt, fee)
	tx.Sign(testChainID, testPrivAccounts[1])
	if err := execTxWithStateNewBlock(state, blockchain, tx); err != nil {
		t.Fatal(err)
	}
	entry, err = state.GetNameRegEntry(name)
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
	acc0PubKey := privAccounts[0].PublicKey()
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())

	newAcc1 := getAccount(state, acc1.Address())
	newAcc1.SetCode(preFactoryCode)
	newAcc2 := getAccount(state, acc2.Address())
	newAcc2.SetCode(factoryCode)

	state.UpdateAccount(newAcc1)
	state.UpdateAccount(newAcc2)

	createData = append(createData, acc2.Address().Word256().Bytes()...)

	// call the pre-factory, triggering the factory to run a create
	tx := &txs.CallTx{
		Input: &txs.TxInput{
			Address:   acc0.Address(),
			Amount:    1,
			Sequence:  acc0.Sequence() + 1,
			PublicKey: acc0PubKey,
		},
		Address:  addressPtr(acc1),
		GasLimit: 10000,
		Data:     createData,
	}

	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	err := execTxWithState(state, tx)
	if err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	acc1 = getAccount(state, acc1.Address())
	firstCreatedAddress, err := state.GetStorage(acc1.Address(), LeftPadWord256(nil))
	require.NoError(t, err)

	acc0 = getAccount(state, acc0.Address())
	// call the pre-factory, triggering the factory to run a create
	tx = &txs.CallTx{
		Input: &txs.TxInput{
			Address:   acc0.Address(),
			Amount:    1,
			Sequence:  acc0.Sequence() + 1,
			PublicKey: acc0PubKey,
		},
		Address:  addressPtr(acc1),
		GasLimit: 100000,
		Data:     createData,
	}

	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	err = execTxWithState(state, tx)
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
	acc0PubKey := privAccounts[0].PublicKey()
	acc1 := getAccount(state, privAccounts[1].Address())
	acc2 := getAccount(state, privAccounts[2].Address())

	newAcc1 := getAccount(state, acc1.Address())
	newAcc1.SetCode(callerCode)
	state.UpdateAccount(newAcc1)

	sendData = append(sendData, acc2.Address().Word256().Bytes()...)
	sendAmt := uint64(10)
	acc2Balance := acc2.Balance()

	// call the contract, triggering the send
	tx := &txs.CallTx{
		Input: &txs.TxInput{
			Address:   acc0.Address(),
			Amount:    sendAmt,
			Sequence:  acc0.Sequence() + 1,
			PublicKey: acc0PubKey,
		},
		Address:  addressPtr(acc1),
		GasLimit: 1000,
		Data:     sendData,
	}

	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	err := execTxWithState(state, tx)
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
	acc0PubKey := privAccounts[0].PublicKey()
	acc1 := getAccount(state, privAccounts[1].Address())

	state.Save()
	// SendTx.
	{
		stateSendTx := state.Copy(dbm.NewMemDB())
		tx := &txs.SendTx{
			Inputs: []*txs.TxInput{
				{
					Address:   acc0.Address(),
					Amount:    1,
					Sequence:  acc0.Sequence() + 1,
					PublicKey: acc0PubKey,
				},
			},
			Outputs: []*txs.TxOutput{
				{
					Address: acc1.Address(),
					Amount:  1,
				},
			},
		}

		tx.Inputs[0].Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateSendTx, tx)
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
		stateCallTx.UpdateAccount(newAcc1)
		tx := &txs.CallTx{
			Input: &txs.TxInput{
				Address:   acc0.Address(),
				Amount:    1,
				Sequence:  acc0.Sequence() + 1,
				PublicKey: acc0PubKey,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateCallTx, tx)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
	}
	state.Save()
	trygetacc0 := getAccount(state, privAccounts[0].Address())
	fmt.Println(trygetacc0.Address())
}

// TODO: test overflows.
// TODO: test for unbonding validators.
func TestTxs(t *testing.T) {
	state, privAccounts := makeGenesisState(3, true, 1000, 1, true, 1000)

	//val0 := state.GetValidatorInfo(privValidators[0].Address())
	acc0 := getAccount(state, privAccounts[0].Address())
	acc0PubKey := privAccounts[0].PublicKey()
	acc1 := getAccount(state, privAccounts[1].Address())

	// SendTx.
	{
		stateSendTx := state.Copy(dbm.NewMemDB())
		tx := &txs.SendTx{
			Inputs: []*txs.TxInput{
				{
					Address:   acc0.Address(),
					Amount:    1,
					Sequence:  acc0.Sequence() + 1,
					PublicKey: acc0PubKey,
				},
			},
			Outputs: []*txs.TxOutput{
				{
					Address: acc1.Address(),
					Amount:  1,
				},
			},
		}

		tx.Inputs[0].Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateSendTx, tx)
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
		stateCallTx.UpdateAccount(newAcc1)
		tx := &txs.CallTx{
			Input: &txs.TxInput{
				Address:   acc0.Address(),
				Amount:    1,
				Sequence:  acc0.Sequence() + 1,
				PublicKey: acc0PubKey,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err := execTxWithState(stateCallTx, tx)
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
		tx := &txs.NameTx{
			Input: &txs.TxInput{
				Address:   acc0.Address(),
				Amount:    entryAmount,
				Sequence:  acc0.Sequence() + 1,
				PublicKey: acc0PubKey,
			},
			Name: entryName,
			Data: entryData,
		}

		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)

		err := execTxWithState(stateNameTx, tx)
		if err != nil {
			t.Errorf("Got error in executing call transaction, %v", err)
		}
		newAcc0 := getAccount(stateNameTx, acc0.Address())
		if acc0.Balance()-entryAmount != newAcc0.Balance() {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance()-entryAmount, newAcc0.Balance())
		}
		entry, err := stateNameTx.GetNameRegEntry(entryName)
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
		tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
		err = execTxWithState(stateNameTx, tx)
		if _, ok := err.(txs.ErrTxInvalidString); !ok {
			t.Errorf("Expected invalid string error. Got: %s", err.Error())
		}
	}

	// BondTx. TODO
	/*
		{
			state := state.Copy()
			tx := &txs.BondTx{
				PublicKey: acc0PubKey.(acm.PublicKeyEd25519),
				Inputs: []*txs.TxInput{
					&txs.TxInput{
						Address:  acc0.Address(),
						Amount:   1,
						Sequence: acc0.Sequence() + 1,
						PublicKey:   acc0PubKey,
					},
				},
				UnbondTo: []*txs.TxOutput{
					&txs.TxOutput{
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
			if acc0Val.VotingPower != 1 {
				t.Errorf("Unexpected voting power. Expected %v, got %v",
					acc0Val.VotingPower, acc0.Balance())
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
	state.UpdateAccount(newAcc1)

	// send call tx with no data, cause self-destruct
	tx := txs.NewCallTxWithSequence(acc0PubKey, addressPtr(acc1), nil, sendingAmount, 1000, 0, acc0.Sequence()+1)
	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)

	// we use cache instead of execTxWithState so we can run the tx twice
	exe := NewBatchCommitter(state, testChainID, bcm.NewBlockchain(nil, testGenesisDoc), event.NewNoOpPublisher(), logger)
	if err := exe.Execute(tx); err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

	// if we do it again, we won't get an error, but the self-destruct
	// shouldn't happen twice and the caller should lose fee
	tx.Input.Sequence += 1
	tx.Input.Signature = acm.ChainSign(privAccounts[0], testChainID, tx)
	if err := exe.Execute(tx); err != nil {
		t.Errorf("Got error in executing call transaction, %v", err)
	}

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

func execTxWithStateAndBlockchain(state *State, tip bcm.Tip, tx txs.Tx) error {
	exe := newExecutor(true, state, testChainID, tip, event.NewNoOpPublisher(), logger)
	if err := exe.Execute(tx); err != nil {
		return err
	} else {
		exe.Commit()
		return nil
	}
}

func execTxWithState(state *State, tx txs.Tx) error {
	return execTxWithStateAndBlockchain(state, bcm.NewBlockchain(nil, testGenesisDoc), tx)
}

func commitNewBlock(state *State, blockchain bcm.MutableBlockchain) {
	blockchain.CommitBlock(blockchain.LastBlockTime().Add(time.Second), sha3.Sha3(blockchain.LastBlockHash()),
		state.Hash())
}

func execTxWithStateNewBlock(state *State, blockchain bcm.MutableBlockchain, tx txs.Tx) error {
	if err := execTxWithStateAndBlockchain(state, blockchain, tx); err != nil {
		return err
	}
	commitNewBlock(state, blockchain)
	return nil
}

func makeGenesisState(numAccounts int, randBalance bool, minBalance uint64, numValidators int, randBonded bool,
	minBonded int64) (*State, []acm.PrivateAccount) {
	testGenesisDoc, privAccounts, _ := deterministicGenesis.GenesisDoc(numAccounts, randBalance, minBalance,
		numValidators, randBonded, minBonded)
	s0, err := MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	if err != nil {
		panic(fmt.Errorf("could not make genesis state: %v", err))
	}
	s0.Save()
	return s0, privAccounts
}

func getAccount(state acm.Getter, address acm.Address) acm.MutableAccount {
	acc, _ := acm.GetMutableAccount(state, address)
	return acc
}

func addressPtr(account acm.Account) *acm.Address {
	if account == nil {
		return nil
	}
	accountAddresss := account.Address()
	return &accountAddresss
}

//-------------------------------------------------------------------------------------
// helpers

var ExceptionTimeOut = "timed out waiting for event"

// run ExecTx and wait for the Call event on given addr
// returns the msg data and an error/exception
func execTxWaitEvent(t *testing.T, batchCommitter *executor, tx txs.Tx, eventid string) (interface{}, string) {
	emitter := event.NewEmitter(logger)
	ch := make(chan interface{})
	emitter.Subscribe(context.Background(), "test", event.QueryForEventID(eventid), ch)
	evc := event.NewEventCache(emitter)
	batchCommitter.eventCache = evc
	go func() {
		if err := batchCommitter.Execute(tx); err != nil {
			ch <- err.Error()
		}
		evc.Flush()
	}()
	ticker := time.NewTicker(5 * time.Second)

	select {
	case msg := <-ch:
		switch ev := msg.(type) {
		case *exe_events.EventDataTx:
			return ev, ev.Exception
		case *evm_events.EventDataCall:
			return ev, ev.Exception
		case string:
			return nil, ev
		default:
			return ev, ""
		}
	case <-ticker.C:
		return nil, ExceptionTimeOut
	}

}

// give a contract perms for an snative, call it, it calls the snative, but shouldn't have permission
func testSNativeCALLExpectFail(t *testing.T, batchCommitter *executor, doug acm.MutableAccount,
	snativeAddress acm.Address, data []byte) {
	testSNativeCALL(t, false, batchCommitter, doug, 0, snativeAddress, data, nil)
}

// give a contract perms for an snative, call it, it calls the snative, ensure the check funciton (f) succeeds
func testSNativeCALLExpectPass(t *testing.T, batchCommitter *executor, doug acm.MutableAccount, snativePerm ptypes.PermFlag,
	snativeAddress acm.Address, data []byte, f func([]byte) error) {
	testSNativeCALL(t, true, batchCommitter, doug, snativePerm, snativeAddress, data, f)
}

func testSNativeCALL(t *testing.T, expectPass bool, batchCommitter *executor, doug acm.MutableAccount,
	snativePerm ptypes.PermFlag, snativeAddress acm.Address, data []byte, f func([]byte) error) {
	if expectPass {
		doug.MutablePermissions().Base.Set(snativePerm, true)
	}

	doug.SetCode(callContractCode(snativeAddress))
	dougAddress := doug.Address()

	batchCommitter.stateCache.UpdateAccount(doug)
	tx, _ := txs.NewCallTx(batchCommitter.stateCache, users[0].PublicKey(), &dougAddress, data, 100, 10000, 100)
	tx.Sign(testChainID, users[0])
	fmt.Println("subscribing to", evm_events.EventStringAccountCall(snativeAddress))
	ev, exception := execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccountCall(snativeAddress))
	if exception == ExceptionTimeOut {
		t.Fatal("Timed out waiting for event")
	}
	if expectPass {
		if exception != "" {
			t.Fatal("Unexpected exception", exception)
		}
		evv := ev.(*evm_events.EventDataCall)
		ret := evv.Return
		if err := f(ret); err != nil {
			t.Fatal(err)
		}
	} else {
		if exception == "" {
			t.Fatal("Expected exception")
		}
	}
}

func testSNativeTxExpectFail(t *testing.T, batchCommitter *executor, snativeArgs permission.PermArgs) {
	testSNativeTx(t, false, batchCommitter, 0, snativeArgs)
}

func testSNativeTxExpectPass(t *testing.T, batchCommitter *executor, perm ptypes.PermFlag, snativeArgs permission.PermArgs) {
	testSNativeTx(t, true, batchCommitter, perm, snativeArgs)
}

func testSNativeTx(t *testing.T, expectPass bool, batchCommitter *executor, perm ptypes.PermFlag, snativeArgs permission.PermArgs) {
	if expectPass {
		acc := getAccount(batchCommitter.stateCache, users[0].Address())
		acc.MutablePermissions().Base.Set(perm, true)
		batchCommitter.stateCache.UpdateAccount(acc)
	}
	tx, _ := txs.NewPermissionsTx(batchCommitter.stateCache, users[0].PublicKey(), snativeArgs)
	tx.Sign(testChainID, users[0])
	err := batchCommitter.Execute(tx)
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

func snativePermTestInputCALL(name string, user acm.PrivateAccount, perm ptypes.PermFlag,
	val bool) (addr acm.Address, pF ptypes.PermFlag, data []byte) {
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
	if pF, err = permission.PermStringToFlag(name); err != nil {
		panic(fmt.Sprintf("failed to convert perm string (%s) to flag", name))
	}
	return
}

func snativePermTestInputTx(name string, user acm.PrivateAccount, perm ptypes.PermFlag, val bool) (snativeArgs permission.PermArgs) {
	switch name {
	case "hasBase":
		snativeArgs = permission.HasBaseArgs(user.Address(), perm)
	case "unsetBase":
		snativeArgs = permission.UnsetBaseArgs(user.Address(), perm)
	case "setBase":
		snativeArgs = permission.SetBaseArgs(user.Address(), perm, val)
	case "setGlobal":
		snativeArgs = permission.SetGlobalArgs(perm, val)
	}
	return
}

func snativeRoleTestInputCALL(name string, user acm.PrivateAccount,
	role string) (addr acm.Address, pF ptypes.PermFlag, data []byte) {
	addr = permissionsContract.Address()
	data = user.Address().Word256().Bytes()
	data = append(data, RightPadBytes([]byte(role), 32)...)
	data = append(permNameToFuncID(name), data...)

	var err error
	if pF, err = permission.PermStringToFlag(name); err != nil {
		panic(fmt.Sprintf("failed to convert perm string (%s) to flag", name))
	}
	return
}

func snativeRoleTestInputTx(name string, user acm.PrivateAccount, role string) (snativeArgs permission.PermArgs) {
	switch name {
	case "hasRole":
		snativeArgs = permission.HasRoleArgs(user.Address(), role)
	case "addRole":
		snativeArgs = permission.AddRoleArgs(user.Address(), role)
	case "removeRole":
		snativeArgs = permission.RemoveRoleArgs(user.Address(), role)
	}
	return
}

// convenience function for contract that calls a given address
func callContractCode(contractAddr acm.Address) []byte {
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
