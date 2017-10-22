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
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging/loggers"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
	"github.com/hyperledger/burrow/txs"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	dbBackend           = "memdb"
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

func makeUsers(n int) []acm.PrivateAccount {
	users := make([]acm.PrivateAccount, n)
	for i := 0; i < n; i++ {
		secret := "mysecret" + strconv.Itoa(i)
		users[i] = acm.GeneratePrivateAccountFromSecret(secret)
	}
	return users
}

func makeExecutor(state *State) *executor {
	return newExecutor(true, state, testChainID, bcm.NewBlockchain(testGenesisDoc), event.NewEmitter(logger),
		logger)
}

func newBaseGenDoc(globalPerm, accountPerm ptypes.AccountPermissions) genesis.GenesisDoc {
	genAccounts := []genesis.GenesisAccount{}
	for _, user := range users[:5] {
		accountPermCopy := accountPerm // Create new instance for custom overridability.
		genAccounts = append(genAccounts, genesis.GenesisAccount{
			BasicAccount: genesis.BasicAccount{
				Address: user.Address(),
				Amount:  1000000,
			},
			Permissions: accountPermCopy,
		})
	}

	return genesis.GenesisDoc{
		GenesisTime: time.Now(),
		ChainName:   testGenesisDoc.ChainName,
		Params: genesis.GenesisParams{
			GlobalPermissions: globalPerm,
		},
		Accounts: genAccounts,
		Validators: []genesis.GenesisValidator{
			{
				PubKey: users[0].PubKey(),
				Amount: 10,
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
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(permission.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(permission.CreateContract, true)
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	//-------------------
	// send txs

	// simple send tx should fail
	tx := txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
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
	if err := tx.AddInput(batchCommitter.blockCache, users[2].PubKey(), 5); err != nil {
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
	if err := tx.AddInput(batchCommitter.blockCache, users[3].PubKey(), 5); err != nil {
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
	acc := getAccount(batchCommitter.blockCache, users[3].Address())
	acc.MutablePermissions().Base.Set(permission.Send, true)
	batchCommitter.blockCache.UpdateAccount(acc)
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[3].PubKey(), 5); err != nil {
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
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Name, true)
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	//-------------------
	// name txs

	// simple name tx without perm should fail
	tx, err := txs.NewNameTx(st, users[0].PubKey(), "somename", "somedata", 10000, 100)
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
	tx, err = txs.NewNameTx(st, users[1].PubKey(), "somename", "somedata", 10000, 100)
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
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(permission.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(permission.CreateContract, true)
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	//-------------------
	// call txs

	address4 := users[4].Address()
	// simple call tx should fail
	tx, _ := txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &address4, nil, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call tx with send permission should fail
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[1].PubKey(), &address4, nil, 100, 100, 100)
	tx.Sign(testChainID, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call tx with create permission should fail
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[3].PubKey(), &address4, nil, 100, 100, 100)
	tx.Sign(testChainID, users[3])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	//-------------------
	// create txs

	// simple call create tx should fail
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), nil, nil, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call create tx with send perm should fail
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[1].PubKey(), nil, nil, 100, 100, 100)
	tx.Sign(testChainID, users[1])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// simple call create tx with call perm should fail
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[2].PubKey(), nil, nil, 100, 100, 100)
	tx.Sign(testChainID, users[2])
	if err := batchCommitter.Execute(tx); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}
}

func TestSendPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true) // give the 0 account permission
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	// A single input, having the permission, should succeed
	tx := txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}

	// Two inputs, one with permission, one without, should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
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
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	st := MakeGenesisState(stateDB, &genDoc)
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
	tx, _ := txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &simpleContractAddr, nil, 100, 100, 100)
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
	batchCommitter.blockCache.UpdateAccount(caller1Acc)

	// A single input, having the permission, but the contract doesn't have permission
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception := execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(caller1ContractAddr)) //
	if exception == "" {
		t.Fatal("Expected exception")
	}

	//----------------------------------------------------------
	// call to contract that calls simple contract - with perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (PASS)")

	// A single input, having the permission, and the contract has permission
	caller1Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.blockCache.UpdateAccount(caller1Acc)
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(caller1ContractAddr)) //
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
	batchCommitter.blockCache.UpdateAccount(caller1Acc)
	batchCommitter.blockCache.UpdateAccount(caller2Acc)

	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(caller1ContractAddr)) //
	if exception == "" {
		t.Fatal("Expected exception")
	}

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// both caller1 and caller2 have permission
	fmt.Println("\n##### CALL TO CONTRACT CALLING SIMPLE CONTRACT (PASS)")

	caller1Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.blockCache.UpdateAccount(caller1Acc)

	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	tx.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(caller1ContractAddr)) //
	if exception != "" {
		t.Fatal("Unexpected exception", exception)
	}
}

func TestCreatePermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.CreateContract, true) // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true)           // give the 0 account permission
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	//------------------------------
	// create a simple contract
	fmt.Println("\n##### CREATE SIMPLE CONTRACT")

	contractCode := []byte{0x60}
	createCode := wrapContractForCreate(contractCode)

	// A single input, having the permission, should succeed
	tx, _ := txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), nil, createCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}
	// ensure the contract is there
	contractAddr := acm.NewContractAddress(tx.Input.Address, tx.Input.Sequence)
	contractAcc := getAccount(batchCommitter.blockCache, contractAddr)
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
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), nil, createFactoryCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}
	// ensure the contract is there
	contractAddr = acm.NewContractAddress(tx.Input.Address, tx.Input.Sequence)
	contractAcc = getAccount(batchCommitter.blockCache, contractAddr)
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
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &contractAddr, createCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	// we need to subscribe to the Call event to detect the exception
	_, exception := execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(contractAddr)) //
	if exception == "" {
		t.Fatal("expected exception")
	}

	//------------------------------
	// call the contract (should PASS)
	fmt.Println("\n###### CALL THE FACTORY (PASS)")

	contractAcc.MutablePermissions().Base.Set(permission.CreateContract, true)
	batchCommitter.blockCache.UpdateAccount(contractAcc)

	// A single input, having the permission, should succeed
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &contractAddr, createCode, 100, 100, 100)
	tx.Sign(testChainID, users[0])
	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(contractAddr)) //
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
	batchCommitter.blockCache.UpdateAccount(contractAcc)

	// this should call the 0 address but not create ...
	tx, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &contractAddr, createCode, 100, 10000, 100)
	tx.Sign(testChainID, users[0])
	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(acm.Address{})) //
	if exception != "" {
		t.Fatal("unexpected exception", exception)
	}
	zeroAcc := getAccount(batchCommitter.blockCache, acm.Address{})
	if len(zeroAcc.Code()) != 0 {
		t.Fatal("the zero account was given code from a CALL!")
	}
}

