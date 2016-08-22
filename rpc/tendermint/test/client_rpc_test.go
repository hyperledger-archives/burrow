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

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testHTTPGetAccount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testGetAccount(t, "HTTP")
}

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testHTTPBroadcastTx(t *testing.T) {
	testBroadcastTx(t, "HTTP")
}

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testHTTPGetStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testGetStorage(t, "HTTP")
}

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testHTTPCallCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCallCode(t, "HTTP")
}

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testHTTPCallContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	testCall(t, "HTTP")
}

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testHTTPNameReg(t *testing.T) {
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

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testJSONGetStorage(t *testing.T) {
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

// TODO: test has been disabled and needs to be re-enabled; tracked in issue
// https://github.com/eris-ltd/eris-db/issues/238
func testJSONCallContract(t *testing.T) {
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
