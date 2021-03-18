// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package execution

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/burrow/execution/engine"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/bcm"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	. "github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/names"
	"github.com/hyperledger/burrow/execution/native"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
)

var (
	permissionsContract = native.Permissions.GetByName("Permissions").(*native.Contract)
	dbBackend           = dbm.MemDBBackend
	dbDir               = ""
)

// keys
var users = makeUsers(10)
var logger = logging.NewNoopLogger()
var deterministicGenesis = genesis.NewDeterministicGenesis(34059836243380576)
var testGenesisDoc, testPrivAccounts, _ = deterministicGenesis.
	GenesisDoc(3, 1)
var testChainID = testGenesisDoc.GetChainID()

func TestSendFails(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(permission.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(permission.CreateContract, true)
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//-------------------
	// send txs

	// simple send tx should fail
	tx := payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].GetAddress(), 5)
	err = exe.signExecuteCommit(tx, users[0])
	require.Error(t, err)

	// simple send tx with call perm should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[2].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[4].GetAddress(), 5)

	err = exe.signExecuteCommit(tx, users[2])
	require.Error(t, err)

	// simple send tx with create perm should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[3].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[4].GetAddress(), 5)
	err = exe.signExecuteCommit(tx, users[3])
	require.Error(t, err)

	// simple send tx to unknown account without create_account perm should fail
	acc := getAccount(t, exe.stateCache, users[3].GetAddress())
	err = acc.Permissions.Base.Set(permission.Send, true)
	require.NoError(t, err)
	exe.stateCache.UpdateAccount(acc)
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[3].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].GetAddress(), 5)
	err = exe.signExecuteCommit(tx, users[3])
	require.Error(t, err)
}

func TestName(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Name, true)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Input, true)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Input, true)
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//-------------------
	// name txs

	// simple name tx without perm should fail
	tx, err := payload.NewNameTx(st, users[0].GetPublicKey(), "somename", "somedata", 10000, 100)
	if err != nil {
		t.Fatal(err)
	}
	err = exe.signExecuteCommit(tx, users[0])
	require.Error(t, err)

	// simple name tx with perm should pass
	tx, err = payload.NewNameTx(st, users[1].GetPublicKey(), "somename", "somedata", 10000, 100)
	if err != nil {
		t.Fatal(err)
	}

	err = exe.signExecuteCommit(tx, users[1])
	require.NoError(t, err)
}

func TestCallFails(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)
	genDoc.Accounts[2].Permissions.Base.Set(permission.Call, true)
	genDoc.Accounts[3].Permissions.Base.Set(permission.CreateContract, true)
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//-------------------
	// call txs

	address4 := users[4].GetAddress()
	// simple call tx should fail
	tx, err := payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &address4, nil, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[0])
	require.Error(t, err)

	// simple call tx with send permission should fail
	tx, err = payload.NewCallTx(exe.stateCache, users[1].GetPublicKey(), &address4, nil, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[1])
	require.Error(t, err)

	// simple call tx with create permission should fail
	tx, err = payload.NewCallTx(exe.stateCache, users[3].GetPublicKey(), &address4, nil, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[3])
	require.Error(t, err)

	//-------------------
	// create txs

	// simple call create tx should fail
	tx, err = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), nil, nil, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[0])
	require.Error(t, err)

	// simple call create tx with send perm should fail
	tx, err = payload.NewCallTx(exe.stateCache, users[1].GetPublicKey(), nil, nil, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[1])
	require.Error(t, err)

	// simple call create tx with call perm should fail
	tx, err = payload.NewCallTx(exe.stateCache, users[2].GetPublicKey(), nil, nil, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[2])
	require.Error(t, err)
}

func TestSendPermission(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)  // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Input, true) // give the 0 account permission
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	// A single input, having the permission, should succeed
	tx := payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[1].GetAddress(), 5)
	err = exe.signExecuteCommit(tx, users[0])
	require.NoError(t, err)

	// Two inputs, one with permission, one without, should fail
	tx = payload.NewSendTx()
	require.NoError(t, tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5))
	require.NoError(t, tx.AddInput(exe.stateCache, users[1].GetPublicKey(), 5))
	require.NoError(t, tx.AddOutput(users[2].GetAddress(), 10))
	err = exe.signExecuteCommit(tx, users[:2]...)
	require.Error(t, err)
}

func TestCallPermission(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true)  // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Input, true) // give the 0 account permission
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//------------------------------
	// call to simple contract
	fmt.Println("\n##### SIMPLE CONTRACT")

	// create simple contract
	simpleContractAddr := crypto.NewContractAddress(users[0].GetAddress(), []byte{100})
	simpleAcc := &acm.Account{
		Address:     simpleContractAddr,
		Balance:     0,
		EVMCode:     []byte{0x60},
		Sequence:    0,
		Permissions: permission.ZeroAccountPermissions,
	}
	exe.updateAccounts(t, simpleAcc)

	// A single input, having the permission, should succeed
	tx, _ := payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &simpleContractAddr, nil, 100, 100, 100)
	err = exe.signExecuteCommit(tx, users[0])
	require.NoError(t, err)

	//----------------------------------------------------------
	// call to contract that calls simple contract - without perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (FAIL)")

	// create contract that calls the simple contract
	contractCode := callContractCode(simpleContractAddr)
	caller1ContractAddr := crypto.NewContractAddress(users[0].GetAddress(), []byte{101})
	caller1Acc := &acm.Account{
		Address:     caller1ContractAddr,
		Balance:     10000,
		EVMCode:     contractCode,
		Sequence:    0,
		Permissions: permission.ZeroAccountPermissions,
	}
	exe.stateCache.UpdateAccount(caller1Acc)

	// A single input, having the permission, but the contract doesn't have permission
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr) //
	require.Error(t, err)

	//----------------------------------------------------------
	// call to contract that calls simple contract - with perm
	fmt.Println("\n##### CALL TO SIMPLE CONTRACT (PASS)")

	// A single input, having the permission, and the contract has permission
	caller1Acc.Permissions.Base.Set(permission.Call, true)
	exe.stateCache.UpdateAccount(caller1Acc)
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr)
	require.NoError(t, err)

	//----------------------------------------------------------
	// call to contract that calls contract that calls simple contract - without perm
	// caller1Contract calls simpleContract. caller2Contract calls caller1Contract.
	// caller1Contract does not have call perms, but caller2Contract does.
	fmt.Println("\n##### CALL TO CONTRACT CALLING SIMPLE CONTRACT (FAIL)")

	contractCode2 := callContractCode(caller1ContractAddr)
	caller2ContractAddr := crypto.NewContractAddress(users[0].GetAddress(), []byte{102})
	caller2Acc := &acm.Account{
		Address:     caller2ContractAddr,
		Balance:     1000,
		EVMCode:     contractCode2,
		Sequence:    0,
		Permissions: permission.ZeroAccountPermissions,
	}
	caller1Acc.Permissions.Base.Set(permission.Call, false)
	caller2Acc.Permissions.Base.Set(permission.Call, true)
	exe.stateCache.UpdateAccount(caller1Acc)
	exe.stateCache.UpdateAccount(caller2Acc)

	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &caller2ContractAddr, nil, 100, 10000, 100)
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

	caller1Acc.Permissions.Base.Set(permission.Call, true)
	exe.stateCache.UpdateAccount(caller1Acc)

	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &caller2ContractAddr, nil, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, caller1ContractAddr) //
	require.NoError(t, err)
}

