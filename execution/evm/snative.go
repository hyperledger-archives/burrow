// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package evm

import (
	"fmt"
	"reflect"
	"runtime"

	"strings"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
)

//
// SNative (from 'secure natives') are native (go) contracts that are dispatched
// based on account permissions and can access and modify an account's permissions
//

// Instructions on adding an SNative function. First declare a function like so:
//
// func unsetBase(context SNativeContext, args unsetBaseArgs) (unsetBaseRets, error) {
// }
//
// The name of the function will be used as the name of the function in solidity. The
// first arguments is SNativeContext; this will give you access to state, and the logger
// etc. The second arguments must be a struct type. The members of this struct must be
// exported (start with uppercase letter), and they will be converted into arguments
// for the solidity function, with the same types. The first return value is a struct
// which defines the return values from solidity just like the arguments.
//
// The second return value must be error. If non-nil is returned for error, then
// the current transaction will be aborted and the execution will stop.
//
// For each contract you will need to create a SNativeContractDescription{} struct,
// with the function listed. Only the PermFlag and the function F needs to be filled
// in for each SNativeFunctionDescription. Add this to the SNativeContracts() function.

// SNativeContext is the first argument to any snative function. This struct carries
// all the context an snative needs to access e.g. state in burrow.
type SNativeContext struct {
	State  Interface
	Caller crypto.Address
	Gas    *uint64
	Logger *logging.Logger
}

// SNativeContractDescription is metadata for SNative contract. Acts as a call target
// from the EVM. Can be used to generate bindings in a smart contract languages.
type SNativeContractDescription struct {
	// Comment describing purpose of SNative contract and reason for assembling
	// the particular functions
	Comment string
	// Name of the SNative contract
	Name          string
	functionsByID map[abi.FunctionID]*SNativeFunctionDescription
	functions     []*SNativeFunctionDescription
}

// SNativeFunctionDescription is metadata for SNative functions. Act as call targets
// for the EVM when collected into an SNativeContractDescription. Can be used to generate
// bindings in a smart contract languages.
type SNativeFunctionDescription struct {
	// Comment describing function's purpose, parameters, and return value
	Comment string
	// Permissions required to call function
	PermFlag permission.PermFlag
	// Native function to which calls will be dispatched when a containing
	F interface{}

	// Following fields are for only for memoization

	// Function name (used to form signature)
	name string
	// The abi
	abi abi.FunctionSpec
}

func registerSNativeContracts() {
	for _, contract := range SNativeContracts() {
		if !RegisterNativeContract(contract.Address(), contract.Dispatch) {
			panic(fmt.Errorf("could not register SNative contract %s because address %s already registered",
				contract.Address(), contract.Name))
		}
	}
}

// SNativeContracts returns a map of all SNative contracts defined indexed by name
func SNativeContracts() map[string]*SNativeContractDescription {
	contracts := []*SNativeContractDescription{
		NewSNativeContract(`
		* Interface for managing Secure Native authorizations.
		* @dev This interface describes the functions exposed by the SNative permissions layer in burrow.
		`,
			"Permissions",
			&SNativeFunctionDescription{Comment: `
			* @notice Adds a role to an account
			* @param Account account address
			* @param Role role name
			* @return result whether role was added
			`,
				PermFlag: permission.AddRole,
				F:        addRole},

			&SNativeFunctionDescription{Comment: `
			* @notice Removes a role from an account
			* @param Account account address
			* @param Role role name
			* @return result whether role was removed
			`,
				PermFlag: permission.RemoveRole,
				F:        removeRole},

			&SNativeFunctionDescription{Comment: `
			* @notice Indicates whether an account has a role
			* @param Account account address
			* @param Role role name
			* @return result whether account has role
			`,
				PermFlag: permission.HasRole,
				F:        hasRole},

			&SNativeFunctionDescription{Comment: `
			* @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
			* @param Account account address
			* @param Permission the base permissions flags to set for the account
			* @param Set whether to set or unset the permissions flags at the account level
			* @return The permission flag that was set as uint64
			`,
				PermFlag: permission.SetBase,
				F:        setBase},

			&SNativeFunctionDescription{Comment: `
			* @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
      		* @param Account account address
      		* @param Permission the permissions flags to unset for the account
			* @return The permission flag that was unset as uint64
      `,
				PermFlag: permission.UnsetBase,
				F:        unsetBase},

			&SNativeFunctionDescription{Comment: `
			* @notice Indicates whether an account has a subset of permissions set
			* @param Account account address
			* @param Permission the permissions flags (mask) to check whether enabled against base permissions for the account
			* @return result whether account has the passed permissions flags set
			`,
				PermFlag: permission.HasBase,
				F:        hasBase},

			&SNativeFunctionDescription{Comment: `
			* @notice Sets the global (default) permissions flags for the entire chain
			* @param Permission the permissions flags to set
			* @param Set whether to set (or unset) the permissions flags
			* @return The permission flag that was set as uint64
			`,
				PermFlag: permission.SetGlobal,
				F:        setGlobal},
		),
	}

	contractMap := make(map[string]*SNativeContractDescription, len(contracts))
	for _, contract := range contracts {
		if _, ok := contractMap[contract.Name]; ok {
			// If this happens we have a pseudo compile time error that will be caught
			// on native.go init()
			panic(fmt.Errorf("duplicate contract with name %s defined. "+
				"Contract names must be unique", contract.Name))
		}
		contractMap[contract.Name] = contract
	}
	return contractMap
}