/* TODO
func TestBondPermission(t *testing.T) {
	stateDB := dbm.NewDB("state",dbBackend,dbDir)
	genDoc := newBaseGenDoc(PermsAllFalse, PermsAllFalse)
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)
	var bondAcc *acm.Account

	//------------------------------
	// one bonder without permission should fail
	tx, _ := txs.NewBondTx(users[1].PubKey())
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[1])
	tx.SignBond(testChainID, users[1])
	if err := ExecTx(batchCommitter.blockCache, tx, true, nil); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	//------------------------------
	// one bonder with permission should pass
	bondAcc = batchCommitter.blockCache.GetAccount(users[1].Address())
	bondAcc.Permissions.Base.Set(permission.Bond, true)
	batchCommitter.blockCache.UpdateAccount(bondAcc)
	if err := ExecTx(batchCommitter.blockCache, tx, true, nil); err != nil {
		t.Fatal("Unexpected error", err)
	}

	// reset state (we can only bond with an account once ..)
	genDoc = newBaseGenDoc(PermsAllFalse, PermsAllFalse)
	st = MakeGenesisState(stateDB, &genDoc)
	batchCommitter.blockCache = NewBlockCache(st)
	bondAcc = batchCommitter.blockCache.GetAccount(users[1].Address())
	bondAcc.Permissions.Base.Set(permission.Bond, true)
	batchCommitter.blockCache.UpdateAccount(bondAcc)
	//------------------------------
	// one bonder with permission and an input without send should fail
	tx, _ = txs.NewBondTx(users[1].PubKey())
	if err := tx.AddInput(batchCommitter.blockCache, users[2].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[2])
	tx.SignBond(testChainID, users[1])
	if err := ExecTx(batchCommitter.blockCache, tx, true, nil); err == nil {
		t.Fatal("Expected error")
	} else {
		fmt.Println(err)
	}

	// reset state (we can only bond with an account once ..)
	genDoc = newBaseGenDoc(PermsAllFalse, PermsAllFalse)
	st = MakeGenesisState(stateDB, &genDoc)
	batchCommitter.blockCache = NewBlockCache(st)
	bondAcc = batchCommitter.blockCache.GetAccount(users[1].Address())
	bondAcc.Permissions.Base.Set(permission.Bond, true)
	batchCommitter.blockCache.UpdateAccount(bondAcc)
	//------------------------------
	// one bonder with permission and an input with send should pass
	sendAcc := batchCommitter.blockCache.GetAccount(users[2].Address())
	sendAcc.Permissions.Base.Set(permission.Send, true)
	batchCommitter.blockCache.UpdateAccount(sendAcc)
	tx, _ = txs.NewBondTx(users[1].PubKey())
	if err := tx.AddInput(batchCommitter.blockCache, users[2].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[2])
	tx.SignBond(testChainID, users[1])
	if err := ExecTx(batchCommitter.blockCache, tx, true, nil); err != nil {
		t.Fatal("Unexpected error", err)
	}

	// reset state (we can only bond with an account once ..)
	genDoc = newBaseGenDoc(PermsAllFalse, PermsAllFalse)
	st = MakeGenesisState(stateDB, &genDoc)
	batchCommitter.blockCache = NewBlockCache(st)
	bondAcc = batchCommitter.blockCache.GetAccount(users[1].Address())
	bondAcc.Permissions.Base.Set(permission.Bond, true)
	batchCommitter.blockCache.UpdateAccount(bondAcc)
	//------------------------------
	// one bonder with permission and an input with bond should pass
	sendAcc.Permissions.Base.Set(permission.Bond, true)
	batchCommitter.blockCache.UpdateAccount(sendAcc)
	tx, _ = txs.NewBondTx(users[1].PubKey())
	if err := tx.AddInput(batchCommitter.blockCache, users[2].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[2])
	tx.SignBond(testChainID, users[1])
	if err := ExecTx(batchCommitter.blockCache, tx, true, nil); err != nil {
		t.Fatal("Unexpected error", err)
	}

	// reset state (we can only bond with an account once ..)
	genDoc = newBaseGenDoc(PermsAllFalse, PermsAllFalse)
	st = MakeGenesisState(stateDB, &genDoc)
	batchCommitter.blockCache = NewBlockCache(st)
	bondAcc = batchCommitter.blockCache.GetAccount(users[1].Address())
	bondAcc.Permissions.Base.Set(permission.Bond, true)
	batchCommitter.blockCache.UpdateAccount(bondAcc)
	//------------------------------
	// one bonder with permission and an input from that bonder and an input without send or bond should fail
	tx, _ = txs.NewBondTx(users[1].PubKey())
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.blockCache, users[2].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].Address(), 5)
	tx.SignInput(testChainID, 0, users[1])
	tx.SignInput(testChainID, 1, users[2])
	tx.SignBond(testChainID, users[1])
	if err := ExecTx(batchCommitter.blockCache, tx, true, nil); err == nil {
		t.Fatal("Expected error")
	}
}
*/