func TestCreatePermission(t *testing.T) {
	stateDB := dbm.NewMemDB()
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.CreateContract, true) // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true)           // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Input, true)          // give the 0 account permission
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//------------------------------
	// create a simple contract
	fmt.Println("\n##### CREATE SIMPLE CONTRACT")

	contractCode := []byte{0x60}
	createCode := wrapContractForCreate(contractCode)

	// A single input, having the permission, should succeed
	tx, err := payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), nil, createCode, 100, 100, 100)
	require.NoError(t, err)
	err = exe.signExecuteCommit(tx, users[0])
	require.NoError(t, err)

	// ensure the contract is there
	hash := getTxHash(tx)
	contractAddr := crypto.NewContractAddress(tx.Input.Address, hash)
	contractAcc := getAccount(t, exe.stateCache, contractAddr)
	if contractAcc == nil {
		t.Fatalf("failed to create contract %s", contractAddr)
	}
	if !bytes.Equal(contractAcc.EVMCode, contractCode) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.EVMCode, contractCode)
	}

	//------------------------------
	// create contract that uses the CREATE op
	fmt.Println("\n##### CREATE FACTORY")

	contractCode = []byte{0x60}
	createCode = wrapContractForCreate(contractCode)
	factoryCode := createContractCode()
	createFactoryCode := wrapContractForCreate(factoryCode)

	// A single input, having the permission, should succeed
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), nil, createFactoryCode, 100, 100, 100)
	err = exe.signExecuteCommit(tx, users[0])
	require.NoError(t, err)

	// ensure the contract is there
	hash = getTxHash(tx)
	contractAddr = crypto.NewContractAddress(tx.Input.Address, hash)
	contractAcc = getAccount(t, exe.stateCache, contractAddr)
	if contractAcc == nil {
		t.Fatalf("failed to create contract %s", contractAddr)
	}
	if !bytes.Equal(contractAcc.EVMCode, factoryCode) {
		t.Fatalf("contract does not have correct code. Got %X, expected %X", contractAcc.EVMCode, factoryCode)
	}

	//------------------------------
	// call the contract (should FAIL)
	fmt.Println("\n###### CALL THE FACTORY (FAIL)")

	// A single input, having the permission, should succeed
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &contractAddr, createCode, 100, 100, 100)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, contractAddr) //
	require.Error(t, err)

	//------------------------------
	// call the contract (should PASS)
	fmt.Println("\n###### CALL THE FACTORY (PASS)")

	contractAcc.Permissions.Base.Set(permission.CreateContract, true)
	exe.stateCache.UpdateAccount(contractAcc)

	// A single input, having the permission, should succeed
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &contractAddr, createCode, 100, 100, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, contractAddr) //
	require.NoError(t, err)

	//--------------------------------
	fmt.Println("\n##### CALL to empty address")
	code := callContractCode(crypto.Address{})

	contractAddr = crypto.NewContractAddress(users[0].GetAddress(), []byte{110})
	contractAcc = &acm.Account{
		Address:     contractAddr,
		Balance:     1000,
		EVMCode:     code,
		Sequence:    0,
		Permissions: permission.ZeroAccountPermissions,
	}
	contractAcc.Permissions.Base.Set(permission.Call, true)
	contractAcc.Permissions.Base.Set(permission.CreateContract, true)
	exe.stateCache.UpdateAccount(contractAcc)

	// this should call the 0 address but not create ...
	tx, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &contractAddr, createCode, 100, 10000, 100)
	txEnv = txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txEnv, crypto.Address{}) //
	require.NoError(t, err)
	zeroAcc := getAccount(t, exe.stateCache, crypto.Address{})
	if len(zeroAcc.EVMCode) != 0 {
		t.Fatal("the zero account was given code from a CALL!")
	}
}

func TestCreateAccountPermission(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Send, true)          // give the 0 account permission
	genDoc.Accounts[1].Permissions.Base.Set(permission.Send, true)          // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.CreateAccount, true) // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Input, true)         // give the 0 account permission
	genDoc.Accounts[1].Permissions.Base.Set(permission.Input, true)         // give the 0 account permission
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//----------------------------------------------------------
	// SendTx to unknown account

	// A single input, having the permission, should succeed
	tx := payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[6].GetAddress(), 5)
	err = exe.signExecuteCommit(tx, users[0])
	require.NoError(t, err)

	// Two inputs, both with send, one with create, one without, should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(exe.stateCache, users[1].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].GetAddress(), 10)
	err = exe.signExecuteCommit(tx, users[:2]...)
	require.Error(t, err)

	// Two inputs, both with send, one with create, one without, two ouputs (one known, one unknown) should fail
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(exe.stateCache, users[1].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].GetAddress(), 4)
	tx.AddOutput(users[4].GetAddress(), 6)
	err = exe.signExecuteCommit(tx, users[:2]...)
	require.Error(t, err)

	// Two inputs, both with send, both with create, should pass
	acc := getAccount(t, exe.stateCache, users[1].GetAddress())
	acc.Permissions.Base.Set(permission.CreateAccount, true)
	exe.stateCache.UpdateAccount(acc)
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(exe.stateCache, users[1].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].GetAddress(), 10)
	err = exe.signExecuteCommit(tx, users[:2]...)
	require.NoError(t, err)

	// Two inputs, both with send, both with create, two outputs (one known, one unknown) should pass
	tx = payload.NewSendTx()
	if err := tx.AddInput(exe.stateCache, users[0].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddInput(exe.stateCache, users[1].GetPublicKey(), 5); err != nil {
		t.Fatal(err)
	}
	tx.AddOutput(users[7].GetAddress(), 7)
	tx.AddOutput(users[4].GetAddress(), 3)
	err = exe.signExecuteCommit(tx, users[:2]...)
	require.NoError(t, err)

	//----------------------------------------------------------
	// CALL to unknown account

	acc = getAccount(t, exe.stateCache, users[0].GetAddress())
	acc.Permissions.Base.Set(permission.Call, true)
	err = exe.stateCache.UpdateAccount(acc)
	require.NoError(t, err)

	// call to contract that calls unknown account - without create_account perm
	// create contract that calls the simple contract
	contractCode := callContractCode(users[9].GetAddress())
	caller1ContractAddr := crypto.NewContractAddress(users[4].GetAddress(), []byte{101})
	caller1Acc := &acm.Account{
		Address:     caller1ContractAddr,
		Balance:     0,
		EVMCode:     contractCode,
		Sequence:    0,
		Permissions: permission.ZeroAccountPermissions,
	}
	err = exe.stateCache.UpdateAccount(caller1Acc)
	require.NoError(t, err)

	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ := payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCallEnv := txs.Enclose(testChainID, txCall)
	err = txCallEnv.Sign(users[0])
	require.NoError(t, err)

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txCallEnv, caller1ContractAddr) //
	require.Error(t, err)

	// NOTE: for a contract to be able to CreateAccount, it must be able to call
	// NOTE: for a users to be able to CreateAccount, it must be able to send!
	caller1Acc.Permissions.Base.Set(permission.CreateAccount, true)
	caller1Acc.Permissions.Base.Set(permission.Call, true)
	err = exe.stateCache.UpdateAccount(caller1Acc)
	require.NoError(t, err)
	// A single input, having the permission, but the contract doesn't have permission
	txCall, _ = payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &caller1ContractAddr, nil, 100, 10000, 100)
	txCallEnv = txs.Enclose(testChainID, txCall)
	err = txCallEnv.Sign(users[0])
	require.NoError(t, err)

	// we need to subscribe to the Call event to detect the exception
	_, err = execTxWaitAccountCall(t, exe, txCallEnv, caller1ContractAddr) //
	require.NoError(t, err)

}

