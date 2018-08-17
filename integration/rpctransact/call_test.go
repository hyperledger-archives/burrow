// +build integration

package rpctransact

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/evm/asm"
	"github.com/hyperledger/burrow/execution/evm/asm/bc"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/integration/rpctest"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallTxNoCode(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)

	// Flip flops between sending private key and input address to test private key and address based signing
	toAddress := rpctest.PrivateAccounts[2].Address()

	numCreates := 1000
	countCh := rpctest.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numCreates; i++ {
		receipt, err := cli.CallTxAsync(context.Background(), &payload.CallTx{
			Input: &payload.TxInput{
				Address: inputAddress,
				Amount:  2,
			},
			Address:  &toAddress,
			Data:     []byte{},
			Fee:      2,
			GasLimit: 10000 + uint64(i),
		})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract)
		assert.Equal(t, toAddress, receipt.ContractAddress)
	}
	require.Equal(t, numCreates, <-countCh)
}

func TestCreateContract(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	numGoroutines := 100
	numCreates := 50
	wg := new(sync.WaitGroup)
	wg.Add(numGoroutines)
	countCh := rpctest.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numCreates; j++ {
				receipt, err := cli.CallTxAsync(context.Background(), &payload.CallTx{
					Input: &payload.TxInput{
						Address: inputAddress,
						Amount:  2,
					},
					Address:  nil,
					Data:     rpctest.Bytecode_strange_loop,
					Fee:      2,
					GasLimit: 10000,
				})
				if assert.NoError(t, err) {
					assert.True(t, receipt.CreatesContract)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	require.Equal(t, numGoroutines*numCreates, <-countCh)
}

func BenchmarkCreateContract(b *testing.B) {
	cli := rpctest.NewTransactClient(b, testConfig.RPC.GRPC.ListenAddress)
	for i := 0; i < b.N; i++ {
		create, err := cli.CallTxAsync(context.Background(), &payload.CallTx{
			Input: &payload.TxInput{
				Address: inputAddress,
				Amount:  2,
			},
			Address:  nil,
			Data:     rpctest.Bytecode_strange_loop,
			Fee:      2,
			GasLimit: 10000,
		})
		require.NoError(b, err)
		assert.True(b, create.CreatesContract)
	}
}

func TestCallTxSync(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	numGoroutines := 40
	numRuns := 5
	functionID := abi.GetFunctionID("UpsieDownsie()")
	countCh := rpctest.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numRuns; j++ {
				createTxe := rpctest.CreateContract(t, cli, inputAddress, rpctest.Bytecode_strange_loop)
				callTxe := rpctest.CallContract(t, cli, inputAddress, lastCall(createTxe.Events).CallData.Callee,
					functionID[:])
				depth := binary.Uint64FromWord256(binary.LeftPadWord256(lastCall(callTxe.Events).Return))
				// Would give 23 if taken from wrong frame (i.e. not the outer stackdepth == 0 one)
				assert.Equal(t, 18, int(depth))
			}
		}()
	}
	require.Equal(t, numGoroutines*numRuns*2, <-countCh)
}

func TestSendTxAsync(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	numSends := 1000
	countCh := rpctest.CommittedTxCount(t, kern.Emitter)
	for i := 0; i < numSends; i++ {
		receipt, err := cli.SendTxAsync(context.Background(), &payload.SendTx{
			Inputs: []*payload.TxInput{{
				Address: inputAddress,
				Amount:  2003,
			}},
			Outputs: []*payload.TxOutput{{
				Address: rpctest.PrivateAccounts[3].Address(),
				Amount:  2003,
			}},
		})
		require.NoError(t, err)
		assert.False(t, receipt.CreatesContract)
	}
	require.Equal(t, numSends, <-countCh)
}

