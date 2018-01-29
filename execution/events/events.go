package events

import (
	"encoding/json"
	"fmt"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/txs"
)

func EventStringAccInput(addr acm.Address) string  { return fmt.Sprintf("Acc/%s/Input", addr) }
func EventStringAccOutput(addr acm.Address) string { return fmt.Sprintf("Acc/%s/Output", addr) }
func EventStringNameReg(name string) string        { return fmt.Sprintf("NameReg/%s", name) }
func EventStringPermissions(name string) string    { return fmt.Sprintf("Permissions/%s", name) }
func EventStringBond() string                      { return "Bond" }
func EventStringUnbond() string                    { return "Unbond" }
func EventStringRebond() string                    { return "Rebond" }

// All txs fire EventDataTx, but only CallTx might have Return or Exception
type EventDataTx struct {
	Tx        txs.Tx `json:"tx"`
	Return    []byte `json:"return"`
	Exception string `json:"exception"`
}

type eventDataTx struct {
	Tx        txs.Wrapper `json:"tx"`
	Return    []byte      `json:"return"`
	Exception string      `json:"exception"`
}

func (edTx EventDataTx) MarshalJSON() ([]byte, error) {
	model := eventDataTx{
		Tx:        txs.Wrap(edTx.Tx),
		Exception: edTx.Exception,
		Return:    edTx.Return,
	}
	return json.Marshal(model)
}

func (edTx *EventDataTx) UnmarshalJSON(data []byte) error {
	model := new(eventDataTx)
	err := json.Unmarshal(data, model)
	if err != nil {
		return err
	}
	edTx.Tx = model.Tx.Unwrap()
	edTx.Return = model.Return
	edTx.Exception = model.Exception
	return nil
}