// holla at my boy
var DougAddress crypto.Address

func init() {
	copy(DougAddress[:], "THISISDOUG")
}

func TestSNativeCALL(t *testing.T) {
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true)  // give the 0 account permission
	genDoc.Accounts[0].Permissions.Base.Set(permission.Input, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(permission.Bond, true)  // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")

	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	exe := makeExecutor(st)

	//----------------------------------------------------------
	// Test CALL to SNative contracts

	// make the main contract once
	doug := &acm.Account{
		Address:     DougAddress,
		Balance:     0,
		EVMCode:     nil,
		Sequence:    0,
		Permissions: permission.ZeroAccountPermissions,
	}

	doug.Permissions.Base.Set(permission.Call, true)
	//doug.Permissions.Base.Set(permission.HasBase, true)
	exe.updateAccounts(t, doug)

	fmt.Println("\n#### HasBase")
	// HasBase
	snativeAddress, pF, data := snativePermTestInputCALL("hasBase", users[3], permission.Bond, false)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Printf("Doug: %s\n", exe.permString(t, users[3].GetAddress()))
	fmt.Println("\n#### SetBase")
	// SetBase

	snativeAddress, pF, data = snativePermTestInputCALL("setBase", users[3], permission.Bond, false)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })

	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.Bond, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})
	snativeAddress, pF, data = snativePermTestInputCALL("setBase", users[3], permission.CreateContract, true)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Printf("Doug: %s\n", exe.permString(t, users[3].GetAddress()))
	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeAddress, pF, data = snativePermTestInputCALL("unsetBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		if !IsZeros(ret) {
			return fmt.Errorf("Expected 0. Got %X", ret)
		}
		return nil
	})

	fmt.Printf("Doug: %s\n", exe.permString(t, users[3].GetAddress()))
	fmt.Printf("Global: %s\n", exe.permString(t, acm.GlobalPermissionsAddress))
	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeAddress, pF, data = snativePermTestInputCALL("setGlobal", users[3], permission.CreateContract, true)
	testSNativeCALLExpectFail(t, exe, doug, snativeAddress, data)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error { return nil })
	snativeAddress, pF, data = snativePermTestInputCALL("hasBase", users[3], permission.CreateContract, false)
	testSNativeCALLExpectPass(t, exe, doug, pF, snativeAddress, data, func(ret []byte) error {
		// return value should be true or false as a 32 byte array...
		if !IsZeros(ret[:31]) || ret[31] != byte(1) {
			return fmt.Errorf("Expected 1. Got %X", ret)
		}
		return nil
	})

	fmt.Printf("Doug: %s\n", exe.permString(t, users[3].GetAddress()))
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
	stateDB, err := dbm.NewDB("state", dbBackend, dbDir)
	require.NoError(t, err)
	defer stateDB.Close()
	genDoc := newBaseGenDoc(permission.ZeroAccountPermissions, permission.ZeroAccountPermissions)
	genDoc.Accounts[0].Permissions.Base.Set(permission.Call, true) // give the 0 account permission
	genDoc.Accounts[3].Permissions.Base.Set(permission.Bond, true) // some arbitrary permission to play with
	genDoc.Accounts[3].Permissions.AddRole("bumble")
	genDoc.Accounts[3].Permissions.AddRole("bee")
	st, err := state.MakeGenesisState(stateDB, &genDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)
	batchCommitter := makeExecutor(st)

	//----------------------------------------------------------
	// Test SNativeTx

	fmt.Println("\n#### SetBase")
	// SetBase
	snativeArgs := snativePermTestInputTx("setBase", users[3], permission.Bond, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetBase, snativeArgs)
	acc := getAccount(t, batchCommitter.stateCache, users[3].GetAddress())
	if v, _ := acc.Permissions.Base.Get(permission.Bond); v {
		t.Fatal("expected permission to be set false")
	}
	snativeArgs = snativePermTestInputTx("setBase", users[3], permission.CreateContract, true)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetBase, snativeArgs)
	acc = getAccount(t, batchCommitter.stateCache, users[3].GetAddress())
	if v, _ := acc.Permissions.Base.Get(permission.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### UnsetBase")
	// UnsetBase
	snativeArgs = snativePermTestInputTx("unsetBase", users[3], permission.CreateContract, false)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.UnsetBase, snativeArgs)
	acc = getAccount(t, batchCommitter.stateCache, users[3].GetAddress())
	if v, _ := acc.Permissions.Base.Get(permission.CreateContract); v {
		t.Fatal("expected permission to be set false")
	}

	fmt.Println("\n#### SetGlobal")
	// SetGlobalPerm
	snativeArgs = snativePermTestInputTx("setGlobal", users[3], permission.CreateContract, true)
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.SetGlobal, snativeArgs)
	acc = getAccount(t, batchCommitter.stateCache, acm.GlobalPermissionsAddress)
	if v, _ := acc.Permissions.Base.Get(permission.CreateContract); !v {
		t.Fatal("expected permission to be set true")
	}

	fmt.Println("\n#### AddRole")
	// AddRole
	snativeArgs = snativeRoleTestInputTx("addRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.AddRole, snativeArgs)
	acc = getAccount(t, batchCommitter.stateCache, users[3].GetAddress())
	if v := acc.Permissions.HasRole("chuck"); !v {
		t.Fatal("expected role to be added")
	}

	fmt.Println("\n#### RemoveRole")
	// RemoveRole
	snativeArgs = snativeRoleTestInputTx("removeRole", users[3], "chuck")
	testSNativeTxExpectFail(t, batchCommitter, snativeArgs)
	testSNativeTxExpectPass(t, batchCommitter, permission.RemoveRole, snativeArgs)
	acc = getAccount(t, batchCommitter.stateCache, users[3].GetAddress())
	if v := acc.Permissions.HasRole("chuck"); v {
		t.Fatal("expected role to be removed")
	}
}

