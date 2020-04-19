package contexts

import (
	"testing"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
)

func TestSendContext(t *testing.T) {
	accountState := acmstate.NewMemoryState()

	originPrivKey := newPrivKey(t)
	originAccount := newAccountFromPrivKey(originPrivKey)

	targetPrivKey := newPrivKey(t)
	targetAccount := newAccountFromPrivKey(targetPrivKey)

	ctx := &SendContext{
		State:  accountState,
		Logger: logging.NewNoopLogger(),
	}

	callTx := &payload.CallTx{}
	err := ctx.Execute(execFromTx(callTx), callTx)
	require.Error(t, err, "should not continue with incorrect payload")

	accountState.Accounts[originAccount.Address] = originAccount
	accountState.Accounts[targetAccount.Address] = targetAccount

	tests := []struct {
		tx  *payload.SendTx
		exp func(t *testing.T, err error)
	}{
		{
			tx: &payload.SendTx{
				Inputs: []*payload.TxInput{
					&payload.TxInput{
						Address: originAccount.Address,
					},
				},
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.Error(t, err, "should not allow zero payment")
			}),
		},
		{
			tx: &payload.SendTx{
				Inputs: []*payload.TxInput{
					&payload.TxInput{
						Address: originAccount.Address,
						Amount:  100,
					},
				},
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.Error(t, err, "should not allow overpayment (i.e. inputs > outputs)")
			}),
		},
		{
			tx: &payload.SendTx{
				Inputs: []*payload.TxInput{
					&payload.TxInput{
						Address: originAccount.Address,
						Amount:  100,
					},
				},
				Outputs: []*payload.TxOutput{
					&payload.TxOutput{
						Address: originAccount.Address,
						Amount:  100,
					},
				},
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.Error(t, err, "should not allow self payment")
			}),
		},
		{
			tx: &payload.SendTx{
				Inputs: []*payload.TxInput{
					&payload.TxInput{
						Address: originAccount.Address,
						Amount:  100,
					},
				},
				Outputs: []*payload.TxOutput{
					&payload.TxOutput{
						Address: targetAccount.Address,
						Amount:  100,
					},
				},
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.NoError(t, err, "should allow payment")
			}),
		},
		{
			tx: &payload.SendTx{
				Inputs: []*payload.TxInput{
					&payload.TxInput{
						Address: originAccount.Address,
						Amount:  10000,
					},
				},
				Outputs: []*payload.TxOutput{
					&payload.TxOutput{
						Address: targetAccount.Address,
						Amount:  10000,
					},
				},
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.Error(t, err, "should not allow send with insufficient funds")
			}),
		},
	}

	for _, tt := range tests {
		err = ctx.Execute(execFromTx(tt.tx), tt.tx)
		tt.exp(t, err)
	}
}
