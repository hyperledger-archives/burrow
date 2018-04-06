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

	"strings"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	. "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/permission"
	ptypes "github.com/hyperledger/burrow/permission/types"
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
	functionsByID map[abi.FunctionSelector]*SNativeFunctionDescription
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
	// Function arguments (used to form signature)
	Args []abi.Arg
	// Function return value
	Return abi.Return
	// Permissions required to call function
	PermFlag ptypes.PermFlag
	// Native function to which calls will be dispatched when a containing
	// contract is called with a FunctionSelector matching this NativeContract
	F NativeContract
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
	permFlagTypeName := abi.Uint64TypeName
	roleTypeName := abi.Bytes32TypeName
	contracts := []*SNativeContractDescription{
		NewSNativeContract(`
		* Interface for managing Secure Native authorizations.
		* @dev This interface describes the functions exposed by the SNative permissions layer in burrow.
		`,
			"Permissions",
			&SNativeFunctionDescription{`
			* @notice Adds a role to an account
			* @param _account account address
			* @param _role role name
			* @return result whether role was added
			`,
				"addRole",
				[]abi.Arg{
					abiArg("_account", abi.AddressTypeName),
					abiArg("_role", roleTypeName),
				},
				abiReturn("result", abi.BoolTypeName),
				permission.AddRole,
				addRole},

			&SNativeFunctionDescription{`
			* @notice Removes a role from an account
			* @param _account account address
			* @param _role role name
			* @return result whether role was removed
			`,
				"removeRole",
				[]abi.Arg{
					abiArg("_account", abi.AddressTypeName),
					abiArg("_role", roleTypeName),
				},
				abiReturn("result", abi.BoolTypeName),
				permission.RemoveRole,
				removeRole},

			&SNativeFunctionDescription{`
			* @notice Indicates whether an account has a role
			* @param _account account address
			* @param _role role name
			* @return result whether account has role
			`,
				"hasRole",
				[]abi.Arg{
					abiArg("_account", abi.AddressTypeName),
					abiArg("_role", roleTypeName),
				},
				abiReturn("result", abi.BoolTypeName),
				permission.HasRole,
				hasRole},

			&SNativeFunctionDescription{`
			* @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
			* @param _account account address
			* @param _permission the base permissions flags to set for the account
			* @param _set whether to set or unset the permissions flags at the account level
			* @return result the effective permissions flags on the account after the call
			`,
				"setBase",
				[]abi.Arg{
					abiArg("_account", abi.AddressTypeName),
					abiArg("_permission", permFlagTypeName),
					abiArg("_set", abi.BoolTypeName),
				},
				abiReturn("result", permFlagTypeName),
				permission.SetBase,
				setBase},

			&SNativeFunctionDescription{`
			* @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
      		* @param _account account address
      		* @param _permission the permissions flags to unset for the account
			* @return result the effective permissions flags on the account after the call
      `,
				"unsetBase",
				[]abi.Arg{
					abiArg("_account", abi.AddressTypeName),
					abiArg("_permission", permFlagTypeName)},
				abiReturn("result", permFlagTypeName),
				permission.UnsetBase,
				unsetBase},

			&SNativeFunctionDescription{`
			* @notice Indicates whether an account has a subset of permissions set
			* @param _account account address
			* @param _permission the permissions flags (mask) to check whether enabled against base permissions for the account
			* @return result whether account has the passed permissions flags set
			`,
				"hasBase",
				[]abi.Arg{
					abiArg("_account", abi.AddressTypeName),
					abiArg("_permission", permFlagTypeName)},
				abiReturn("result", abi.BoolTypeName),
				permission.HasBase,
				hasBase},

			&SNativeFunctionDescription{`
			* @notice Sets the global (default) permissions flags for the entire chain
			* @param _permission the permissions flags to set
			* @param _set whether to set (or unset) the permissions flags
			* @return result the global permissions flags after the call
			`,
				"setGlobal",
				[]abi.Arg{
					abiArg("_permission", permFlagTypeName),
					abiArg("_set", abi.BoolTypeName)},
				abiReturn("result", permFlagTypeName),
				permission.SetGlobal,
				setGlobal},
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

	functionsByID := make(map[abi.FunctionSelector]*SNativeFunctionDescription, len(functions))
	for _, f := range functions {
		fid := f.ID()
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

type ErrLacksSNativePermission struct {
	Address acm.Address
	SNative string
}

func (e ErrLacksSNativePermission) Error() string {
	return fmt.Sprintf("account %s does not have SNative function call permission: %s", e.Address, e.SNative)
}

// This function is designed to be called from the EVM once a SNative contract
// has been selected. It is also placed in a registry by registerSNativeContracts
// So it can be looked up by SNative address
func (contract *SNativeContractDescription) Dispatch(state state.Writer, caller acm.Account,
	args []byte, gas *uint64, logger *logging.Logger) (output []byte, err error) {

	logger = logger.With(structure.ScopeKey, "Dispatch", "contract_name", contract.Name)

	if len(args) < abi.FunctionSelectorLength {
		return nil, fmt.Errorf("SNatives dispatch requires a 4-byte function "+
			"identifier but arguments are only %v bytes long", len(args))
	}

	function, err := contract.FunctionByID(abi.FirstFourBytes(args))
	if err != nil {
		return nil, err
	}

	logger.TraceMsg("Dispatching to function",
		"caller", caller.Address(),
		"function_name", function.Name)

	remainingArgs := args[abi.FunctionSelectorLength:]

	// check if we have permission to call this function
	if !HasPermission(state, caller, function.PermFlag) {
		return nil, ErrLacksSNativePermission{caller.Address(), function.Name}
	}

	// ensure there are enough arguments
	if len(remainingArgs) != function.NArgs()*Word256Length {
		return nil, fmt.Errorf("%s() takes %d arguments but got %d (with %d bytes unconsumed - should be 0)",
			function.Name, function.NArgs(), len(remainingArgs)/Word256Length, len(remainingArgs)%Word256Length)
	}

	// call the function
	return function.F(state, caller, remainingArgs, gas, logger)
}

// We define the address of an SNative contact as the last 20 bytes of the sha3
// hash of its name
func (contract *SNativeContractDescription) Address() (address acm.Address) {
	hash := sha3.Sha3([]byte(contract.Name))
	copy(address[:], hash[len(hash)-abi.AddressLength:])
	return
}

// Get function by calling identifier FunctionSelector
func (contract *SNativeContractDescription) FunctionByID(id abi.FunctionSelector) (*SNativeFunctionDescription, error) {
	f, ok := contract.functionsByID[id]
	if !ok {
		return nil,
			fmt.Errorf("unknown SNative function with ID %x", id)
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
	argTypeNames := make([]string, len(function.Args))
	for i, arg := range function.Args {
		argTypeNames[i] = string(arg.TypeName)
	}
	return fmt.Sprintf("%s(%s)", function.Name,
		strings.Join(argTypeNames, ","))
}

// Get function calling identifier FunctionSelector
func (function *SNativeFunctionDescription) ID() abi.FunctionSelector {
	return abi.FunctionID(function.Signature())
}

// Get number of function arguments
func (function *SNativeFunctionDescription) NArgs() int {
	return len(function.Args)
}

func abiArg(name string, abiTypeName abi.TypeName) abi.Arg {
	return abi.Arg{
		Name:     name,
		TypeName: abiTypeName,
	}
}

func abiReturn(name string, abiTypeName abi.TypeName) abi.Return {
	return abi.Return{
		Name:     name,
		TypeName: abiTypeName,
	}
}

// Permission function defintions

// TODO: catch errors, log em, return 0s to the vm (should some errors cause exceptions though?)
func hasBase(state state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, permNum := returnTwoArgs(args)
	address := acm.AddressFromWord256(addrWord256)
	acc, err := state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", address)
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum)) // already shifted
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	hasPermission := HasPermission(state, acc, permN)
	permInt := byteFromBool(hasPermission)
	logger.Trace.Log("function", "hasBase",
		"address", address.String(),
		"account_base_permissions", acc.Permissions().Base,
		"perm_flag", fmt.Sprintf("%b", permN),
		"has_permission", hasPermission)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func setBase(stateWriter state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, permNum, permVal := returnThreeArgs(args)
	address := acm.AddressFromWord256(addrWord256)
	acc, err := state.GetMutableAccount(stateWriter, address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", address)
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum))
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	permV := !permVal.IsZero()
	if err = acc.MutablePermissions().Base.Set(permN, permV); err != nil {
		return nil, err
	}
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "setBase", "address", address.String(),
		"permission_flag", fmt.Sprintf("%b", permN),
		"permission_value", permV)
	return effectivePermBytes(acc.Permissions().Base, globalPerms(stateWriter)), nil
}

func unsetBase(stateWriter state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, permNum := returnTwoArgs(args)
	address := acm.AddressFromWord256(addrWord256)
	acc, err := state.GetMutableAccount(stateWriter, address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", address)
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum))
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	if err = acc.MutablePermissions().Base.Unset(permN); err != nil {
		return nil, err
	}
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "unsetBase", "address", address.String(),
		"perm_flag", fmt.Sprintf("%b", permN),
		"permission_flag", fmt.Sprintf("%b", permN))

	return effectivePermBytes(acc.Permissions().Base, globalPerms(stateWriter)), nil
}