func TestTxSequence(t *testing.T) {
	st, privAccounts := makeGenesisState(3, 1)
	acc0 := getAccount(t, st, privAccounts[0].GetAddress())
	acc0PubKey := privAccounts[0].GetPublicKey()
	acc1 := getAccount(t, st, privAccounts[1].GetAddress())

	// Test a variety of sequence numbers for the tx.
	// The tx should only pass when i == 1.
	for i := uint64(0); i < 3; i++ {
		sequence := acc0.Sequence + i
		tx := payload.NewSendTx()
		tx.AddInputWithSequence(acc0PubKey, 1, sequence)
		tx.AddOutput(acc1.Address, 1)

		exe := makeExecutor(copyState(t, st))
		err := exe.signExecuteCommit(tx, privAccounts[0])
		if i == 1 {
			// Sequence is good.
			if err != nil {
				t.Errorf("Expected good sequence to pass: %v", err)
			}
			// Check acc.Sequence.
			newAcc0 := getAccount(t, exe.state, acc0.Address)
			if newAcc0.Sequence != sequence {
				t.Errorf("Expected account sequence to change to %v, got %v",
					sequence, newAcc0.Sequence)
			}
		} else {
			// Sequence is bad.
			if err == nil {
				t.Errorf("Expected bad sequence to fail")
			}
			// Check acc.Sequence. (shouldn't have changed)
			newAcc0 := getAccount(t, exe.state, acc0.Address)
			if newAcc0.Sequence != acc0.Sequence {
				t.Errorf("Expected account sequence to not change from %v, got %v",
					acc0.Sequence, newAcc0.Sequence)
			}
		}
	}
}

func TestNameTxs(t *testing.T) {
	st, err := state.MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	require.NoError(t, err)
	err = st.InitialCommit()
	require.NoError(t, err)

	names.MinNameRegistrationPeriod = 5
	exe := makeExecutor(st)
	startingBlock := exe.Blockchain.LastBlockHeight()

	// try some bad names. these should all fail
	nameStrings := []string{"", "\n", "123#$%", "\x00", string([]byte{20, 40, 60, 80}),
		"baffledbythespectacleinallofthisyouseeehesaidwithouteyessurprised", "no spaces please"}
	data := "something about all this just doesn't feel right."
	fee := uint64(1000)
	numDesiredBlocks := uint64(5)
	for _, name := range nameStrings {
		amt := fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*
			names.NameBaseCost(name, data)
		tx, _ := payload.NewNameTx(st, testPrivAccounts[0].GetPublicKey(), name, data, amt, fee)

		if err = exe.signExecuteCommit(tx, testPrivAccounts[0]); err == nil {
			t.Fatalf("Expected invalid name error from %s", name)
		}
	}

	// try some bad data. these should all fail
	name := "hold_it_chum"
	datas := []string{"cold&warm", "!@#$%^&*()", "<<<>>>>", "because why would you ever need a ~ or a & or even a % in a json file? make your case and we'll talk"}
	for _, data := range datas {
		amt := fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*
			names.NameBaseCost(name, data)
		tx, _ := payload.NewNameTx(st, testPrivAccounts[0].GetPublicKey(), name, data, amt, fee)

		if err = exe.signExecuteCommit(tx, testPrivAccounts[0]); err == nil {
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
	tx, _ := payload.NewNameTx(st, testPrivAccounts[0].GetPublicKey(), name, data, amt, fee)

	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[0]))

	entry, err := st.GetName(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].GetAddress(), startingBlock+numDesiredBlocks)

	// fail to update it as non-owner, in same block
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].GetPublicKey(), name, data, amt, fee)
	require.Error(t, exe.signExecuteCommit(tx, testPrivAccounts[1]))

	// update it as owner, just to increase expiry, in same block
	// NOTE: we have to resend the data or it will clear it (is this what we want?)
	tx, _ = payload.NewNameTx(st, testPrivAccounts[0].GetPublicKey(), name, data, amt, fee)

	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[0]))

	entry, err = st.GetName(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].GetAddress(), startingBlock+numDesiredBlocks*2)

	// update it as owner, just to increase expiry, in next block
	tx, _ = payload.NewNameTx(st, testPrivAccounts[0].GetPublicKey(), name, data, amt, fee)
	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[0]))

	entry, err = st.GetName(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].GetAddress(), startingBlock+numDesiredBlocks*3)

	// fail to update it as non-owner
	// Fast forward
	for exe.Blockchain.LastBlockHeight() < entry.Expires-1 {
		_, err = exe.Commit(nil)
		require.NoError(t, err)
	}
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].GetPublicKey(), name, data, amt, fee)
	require.Error(t, exe.signExecuteCommit(tx, testPrivAccounts[1]))
	_, err = exe.Commit(nil)
	require.NoError(t, err)

	// once expires, non-owner succeeds
	startingBlock = exe.Blockchain.LastBlockHeight()
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].GetPublicKey(), name, data, amt, fee)
	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[1]))

	entry, err = st.GetName(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[1].GetAddress(), startingBlock+numDesiredBlocks)

	// update it as new owner, with new data (longer), but keep the expiry!
	data = "In the beginning there was no thing, not even the beginning. It hadn't been here, no there, nor for that matter anywhere, not especially because it had not to even exist, let alone to not. Nothing especially odd about that."
	oldCredit := amt - fee
	numDesiredBlocks = 10
	amt = fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*names.NameBaseCost(name, data) - oldCredit
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].GetPublicKey(), name, data, amt, fee)
	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[1]))

	entry, err = st.GetName(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[1].GetAddress(), startingBlock+numDesiredBlocks)

	// test removal
	amt = fee
	data = ""
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].GetPublicKey(), name, data, amt, fee)
	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[1]))

	entry, err = st.GetName(name)
	require.NoError(t, err)
	if entry != nil {
		t.Fatal("Expected removed entry to be nil")
	}

	// create entry by key0,
	// test removal by key1 after expiry
	startingBlock = exe.Blockchain.LastBlockHeight()
	name = "looking_good/karaoke_bar"
	data = "some data"
	amt = fee + numDesiredBlocks*names.NameByteCostMultiplier*names.NameBlockCostMultiplier*names.NameBaseCost(name, data)
	tx, _ = payload.NewNameTx(st, testPrivAccounts[0].GetPublicKey(), name, data, amt, fee)
	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[0]))

	entry, err = st.GetName(name)
	require.NoError(t, err)
	validateEntry(t, entry, name, data, testPrivAccounts[0].GetAddress(), startingBlock+numDesiredBlocks)
	// Fast forward
	for exe.Blockchain.LastBlockHeight() < entry.Expires {
		_, err = exe.Commit(nil)
		require.NoError(t, err)
	}

	amt = fee
	data = ""
	tx, _ = payload.NewNameTx(st, testPrivAccounts[1].GetPublicKey(), name, data, amt, fee)
	require.NoError(t, exe.signExecuteCommit(tx, testPrivAccounts[1]))

	entry, err = st.GetName(name)
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
var preFactoryCode = hex.MustDecodeString("60606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")
var factoryCode = hex.MustDecodeString("60606040526000357C010000000000000000000000000000000000000000000000000000000090048063EFC81A8C146037576035565B005B60426004805050606E565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B6000604051610153806100E0833901809050604051809103906000F0600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905060DD565B90566060604052610141806100126000396000F360606040526000357C0100000000000000000000000000000000000000000000000000000000900480639ED933181461003957610037565B005B61004F600480803590602001909190505061007B565B604051808273FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16815260200191505060405180910390F35B60008173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1663EFC81A8C604051817C01000000000000000000000000000000000000000000000000000000000281526004018090506020604051808303816000876161DA5A03F1156100025750505060405180519060200150600060006101000A81548173FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF02191690830217905550600060009054906101000A900473FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF16905061013C565B91905056")

