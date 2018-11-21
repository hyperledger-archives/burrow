package def

import (
	"testing"

	"github.com/hyperledger/burrow/crypto"
	"github.com/stretchr/testify/require"
)

func TestPackage_Validate(t *testing.T) {
	address := crypto.Address{3, 4}.String()
	pkgs := &DeployScript{
		Jobs: []*Job{{
			Name: "CallJob",
			Call: &Call{
				Sequence:    "13",
				Destination: address,
			},
		}},
	}
	err := pkgs.Validate()
	require.NoError(t, err)

	pkgs.Jobs = append(pkgs.Jobs, &Job{
		Name: "Foo",
		Account: &Account{
			Address: address,
		},
	})
	err = pkgs.Validate()
	require.NoError(t, err)

	// cannot set two job fields
	pkgs.Jobs[1].QueryAccount = &QueryAccount{
		Account: address,
		Field:   "Foo",
	}
	err = pkgs.Validate()
	require.Error(t, err)

	pkgs = &DeployScript{
		Jobs: []*Job{{
			Name: "UpdateAccount",
			UpdateAccount: &UpdateAccount{
				Target:   address,
				Sequence: "13",
				Native:   "333",
			},
		}},
	}
	err = pkgs.Validate()
	require.NoError(t, err)
}
