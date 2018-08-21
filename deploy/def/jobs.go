package def

import (
	"regexp"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/hyperledger/burrow/deploy/def/rule"
	"github.com/hyperledger/burrow/execution/evm/abi"
)

// ------------------------------------------------------------------------
// Meta Jobs
// ------------------------------------------------------------------------

// Used in the Target of UpdateAccount to determine whether to create a new account, e.g. new() or new(key1,ed25519)
var NewKeyRegex = regexp.MustCompile(`new\((?P<keyName>[[:alnum:]]+)?(,(?P<curveType>[[:alnum:]]+))?\)`)

func KeyNameCurveType(newKeyMatch []string) (keyName, curveType string) {
	for i, name := range NewKeyRegex.SubexpNames() {
		switch name {
		case "keyName":
			keyName = newKeyMatch[i]
		case "curveType":
			curveType = newKeyMatch[i]
		}
	}
	return
}

type Meta struct {
	// (Required) the file path of the sub yaml to run
	File string `mapstructure:"file" json:"file" yaml:"file" toml:"file"`
}

func (job *Meta) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.File, validation.Required),
	)
}

// ------------------------------------------------------------------------
// Governance Jobs
// ------------------------------------------------------------------------

type PermissionString string

func (ps PermissionString) Validate() error {
	return rule.PermissionOrPlaceholder.Validate(ps)
}

// UpdateAccount updates an account by overwriting the given values, where values are omitted the existing values
// are preserved. Currently requires Root permission on Source account
type UpdateAccount struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required) The target account that will be governed - either an address or public key (its type will be determined by it's length)
	// if altering power then either a public key must be provided or the requisite public key associated with the address
	// must be available in an connected keys Signer
	Target string `mapstructure:"target" json:"target" yaml:"target" toml:"target"`
	// (Optional) the Tendermint validator power to set for this account
	Power string `mapstructure:"power" json:"power" yaml:"power" toml:"power"`
	// (Optional) The Burrow native token balance to set for this account
	Native string `mapstructure:"native" json:"native" yaml:"native" toml:"native"`
	// (Optional) the permissions to set for this account
	Permissions []PermissionString `mapstructure:"permissions" json:"permissions" yaml:"permissions" toml:"permissions"`
	// (Optional) the account permission roles to set for this account
	Roles []string `mapstructure:"roles" json:"roles" yaml:"roles" toml:"roles"`
	// (Optional, advanced only) sequence to use when burrow keys signs the transaction (do not use unless you
	// know what you're doing)
	Sequence string `mapstructure:"sequence" json:"sequence" yaml:"sequence" toml:"sequence"`
}

func (job *UpdateAccount) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Source, rule.AddressOrPlaceholder),
		validation.Field(&job.Target, validation.Required, rule.Or(rule.Placeholder, is.Hexadecimal,
			validation.Match(NewKeyRegex))),
		validation.Field(&job.Permissions),
		validation.Field(&job.Power, rule.Uint64OrPlaceholder),
		validation.Field(&job.Native, rule.Uint64OrPlaceholder),
		validation.Field(&job.Sequence, rule.Uint64OrPlaceholder),
	)
}

// ------------------------------------------------------------------------
// Util Jobs
// ------------------------------------------------------------------------

type Account struct {
	// (Required) address of the account which should be used as the default (if source) is
	// not given for future transactions. Will make sure the burrow keys has the public key
	// for the account. Generally account should be the first job called unless it is used
	// via a flag or environment variables to establish what default to use.
	Address string `mapstructure:"address" json:"address" yaml:"address" toml:"address"`
}

func (job *Account) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Address, validation.Required, rule.AddressOrPlaceholder),
	)
}

type Set struct {
	// (Required) value which should be saved along with the jobName (which will be the key)
	// this is useful to set variables which can be used throughout the jobs definition file (deploy.yaml).
	// It should be noted that arrays and bools must be defined using strings as such "[1,2,3]"
	// if they are intended to be used further in a assert job.
	Value string `mapstructure:"val" json:"val" yaml:"val" toml:"val"`
}

func (job *Set) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Value, validation.Required),
	)
}

// ------------------------------------------------------------------------
// Transaction Jobs
// ------------------------------------------------------------------------

type Send struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required) address of the account to send the tokens
	Destination string `mapstructure:"destination" json:"destination" yaml:"destination" toml:"destination"`
	// (Required) amount of tokens to send from the `source` to the `destination`
	Amount string `mapstructure:"amount" json:"amount" yaml:"amount" toml:"amount"`
	// (Optional, advanced only) sequence to use when burrow keys signs the transaction (do not use unless you
	// know what you're doing)
	Sequence string `mapstructure:"sequence" json:"sequence" yaml:"sequence" toml:"sequence"`
}