// Create a new SNative contract description object by passing a comment, name
// and a list of member functions descriptions
func NewSNativeContract(comment, name string,
	functions ...*SNativeFunctionDescription) *SNativeContractDescription {

	functionsByID := make(map[abi.FunctionID]*SNativeFunctionDescription, len(functions))
	for _, f := range functions {
		// Get name of function
		t := reflect.TypeOf(f.F)
		v := reflect.ValueOf(f.F)
		// v.String() for functions returns the empty string
		fullyqualifiedname := runtime.FuncForPC(v.Pointer()).Name()
		a := strings.Split(fullyqualifiedname, ".")
		f.name = a[len(a)-1]

		if t.NumIn() != 2 {
			panic(fmt.Sprintf("%s must have two arguments", fullyqualifiedname))
		}

		if t.NumOut() != 2 {
			panic(fmt.Sprintf("%s must have two return values", fullyqualifiedname))
		}

		if t.In(0) != reflect.TypeOf(SNativeContext{}) {
			panic(fmt.Sprintf("first agument of %s must be struct SNativeContext", fullyqualifiedname))
		}

		f.abi = *abi.SpecFromStructReflect(f.name, t.In(1), t.Out(0))
		fid := f.abi.FunctionID
		otherF, ok := functionsByID[fid]
		if ok {
			panic(fmt.Errorf("function with ID %x already defined: %s", fid, otherF.Signature()))
		}
		functionsByID[fid] = f
	}
	return &SNativeContractDescription{
		Comment:       comment,
		Name:          name,
		functionsByID: functionsByID,
		functions:     functions,
	}
}

// Dispatch is designed to be called from the EVM once a SNative contract
// has been selected. It is also placed in a registry by registerSNativeContracts
// So it can be looked up by SNative address
func (contract *SNativeContractDescription) Dispatch(st Interface, caller crypto.Address,
	args []byte, gas *uint64, logger *logging.Logger) (output []byte, err error) {

	logger = logger.With(structure.ScopeKey, "Dispatch", "contract_name", contract.Name)

	if len(args) < abi.FunctionIDSize {
		return nil, errors.ErrorCodef(errors.ErrorCodeNativeFunction,
			"SNatives dispatch requires a 4-byte function identifier but arguments are only %v bytes long",
			len(args))
	}

	var id abi.FunctionID
	copy(id[:], args)
	function, err := contract.FunctionByID(id)
	if err != nil {
		return nil, err
	}

	logger.TraceMsg("Dispatching to function",
		"caller", caller,
		"function_name", function.name)

	remainingArgs := args[abi.FunctionIDSize:]

	// check if we have permission to call this function
	if !HasPermission(st, caller, function.PermFlag) {
		return nil, errors.LacksSNativePermission{Address: caller, SNative: function.name}
	}

	arguments := reflect.New(reflect.TypeOf(function.F).In(1))
	err = abi.Unpack(function.abi.Inputs, remainingArgs, arguments.Interface())
	if err != nil {
		return nil, err
	}

	ctx := SNativeContext{
		State:  st,
		Caller: caller,
		Gas:    gas,
		Logger: logger,
	}

	fn := reflect.ValueOf(function.F)
	rets := fn.Call([]reflect.Value{reflect.ValueOf(ctx), arguments.Elem()})
	if !rets[1].IsNil() {
		return nil, rets[1].Interface().(error)
	}
	err = st.Error()
	if err != nil {
		return nil, fmt.Errorf("state error in %v: %v", function, err)
	}

	return abi.Pack(function.abi.Outputs, rets[0].Interface())
}