func TestCreates(t *testing.T) {
	//evm.SetDebug(true)
	st, privAccounts := makeGenesisState(3, 1)

	//val0 := state.GetValidatorInfo(privValidators[0].Address)
	acc0 := getAccount(t, st, privAccounts[0].GetAddress())
	acc1 := getAccount(t, st, privAccounts[1].GetAddress())
	acc2 := getAccount(t, st, privAccounts[2].GetAddress())

	exe := makeExecutor(st)

	newAcc1 := getAccount(t, st, acc1.Address)
	newAcc1.EVMCode = preFactoryCode
	newAcc2 := getAccount(t, st, acc2.Address)
	newAcc2.EVMCode = factoryCode

	exe.updateAccounts(t, newAcc1, newAcc2)

	createData := hex.MustDecodeString("9ed93318")
	createData = append(createData, acc2.Address.Word256().Bytes()...)

	// call the pre-factory, triggering the factory to run a create
	tx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address,
			Amount:   1,
			Sequence: acc0.Sequence + 1,
		},
		Address:  addressPtr(acc1),
		GasLimit: 100000,
		Data:     createData,
	}

	require.NoError(t, exe.signExecuteCommit(tx, privAccounts[0]))

	acc1 = getAccount(t, st, acc1.Address)
	firstCreatedAddress, err := st.GetStorage(acc1.Address, LeftPadWord256(nil))
	require.NoError(t, err)
	require.NotEqual(t, Zero256, firstCreatedAddress, "should not be zero address")

	acc0 = getAccount(t, st, acc0.Address)
	// call the pre-factory, triggering the factory to run a create
	tx = &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address,
			Amount:   1,
			Sequence: tx.Input.Sequence + 1,
		},
		Address:  addressPtr(acc1),
		GasLimit: 100000,
		Data:     createData,
	}

	require.NoError(t, exe.signExecuteCommit(tx, privAccounts[0]))

	acc1 = getAccount(t, st, acc1.Address)
	secondCreatedAddress, err := st.GetStorage(acc1.Address, LeftPadWord256(nil))
	require.NoError(t, err)
	require.NotEqual(t, Zero256, secondCreatedAddress, "should not be zero address")

	if bytes.Equal(firstCreatedAddress, secondCreatedAddress) {
		t.Errorf("Multiple contracts created with the same address!")
	}
}

func TestContractSend(t *testing.T) {
	st, privAccounts := makeGenesisState(3, 1)
	/*
		contract Caller {
		   function send(address x){
			   x.send(msg.value);
		   }
		}
	*/
	callerCode := hex.MustDecodeString("60606040526000357c0100000000000000000000000000000000000000000000000000000000900480633e58c58c146037576035565b005b604b6004808035906020019091905050604d565b005b8073ffffffffffffffffffffffffffffffffffffffff16600034604051809050600060405180830381858888f19350505050505b5056")
	sendData := abi.GetFunctionID("send(address)").Bytes()

	acc0 := getAccount(t, st, privAccounts[0].GetAddress())
	acc1 := getAccount(t, st, privAccounts[1].GetAddress())
	acc2 := getAccount(t, st, privAccounts[2].GetAddress())

	newAcc1 := getAccount(t, st, acc1.Address)
	newAcc1.EVMCode = callerCode
	_, _, err := st.Update(func(up state.Updatable) error {
		return up.UpdateAccount(newAcc1)
	})
	require.NoError(t, err)

	sendAmt := uint64(10)
	acc2Balance := acc2.Balance

	// call the contract, triggering the send
	tx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address,
			Amount:   sendAmt,
			Sequence: acc0.Sequence + 1,
		},
		Address:  addressPtr(acc1),
		GasLimit: 1000,
		Data:     append(sendData, acc2.Address.Word256().Bytes()...),
	}

	exe := makeExecutor(st)
	err = exe.signExecuteCommit(tx, privAccounts[0])
	require.NoError(t, err)

	acc2 = getAccount(t, st, acc2.Address)
	require.Equal(t, sendAmt+acc2Balance, acc2.Balance, "value should be transferred")

	addressNonExistent := newAddress("nobody")

	// Send value to non-existent account
	tx = &payload.CallTx{
		Input: &payload.TxInput{
			Address:  acc0.Address,
			Amount:   sendAmt,
			Sequence: acc0.Sequence + 2,
		},
		Address:  addressPtr(acc1),
		GasLimit: 1000,
		Data:     append(sendData, addressNonExistent.Word256().Bytes()...),
	}

	err = exe.signExecuteCommit(tx, privAccounts[0])
	require.NoError(t, err)

	accNonExistent := getAccount(t, st, addressNonExistent)
	assert.Equal(t, sendAmt, accNonExistent.Balance, "value should have been transferred")
}

