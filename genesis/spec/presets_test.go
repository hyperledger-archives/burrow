package spec

import (
	"testing"

	"github.com/hyperledger/burrow/keys/mock"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeGenesisSpecAccounts(t *testing.T) {
	keyClient := mock.NewMockKeyClient()
	gs := MergeGenesisSpecs(FullAccount(0), ParticipantAccount(1), ParticipantAccount(2))
	gd, err := gs.GenesisDoc(keyClient)
	require.NoError(t, err)
	assert.Len(t, gd.Validators, 1)
	assert.Len(t, gd.Accounts, 3)
	//bs, err := gd.JSONBytes()
	//require.NoError(t, err)
	//fmt.Println(string(bs))
}

func TestMergeGenesisSpecGlobalPermissions(t *testing.T) {
	gs1 := GenesisSpec{
		GlobalPermissions: []string{permission.CreateAccountString, permission.CreateAccountString},
	}
	gs2 := GenesisSpec{
		GlobalPermissions: []string{permission.SendString, permission.CreateAccountString, permission.HasRoleString},
	}

	gsMerged := MergeGenesisSpecs(gs1, gs2)
	assert.Equal(t, []string{permission.CreateAccountString, permission.HasRoleString, permission.SendString},
		gsMerged.GlobalPermissions)
}
