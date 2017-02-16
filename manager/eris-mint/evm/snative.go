package vm

import (
	"fmt"

	"github.com/eris-ltd/eris-db/common/sanity"
	"github.com/eris-ltd/eris-db/manager/eris-mint/evm/sha3"
	ptypes "github.com/eris-ltd/eris-db/permission/types"
	. "github.com/eris-ltd/eris-db/word256"

	"bytes"
	"strings"
)

//------------------------------------------------------------------------------------------------
// Registered SNative contracts

type SNativeContractDescription struct {
	Comment   string
	Name      string
	functions map[FuncID]SNativeFuncDescription
}

type SNativeFuncDescription struct {
	Comment  string
	Name     string
	Args     []SolidityArg
	Return   SolidityReturn
	PermFlag ptypes.PermFlag
	F        NativeContract
}

type SolidityType string

type SolidityArg struct {
	Name string
	Type SolidityType
}

type SolidityReturn struct {
	Name string
	Type SolidityType
}

const (
	// We don't need to be exhaustive here, just make what we used strongly typed
	SolidityAddress SolidityType = "address"
	SolidityInt     SolidityType = "int"
	SolidityUint64  SolidityType = "uint64"
	SolidityBytes32 SolidityType = "bytes32"
	SolidityString  SolidityType = "string"
	SolidityBool    SolidityType = "bool"
)

func registerSNativeContracts() {
	for _, contract := range SNativeContracts() {
		registeredNativeContracts[contract.Address()] = contract.Dispatch
	}
}

// Returns a map of all SNative contracts defined indexed by name
func SNativeContracts() map[string]SNativeContractDescription {
	contracts := []SNativeContractDescription{
		NewSNativeContract(`
		* Interface for managing Secure Native authorizations.
		* @dev This Solidity interface describes the functions exposed by the SNative permissions layer in the Monax blockchain (ErisDB).
		`,
			"permissions_contract",
			SNativeFuncDescription{`
			* @notice Adds a role to an account
			* @param _account account
			* @param _role role
			* @return result whether role was added
			`,
				"add_role",
				[]SolidityArg{
					arg("_account", SolidityAddress),
					arg("_role", SolidityBytes32),
				},
				ret("result", SolidityBool),
				ptypes.AddRole,
				add_role},

			SNativeFuncDescription{`
			* @notice Indicates whether an account has a role
			* @param _account account
			* @param _role role
			* @return result whether account has role
			`,
				"has_role",
				[]SolidityArg{
					arg("_account", SolidityAddress),
					arg("_role", SolidityBytes32),
				},
				ret("result", SolidityBool),
				ptypes.HasRole,
				has_role},

			SNativeFuncDescription{`
			* @notice Removes a role from an account
			* @param _account account
			* @param _role role
			* @return result whether role was removed
			`,
				"rm_role",
				[]SolidityArg{
					arg("_account", SolidityAddress),
					arg("_role", SolidityBytes32),
				},
				ret("result", SolidityBool),
				ptypes.RmRole,
				rm_role},

			SNativeFuncDescription{`
			* @notice Sets a base authorization for an account
			* @param _account account
			* @param _authorization base authorization
			* @param _value value of base authorization
			* @return result value passed in
			`,
				"set_base",
				[]SolidityArg{arg("_account", SolidityAddress),
					arg("_authorization", SolidityInt),
					arg("_value", SolidityInt)},
				ret("result", SolidityBool),
				ptypes.SetBase,
				set_base},

			SNativeFuncDescription{`
			* @notice Indicates whether an account has a base authorization
			* @param _account account
			* @param _authorization base authorization
			* @return result whether account has base authorization set
			`,
				"has_base",
				[]SolidityArg{arg("_account", SolidityAddress),
					arg("_authorization", SolidityInt)},
				ret("result", SolidityBool),
				ptypes.HasBase,
				has_base},

			SNativeFuncDescription{`
			* @notice Sets a base authorization for an account to the global (default) value of the base authorization
      * @param _account account
      * @param _authorization base authorization
      * @return authorization base authorization passed in
      `,
				"unset_base",
				[]SolidityArg{arg("_account", SolidityAddress),
					arg("_authorization", SolidityInt)},
				ret("authorization", SolidityInt),
				ptypes.UnsetBase,
				unset_base},

			SNativeFuncDescription{`
			* @notice Sets global (default) value for a base authorization
			* @param _account account
			* @param _authorization base authorization
			* @param _value value of base authorization
			* @return authorization base authorization passed in
			`,
				"set_global",
				[]SolidityArg{arg("_account", SolidityAddress),
					arg("_authorization", SolidityInt),
					arg("_value", SolidityInt)},
				ret("authorization", SolidityInt),
				ptypes.SetGlobal,
				set_global},
		),
	}

	contractMap := make(map[string]SNativeContractDescription, len(contracts))
	for _, contract := range contracts {
		contractMap[contract.Name] = contract
	}
	return contractMap
}