// We define the address of an SNative contact as the last 20 bytes of the sha3
// hash of its name
func (contract *SNativeContractDescription) Address() (address crypto.Address) {
	hash := sha3.Sha3([]byte(contract.Name))
	copy(address[:], hash[len(hash)-crypto.AddressLength:])
	return
}

// Get function by calling identifier FunctionSelector
func (contract *SNativeContractDescription) FunctionByID(id abi.FunctionID) (*SNativeFunctionDescription, errors.CodedError) {
	f, ok := contract.functionsByID[id]
	if !ok {
		return nil,
			errors.ErrorCodef(errors.ErrorCodeNativeFunction, "unknown SNative function with ID %x", id)
	}
	return f, nil
}

// Get function by name
func (contract *SNativeContractDescription) FunctionByName(name string) (*SNativeFunctionDescription, error) {
	for _, f := range contract.functions {
		if f.name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("unknown SNative function with name %s", name)
}

// Get functions in order of declaration
func (contract *SNativeContractDescription) Functions() []*SNativeFunctionDescription {
	functions := make([]*SNativeFunctionDescription, len(contract.functions))
	copy(functions, contract.functions)
	return functions
}

//
// SNative functions
//

// Signature returns the function signature as would be used for ABI hashing
func (function *SNativeFunctionDescription) Signature() string {
	argTypeNames := make([]string, len(function.abi.Inputs))
	for i, arg := range function.abi.Inputs {
		argTypeNames[i] = arg.EVM.GetSignature()
	}
	return fmt.Sprintf("%s(%s)", function.name,
		strings.Join(argTypeNames, ","))
}

// NArgs returns the number of function arguments
func (function *SNativeFunctionDescription) NArgs() int {
	return len(function.abi.Inputs)
}

// Name returns the name for this function
func (function *SNativeFunctionDescription) Name() string {
	return function.name
}

// Abi returns the FunctionSpec for this function
func (function *SNativeFunctionDescription) Abi() abi.FunctionSpec {
	return function.abi
}

func (fn *SNativeFunctionDescription) String() string {
	return fmt.Sprintf("SNativeFunction{Name: %s; Inputs: %d; Outputs: %d}",
		fn.name, len(fn.abi.Inputs), len(fn.abi.Outputs))
}

// Permission function defintions

// TODO: catch errors, log em, return 0s to the vm (should some errors cause exceptions though?)
type hasBaseArgs struct {
	Account    crypto.Address
	Permission uint64
}

type hasBaseRets struct {
	Result bool
}

func hasBase(context SNativeContext, args hasBaseArgs) (hasBaseRets, error) {
	if !context.State.Exists(args.Account) {
		return hasBaseRets{}, fmt.Errorf("unknown account %s", args.Account)
	}
	permN := permission.PermFlag(args.Permission) // already shifted
	if !permN.IsValid() {
		return hasBaseRets{}, permission.ErrInvalidPermission(permN)
	}
	hasPermission := HasPermission(context.State, args.Account, permN)
	context.Logger.Trace.Log("function", "hasBase",
		"address", args.Account.String(),
		"perm_flag", fmt.Sprintf("%b", permN),
		"has_permission", hasPermission)
	return hasBaseRets{Result: hasPermission}, nil
}

type setBaseArgs struct {
	Account    crypto.Address
	Permission uint64
	Set        bool
}

type setBaseRets struct {
	Result uint64
}

func setBase(context SNativeContext, args setBaseArgs) (setBaseRets, error) {
	exists := context.State.Exists(args.Account)
	if !exists {
		return setBaseRets{}, fmt.Errorf("unknown account %s", args.Account)
	}
	permN := permission.PermFlag(args.Permission)
	if !permN.IsValid() {
		return setBaseRets{}, permission.ErrInvalidPermission(permN)
	}
	context.State.SetPermission(args.Account, permN, args.Set)
	context.Logger.Trace.Log("function", "setBase", "address", args.Account.String(),
		"permission_flag", fmt.Sprintf("%b", permN),
		"permission_value", args.Permission)
	return setBaseRets{Result: uint64(permN)}, nil
}

type unsetBaseArgs struct {
	Account    crypto.Address
	Permission uint64
}

type unsetBaseRets struct {
	Result uint64
}

func unsetBase(context SNativeContext, args unsetBaseArgs) (unsetBaseRets, error) {
	if !context.State.Exists(args.Account) {
		return unsetBaseRets{}, fmt.Errorf("unknown account %s", args.Account)
	}
	permN := permission.PermFlag(args.Permission)
	if !permN.IsValid() {
		return unsetBaseRets{}, permission.ErrInvalidPermission(permN)
	}
	context.State.UnsetPermission(args.Account, permN)
	context.Logger.Trace.Log("function", "unsetBase", "address", args.Account.String(),
		"perm_flag", fmt.Sprintf("%b", permN),
		"permission_flag", fmt.Sprintf("%b", permN))

	return unsetBaseRets{Result: uint64(permN)}, nil
}

type setGlobalArgs struct {
	Permission uint64
	Set        bool
}

type setGlobalRets struct {
	Result uint64
}

func setGlobal(context SNativeContext, args setGlobalArgs) (setGlobalRets, error) {
	permN := permission.PermFlag(args.Permission)
	if !permN.IsValid() {
		return setGlobalRets{}, permission.ErrInvalidPermission(permN)
	}
	context.State.SetPermission(acm.GlobalPermissionsAddress, permN, args.Set)
	context.Logger.Trace.Log("function", "setGlobal",
		"permission_flag", fmt.Sprintf("%b", permN),
		"permission_value", args.Set)
	return setGlobalRets{Result: uint64(permN)}, nil
}

type hasRoleArgs struct {
	Account crypto.Address
	Role    string
}

type hasRoleRets struct {
	Result bool
}

func hasRole(context SNativeContext, args hasRoleArgs) (hasRoleRets, error) {
	perms := context.State.GetPermissions(args.Account)
	if err := context.State.Error(); err != nil {
		return hasRoleRets{}, fmt.Errorf("hasRole could not get permissions: %v", err)
	}
	hasRole := perms.HasRole(args.Role)
	context.Logger.Trace.Log("function", "hasRole", "address", args.Account.String(),
		"role", args.Role,
		"has_role", hasRole)
	return hasRoleRets{Result: hasRole}, nil
}

type addRoleArgs struct {
	Account crypto.Address
	Role    string
}

type addRoleRets struct {
	Result bool
}

func addRole(context SNativeContext, args addRoleArgs) (addRoleRets, error) {
	roleAdded := context.State.AddRole(args.Account, args.Role)
	context.Logger.Trace.Log("function", "addRole", "address", args.Account.String(),
		"role", args.Role,
		"role_added", roleAdded)
	return addRoleRets{Result: roleAdded}, nil
}

type removeRoleArgs struct {
	Account crypto.Address
	Role    string
}

type removeRoleRets struct {
	Result bool
}

func removeRole(context SNativeContext, args removeRoleArgs) (removeRoleRets, error) {
	roleRemoved := context.State.RemoveRole(args.Account, args.Role)
	context.Logger.Trace.Log("function", "removeRole", "address", args.Account.String(),
		"role", args.Role,
		"role_removed", roleRemoved)
	return removeRoleRets{Result: roleRemoved}, nil
}