func TestCallCodeSim(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	// add two integers and return the result
	var i, j byte = 123, 21
	_, contractCode, expectedReturn := simpleContract(i, j)
	txe, err := cli.CallCodeSim(context.Background(), &rpctransact.CallCodeParam{
		FromAddress: inputAddress,
		Code:        contractCode,
	})
	require.NoError(t, err)
	assert.Equal(t, expectedReturn, txe.Result.Return)

	// pass two ints as calldata, add, and return the result
	txe, err = cli.CallCodeSim(context.Background(), &rpctransact.CallCodeParam{
		FromAddress: inputAddress,
		Code: bc.MustSplice(asm.PUSH1, 0x0, asm.CALLDATALOAD, asm.PUSH1, 0x20, asm.CALLDATALOAD, asm.ADD, asm.PUSH1,
			0x0, asm.MSTORE, asm.PUSH1, 0x20, asm.PUSH1, 0x0, asm.RETURN),
		Data: bc.MustSplice(binary.LeftPadWord256([]byte{i}), binary.LeftPadWord256([]byte{j})),
	})
	require.NoError(t, err)
	assert.Equal(t, expectedReturn, txe.Result.Return)
}

func TestCallContract(t *testing.T) {
	initCode, _, expectedReturn := simpleContract(43, 1)
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	txe, err := cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  uint64(6969),
		},
		Address:  nil,
		Data:     initCode,
		Fee:      uint64(1000),
		GasLimit: uint64(1000),
	})
	require.NoError(t, err)
	assert.Equal(t, true, txe.Receipt.CreatesContract, "This transaction should"+
		" create a contract")
	assert.NotEqual(t, 0, len(txe.TxHash), "Receipt should contain a"+
		" transaction hash")
	contractAddress := txe.Receipt.ContractAddress
	assert.NotEqual(t, 0, len(contractAddress), "Transactions claims to have"+
		" created a contract but the contract address is empty")

	txe, err = cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  uint64(6969),
		},
		Address:  &contractAddress,
		Fee:      uint64(1000),
		GasLimit: uint64(1000),
	})
	require.NoError(t, err)

	assert.Equal(t, expectedReturn, txe.Result.Return)
}

// create two contracts, one of which calls the other
func TestNestedCall(t *testing.T) {
	code, _, expectedReturn := simpleContract(5, 6)

	// Deploy callee contract
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	txe, err := cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  uint64(6969),
		},
		Data:     code,
		GasLimit: 10000,
	})
	require.NoError(t, err)
	assert.True(t, txe.Receipt.CreatesContract)
	calleeContractAddress := txe.Receipt.ContractAddress

	// Deploy caller contract
	code, _, _ = simpleCallContract(calleeContractAddress)
	txe, err = cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  uint64(6969),
		},
		Data:     code,
		GasLimit: 10000,
	})
	require.NoError(t, err)
	assert.True(t, txe.Receipt.CreatesContract)
	callerContractAddress := txe.Receipt.ContractAddress

	// Call caller contract
	txe, err = cli.CallTxSync(context.Background(), &payload.CallTx{
		Input: &payload.TxInput{
			Address: inputAddress,
			Amount:  uint64(6969),
		},
		Address:  &callerContractAddress,
		GasLimit: 10000,
	})
	require.NoError(t, err)
	assert.Equal(t, expectedReturn, txe.Result.Return)
}

func TestCallEvents(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	createTxe := rpctest.CreateContract(t, cli, inputAddress, rpctest.Bytecode_strange_loop)
	address := lastCall(createTxe.Events).CallData.Callee
	functionID := abi.GetFunctionID("UpsieDownsie()")
	callTxe := rpctest.CallContract(t, cli, inputAddress, address, functionID[:])
	callEvents := filterCalls(callTxe.Events)
	require.Len(t, callEvents, rpctest.UpsieDownsieCallCount, "should see 30 recursive call events")
	for i, ev := range callEvents {
		assert.Equal(t, uint64(rpctest.UpsieDownsieCallCount-i-1), ev.StackDepth)
	}
}

func TestLogEvents(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	createTxe := rpctest.CreateContract(t, cli, inputAddress, rpctest.Bytecode_strange_loop)
	address := lastCall(createTxe.Events).CallData.Callee
	functionID := abi.GetFunctionID("UpsieDownsie()")
	callTxe := rpctest.CallContract(t, cli, inputAddress, address, functionID[:])
	evs := filterLogs(callTxe.Events)
	require.Len(t, evs, rpctest.UpsieDownsieCallCount-2)
	log := evs[0]
	depth := binary.Int64FromWord256(log.Topics[2])
	direction := strings.TrimRight(string(log.Topics[1][:]), "\x00")
	assert.Equal(t, int64(18), depth)
	assert.Equal(t, "Upsie!", direction)
}

