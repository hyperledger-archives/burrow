package erisdb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

// Filter for account code.
// Ops: == or !=
// Could be used to match against nil, to see if an account is a contract account.
type AccountCallTxHashFilter struct {
	op    string
	value []byte
	match func([]byte, []byte) bool
}

func (this *AccountCallTxHashFilter) Configure(fd *ep.FilterData) error {
	op := fd.Op
	val, err := hex.DecodeString(fd.Value)

	if err != nil {
		return fmt.Errorf("Wrong value type.")
	}
	if op == "==" {
		this.match = func(a, b []byte) bool {
			return bytes.Equal(a, b)
		}
	} else if op == "!=" {
		this.match = func(a, b []byte) bool {
			return !bytes.Equal(a, b)
		}
	} else {
		return fmt.Errorf("Op: " + this.op + " is not supported for 'code' filtering")
	}
	this.op = op
	this.value = val
	return nil
}

func (this *AccountCallTxHashFilter) Match(v interface{}) bool {
	emct, ok := v.(*types.EventMsgCall)
	if !ok {
		return false
	}
	return this.match(emct.TxID, this.value)
}