func TestCreateAccountPermission(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)          // give the 0 account permission
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)          // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.CreateAccount, true) // give the 0 account permission
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// SendTx to unknown account

	// A single input, having the permission, should succeed
	tx := txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].Address(), 5)
	tx.SignInput(testChainID, 0, users[0])
	if err := batchCommitter.Execute(tx); err != nil {
		t.Fatal("Transaction failed", err)
	}

	// Two inputs, both with send, one with create, one without, should fail
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
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
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
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
	acc := getAccount(batchCommitter.blockCache, users[1].Address())
	acc.MutablePermissions().Base.Set(permission.CreateAccount, true)
	batchCommitter.blockCache.UpdateAccount(acc)
	tx = txs.NewSendTx()
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
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
	if err := tx.AddInput(batchCommitter.blockCache, users[0].PubKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(batchCommitter.blockCache, users[1].PubKey(), 5); err != nil {
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

	acc = getAccount(batchCommitter.blockCache, users[0].Address())
	acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.blockCache.UpdateAccount(acc)

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
	batchCommitter.blockCache.UpdateAccount(caller1Acc)

	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ := txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCall.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception := execTxWaitEvent(t, batchCommitter, txCall, evm_events.EventStringAccCall(caller1ContractAddr)) //
	if exception == "" {
		t.Fatal("Expected exception")
	}

	// NOTE: for a contract to be able to CreateAccount, it must be able to call
	// NOTE: for a users to be able to CreateAccount, it must be able to send!
	caller1Acc.MutablePermissions().Base.Set(permission.CreateAccount, true)
	caller1Acc.MutablePermissions().Base.Set(permission.Call, true)
	batchCommitter.blockCache.UpdateAccount(caller1Acc)
	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ = txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCall.Sign(testChainID, users[0])

	// we need to subscribe to the Call event to detect the exception
	_, exception = execTxWaitEvent(t, batchCommitter, txCall, evm_events.EventStringAccCall(caller1ContractAddr)) //
	if exception != "" {
		t.Fatal("Unexpected exception", exception)
	}

}

// holla at my boy
var DougAddress = acm.AddressFromString("THISISDOUG")

func TestSNativeCALL(t *testing.T) {
	stateDB := dbm.NewDB("state", dbBackend, dbDir)
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(permission.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st := MakeGenesisState(stateDB, &genDoc)
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
	batchCommitter.blockCache.UpdateAccount(doug)

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

	fmt.Println("\n#### RmRole")
	// RmRole
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
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(permission.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st := MakeGenesisState(stateDB, &genDoc)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// Test SNativeTx

	fmt.Println("\n#### SetBase")
	// SetBase
	snativeArgs := snativePermTestInputTx("setBase", users[3], permission.Bond, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetBase, snativeArgs)
	acc := getAccount(batchCommitter.blockCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(permission.Bond); v {
		t.Fatal("expected permission to be set false")
	}
	snativeArgs = snativePermTestInputTx("setBase", users[3], permission.CreateContract, true)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetBase, snativeArgs)
	acc = getAccount(batchCommitter.blockCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(permission.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeArgs = snativePermTestInputTx("unsetBase", users[3], permission.CreateContract, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.UnsetBase, snativeArgs)
	acc = getAccount(batchCommitter.blockCache, users[3].Address())
	if v, _ := acc.MutablePermissions().Base.Get(permission.CreateContract); v {
		t.Fatal("expected permission to be set false")
	}

	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeArgs = snativePermTestInputTx("setGlobal", users[3], permission.CreateContract, true)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetGlobal, snativeArgs)
	acc = getAccount(batchCommitter.blockCache, permission.GlobalPermissionsAddress)
	if v, _ := acc.MutablePermissions().Base.Get(permission.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### AddRole")
	// AddRole
	snativeArgs = snativeRoleTestInputTx("addRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.AddRole, snativeArgs)
	acc = getAccount(batchCommitter.blockCache, users[3].Address())
	if v := acc.Permissions().HasRole("chuck"); !v {
		t.Fatal("expected role to be added")
	}

	fmt.Println("\n#### RmRole")
	// RmRole
	snativeArgs = snativeRoleTestInputTx("removeRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.RmRole, snativeArgs)
	acc = getAccount(batchCommitter.blockCache, users[3].Address())
	if v := acc.Permissions().HasRole("chuck"); v {
		t.Fatal("expected role to be removed")
	}
}

//-------------------------------------------------------------------------------------
// helpers

var ExceptionTimeOut = "timed out waiting for event"

// run ExecTx and wait for the Call event on given addr
// returns the msg data and an error/exception
func execTxWaitEvent(t *testing.T, batchCommitter *executor, tx txs.Tx, eventid string) (interface{}, string) {
	evsw := event.NewEmitter(logger)
	ch := make(chan event.AnyEventData)
	evsw.Subscribe("test", eventid, func(msg event.AnyEventData) {
		ch <- msg
	})
	evc := event.NewEventCache(evsw)
	batchCommitter.eventCache = evc
	go func() {
		if err := batchCommitter.Execute(tx); err != nil {
			errStr := err.Error()
			ch <- event.AnyEventData{Err: &errStr}
		}
		evc.Flush()
	}()
	ticker := time.NewTicker(5 * time.Second)
	var msg event.AnyEventData
	select {
	case msg = <-ch:
	case <-ticker.C:
		return nil, ExceptionTimeOut
	}

	switch ev := msg.Get().(type) {
	case exe_events.EventDataTx:
		return ev, ev.Exception
	case evm_events.EventDataCall:
		return ev, ev.Exception
	case string:
		return nil, ev
	default:
		return ev, ""
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

	batchCommitter.blockCache.UpdateAccount(doug)
	tx, _ := txs.NewCallTx(batchCommitter.blockCache, users[0].PubKey(), &dougAddress, data, 100, 10000, 100)
	tx.Sign(testChainID, users[0])
	fmt.Println("subscribing to", evm_events.EventStringAccCall(snativeAddress))
	ev, exception := execTxWaitEvent(t, batchCommitter, tx, evm_events.EventStringAccCall(snativeAddress))
	if exception == ExceptionTimeOut {
		t.Fatal("Timed out waiting for event")
	}
	if expectPass {
		if exception != "" {
			t.Fatal("Unexpected exception", exception)
		}
		evv := ev.(evm_events.EventDataCall)
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
		acc := getAccount(batchCommitter.blockCache, users[0].Address())
		acc.MutablePermissions().Base.Set(perm, true)
		batchCommitter.blockCache.UpdateAccount(acc)
	}
	tx, _ := txs.NewPermissionsTx(batchCommitter.blockCache, users[0].PubKey(), snativeArgs)
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
		snativeArgs = &permission.HasBaseArgs{user.Address(), perm}
	case "unsetBase":
		snativeArgs = &permission.UnsetBaseArgs{user.Address(), perm}
	case "setBase":
		snativeArgs = &permission.SetBaseArgs{user.Address(), perm, val}
	case "setGlobal":
		snativeArgs = &permission.SetGlobalArgs{perm, val}
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
		snativeArgs = &permission.HasRoleArgs{user.Address(), role}
	case "addRole":
		snativeArgs = &permission.AddRoleArgs{user.Address(), role}
	case "removeRole":
		snativeArgs = &permission.RmRoleArgs{user.Address(), role}
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
	return bc.Splice(CALLDATASIZE, PUSH1, inputOff, PUSH1, memOff,
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