func setGlobal(stateWriter state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	permNum, permVal := returnTwoArgs(args)
	acc, err := state.GetMutableAccount(stateWriter, acm.GlobalPermissionsAddress)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		panic("cant find the global permissions account")
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum))
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	permV := !permVal.IsZero()
	if err = acc.MutablePermissions().Base.Set(permN, permV); err != nil {
		return nil, err
	}
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "setGlobal",
		"permission_flag", fmt.Sprintf("%b", permN),
		"permission_value", permV)
	return permBytes(acc.Permissions().Base.ResultantPerms()), nil
}

func hasRole(state state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, role := returnTwoArgs(args)
	address := acm.AddressFromWord256(addrWord256)
	acc, err := state.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", address)
	}
	roleS := string(role.Bytes())
	hasRole := acc.Permissions().HasRole(roleS)
	permInt := byteFromBool(hasRole)
	logger.Trace.Log("function", "hasRole", "address", address.String(),
		"role", roleS,
		"has_role", hasRole)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func addRole(stateWriter state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, role := returnTwoArgs(args)
	address := acm.AddressFromWord256(addrWord256)
	acc, err := state.GetMutableAccount(stateWriter, address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", address)
	}
	roleS := string(role.Bytes())
	roleAdded := acc.MutablePermissions().AddRole(roleS)
	permInt := byteFromBool(roleAdded)
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "addRole", "address", address.String(),
		"role", roleS,
		"role_added", roleAdded)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func removeRole(stateWriter state.Writer, caller acm.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, role := returnTwoArgs(args)
	address := acm.AddressFromWord256(addrWord256)
	acc, err := state.GetMutableAccount(stateWriter, address)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("unknown account %s", address)
	}
	roleS := string(role.Bytes())
	roleRemoved := acc.MutablePermissions().RmRole(roleS)
	permInt := byteFromBool(roleRemoved)
	stateWriter.UpdateAccount(acc)
	logger.Trace.Log("function", "removeRole", "address", address.String(),
		"role", roleS,
		"role_removed", roleRemoved)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

