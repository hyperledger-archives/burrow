package def

import (
	"strings"
	"testing"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJob_Validate(t *testing.T) {
	address := acm.GeneratePrivateAccountFromSecret("blah").GetAddress()
	job := &Job{
		Result: "brian",
		// This should pass emptiness validation
		Variables: []*abi.Variable{},
		QueryAccount: &QueryAccount{
			Account: address.String(),
			Field:   "bar",
		},
	}
	err := job.Validate()
	require.Error(t, err)
	errs := strings.Split(err.Error(), ";")
	if !assert.Len(t, errs, 2, "Should have two validation error from omitted name and included result") {
		t.Logf("Validation error was: %v", err)
	}

	job = &Job{
		Name: "Any kind of job",
		Account: &Account{
			Address: address.String(),
		},
	}
	err = job.Validate()
	require.NoError(t, err)

	job.Account.Address = "blah"
	err = job.Validate()
	require.NoError(t, err)
}
