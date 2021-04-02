package execution

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/solidity"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
	"golang.org/x/sync/errgroup"
)

var genesisDoc, _, _ = genesis.NewDeterministicGenesis(100).GenesisDoc(1, 1)

// This test looks at caching problems that arise when doing concurrent reads via CallSim. It requires a cold cache
// for bug to be exhibited.
// The root cause of the original bug was a race condition that could lead to reading a uninitialised tree from the
// MutableForest tree cache.
func TestCallSimDelegate(t *testing.T) {
	// Roll up our sleeves and swear fealty to the witch-king
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	db := dbm.NewMemDB()
	st, err := state.MakeGenesisState(db, genesisDoc)
	require.NoError(t, err)

	from := crypto.PrivateKeyFromSecret("raaah", crypto.CurveTypeEd25519)
	contractAddress := crypto.Address{1, 2, 3, 4, 5}
	blockchain := &bcm.Blockchain{}
	sink := exec.NewNoopEventSink()

	// Function to set storage value for later
	setDelegate := func(up state.Updatable, value crypto.Address) error {
		call, _, err := abi.EncodeFunctionCall(string(solidity.Abi_DelegateProxy), "setDelegate", logger, value)
		if err != nil {
			return err
		}

		cache := acmstate.NewCache(st)
		_, err = evm.Default().Execute(cache, blockchain, sink,
			engine.CallParams{
				CallType: exec.CallTypeCall,
				Origin:   from.GetAddress(),
				Caller:   from.GetAddress(),
				Callee:   contractAddress,
				Input:    call,
				Gas:      big.NewInt(9999999),
			}, solidity.DeployedBytecode_DelegateProxy)

		if err != nil {
			return err
		}
		return cache.Sync(up)
	}

	// Initialise state
	_, _, err = st.Update(func(up state.Updatable) error {
		err = up.UpdateAccount(&acm.Account{
			Address:     from.GetAddress(),
			PublicKey:   from.GetPublicKey(),
			Balance:     9999999,
			Permissions: permission.DefaultAccountPermissions,
		})
		if err != nil {
			return err
		}
		return up.UpdateAccount(&acm.Account{
			Address: contractAddress,
			EVMCode: solidity.DeployedBytecode_DelegateProxy,
		})
	})
	require.NoError(t, err)

	// Set a series of values of storage slot so we get a deep version tree (which we need to trigger the bug)
	delegate := crypto.Address{0xBE, 0xEF, 0, 0xFA, 0xCE, 0, 0xBA, 0}
	for i := 0; i < 0xBF; i++ {
		delegate[7] = byte(i)
		_, _, err = st.Update(func(up state.Updatable) error {
			return setDelegate(up, delegate)
		})
		require.NoError(t, err)
	}

	st, err = state.LoadState(db, st.Version())

	getIntCall, _, err := abi.EncodeFunctionCall(string(solidity.Abi_DelegateProxy), "getDelegate", logger)
	require.NoError(t, err)
	n := 1000

	for i := 0; i < n; i++ {
		g.Go(func() error {
			txe, err := CallSim(st, blockchain, from.GetAddress(), contractAddress, getIntCall, logger)
			if err != nil {
				return err
			}
			err = txe.GetException().AsError()
			if err != nil {
				return err
			}
			address, err := crypto.AddressFromBytes(txe.GetResult().Return[12:])
			if err != nil {
				return err
			}
			if address != delegate {
				return fmt.Errorf("getDelegate returned %v but expected %v", address, delegate)
			}
			return nil
		})
	}

	require.NoError(t, g.Wait())
}