func TestMerklePanic(t *testing.T) {
	st, privAccounts := makeGenesisState(3, 1)

	acc0 := getAccount(t, st, privAccounts[0].GetAddress())
	acc1 := getAccount(t, st, privAccounts[1].GetAddress())

	// SendTx.
	{
		tx := &payload.SendTx{
			Inputs: []*payload.TxInput{
				{
					Address:  acc0.Address,
					Amount:   1,
					Sequence: acc0.Sequence + 1,
				},
			},
			Outputs: []*payload.TxOutput{
				{
					Address: acc1.Address,
					Amount:  1,
				},
			},
		}

		err := makeExecutor(copyState(t, st)).signExecuteCommit(tx, privAccounts[0])
		require.NoError(t, err)
	}

	// CallTx. Just runs through it and checks the transfer. See vm, rpc tests for more
	{
		stateCallTx := makeExecutor(copyState(t, st))
		newAcc1 := getAccount(t, stateCallTx, acc1.Address)
		newAcc1.EVMCode = []byte{0x60}
		err := stateCallTx.stateCache.UpdateAccount(newAcc1)
		require.NoError(t, err)
		tx := &payload.CallTx{
			Input: &payload.TxInput{
				Address:  acc0.Address,
				Amount:   1,
				Sequence: acc0.Sequence + 1,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		err = stateCallTx.signExecuteCommit(tx, privAccounts[0])
		require.NoError(t, err)
	}
}

// Ensure we are setting the origin
func TestOrigin(t *testing.T) {
	st, privAccounts := makeGenesisState(3, 1)
	exe := makeExecutor(st)

	origin := privAccounts[2]
	calleeAddress := privAccounts[1].GetAddress()

	// Set a contract that stores the origin address in storage at loc
	loc := []byte{3}
	err := engine.UpdateAccount(exe.stateCache, calleeAddress, func(acc *acm.Account) error {
		acc.EVMCode = bc.MustSplice(ORIGIN, PUSH1, loc, SSTORE)
		return nil
	})
	require.NoError(t, err)

	// Call the origin-storing contract
	tx := &payload.CallTx{
		Input: &payload.TxInput{
			Address:  origin.GetAddress(),
			Sequence: 1,
		},
		Address:  &calleeAddress,
		GasLimit: 1000,
	}

	err = exe.signExecuteCommit(tx, origin)
	require.NoError(t, err)

	// Check the correct origin value is stored
	val, err := st.GetStorage(calleeAddress, LeftPadWord256(loc))
	require.NoError(t, err)

	originAddress := LeftPadWord256(val)
	assert.Equal(t, origin.GetAddress(), crypto.AddressFromWord256(originAddress))
}

func TestTxs(t *testing.T) {
	st, privAccounts := makeGenesisState(3, 1)

	acc0 := getAccount(t, st, privAccounts[0].GetAddress())
	acc1 := getAccount(t, st, privAccounts[1].GetAddress())

	// SendTx.
	{
		stateSendTx := copyState(t, st)
		tx := &payload.SendTx{
			Inputs: []*payload.TxInput{
				{
					Address:  acc0.Address,
					Amount:   1,
					Sequence: acc0.Sequence + 1,
				},
			},
			Outputs: []*payload.TxOutput{
				{
					Address: acc1.Address,
					Amount:  1,
				},
			},
		}

		err := makeExecutor(stateSendTx).signExecuteCommit(tx, privAccounts[0])
		require.NoError(t, err)
		newAcc0 := getAccount(t, stateSendTx, acc0.Address)
		if acc0.Balance-1 != newAcc0.Balance {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance-1, newAcc0.Balance)
		}
		newAcc1 := getAccount(t, stateSendTx, acc1.Address)
		if acc1.Balance+1 != newAcc1.Balance {
			t.Errorf("Unexpected newAcc1 balance. Expected %v, got %v",
				acc1.Balance+1, newAcc1.Balance)
		}
	}

	// CallTx. Just runs through it and checks the transfer. See vm, rpc tests for more
	{
		stateCallTx := copyState(t, st)
		newAcc1 := getAccount(t, stateCallTx, acc1.Address)
		newAcc1.EVMCode = []byte{0x60}
		_, _, err := stateCallTx.Update(func(up state.Updatable) error {
			return up.UpdateAccount(newAcc1)
		})
		require.NoError(t, err)
		tx := &payload.CallTx{
			Input: &payload.TxInput{
				Address:  acc0.Address,
				Amount:   1,
				Sequence: acc0.Sequence + 1,
			},
			Address:  addressPtr(acc1),
			GasLimit: 10,
		}

		err = makeExecutor(stateCallTx).signExecuteCommit(tx, privAccounts[0])
		require.NoError(t, err)
		newAcc0 := getAccount(t, stateCallTx, acc0.Address)
		if acc0.Balance-1 != newAcc0.Balance {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance-1, newAcc0.Balance)
		}
		newAcc1 = getAccount(t, stateCallTx, acc1.Address)
		if acc1.Balance+1 != newAcc1.Balance {
			t.Errorf("Unexpected newAcc1 balance. Expected %v, got %v",
				acc1.Balance+1, newAcc1.Balance)
		}
	}

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

		stateNameTx := copyState(t, st)
		tx := &payload.NameTx{
			Input: &payload.TxInput{
				Address:  acc0.Address,
				Amount:   entryAmount,
				Sequence: acc0.Sequence + 1,
			},
			Name: entryName,
			Data: entryData,
		}

		exe := makeExecutor(stateNameTx)
		err := exe.signExecuteCommit(tx, privAccounts[0])
		require.NoError(t, err)

		newAcc0 := getAccount(t, stateNameTx, acc0.Address)
		if acc0.Balance-entryAmount != newAcc0.Balance {
			t.Errorf("Unexpected newAcc0 balance. Expected %v, got %v",
				acc0.Balance-entryAmount, newAcc0.Balance)
		}
		entry, err := stateNameTx.GetName(entryName)
		require.NoError(t, err)
		require.NotNil(t, entry, "Expected entry")
		if entry.Data != entryData {
			t.Errorf("Wrong data stored")
		}

		// test a bad string
		tx.Data = string([]byte{0, 1, 2, 3, 127, 128, 129, 200, 251})
		tx.Input.Sequence += 1
		err = exe.signExecuteCommit(tx, privAccounts[0])
		require.Error(t, err)
		if errors.AsException(err).ErrorCode() != errors.Codes.InvalidString {
			t.Errorf("Expected invalid string error. Got: %v", err)
		}
	}

}

func TestSelfDestruct(t *testing.T) {
	st, privAccounts := makeGenesisState(3, 1)

	acc0 := getAccount(t, st, privAccounts[0].GetAddress())
	acc0PubKey := privAccounts[0].GetPublicKey()
	acc1 := getAccount(t, st, privAccounts[1].GetAddress())
	acc2 := getAccount(t, st, privAccounts[2].GetAddress())
	sendingAmount, refundedBalance, oldBalance := uint64(1), acc1.Balance, acc2.Balance

	newAcc1 := getAccount(t, st, acc1.Address)

	// store 0x1 at 0x1, push an address, then self-destruct:)
	contractCode := []byte{0x60, 0x01, 0x60, 0x01, 0x55, 0x73}
	contractCode = append(contractCode, acc2.Address.Bytes()...)
	contractCode = append(contractCode, 0xff)
	newAcc1.EVMCode = contractCode
	_, _, err := st.Update(func(up state.Updatable) error {
		return up.UpdateAccount(newAcc1)
	})
	require.NoError(t, err)

	// send call tx with no data, cause self-destruct
	tx := payload.NewCallTxWithSequence(acc0PubKey, addressPtr(acc1), nil, sendingAmount, 1000, 0, acc0.Sequence+1)

	// we use cache instead of execTxWithState so we can run the tx twice
	exe := makeExecutor(st)
	err = exe.signExecuteCommit(tx, privAccounts[0])
	require.NoError(t, err)

	// if we do it again the self-destruct shouldn't happen twice and the caller should lose fee
	tx.Input.Sequence += 1
	err = exe.signExecuteCommit(tx, privAccounts[0])
	assertErrorCode(t, errors.Codes.InvalidAddress, err)

	// commit the block
	_, err = exe.Commit(nil)
	require.NoError(t, err)

	// acc2 should receive the sent funds and the contracts balance
	newAcc2 := getAccount(t, st, acc2.Address)
	newBalance := sendingAmount + refundedBalance + oldBalance
	if newAcc2.Balance != newBalance {
		t.Errorf("Unexpected newAcc2 balance. Expected %v, got %v",
			newAcc2.Balance, newBalance)
	}
	accRemoved, err := st.GetAccount(acc1.Address)
	require.NoError(t, err)
	require.Nil(t, accRemoved, "Expected account to be removed")
}