func (job *Send) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Source, rule.AddressOrPlaceholder),
		validation.Field(&job.Destination, validation.Required, rule.AddressOrPlaceholder),
		validation.Field(&job.Amount, rule.Uint64OrPlaceholder),
		validation.Field(&job.Sequence, rule.Uint64OrPlaceholder),
	)
}

type RegisterName struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required - unless providing data file) name which will be registered
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	// (Optional, if data_file is used; otherwise required) data which will be stored at the `name` key
	Data string `mapstructure:"data" json:"data" yaml:"data" toml:"data"`
	// (Optional) csv file in the form (name,data[,amount]) which can be used to bulk register names
	DataFile string `mapstructure:"data_file" json:"data_file" yaml:"data_file" toml:"data_file"`
	// (Optional) amount of blocks which the name entry will be reserved for the registering user
	Amount string `mapstructure:"amount" json:"amount" yaml:"amount" toml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" json:"fee" yaml:"fee" toml:"fee"`
	// (Optional, advanced only) sequence to use when burrow keys signs the transaction (do not use unless you
	// know what you're doing)
	Sequence string `mapstructure:"sequence" json:"sequence" yaml:"sequence" toml:"sequence"`
}

func (job *RegisterName) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Source, rule.AddressOrPlaceholder),
		validation.Field(&job.Amount, rule.Uint64OrPlaceholder),
		validation.Field(&job.Fee, rule.Uint64OrPlaceholder),
		validation.Field(&job.Sequence, rule.Uint64OrPlaceholder),
	)
}

type Permission struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required) actions must be in the set ["set_base", "unset_base", "set_global", "add_role" "rm_role"]
	Action string `mapstructure:"action" json:"action" yaml:"action" toml:"action"`
	// (Required, unless add_role or rm_role action selected) the name of the permission flag which is to
	// be updated
	Permission string `mapstructure:"permission" json:"permission" yaml:"permission" toml:"permission"`
	// (Required) the value of the permission or role which is to be updated
	Value string `mapstructure:"value" json:"value" yaml:"value" toml:"value"`
	// (Required) the target account which is to be updated
	Target string `mapstructure:"target" json:"target" yaml:"target" toml:"target"`
	// (Required, if add_role or rm_role action selected) the role which should be given to the account
	Role string `mapstructure:"role" json:"role" yaml:"role" toml:"role"`
	// (Optional, advanced only) sequence to use when burrow keys signs the transaction (do not use unless you
	// know what you're doing)
	Sequence string `mapstructure:"sequence" json:"sequence" yaml:"sequence" toml:"sequence"`
}

func (job *Permission) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Source, rule.AddressOrPlaceholder),
		validation.Field(&job.Value, validation.In("true", "false", "")),
		validation.Field(&job.Sequence, rule.Uint64OrPlaceholder),
	)
}

// ------------------------------------------------------------------------
// Contracts Jobs
// ------------------------------------------------------------------------

type PackageDeploy struct {
	// TODO
}

type Build struct {
	// (Required) the filepath to the contract file. this should be relative to the current path **or**
	// relative to the contracts path established via the --contracts-path flag or the $EPM_CONTRACTS_PATH
	// environment variable. If contract has a "bin" file extension then it will not be sent to the
	// compilers but rather will just be sent to the chain. Note, if you use a "call" job after deploying
	// a binary contract then you will be **required** to utilize an abi field in the call job.
	Contract string `mapstructure:"contract" json:"contract" yaml:"contract" toml:"contract"`
	// (Optional) where to save the result of the compilation
	BinPath string `mapstructure:"binpath" json:"binpath" yaml:"binpath" toml:"binpath"`
	// (Optional) the name of contract to instantiate (it has to be one of the contracts present)
	// in the file defined in Contract above.
	// When none is provided, the system will choose the contract with the same name as that file.
	// use "all" to override and deploy all contracts in order. if "all" is selected the result
	// of the job will default to the address of the contract which was deployed that matches
	// the name of the file (or the last one deployed if there are no matching names; not the "last"
	// one deployed" strategy is non-deterministic and should not be used).
	Instance string `mapstructure:"instance" json:"instance" yaml:"instance" toml:"instance"`
}

func (job *Build) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Contract, validation.Required),
	)
}

