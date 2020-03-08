package contexts

import (
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
)

func TestBondContext(t *testing.T) {
	t.Run("CurveType", func(t *testing.T) {
		privKey, err := crypto.GeneratePrivateKey(nil, crypto.CurveTypeSecp256k1)
		require.NoError(t, err)
		pubKey := privKey.GetPublicKey()
		address := pubKey.GetAddress()

		accountState := acmstate.NewMemoryState()
		accountState.Accounts[address] = &acm.Account{
			Address:   address,
			PublicKey: pubKey,
			Balance:   1337,
		}

		bondContext := &BondContext{
			State:        accountState,
			ValidatorSet: validator.NewSet(),
			Logger:       logging.NewNoopLogger(),
		}

		err = bondContext.Execute(&exec.TxExecution{}, &payload.BondTx{
			Input: &payload.TxInput{
				Address: address,
				Amount:  1337,
			},
		})
		require.Error(t, err)
	})
}