func TestPredecessorTracking(t *testing.T) {
	st, signers := makeGenesisState(3, 1)
	exe := makeExecutor(st)

	mkTx := func() *payload.SendTx {
		tx := payload.NewSendTx()
		err := tx.AddInput(st, signers[0].GetPublicKey(), 100)
		require.NoError(t, err)
		tx.AddOutput(signers[1].GetAddress(), 100)
		require.NoError(t, err)
		return tx
	}

	// Empty block - predecessor does not advance
	_, err := exe.Commit(nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), exe.block.PredecessorHeight)
	require.Equal(t, uint64(2), exe.block.Height)

	// Tx in block - predecessor becomes this block
	err = exe.signExecuteCommit(mkTx(), signers[0])
	require.NoError(t, err)
	require.Equal(t, uint64(2), exe.block.PredecessorHeight)
	require.Equal(t, uint64(3), exe.block.Height)

	//  Empty again
	_, err = exe.Commit(nil)
	require.NoError(t, err)
	require.Equal(t, uint64(2), exe.block.PredecessorHeight)
	require.Equal(t, uint64(4), exe.block.Height)

	// Non-empty - back to consecutive predecessor
	err = exe.signExecuteCommit(mkTx(), signers[0])
	require.NoError(t, err)
	require.Equal(t, uint64(4), exe.block.PredecessorHeight)
	require.Equal(t, uint64(5), exe.block.Height)
}

// Helpers

func makeUsers(n int) []acm.AddressableSigner {
	users := make([]acm.AddressableSigner, n)
	for i := 0; i < n; i++ {
		secret := "mysecret" + strconv.Itoa(i)
		users[i] = acm.GeneratePrivateAccountFromSecret(secret)
	}
	return users
}

func newBaseGenDoc(globalPerm, accountPerm permission.AccountPermissions) genesis.GenesisDoc {
	var genAccounts []genesis.Account
	for _, user := range users[:5] {
		accountPermCopy := accountPerm // Create new instance for custom overridability.
		genAccounts = append(genAccounts, genesis.Account{
			BasicAccount: genesis.BasicAccount{
				Address: user.GetAddress(),
				Amount:  1000000,
			},
			Permissions: accountPermCopy,
		})
	}

	validatorAccount := genesis.BasicAccount{
		Address:   users[0].GetPublicKey().GetAddress(),
		PublicKey: users[0].GetPublicKey(),
		Amount:    10,
	}
	return genesis.GenesisDoc{
		GenesisTime:       time.Now(),
		ChainName:         testGenesisDoc.ChainName,
		GlobalPermissions: globalPerm,
		Accounts:          genAccounts,
		Validators: []genesis.Validator{
			{
				BasicAccount: validatorAccount,
				UnbondTo:     []genesis.BasicAccount{validatorAccount},
			},
		},
	}
}

func makeGenesisState(numAccounts int, numValidators int) (*state.State, []*acm.PrivateAccount) {
	testGenesisDoc, privAccounts, _ := deterministicGenesis.GenesisDoc(numAccounts, numValidators)
	s0, err := state.MakeGenesisState(dbm.NewMemDB(), testGenesisDoc)
	if err != nil {
		panic(fmt.Errorf("could not make genesis state: %v", err))
	}
	err = s0.InitialCommit()
	if err != nil {
		panic(fmt.Errorf("could not commit genesis state: %v", err))
	}
	return s0, privAccounts
}

func getAccount(t testing.TB, accountGetter acmstate.AccountGetter, address crypto.Address) *acm.Account {
	acc, err := accountGetter.GetAccount(address)
	require.NoError(t, err)
	require.NotNil(t, acc)
	return acc
}

func addressPtr(account *acm.Account) *crypto.Address {
	if account == nil {
		return nil
	}
	accountAddresss := account.GetAddress()
	return &accountAddresss
}

type testExecutor struct {
	*executor
	*bcm.Blockchain
}

func makeExecutor(state *state.State) *testExecutor {
	testDB, err := dbm.NewDB("test", dbBackend, ".")
	if err != nil {
		panic(err)
	}
	blockchain, _, _ := bcm.LoadOrNewBlockchain(testDB, testGenesisDoc, logger)
	err = blockchain.CommitBlockAtHeight(time.Now(), []byte("hashily"), state.Hash(), HeightAtVersion(state.Version()))
	if err != nil {
		panic(err)
	}
	executor, err := newExecutor("makeExecutorCache", true, ParamsFromGenesis(testGenesisDoc), state,
		blockchain, nil, logger)
	if err != nil {
		panic(err)
	}
	return &testExecutor{
		Blockchain: blockchain,
		executor:   executor,
	}
}

func copyState(t testing.TB, st *state.State) *state.State {
	cpy, err := st.Copy(dbm.NewMemDB())
	require.NoError(t, err)
	return cpy
}

func (te *testExecutor) Commit(header *tmproto.Header) ([]byte, error) {
	appHash, err := te.executor.Commit(header)
	if err != nil {
		return nil, err
	}
	err = te.Blockchain.CommitBlock(time.Now(), nil, appHash)
	if err != nil {
		return nil, err
	}
	return appHash, nil
}

func (te *testExecutor) getAccount(t *testing.T, address crypto.Address) *acm.Account {
	acc, err := te.GetAccount(address)
	require.NoError(t, err)
	return acc
}

func (te *testExecutor) permString(t *testing.T, address crypto.Address) string {
	acc := te.getAccount(t, address)
	return fmt.Sprintf("%v Perms: %v Set: %v \n", address,
		permission.PermFlagToStringList(acc.Permissions.Base.Perms),
		permission.PermFlagToStringList(acc.Permissions.Base.SetBit))
}

func (te *testExecutor) updateAccounts(t *testing.T, accounts ...*acm.Account) {
	for _, acc := range accounts {
		err := te.stateCache.UpdateAccount(acc)
		require.NoError(t, err)
		_, err = te.Commit(nil)
		require.NoError(t, err)
	}
}

func getTxHash(tx payload.Payload) []byte {
	txEnv := txs.Enclose(testChainID, tx)
	return txEnv.Tx.Hash()
}

func (te *testExecutor) signExecuteCommit(tx payload.Payload, signers ...acm.AddressableSigner) error {
	txEnv := txs.Enclose(testChainID, tx)
	err := txEnv.Sign(signers...)
	if err != nil {
		return err
	}
	txe, err := te.Execute(txEnv)
	if err != nil {
		return err
	}
	if txe.Exception != nil {
		return txe.Exception
	}
	_, err = te.Commit(nil)
	return err
}

