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

	"strings"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
)

//
// SNative (from 'secure natives') are native (go) contracts that are dispatched
// based on account permissions and can access and modify an account's permissions
//

// Metadata for SNative contract. Acts as a call target from the EVM. Can be
// used to generate bindings in a smart contract languages.
type SNativeContractDescription struct {
	// Comment describing purpose of SNative contract and reason for assembling
	// the particular functions
	Comment string
	// Name of the SNative contract
	Name          string
	functionsByID map[abi.FunctionID]*SNativeFunctionDescription
	functions     []*SNativeFunctionDescription
}

// Metadata for SNative functions. Act as call targets for the EVM when
// collected into an SNativeContractDescription. Can be used to generate
// bindings in a smart contract languages.
type SNativeFunctionDescription struct {
	// Comment describing function's purpose, parameters, and return value
	Comment string
	// Function name (used to form signature)
	Name string
	// Function arguments
	Arguments reflect.Type
	// Function return values
	Returns reflect.Type
	// The abi
	Abi abi.FunctionSpec
	// Permissions required to call function
	PermFlag permission.PermFlag
	// Native function to which calls will be dispatched when a containing
	F func(stateWriter state.ReaderWriter, caller acm.Account, gas *uint64,
		logger *logging.Logger, v interface{}) (interface{}, error)
}

func registerSNativeContracts() {
	for _, contract := range SNativeContracts() {
		if !RegisterNativeContract(contract.Address().Word256(), contract.Dispatch) {
			panic(fmt.Errorf("could not register SNative contract %s because address %s already registered",
				contract.Address(), contract.Name))
		}
	}
}

// Returns a map of all SNative contracts defined indexed by name
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
				Name:      "addRole",
				PermFlag:  permission.AddRole,
				Arguments: reflect.TypeOf(addRoleArgs{}),
				Returns:   reflect.TypeOf(addRoleRets{}),
				F:         addRole},

			&SNativeFunctionDescription{Comment: `
			* @notice Removes a role from an account
			* @param Account account address
			* @param Role role name
			* @return result whether role was removed
			`,
				Name:      "removeRole",
				PermFlag:  permission.RemoveRole,
				Arguments: reflect.TypeOf(removeRoleArgs{}),
				Returns:   reflect.TypeOf(removeRoleRets{}),
				F:         removeRole},

			&SNativeFunctionDescription{Comment: `
			* @notice Indicates whether an account has a role
			* @param Account account address
			* @param Role role name
			* @return result whether account has role
			`,
				Name:      "hasRole",
				PermFlag:  permission.HasRole,
				Arguments: reflect.TypeOf(hasRoleArgs{}),
				Returns:   reflect.TypeOf(hasRoleRets{}),
				F:         hasRole},

			&SNativeFunctionDescription{Comment: `
			* @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
			* @param Account account address
			* @param Permission the base permissions flags to set for the account
			* @param Set whether to set or unset the permissions flags at the account level
			* @return result the effective permissions flags on the account after the call
			`,
				Name:      "setBase",
				PermFlag:  permission.SetBase,
				Arguments: reflect.TypeOf(setBaseArgs{}),
				Returns:   reflect.TypeOf(setBaseRets{}),
				F:         setBase},

			&SNativeFunctionDescription{Comment: `
			* @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
      		* @param Account account address
      		* @param Permission the permissions flags to unset for the account
			* @return result the effective permissions flags on the account after the call
      `,
				Name:      "unsetBase",
				PermFlag:  permission.UnsetBase,
				Arguments: reflect.TypeOf(unsetBaseArgs{}),
				Returns:   reflect.TypeOf(unsetBaseRets{}),
				F:         unsetBase},

			&SNativeFunctionDescription{Comment: `
			* @notice Indicates whether an account has a subset of permissions set
			* @param Account account address
			* @param Permission the permissions flags (mask) to check whether enabled against base permissions for the account
			* @return result whether account has the passed permissions flags set
			`,
				Name:      "hasBase",
				PermFlag:  permission.HasBase,
				Arguments: reflect.TypeOf(hasBaseArgs{}),
				Returns:   reflect.TypeOf(hasBaseRets{}),
				F:         hasBase},

			&SNativeFunctionDescription{Comment: `
			* @notice Sets the global (default) permissions flags for the entire chain
			* @param Permission the permissions flags to set
			* @param Set whether to set (or unset) the permissions flags
			* @return result the global permissions flags after the call
			`,
				Name:      "setGlobal",
				PermFlag:  permission.SetGlobal,
				Arguments: reflect.TypeOf(setGlobalArgs{}),
				Returns:   reflect.TypeOf(setGlobalRets{}),
				F:         setGlobal},
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

		f.Abi = *abi.SpecFromStructReflect(f.Name, f.Arguments, f.Returns)
		fid := f.Abi.FunctionID
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