//------------------------------------------------------------------------------------------------
// Errors and utility funcs

// Checks if a permission flag is valid (a known base chain or snative permission)
func ValidPermN(n ptypes.PermFlag) bool {
	return n <= permission.AllPermFlags
}

// Get the global BasePermissions
func globalPerms(stateWriter state.Writer) ptypes.BasePermissions {
	return state.GlobalAccountPermissions(stateWriter).Base
}

// Compute the effective permissions from an acm.Account's BasePermissions by
// taking the bitwise or with the global BasePermissions resultant permissions
func effectivePermBytes(basePerms ptypes.BasePermissions,
	globalPerms ptypes.BasePermissions) []byte {
	return permBytes(basePerms.ResultantPerms() | globalPerms.ResultantPerms())
}

func permBytes(basePerms ptypes.PermFlag) []byte {
	return Uint64ToWord256(uint64(basePerms)).Bytes()
}

// CONTRACT: length has already been checked
func returnTwoArgs(args []byte) (a Word256, b Word256) {
	copy(a[:], args[:32])
	copy(b[:], args[32:64])
	return
}

// CONTRACT: length has already been checked
func returnThreeArgs(args []byte) (a Word256, b Word256, c Word256) {
	copy(a[:], args[:32])
	copy(b[:], args[32:64])
	copy(c[:], args[64:96])
	return
}

func byteFromBool(b bool) byte {
	if b {
		return 0x1
	}
	return 0x0
}