type Deploy struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required) the filepath to the contract file. this should be relative to the current path **or**
	// relative to the contracts path established via the --contracts-path flag or the $EPM_CONTRACTS_PATH
	// environment variable. If contract has a "bin" file extension then it will not be sent to the
	// compilers but rather will just be sent to the chain. Note, if you use a "call" job after deploying
	// a binary contract then you will be **required** to utilize an abi field in the call job.
	Contract string `mapstructure:"contract" json:"contract" yaml:"contract" toml:"contract"`
	// (Optional) the name of contract to instantiate (it has to be one of the contracts present)
	// in the file defined in Contract above.
	// When none is provided, the system will choose the contract with the same name as that file.
	// use "all" to override and deploy all contracts in order. if "all" is selected the result
	// of the job will default to the address of the contract which was deployed that matches
	// the name of the file (or the last one deployed if there are no matching names; not the "last"
	// one deployed" strategy is non-deterministic and should not be used).
	Instance string `mapstructure:"instance" json:"instance" yaml:"instance" toml:"instance"`
	// (Optional) the file path for the linkReferences for contract
	Libraries string `mapstructure:"libraries" json:"libraries" yaml:"libraries" toml:"libraries"`
	// (Optional) TODO: additional arguments to send along with the contract code
	Data interface{} `mapstructure:"data" json:"data" yaml:"data" toml:"data"`
	// (Optional) amount of tokens to send to the contract which will (after deployment) reside in the
	// contract's account
	Amount string `mapstructure:"amount" json:"amount" yaml:"amount" toml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" json:"fee" yaml:"fee" toml:"fee"`
	// (Optional) amount of gas which should be sent along with the contract deployment transaction
	Gas string `mapstructure:"gas" json:"gas" yaml:"gas" toml:"gas"`
	// (Optional, advanced only) sequence to use when burrow keys signs the transaction (do not use unless you
	// know what you're doing)
	Sequence string `mapstructure:"sequence" json:"sequence" yaml:"sequence" toml:"sequence"`
	// (Optional) todo
	Variables []*abi.Variable
}

func (job *Deploy) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Contract, validation.Required),
		validation.Field(&job.Amount, rule.Uint64OrPlaceholder),
		validation.Field(&job.Fee, rule.Uint64OrPlaceholder),
		validation.Field(&job.Gas, rule.Uint64OrPlaceholder),
		validation.Field(&job.Sequence, rule.Uint64OrPlaceholder),
	)
}

type Call struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required) address of the contract which should be called
	Destination string `mapstructure:"destination" json:"destination" yaml:"destination" toml:"destination"`
	// (Required unless testing fallback function) function inside the contract to be called
	Function string `mapstructure:"function" json:"function" yaml:"function" toml:"function"`
	// (Optional) data which should be called. will use the monax-abi tooling under the hood to formalize the
	// transaction
	Data interface{} `mapstructure:"data" json:"data" yaml:"data" toml:"data"`
	// (Optional) amount of tokens to send to the contract
	Amount string `mapstructure:"amount" json:"amount" yaml:"amount" toml:"amount"`
	// (Optional) validators' fee
	Fee string `mapstructure:"fee" json:"fee" yaml:"fee" toml:"fee"`
	// (Optional) amount of gas which should be sent along with the call transaction
	Gas string `mapstructure:"gas" json:"gas" yaml:"gas" toml:"gas"`
	// (Optional, advanced only) sequence to use when burrow keys signs the transaction (do not use unless you
	// know what you're doing)
	Sequence string `mapstructure:"sequence" json:"sequence" yaml:"sequence" toml:"sequence"`
	// (Optional) location of the bin file to use (can be relative path or in bin path)
	// deployed contracts save ABI artifacts in the bin folder as *both* the name of the contract
	// and the address where the contract was deployed to
	Bin string `mapstructure:"bin" json:"bin" yaml:"bin" toml:"bin"`
	// (Optional) by default the call job will "store" the return from the contract as the
	// result of the job. If you would like to store the transaction hash instead of the
	// return from the call job as the result of the call job then select "tx" on the save
	// variable. Anything other than "tx" in this field will use the default.
	Save string `mapstructure:"save" json:"save" yaml:"save" toml:"save"`
	// (Optional) the call job's returned variables
	Variables []*abi.Variable
}

func (job *Call) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Destination, validation.Required, rule.AddressOrPlaceholder),
		validation.Field(&job.Amount, rule.Uint64OrPlaceholder),
		validation.Field(&job.Fee, rule.Uint64OrPlaceholder),
		validation.Field(&job.Gas, rule.Uint64OrPlaceholder),
		validation.Field(&job.Sequence, rule.Uint64OrPlaceholder),
	)
}

// ------------------------------------------------------------------------
// State Jobs
// ------------------------------------------------------------------------

type DumpState struct {
	WithValidators bool   `mapstructure:"include-validators" json:"include-validators" yaml:"include-validators" toml:"include-validators"`
	ToIPFS         bool   `mapstructure:"to-ipfs" json:"to-ipfs" yaml:"to-ipfs" toml:"to-ipfs"`
	ToFile         bool   `mapstructure:"to-file" json:"to-file" yaml:"to-file" toml:"to-file"`
	IPFSHost       string `mapstructure:"ipfs-host" json:"ipfs-host" yaml:"ipfs-host" toml:"ipfs-host"`
	FilePath       string `mapstructure:"file" json:"file" yaml:"file" toml:"file"`
}