func TestRevert(t *testing.T) {
	cli := rpctest.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	txe := rpctest.CreateContract(t, cli, inputAddress, rpctest.Bytecode_revert)
	functionID := abi.GetFunctionID("RevertAt(uint32)")
	txe = rpctest.CallContract(t, cli, inputAddress, txe.Receipt.ContractAddress,
		bc.MustSplice(functionID, binary.Int64ToWord256(4)))
	assert.Equal(t, errors.ErrorCodeExecutionReverted, txe.Exception.Code)

	revertReason := "I have reverted"
	expectedReturn := bc.MustSplice(abi.GetFunctionID("Error(string)"), binary.Int64ToWord256(binary.Word256Length),
		binary.Int64ToWord256(int64(len(revertReason))), binary.RightPadWord256([]byte(revertReason)))
	assert.Equal(t, expectedReturn, txe.Result.Return)
}

func filterCalls(evs []*exec.Event) []*exec.CallEvent {
	var callEvs []*exec.CallEvent
	for _, ev := range evs {
		if ev.Call != nil {
			callEvs = append(callEvs, ev.Call)
		}
	}
	return callEvs
}

func filterLogs(evs []*exec.Event) []*exec.LogEvent {
	var logEvs []*exec.LogEvent
	for _, ev := range evs {
		if ev.Log != nil {
			logEvs = append(logEvs, ev.Log)
		}
	}
	return logEvs
}

func lastCall(evs []*exec.Event) *exec.CallEvent {
	callEvs := filterCalls(evs)
	return callEvs[len(callEvs)-1]
}

func TestName(t *testing.T) {
	initcode, contractcode, returncode := simpleContract(5, 6)
	fmt.Println(asm.OpCode(0x7f).Name(), initcode, contractcode, returncode)

}

// simple contract returns 5 + 6 = 0xb
func simpleContract(i, j byte) ([]byte, []byte, []byte) {
	// this is the code we want to run when the contract is called
	contractCode := bc.MustSplice(asm.PUSH1, i, asm.PUSH1, j, asm.ADD, asm.PUSH1, 0x0, asm.MSTORE, asm.PUSH1, 0x20, asm.PUSH1,
		0x0, asm.RETURN)
	// the is the code we need to return the contractCode when the contract is initialized
	lenCode := len(contractCode)
	// push code to the stack
	initCode := bc.MustSplice(asm.PUSH32, binary.RightPadWord256(contractCode),
		// store it in memory
		asm.PUSH1, 0x0, asm.MSTORE,
		// return whats in memory
		asm.PUSH1, lenCode, asm.PUSH1, 0x0, asm.RETURN)
	// return init code, contract code, expected return
	return initCode, contractCode, binary.LeftPadBytes([]byte{i + j}, 32)
}

func simpleCallContract(address crypto.Address) ([]byte, []byte, []byte) {
	gas1, gas2 := byte(0x1), byte(0x1)
	value := byte(0x1)
	inOff, inSize := byte(0x0), byte(0x0) // no call data
	retOff, retSize := byte(0x0), byte(0x20)
	// this is the code we want to run (call a contract and return)
	contractCode := []byte{0x60, retSize, 0x60, retOff, 0x60, inSize, 0x60, inOff,
		0x60, value, 0x73}
	contractCode = append(contractCode, address.Bytes()...)
	contractCode = append(contractCode, []byte{0x61, gas1, gas2, 0xf1, 0x60, 0x20,
		0x60, 0x0, 0xf3}...)

	// the is the code we need to return; the contractCode when the contract is initialized
	// it should copy the code from the input into memory
	lenCode := len(contractCode)
	memOff := byte(0x0)
	inOff = byte(0xc) // length of code before codeContract
	length := byte(lenCode)

	code := []byte{0x60, length, 0x60, inOff, 0x60, memOff, 0x37}
	// return whats in memory
	code = append(code, []byte{0x60, byte(lenCode), 0x60, 0x0, 0xf3}...)
	code = append(code, contractCode...)
	// return init code, contract code, expected return
	return code, contractCode, binary.LeftPadBytes([]byte{0xb}, 32)
}
