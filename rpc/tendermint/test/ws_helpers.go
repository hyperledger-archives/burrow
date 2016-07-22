package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	ctypes "github.com/eris-ltd/eris-db/rpc/tendermint/core/types"
	txtypes "github.com/eris-ltd/eris-db/txs"
	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/go-events"
	client "github.com/tendermint/go-rpc/client"
	rpctypes "github.com/tendermint/go-rpc/types"
	"github.com/tendermint/go-wire"
)

//--------------------------------------------------------------------------------
// Utilities for testing the websocket service

// create a new connection
func newWSClient(t *testing.T) *client.WSClient {
	wsc := client.NewWSClient(websocketAddr, websocketEndpoint)
	if _, err := wsc.Start(); err != nil {
		t.Fatal(err)
	}
	return wsc
}

// subscribe to an event
func subscribe(t *testing.T, wsc *client.WSClient, eventid string) {
	if err := wsc.Subscribe(eventid); err != nil {
		t.Fatal(err)
	}
}

// unsubscribe from an event
func unsubscribe(t *testing.T, wsc *client.WSClient, eventid string) {
	if err := wsc.Unsubscribe(eventid); err != nil {
		t.Fatal(err)
	}
}

// wait for an event; do things that might trigger events, and check them when they are received
// the check function takes an event id and the byte slice read off the ws
func waitForEvent(t *testing.T, wsc *client.WSClient, eventid string, dieOnTimeout bool, f func(), check func(string, interface{}) error) {
	// go routine to wait for webscoket msg
	goodCh := make(chan interface{})
	errCh := make(chan error)

	// Read message
	go func() {
		var err error
	LOOP:
		for {
			select {
			case r := <-wsc.ResultsCh:
				result := new(ctypes.ErisDBResult)
				wire.ReadJSONPtr(result, r, &err)
				if err != nil {
					errCh <- err
					break LOOP
				}
				event, ok := (*result).(*ctypes.ResultEvent)
				if ok && event.Event == eventid {
					goodCh <- event.Data
					break LOOP
				}
			case err := <-wsc.ErrorsCh:
				errCh <- err
				break LOOP
			case <-wsc.Quit:
				break LOOP
			}
		}
	}()

	// do stuff (transactions)
	f()

	// wait for an event or timeout
	timeout := time.NewTimer(10 * time.Second)
	select {
	case <-timeout.C:
		if dieOnTimeout {
			wsc.Stop()
			t.Fatalf("%s event was not received in time", eventid)
		}
		// else that's great, we didn't hear the event
		// and we shouldn't have
	case eventData := <-goodCh:
		if dieOnTimeout {
			// message was received and expected
			// run the check
			if err := check(eventid, eventData); err != nil {
				t.Fatal(err) // Show the stack trace.
			}
		} else {
			wsc.Stop()
			t.Fatalf("%s event was not expected", eventid)
		}
	case err := <-errCh:
		t.Fatal(err)
		panic(err) // Show the stack trace.

	}
}

//--------------------------------------------------------------------------------

func unmarshalResponseNewBlock(b []byte) (*types.Block, error) {
	// unmarshall and assert somethings
	var response rpctypes.RPCResponse
	var err error
	wire.ReadJSON(&response, b, &err)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, fmt.Errorf(response.Error)
	}
	// TODO
	//block := response.Result.(*ctypes.ResultEvent).Data.(types.EventDataNewBlock).Block
	// return block, nil
	return nil, nil
}

func unmarshalResponseNameReg(b []byte) (*txtypes.NameTx, error) {
	// unmarshall and assert somethings
	var response rpctypes.RPCResponse
	var err error
	wire.ReadJSON(&response, b, &err)
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, fmt.Errorf(response.Error)
	}
	_, val := UnmarshalEvent(*response.Result)
	tx := txtypes.DecodeTx(val.(types.EventDataTx).Tx).(*txtypes.NameTx)
	return tx, nil
}

func unmarshalValidateBlockchain(t *testing.T, wsc *client.WSClient, eid string) {
	var initBlockN int
	for i := 0; i < 2; i++ {
		waitForEvent(t, wsc, eid, true, func() {}, func(eid string, b interface{}) error {
			block, err := unmarshalResponseNewBlock(b.([]byte))
			if err != nil {
				return err
			}
			if i == 0 {
				initBlockN = block.Header.Height
			} else {
				if block.Header.Height != initBlockN+i {
					return fmt.Errorf("Expected block %d, got block %d", i, block.Header.Height)
				}
			}

			return nil
		})
	}
}