// This function is designed to be called from the EVM once a SNative contract
// has been selected. It is also placed in a registry by registerSNativeContracts
// So it can be looked up by SNative address
func (contract *SNativeContractDescription) Dispatch(state state.ReaderWriter, caller acm.Account,
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
		"caller", caller.Address(),
		"function_name", function.Name)

	remainingArgs := args[abi.FunctionIDSize:]

	// check if we have permission to call this function
	if !HasPermission(state, caller, function.PermFlag) {
		return nil, errors.LacksSNativePermission{caller.Address(), function.Name}
	}

	nativeArgs := reflect.New(function.Arguments).Interface()
	err = abi.UnpackIntoStruct(function.Abi.Inputs, remainingArgs, nativeArgs)
	if err != nil {
		return nil, err
	}

	nativeRets, err := function.F(state, caller, gas, logger, nativeArgs)
	if err != nil {
		return nil, err
	}

	return abi.PackIntoStruct(function.Abi.Outputs, nativeRets)
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
		if f.Name == name {
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

// Get function signature
func (function *SNativeFunctionDescription) Signature() string {
	argTypeNames := make([]string, len(function.Abi.Inputs))
	for i, arg := range function.Abi.Inputs {
		argTypeNames[i] = arg.EVM.GetSignature()
	}
	return fmt.Sprintf("%s(%s)", function.Name,
		strings.Join(argTypeNames, ","))
}

// Get number of function arguments
func (function *SNativeFunctionDescription) NArgs() int {
	return len(function.Abi.Inputs)
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

func hasBase(state state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, a interface{}) (interface{}, error) {
	args := a.(*hasBaseArgs)

	acc, err := state.GetAccount(args.Account)
	if err != nil {
		return false, err
	}
	if acc == nil {
		return false, fmt.Errorf("unknown account %s", args.Account)
	}
	permN := permission.PermFlag(args.Permission) // already shifted
	if !permN.IsValid() {
		return false, permission.ErrInvalidPermission(permN)
	}
	hasPermission := HasPermission(state, acc, permN)
	logger.Trace.Log("function", "hasBase",
		"address", args.Account.String(),
		"account_base_permissions", acc.Permissions().Base,
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

func setBase(stateWriter state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, a interface{}) (interface{}, error) {
	args := a.(*setBaseArgs)

	acc, err := state.GetMutableAccount(stateWriter, args.Account)
	if err != nil {
		return 0, err
	}
	if acc == nil {
		return 0, fmt.Errorf("unknown account %s", args.Account)
	}
	permN := permission.PermFlag(args.Permission)
	if !permN.IsValid() {
		return 0, permission.ErrInvalidPermission(permN)
	}
	if err = acc.MutablePermissions().Base.Set(permN, args.Set); err != nil {
		return 0, err
	}
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "setBase", "address", args.Account.String(),
		"permission_flag", fmt.Sprintf("%b", permN),
		"permission_value", args.Permission)
	return setBaseRets{Result: uint64(effectivePerm(acc.Permissions().Base, globalPerms(stateWriter)))}, nil
}

type unsetBaseArgs struct {
	Account    crypto.Address
	Permission uint64
}

type unsetBaseRets struct {
	Result uint64
}

func unsetBase(stateWriter state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, a interface{}) (r interface{}, err error) {
	args := a.(*unsetBaseArgs)

	acc, err := state.GetMutableAccount(stateWriter, args.Account)
	if err != nil {
		return 0, err
	}
	if acc == nil {
		return 0, fmt.Errorf("unknown account %s", args.Account)
	}
	permN := permission.PermFlag(args.Permission)
	if !permN.IsValid() {
		return 0, permission.ErrInvalidPermission(permN)
	}
	if err = acc.MutablePermissions().Base.Unset(permN); err != nil {
		return 0, err
	}
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "unsetBase", "address", args.Account.String(),
		"perm_flag", fmt.Sprintf("%b", permN),
		"permission_flag", fmt.Sprintf("%b", permN))

	return unsetBaseRets{Result: uint64(effectivePerm(acc.Permissions().Base, globalPerms(stateWriter)))}, nil
}

type setGlobalArgs struct {
	Permission uint64
	Set        bool
}

type setGlobalRets struct {
	Result uint64
}

func setGlobal(stateWriter state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, a interface{}) (r interface{}, err error) {

	args := a.(*setGlobalArgs)

	acc, err := state.GetMutableAccount(stateWriter, acm.GlobalPermissionsAddress)
	if err != nil {
		return 0, err
	}
	if acc == nil {
		panic("cant find the global permissions account")
	}
	permN := permission.PermFlag(args.Permission)
	if !permN.IsValid() {
		return 0, permission.ErrInvalidPermission(permN)
	}
	if err = acc.MutablePermissions().Base.Set(permN, args.Set); err != nil {
		return 0, err
	}
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "setGlobal",
		"permission_flag", fmt.Sprintf("%b", permN),
		"permission_value", args.Set)
	return setGlobalRets{Result: uint64(acc.Permissions().Base.ResultantPerms())}, nil
}

type hasRoleArgs struct {
	Account crypto.Address
	Role    string
}

type hasRoleRets struct {
	Result bool
}

func hasRole(state state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, a interface{}) (r interface{}, err error) {

	args := a.(*hasRoleArgs)
	acc, err := state.GetAccount(args.Account)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", args.Account)
	}
	hasRole := acc.Permissions().HasRole(args.Role)
	logger.Trace.Log("function", "hasRole", "address", args.Account.String(),
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

func addRole(stateWriter state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, v interface{}) (interface{}, error) {
	args := v.(*addRoleArgs)
	acc, err := state.GetMutableAccount(stateWriter, args.Account)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", args.Account)
	}
	roleAdded := acc.MutablePermissions().AddRole(args.Role)
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "addRole", "address", args.Account.String(),
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

func removeRole(stateWriter state.ReaderWriter, caller acm.Account, gas *uint64,
	logger *logging.Logger, a interface{}) (interface{}, error) {
	args := a.(*removeRoleArgs)

	acc, err := state.GetMutableAccount(stateWriter, args.Account)
	if err != nil {
		return false, err
	}
	if acc == nil {
		return false, fmt.Errorf("unknown account %s", args.Account)
	}
	roleRemoved := acc.MutablePermissions().RmRole(args.Role)
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "removeRole", "address", args.Account.String(),
		"role", args.Role,
		"role_removed", roleRemoved)
	return removeRoleRets{Result: roleRemoved}, nil
}

//------------------------------------------------------------------------------------------------
// Errors and utility funcs

// Get the global BasePermissions
func globalPerms(stateWriter state.ReaderWriter) permission.BasePermissions {
	return state.GlobalAccountPermissions(stateWriter).Base
}

// Compute the effective permissions from an acm.Account's BasePermissions by
// taking the bitwise or with the global BasePermissions resultant permissions
func effectivePerm(basePerms permission.BasePermissions,
	globalPerms permission.BasePermissions) permission.PermFlag {
	return basePerms.ResultantPerms() | globalPerms.ResultantPerms()
}