//-----------------------------------------------------------------------------
// snative are native contracts that can access and modify an account's permissions

func NewSNativeContract(comment, name string, functions ...SNativeFuncDescription) SNativeContractDescription {
	fs := make(map[FuncID]SNativeFuncDescription, len(functions))
	for _, f := range functions {
		fid := f.ID()
		otherF, ok := fs[fid]
		if ok {
			panic(fmt.Errorf("Function with ID %x already defined: %s", fid,
				otherF))
		}
		fs[fid] = f
	}
	return SNativeContractDescription{
		Comment:   comment,
		Name:      name,
		functions: fs,
	}
}

func (contract *SNativeContractDescription) Address() Word256 {
	return LeftPadWord256([]byte(contract.Name))
}

func (contract *SNativeContractDescription) FunctionByID(id FuncID) (*SNativeFuncDescription, error) {
	f, ok := contract.functions[id]
	if !ok {
		return nil,
			fmt.Errorf("Unknown SNative function with ID %x", id)
	}
	return &f, nil
}

func (contract *SNativeContractDescription) FunctionByName(name string) (*SNativeFuncDescription, error) {
	for _, f := range contract.functions {
		if f.Name == name {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("Unknown SNative function with name %s", name)
}

func (contract *SNativeContractDescription) Functions() []SNativeFuncDescription {
	fs := make([]SNativeFuncDescription, 0, len(contract.functions))
	for _, f := range contract.functions {
		fs = append(fs, f)
	}
	return fs
}

func (contract *SNativeContractDescription) Dispatch(appState AppState,
	caller *Account, args []byte, gas *int64) (output []byte, err error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("SNatives dispatch requires a 4-byte function "+
			"identifier but arguments are only %s bytes long", len(args))
	}

	function, err := contract.FunctionByID(firstFourBytes(args))
	if err != nil {
		return nil, err
	}

	remainingArgs := args[4:]

	// check if we have permission to call this function
	if !HasPermission(appState, caller, function.PermFlag) {
		return nil, ErrInvalidPermission{caller.Address, function.Name}
	}

	// ensure there are enough arguments
	if len(remainingArgs) != function.NArgs()*32 {
		return nil, fmt.Errorf("%s() takes %d arguments", function.Name,
			function.NArgs())
	}

	// call the function
	return function.F(appState, caller, remainingArgs, gas)
}

func (contract *SNativeContractDescription) SolidityComment() string {
	return solidityComment(contract.Comment)
}

// Generate solidity code for this SNative contract
func (contract *SNativeContractDescription) Solidity() (string, error) {
	buf := new(bytes.Buffer)
	err := snativeContractTemplate.Execute(buf, contract)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

//
// SNative functions
//
func (function *SNativeFuncDescription) Signature() string {
	argTypes := make([]string, len(function.Args))
	for i, arg := range function.Args {
		argTypes[i] = string(arg.Type)
	}
	return fmt.Sprintf("%s(%s)", function.Name,
		strings.Join(argTypes, ","))
}

func (function *SNativeFuncDescription) ID() FuncID {
	return firstFourBytes(sha3.Sha3([]byte(function.Signature())))
}

func (function *SNativeFuncDescription) NArgs() int {
	return len(function.Args)
}

func (function *SNativeFuncDescription) SolidityArgList() string {
	argList := make([]string, len(function.Args))
	for i, arg := range function.Args {
		argList[i] = fmt.Sprintf("%s %s", arg.Type, arg.Name)
	}
	return strings.Join(argList, ", ")
}

func (function *SNativeFuncDescription) SolidityComment() string {
	return solidityComment(function.Comment)
}

func (function *SNativeFuncDescription) Solidity() (string, error) {
	buf := new(bytes.Buffer)
	err := snativeFuncTemplate.Execute(buf, function)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func solidityComment(comment string) string {
	commentLines := make([]string, 0, 5)
	for _, line := range strings.Split(comment, "\n") {
		trimLine := strings.TrimLeft(line, " \t\n")
		if trimLine != "" {
			commentLines = append(commentLines, trimLine)
		}
	}
	return strings.Join(commentLines, "\n")
}

func arg(name string, solidityType SolidityType) SolidityArg {
	return SolidityArg{
		Name: name,
		Type: solidityType,
	}
}

func ret(name string, solidityType SolidityType) SolidityReturn {
	return SolidityReturn{
		Name: name,
		Type: solidityType,
	}
}

// Permission function defintions

// TODO: catch errors, log em, return 0s to the vm (should some errors cause exceptions though?)
func has_base(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	addr, permNum := returnTwoArgs(args)
	vmAcc := appState.GetAccount(addr)
	if vmAcc == nil {
		return nil, fmt.Errorf("Unknown account %X", addr)
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum)) // already shifted
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	permInt := byteFromBool(HasPermission(appState, vmAcc, permN))
	dbg.Printf("snative.hasBasePerm(0x%X, %b) = %v\n", addr.Postfix(20), permN, permInt)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func set_base(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	addr, permNum, perm := returnThreeArgs(args)
	vmAcc := appState.GetAccount(addr)
	if vmAcc == nil {
		return nil, fmt.Errorf("Unknown account %X", addr)
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum))
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	permV := !perm.IsZero()
	if err = vmAcc.Permissions.Base.Set(permN, permV); err != nil {
		return nil, err
	}
	appState.UpdateAccount(vmAcc)
	dbg.Printf("snative.setBasePerm(0x%X, %b, %v)\n", addr.Postfix(20), permN, permV)
	return perm.Bytes(), nil
}

func unset_base(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	addr, permNum := returnTwoArgs(args)
	vmAcc := appState.GetAccount(addr)
	if vmAcc == nil {
		return nil, fmt.Errorf("Unknown account %X", addr)
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum))
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	if err = vmAcc.Permissions.Base.Unset(permN); err != nil {
		return nil, err
	}
	appState.UpdateAccount(vmAcc)
	dbg.Printf("snative.unsetBasePerm(0x%X, %b)\n", addr.Postfix(20), permN)
	return permNum.Bytes(), nil
}

func set_global(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	permNum, perm := returnTwoArgs(args)
	vmAcc := appState.GetAccount(ptypes.GlobalPermissionsAddress256)
	if vmAcc == nil {
		sanity.PanicSanity("cant find the global permissions account")
	}
	permN := ptypes.PermFlag(Uint64FromWord256(permNum))
	if !ValidPermN(permN) {
		return nil, ptypes.ErrInvalidPermission(permN)
	}
	permV := !perm.IsZero()
	if err = vmAcc.Permissions.Base.Set(permN, permV); err != nil {
		return nil, err
	}
	appState.UpdateAccount(vmAcc)
	dbg.Printf("snative.setGlobalPerm(%b, %v)\n", permN, permV)
	return perm.Bytes(), nil
}

func has_role(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	addr, role := returnTwoArgs(args)
	vmAcc := appState.GetAccount(addr)
	if vmAcc == nil {
		return nil, fmt.Errorf("Unknown account %X", addr)
	}
	roleS := string(role.Bytes())
	permInt := byteFromBool(vmAcc.Permissions.HasRole(roleS))
	dbg.Printf("snative.hasRole(0x%X, %s) = %v\n", addr.Postfix(20), roleS, permInt > 0)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func add_role(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	addr, role := returnTwoArgs(args)
	vmAcc := appState.GetAccount(addr)
	if vmAcc == nil {
		return nil, fmt.Errorf("Unknown account %X", addr)
	}
	roleS := string(role.Bytes())
	permInt := byteFromBool(vmAcc.Permissions.AddRole(roleS))
	appState.UpdateAccount(vmAcc)
	dbg.Printf("snative.addRole(0x%X, %s) = %v\n", addr.Postfix(20), roleS, permInt > 0)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func rm_role(appState AppState, caller *Account, args []byte, gas *int64) (output []byte, err error) {
	addr, role := returnTwoArgs(args)
	vmAcc := appState.GetAccount(addr)
	if vmAcc == nil {
		return nil, fmt.Errorf("Unknown account %X", addr)
	}
	roleS := string(role.Bytes())
	permInt := byteFromBool(vmAcc.Permissions.RmRole(roleS))
	appState.UpdateAccount(vmAcc)
	dbg.Printf("snative.rmRole(0x%X, %s) = %v\n", addr.Postfix(20), roleS, permInt > 0)
	return LeftPadWord256([]byte{permInt}).Bytes(), nil
}

//------------------------------------------------------------------------------------------------
// Errors and utility funcs

type ErrInvalidPermission struct {
	Address Word256
	SNative string
}

func (e ErrInvalidPermission) Error() string {
	return fmt.Sprintf("Account %X does not have permission snative.%s", e.Address.Postfix(20), e.SNative)
}

// Checks if a permission flag is valid (a known base chain or snative permission)
func ValidPermN(n ptypes.PermFlag) bool {
	if n > ptypes.TopPermFlag {
		return false
	}
	return true
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

func firstFourBytes(byteSlice []byte) [4]byte {
	var bs [4]byte
	copy(bs[:], byteSlice[:4])
	return bs
}