func unmarshalValidateSend(amt int64, toAddr []byte) func(string, interface{}) error {
	return func(eid string, b interface{}) error {
		// unmarshal and assert correctness
		var response rpctypes.RPCResponse
		var err error
		wire.ReadJSON(&response, b.([]byte), &err)
		if err != nil {
			return err
		}
		if response.Error != "" {
			return fmt.Errorf(response.Error)
		}
		event, val := UnmarshalEvent(*response.Result)
		if eid != event {
			return fmt.Errorf("Eventid is not correct. Got %s, expected %s", event, eid)
		}
		tx := txtypes.DecodeTx(val.(types.EventDataTx).Tx).(*txtypes.SendTx)
		if !bytes.Equal(tx.Inputs[0].Address, user[0].Address) {
			return fmt.Errorf("Senders do not match up! Got %x, expected %x", tx.Inputs[0].Address, user[0].Address)
		}
		if tx.Inputs[0].Amount != amt {
			return fmt.Errorf("Amt does not match up! Got %d, expected %d", tx.Inputs[0].Amount, amt)
		}
		if !bytes.Equal(tx.Outputs[0].Address, toAddr) {
			return fmt.Errorf("Receivers do not match up! Got %x, expected %x", tx.Outputs[0].Address, user[0].Address)
		}
		return nil
	}
}

func unmarshalValidateTx(amt int64, returnCode []byte) func(string, interface{}) error {
	return func(eid string, b interface{}) error {
		// unmarshall and assert somethings
		var response rpctypes.RPCResponse
		var err error
		wire.ReadJSON(&response, b.([]byte), &err)
		if err != nil {
			return err
		}
		if response.Error != "" {
			return fmt.Errorf(response.Error)
		}
		_, val := UnmarshalEvent(*response.Result)
		var data = val.(txtypes.EventDataTx)
		if data.Exception != "" {
			return fmt.Errorf(data.Exception)
		}
		tx := data.Tx.(*txtypes.CallTx)
		if !bytes.Equal(tx.Input.Address, user[0].Address) {
			return fmt.Errorf("Senders do not match up! Got %x, expected %x",
				tx.Input.Address, user[0].Address)
		}
		if tx.Input.Amount != amt {
			return fmt.Errorf("Amt does not match up! Got %d, expected %d",
				tx.Input.Amount, amt)
		}
		ret := data.Return
		if !bytes.Equal(ret, returnCode) {
			return fmt.Errorf("Tx did not return correctly. Got %x, expected %x", ret, returnCode)
		}
		return nil
	}
}

func unmarshalValidateCall(origin, returnCode []byte, txid *[]byte) func(string, interface{}) error {
	return func(eid string, b interface{}) error {
		// unmarshall and assert somethings
		var response rpctypes.RPCResponse
		var err error
		wire.ReadJSON(&response, b.([]byte), &err)
		if err != nil {
			return err
		}
		if response.Error != "" {
			return fmt.Errorf(response.Error)
		}
		_, val := UnmarshalEvent(*response.Result)
		var data = val.(txtypes.EventDataCall)
		if data.Exception != "" {
			return fmt.Errorf(data.Exception)
		}
		if !bytes.Equal(data.Origin, origin) {
			return fmt.Errorf("Origin does not match up! Got %x, expected %x",
				data.Origin, origin)
		}
		ret := data.Return
		if !bytes.Equal(ret, returnCode) {
			return fmt.Errorf("Call did not return correctly. Got %x, expected %x", ret, returnCode)
		}
		if !bytes.Equal(data.TxID, *txid) {
			return fmt.Errorf("TxIDs do not match up! Got %x, expected %x",
				data.TxID, *txid)
		}
		return nil
	}
}

// Unmarshal a json event
func UnmarshalEvent(b json.RawMessage) (string, events.EventData) {
	var err error
	result := new(ctypes.ErisDBResult)
	wire.ReadJSONPtr(result, b, &err)
	if err != nil {
		panic(err)
	}
	event, ok := (*result).(*ctypes.ResultEvent)
	if !ok {
		return "", nil // TODO: handle non-event messages (ie. return from subscribe/unsubscribe)
		// fmt.Errorf("Result is not type *ctypes.ResultEvent. Got %v", reflect.TypeOf(*result))
	}
	return event.Event, event.Data
}
