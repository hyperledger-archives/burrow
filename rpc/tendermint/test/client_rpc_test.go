package test

import (
	"testing"

	//	ctypes "github.com/eris-ltd/eris-db/rpc/core/types"
	_ "github.com/tendermint/tendermint/config/tendermint_test"
)

// When run with `-test.short` we only run:
// TestHTTPStatus, TestHTTPBroadcast, TestJSONStatus, TestJSONBroadcast, TestWSConnect, TestWSSend

//--------------------------------------------------------------------------------
func TestHTTPStatus(t *testing.T) {
	testStatus(t, "HTTP")
}

func TestHTTPGetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testGetAccount(t, "HTTP")
}

func TestHTTPBroadcastTx(t *testing.T) {
	testBroadcastTx(t, "HTTP")
}

func TestHTTPGetStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testGetStorage(t, "HTTP")
}

func TestHTTPCallCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCallCode(t, "HTTP")
}

func TestHTTPCallContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCall(t, "HTTP")
}

func TestHTTPNameReg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testNameReg(t, "HTTP")
}

//--------------------------------------------------------------------------------
// Test the JSONRPC client

func TestJSONStatus(t *testing.T) {
	testStatus(t, "JSONRPC")
}

func TestJSONGetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testGetAccount(t, "JSONRPC")
}

func TestJSONBroadcastTx(t *testing.T) {
	testBroadcastTx(t, "JSONRPC")
}

func TestJSONGetStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testGetStorage(t, "JSONRPC")
}

func TestJSONCallCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCallCode(t, "JSONRPC")
}

func TestJSONCallContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCall(t, "JSONRPC")
}

func TestJSONNameReg(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testNameReg(t, "JSONRPC")
}
