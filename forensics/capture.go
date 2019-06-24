package forensics

import (
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/pkg/errors"
)

type ReplayCapture struct {
	Height        uint64
	AppHashBefore binary.HexBytes
	AppHashAfter  binary.HexBytes
	TxExecutions  []*exec.TxExecution
}

func (rc *ReplayCapture) String() string {
	return fmt.Sprintf("ReplayCapture[Height %d; AppHash: %v -> %v]",
		rc.Height, rc.AppHashBefore, rc.AppHashAfter)
}

// Compare the app hashes of two block replays
func (exp *ReplayCapture) Compare(act *ReplayCapture) error {
	if exp.AppHashBefore.String() != act.AppHashBefore.String() {
		return fmt.Errorf("app hashes before do not match")
	} else if exp.AppHashAfter.String() != act.AppHashAfter.String() {
		return fmt.Errorf("app hashes after do not match")
	}

	return nil
}

// CompareCaptures of two independent replays
func CompareCaptures(exp, act []*ReplayCapture) (uint64, error) {
	for i, rc := range exp {
		if err := rc.Compare(act[i]); err != nil {
			return rc.Height, errors.Wrapf(err, "mismatch at height %d", rc.Height)
		}
	}
	return 0, nil
}