// run ExecTx and wait for the Call event on given addr
// returns the msg data and an error/exception
func execTxWaitAccountCall(t *testing.T, exe *testExecutor, txEnv *txs.Envelope,
	address crypto.Address) (*exec.CallEvent, error) {

	qry, err := query.NewBuilder().
		AndEquals(event.EventIDKey, exec.EventStringAccountCall(address)).
		AndEquals(event.TxHashKey, txEnv.Tx.Hash()).Query()
	if err != nil {
		return nil, err
	}
	txe, err := exe.Execute(txEnv)
	if err != nil {
		return nil, err
	}
	_, err = exe.Commit(nil)
	require.NoError(t, err)

	for _, ev := range txe.TaggedEvents().Filter(qry) {
		if ev.Call != nil {
			return ev.Call, ev.Header.Exception.AsError()
		}
	}
	return nil, fmt.Errorf("did not receive EventDataCall execution event from execTxWaitAccountCall")
}

// give a contract perms for an snative, call it, it calls the snative, but shouldn't have permission
func testSNativeCALLExpectFail(t *testing.T, exe *testExecutor, doug *acm.Account, snativeAddress crypto.Address,
	data []byte) {
	testSNativeCALL(t, false, exe, doug, 0, snativeAddress, data, nil)
}

// give a contract perms for an snative, call it, it calls the snative, ensure the check function (f) succeeds
func testSNativeCALLExpectPass(t *testing.T, exe *testExecutor, doug *acm.Account, snativePerm permission.PermFlag,
	snativeAddress crypto.Address, data []byte, f func([]byte) error) {
	testSNativeCALL(t, true, exe, doug, snativePerm, snativeAddress, data, f)
}

func testSNativeCALL(t *testing.T, expectPass bool, exe *testExecutor, doug *acm.Account,
	snativePerm permission.PermFlag, snativeAddress crypto.Address, data []byte, f func([]byte) error) {
	if expectPass {
		doug.Permissions.Base.Set(snativePerm, true)
	}

	doug.EVMCode = callContractCode(snativeAddress)
	dougAddress := doug.Address

	exe.updateAccounts(t, doug)
	tx, _ := payload.NewCallTx(exe.stateCache, users[0].GetPublicKey(), &dougAddress, data, 100, 10000, 100)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	ev, err := execTxWaitAccountCall(t, exe, txEnv, snativeAddress)
	if expectPass {
		require.NoError(t, err)
		ret := ev.Return
		require.NoError(t, f(ret))
	} else {
		require.Error(t, err)
	}
}

func testSNativeTxExpectFail(t *testing.T, batchCommitter *testExecutor, snativeArgs permission.PermArgs) {
	testSNativeTx(t, false, batchCommitter, 0, snativeArgs)
}

func testSNativeTxExpectPass(t *testing.T, batchCommitter *testExecutor, perm permission.PermFlag,
	snativeArgs permission.PermArgs) {
	testSNativeTx(t, true, batchCommitter, perm, snativeArgs)
}

func testSNativeTx(t *testing.T, expectPass bool, batchCommitter *testExecutor, perm permission.PermFlag,
	snativeArgs permission.PermArgs) {
	acc := getAccount(t, batchCommitter.stateCache, users[0].GetAddress())
	if expectPass {
		acc.Permissions.Base.Set(perm, true)
	}
	acc.Permissions.Base.Set(permission.Input, true)
	batchCommitter.stateCache.UpdateAccount(acc)
	tx, _ := payload.NewPermsTx(batchCommitter.stateCache, users[0].GetPublicKey(), snativeArgs)
	txEnv := txs.Enclose(testChainID, tx)
	require.NoError(t, txEnv.Sign(users[0]))
	_, err := batchCommitter.Execute(txEnv)
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
	function := permissionsContract.FunctionByName(name)
	if function == nil {
		panic(fmt.Errorf("could not find permission function %s", name))
	}
	id := function.Abi().FunctionID
	return id[:]
}

func snativePermTestInputCALL(name string, user acm.AddressableSigner, perm permission.PermFlag,
	val bool) (addr crypto.Address, pF permission.PermFlag, data []byte) {
	addr = native.Permissions.GetByName("Permissions").Address()
	switch name {
	case "hasBase", "unsetBase":
		data = user.GetAddress().Word256().Bytes()
		data = append(data, Uint64ToWord256(uint64(perm)).Bytes()...)
	case "setBase":
		data = user.GetAddress().Word256().Bytes()
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

func snativePermTestInputTx(name string, user acm.AddressableSigner, perm permission.PermFlag,
	val bool) (snativeArgs permission.PermArgs) {

	switch name {
	case "hasBase":
		snativeArgs = permission.HasBaseArgs(user.GetAddress(), perm)
	case "unsetBase":
		snativeArgs = permission.UnsetBaseArgs(user.GetAddress(), perm)
	case "setBase":
		snativeArgs = permission.SetBaseArgs(user.GetAddress(), perm, val)
	case "setGlobal":
		snativeArgs = permission.SetGlobalArgs(perm, val)
	}
	return
}

func snativeRoleTestInputCALL(name string, user acm.AddressableSigner,
	role string) (addr crypto.Address, pF permission.PermFlag, data []byte) {
	addr = permissionsContract.Address()
	data = user.GetAddress().Word256().Bytes()
	data = append(data, LeftPadBytes([]byte{0x40}, 32)...)
	data = append(data, LeftPadBytes([]byte{byte(len(role))}, 32)...)
	data = append(data, RightPadBytes([]byte(role), 32)...)
	data = append(permNameToFuncID(name), data...)

	var err error
	if pF, err = permission.PermStringToFlag(name); err != nil {
		panic(fmt.Sprintf("failed to convert perm string (%s) to flag", name))
	}
	return
}

func snativeRoleTestInputTx(name string, user acm.AddressableSigner, role string) (snativeArgs permission.PermArgs) {
	switch name {
	case "hasRole":
		snativeArgs = permission.HasRoleArgs(user.GetAddress(), role)
	case "addRole":
		snativeArgs = permission.AddRoleArgs(user.GetAddress(), role)
	case "removeRole":
		snativeArgs = permission.RemoveRoleArgs(user.GetAddress(), role)
	}
	return
}

// convenience function for contract that calls a given address
func callContractCode(contractAddr crypto.Address) []byte {
	// calldatacopy into mem and use as input to call
	memOff, inputOff := byte(0x0), byte(0x0)
	value := byte(0x0)
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

func assertErrorCode(t *testing.T, expectedCode *errors.Code, err error, msgAndArgs ...interface{}) {
	if assert.Error(t, err, msgAndArgs...) {
		actualCode := errors.AsException(err).ErrorCode()
		if !assert.Equal(t, expectedCode, actualCode, "expected error code %v", expectedCode) {
			t.Logf("Expected '%v' but got '%v'", expectedCode, actualCode)
		}
	}
}

func newAddress(name string) crypto.Address {
	hasher := ripemd160.New()
	hasher.Write([]byte(name))
	return crypto.MustAddressFromBytes(hasher.Sum(nil))
}
