package def

import (
	"regexp"

	"reflect"

	"fmt"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/deploy/def/rule"
	"github.com/hyperledger/burrow/execution/evm/abi"
)

//TODO: Interface all the jobs, determine if they should remain in definitions or get their own package

type Job struct {
	// Name of the job
	Name string `mapstructure:"name,omitempty" json:"name,omitempty" yaml:"name,omitempty" toml:"name"`
	// Not marshalled
	Result interface{} `json:"-" yaml:"-" toml:"-"`
	// For multiple values
	Variables []*abi.Variable `json:"-" yaml:"-" toml:"-"`
	// Sets/Resets the primary account to use
	Account *Account `mapstructure:"account,omitempty" json:"account,omitempty" yaml:"account,omitempty" toml:"account"`
	// Set an arbitrary value
	Set *Set `mapstructure:"set,omitempty" json:"set,omitempty" yaml:"set,omitempty" toml:"set"`
	// Run a sequence of other deploy.yamls
	Meta *Meta `mapstructure:"meta,omitempty" json:"meta,omitempty" yaml:"meta,omitempty" toml:"meta"`
	// Issue a governance transaction
	UpdateAccount *UpdateAccount `mapstructure:"update-account,omitempty" json:"update-account,omitempty" yaml:"update-account,omitempty" toml:"update-account"`
	// Contract compile and send to the chain functions
	Deploy *Deploy `mapstructure:"deploy,omitempty" json:"deploy,omitempty" yaml:"deploy,omitempty" toml:"deploy"`
	// Contract compile/build
	Build *Build `mapstructure:"build,omitempty" json:"build,omitempty" yaml:"build,omitempty" toml:"build"`
	// Send tokens from one account to another
	Send *Send `mapstructure:"send,omitempty" json:"send,omitempty" yaml:"send,omitempty" toml:"send"`
	// Utilize monax:db's native name registry to register a name
	RegisterName *RegisterName `mapstructure:"register,omitempty" json:"register,omitempty" yaml:"register,omitempty" toml:"register"`
	// Sends a transaction which will update the permissions of an account. Must be sent from an account which
	// has root permissions on the blockchain (as set by either the genesis.json or in a subsequence transaction)
	Permission *Permission `mapstructure:"permission,omitempty" json:"permission,omitempty" yaml:"permission,omitempty" toml:"permission"`
	// Sends a transaction to a contract. Will utilize monax-abi under the hood to perform all of the heavy lifting
	Call *Call `mapstructure:"call,omitempty" json:"call,omitempty" yaml:"call,omitempty" toml:"call"`
	// Wrapper for mintdump dump. WIP
	DumpState *DumpState `mapstructure:"dump-state,omitempty" json:"dump-state,omitempty" yaml:"dump-state,omitempty" toml:"dump-state"`
	// Wrapper for mintdum restore. WIP
	RestoreState *RestoreState `mapstructure:"restore-state,omitempty" json:"restore-state,omitempty" yaml:"restore-state,omitempty" toml:"restore-state"`
	// Sends a "simulated call,omitempty" to a contract. Predominantly used for accessor functions ("Getters,omitempty" within contracts)
	QueryContract *QueryContract `mapstructure:"query-contract,omitempty" json:"query-contract,omitempty" yaml:"query-contract,omitempty" toml:"query-contract"`
	// Queries information from an account.
	QueryAccount *QueryAccount `mapstructure:"query-account,omitempty" json:"query-account,omitempty" yaml:"query-account,omitempty" toml:"query-account"`
	// Queries information about a name registered with monax:db's native name registry
	QueryName *QueryName `mapstructure:"query-name,omitempty" json:"query-name,omitempty" yaml:"query-name,omitempty" toml:"query-name"`
	// Queries information about the validator set
	QueryVals *QueryVals `mapstructure:"query-vals,omitempty" json:"query-vals,omitempty" yaml:"query-vals,omitempty" toml:"query-vals"`
	// Makes and assertion (useful for testing purposes)
	Assert *Assert `mapstructure:"assert,omitempty" json:"assert,omitempty" yaml:"assert,omitempty" toml:"assert"`
}

type Payload interface {
	validation.Validatable
}

func (job *Job) Validate() error {
	payloadField, err := job.PayloadField()
	if err != nil {
		return err
	}
	return validation.ValidateStruct(job,
		validation.Field(&job.Name, validation.Required, validation.Match(regexp.MustCompile("[[:word:]]+")).
			Error("must contain word characters; alphanumeric plus underscores/hyphens")),
		validation.Field(&job.Result, rule.New(rule.IsOmitted, "internally reserved and should be removed")),
		validation.Field(&job.Variables, rule.New(rule.IsOmitted, "internally reserved and should be removed")),
		validation.Field(payloadField.Addr().Interface()),
	)
}

var payloadType = reflect.TypeOf((*Payload)(nil)).Elem()

func (job *Job) Payload() (Payload, error) {
	field, err := job.PayloadField()
	if err != nil {
		return nil, err
	}
	return field.Interface().(Payload), nil
}

// Ensures only one Job payload is set and returns a pointer to that field or an error if none or multiple
// job payload fields are set
func (job *Job) PayloadField() (reflect.Value, error) {
	rv := reflect.ValueOf(job).Elem()
	rt := rv.Type()

	payloadIndex := -1
	for i := 0; i < rt.NumField(); i++ {
		if rt.Field(i).Type.Implements(payloadType) && !rv.Field(i).IsNil() {
			if payloadIndex >= 0 {
				return reflect.Value{}, fmt.Errorf("only one Job payload field should be set, but both '%v' and '%v' are set",
					rt.Field(payloadIndex).Name, rt.Field(i).Name)
			}
			payloadIndex = i
		}
	}
	if payloadIndex == -1 {
		return reflect.Value{}, fmt.Errorf("Job has no payload, please set at least one job value")
	}

	return rv.Field(payloadIndex), nil
}