func (job *DumpState) Validate() error {
	// TODO: write validation logic
	return nil
}

type RestoreState struct {
	FromIPFS bool   `mapstructure:"from-ipfs" json:"from-ipfs" yaml:"from-ipfs" toml:"from-ipfs"`
	FromFile bool   `mapstructure:"from-file" json:"from-file" yaml:"from-file" toml:"from-file"`
	IPFSHost string `mapstructure:"ipfs-host" json:"ipfs-host" yaml:"ipfs-host" toml:"ipfs-host"`
	FilePath string `mapstructure:"file" json:"file" yaml:"file" toml:"file"`
}

func (job *RestoreState) Validate() error {
	// TODO: write validation logic
	return nil
}

// ------------------------------------------------------------------------
// Testing Jobs
// ------------------------------------------------------------------------

// aka. Simulated Call.
type QueryContract struct {
	// (Optional, if account job or global account set) address of the account from which to send (the
	// public key for the account must be available to burrow keys)
	Source string `mapstructure:"source" json:"source" yaml:"source" toml:"source"`
	// (Required) address of the contract which should be called
	Destination string `mapstructure:"destination" json:"destination" yaml:"destination" toml:"destination"`
	// (Required) data which should be called. will use the monax-abi tooling under the hood to formalize the
	// transaction. QueryContract will usually be used with "accessor" functions in contracts
	Function string `mapstructure:"function" json:"function" yaml:"function" toml:"function"`
	// (Optional) data to be used in the function arguments. Will use the monax-abi tooling under the hood to formalize the
	// transaction.
	Data interface{} `mapstructure:"data" json:"data" yaml:"data" toml:"data"`
	// (Optional) location of the bin file to use (can be relative path or in abi path)
	// deployed contracts save ABI artifacts in the abi folder as *both* the name of the contract
	// and the address where the contract was deployed to
	Bin string `mapstructure:"bin" json:"bin" yaml:"bin" toml:"bin"`

	Variables []*abi.Variable
}

func (job *QueryContract) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Destination, validation.Required, rule.AddressOrPlaceholder),
	)
}

type QueryAccount struct {
	// (Required) address of the account which should be queried
	Account string `mapstructure:"account" json:"account" yaml:"account" toml:"account"`
	// (Required) field which should be queried. If users are trying to query the permissions of the
	// account one can get either the `permissions.base` which will return the base permission of the
	// account, or one can get the `permissions.set` which will return the setBit of the account.
	Field string `mapstructure:"field" json:"field" yaml:"field" toml:"field"`
}

func (job *QueryAccount) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Account, validation.Required, rule.AddressOrPlaceholder),
		validation.Field(&job.Field, validation.Required),
	)
}

type QueryName struct {
	// (Required) name which should be queried
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	// (Required) field which should be quiried (generally will be "data" to get the registered "name")
	Field string `mapstructure:"field" json:"field" yaml:"field" toml:"field"`
}

func (job *QueryName) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Name, validation.Required),
		validation.Field(&job.Field, validation.Required),
	)
}

type QueryVals struct {
	// (Required) should be of the set ["bonded_validators" or "unbonding_validators"] and it will
	// return a comma separated listing of the addresses which fall into one of those categories
	Query string `mapstructure:"field" json:"field" yaml:"field" toml:"field"`
}

func (job *QueryVals) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Query, validation.Required),
	)
}

type Assert struct {
	// (Required) key which should be used for the assertion. This is usually known as the "expected"
	// value in most testing suites
	Key string `mapstructure:"key" json:"key" yaml:"key" toml:"key"`
	// (Required) must be of the set ["eq", "ne", "ge", "gt", "le", "lt", "==", "!=", ">=", ">", "<=", "<"]
	// establishes the relation to be tested by the assertion. If a strings key:value pair is being used
	// only the equals or not-equals relations may be used as the key:value will try to be converted to
	// ints for the remainder of the relations. if strings are passed to them then `monax pkgs do` will return an
	// error
	Relation string `mapstructure:"relation" json:"relation" yaml:"relation" toml:"relation"`
	// (Required) value which should be used for the assertion. This is usually known as the "given"
	// value in most testing suites. Generally it will be a variable expansion from one of the query
	// jobs.
	Value string `mapstructure:"val" json:"val" yaml:"val" toml:"val"`
}

func (job *Assert) Validate() error {
	return validation.ValidateStruct(job,
		validation.Field(&job.Relation, validation.Required, rule.Relation),
	)
}
