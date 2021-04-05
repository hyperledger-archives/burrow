package contexts

import (
	"testing"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/permission"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/stretchr/testify/require"
)

func TestPermissionsContext(t *testing.T) {
	accountState := acmstate.NewMemoryState()

	originPrivKey := newPrivKey(t)
	originAccount := newAccountFromPrivKey(originPrivKey)
	originAccount.Permissions.Base = permission.AllAccountPermissions.GetBase()

	targetPrivKey := newPrivKey(t)
	targetAccount := newAccountFromPrivKey(targetPrivKey)

	ctx := &PermissionsContext{
		State:  accountState,
		Logger: logging.NewNoopLogger(),
	}

	callTx := &payload.CallTx{}
	err := ctx.Execute(execFromTx(callTx), callTx)
	require.Error(t, err, "should not continue with incorrect payload")

	permsTx := &payload.PermsTx{
		Input: &payload.TxInput{
			Address: originAccount.Address,
		},
	}

	err = ctx.Execute(execFromTx(permsTx), permsTx)
	require.Error(t, err, "account should not exist")

	accountState.Accounts[originAccount.Address] = originAccount
	accountState.Accounts[targetAccount.Address] = targetAccount

	value := true
	tests := []struct {
		args permission.PermArgs
		exp  func(t *testing.T, err error)
	}{
		{
			args: permission.PermArgs{
				Action:     1337,
				Target:     &targetAccount.Address,
				Permission: ptrPermFlag(permission.SetBase),
				Value:      &value,
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.Error(t, err, "should error with unknown action")
			}),
		},
		{
			args: permission.PermArgs{
				Action:     permission.SetBase,
				Target:     &targetAccount.Address,
				Permission: ptrPermFlag(permission.SetBase),
				Value:      &value,
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.NoError(t, err)
			}),
		},
		{
			args: permission.PermArgs{
				Action:     permission.UnsetBase,
				Target:     &targetAccount.Address,
				Permission: ptrPermFlag(permission.UnsetBase),
				Value:      &value,
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.NoError(t, err)
			}),
		},
		{
			args: permission.PermArgs{
				Action:     permission.AddRole,
				Target:     &targetAccount.Address,
				Permission: ptrPermFlag(permission.AddRole),
				Value:      &value,
				Role:       ptrRoleString(permission.BondString),
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.NoError(t, err)
			}),
		},
		{
			args: permission.PermArgs{
				Action:     permission.RemoveRole,
				Target:     &targetAccount.Address,
				Permission: ptrPermFlag(permission.RemoveRole),
				Value:      &value,
				Role:       ptrRoleString(permission.BondString),
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.NoError(t, err)
			}),
		},
		{
			args: permission.PermArgs{
				Action:     permission.RemoveRole,
				Target:     &targetAccount.Address,
				Permission: ptrPermFlag(permission.RemoveRole),
				Value:      &value,
				Role:       ptrRoleString(permission.BondString),
			},
			exp: errCallback(func(t *testing.T, err error) {
				require.Error(t, err, "can't remove role that isn't set")
			}),
		},
	}

	for _, tt := range tests {
		permsTx.PermArgs = tt.args
		err = ctx.Execute(execFromTx(permsTx), permsTx)
		tt.exp(t, err)
	}
}

func ptrPermFlag(flag permission.PermFlag) *permission.PermFlag {
	return &flag
}

func ptrRoleString(role string) *string {
	return &role
}

func errCallback(condition func(t *testing.T, err error)) func(t *testing.T, err error) {
	return func(t *testing.T, err error) {
		condition(t, err)
	}
}
