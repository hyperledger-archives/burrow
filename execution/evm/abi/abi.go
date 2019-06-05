package abi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	hex "github.com/tmthrgd/go-hex"

	burrow_binary "github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"

	"os"
	"path"
	"path/filepath"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/logging"
)

// Variable exist to unpack return values into, so have both the return
// value and its name
type Variable struct {
	Name  string
	Value string
}

func init() {
	var err error
	revertAbi, err = ReadSpec([]byte(`[{"name":"Error","type":"function","outputs":[{"type":"string"}],"inputs":[{"type":"string"}]}]`))
	if err != nil {
		panic(fmt.Sprintf("internal error: failed to build revert abi: %v", err))
	}
}

// revertAbi exists to decode reverts. Any contract function call fail using revert(), assert() or require().
// If a function exits this way, the this hardcoded ABI will be used.
var revertAbi *Spec

// EncodeFunctionCallFromFile ABI encodes a function call based on ABI in file, and the
// arguments specified as strings.
// The abiFileName specifies the name of the ABI file, and abiPath the path where it can be found.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCallFromFile(abiFileName, abiPath, funcName string, logger *logging.Logger, args ...interface{}) ([]byte, *FunctionSpec, error) {
	abiSpecBytes, err := readAbi(abiPath, abiFileName, logger)
	if err != nil {
		return []byte{}, nil, err
	}

	return EncodeFunctionCall(abiSpecBytes, funcName, logger, args...)
}

// EncodeFunctionCall ABI encodes a function call based on ABI in string abiData
// and the arguments specified as strings.
// The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func EncodeFunctionCall(abiData, funcName string, logger *logging.Logger, args ...interface{}) ([]byte, *FunctionSpec, error) {
	logger.TraceMsg("Packing Call via ABI",
		"spec", abiData,
		"function", funcName,
		"arguments", fmt.Sprintf("%v", args),
	)

	abiSpec, err := ReadSpec([]byte(abiData))
	if err != nil {
		logger.InfoMsg("Failed to decode abi spec",
			"abi", abiData,
			"error", err.Error(),
		)
		return nil, nil, err
	}

	packedBytes, funcSpec, err := abiSpec.Pack(funcName, args...)
	if err != nil {
		logger.InfoMsg("Failed to encode abi spec",
			"abi", abiData,
			"error", err.Error(),
		)
		return nil, nil, err
	}

	return packedBytes, funcSpec, nil
}

// DecodeFunctionReturnFromFile ABI decodes the return value from a contract function call.
func DecodeFunctionReturnFromFile(abiLocation, binPath, funcName string, resultRaw []byte, logger *logging.Logger) ([]*Variable, error) {
	abiSpecBytes, err := readAbi(binPath, abiLocation, logger)
	if err != nil {
		return nil, err
	}
	logger.TraceMsg("ABI Specification (Decode)", "spec", abiSpecBytes)

	// Unpack the result
	return DecodeFunctionReturn(abiSpecBytes, funcName, resultRaw)
}

func DecodeFunctionReturn(abiData, name string, data []byte) ([]*Variable, error) {
	abiSpec, err := ReadSpec([]byte(abiData))
	if err != nil {
		return nil, err
	}

	var args []Argument

	if name == "" {
		args = abiSpec.Constructor.Outputs
	} else {
		if _, ok := abiSpec.Functions[name]; ok {
			args = abiSpec.Functions[name].Outputs
		} else {
			args = abiSpec.Fallback.Outputs
		}
	}

	if args == nil {
		return nil, fmt.Errorf("no such function")
	}
	vars := make([]*Variable, len(args))

	if len(args) == 0 {
		return nil, nil
	}

	vals := make([]interface{}, len(args))
	for i := range vals {
		vals[i] = new(string)
	}
	err = Unpack(args, data, vals...)
	if err != nil {
		return nil, err
	}

	for i, a := range args {
		if a.Name != "" {
			vars[i] = &Variable{Name: a.Name, Value: *(vals[i].(*string))}
		} else {
			vars[i] = &Variable{Name: fmt.Sprintf("%d", i), Value: *(vals[i].(*string))}
		}
	}

	return vars, nil
}

func readAbi(root, contract string, logger *logging.Logger) (string, error) {
	p := path.Join(root, stripHex(contract))
	if _, err := os.Stat(p); err != nil {
		logger.TraceMsg("abifile not found", "tried", p)
		p = path.Join(root, stripHex(contract)+".bin")
		if _, err = os.Stat(p); err != nil {
			logger.TraceMsg("abifile not found", "tried", p)
			return "", fmt.Errorf("abi doesn't exist for =>\t%s", p)
		}
	}
	logger.TraceMsg("Found ABI file", "path", p)
	sol, err := compile.LoadSolidityContract(p)
	if err != nil {
		return "", err
	}
	return string(sol.Abi), nil
}

// LoadPath loads one abi file or finds all files in a directory
func LoadPath(abiFileOrDirs ...string) (*Spec, error) {
	if len(abiFileOrDirs) == 0 {
		return &Spec{}, fmt.Errorf("no ABI file or directory provided")
	}

	specs := make([]*Spec, 0)

	for _, dir := range abiFileOrDirs {
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error returned while walking abiDir '%s': %v", dir, err)
			}
			ext := filepath.Ext(path)
			if fi.IsDir() || !(ext == ".bin" || ext == ".abi") {
				return nil
			}
			if err == nil {
				abiSpc, err := ReadSpecFile(path)
				if err != nil {
					return errors.Wrap(err, "Error parsing abi file "+path)
				}
				specs = append(specs, abiSpc)
			}
			return nil
		})
		if err != nil {
			return &Spec{}, err
		}
	}
	return MergeSpec(specs), nil
}

func stripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			s = s[2:]
			if len(s)%2 != 0 {
				s = "0" + s
			}
			return s
		}
	}
	return s
}

// Argument is a decoded function parameter, return or event field
type Argument struct {
	Name        string
	EVM         EVMType
	IsArray     bool
	Indexed     bool
	Hashed      bool
	ArrayLength uint64
}

// FunctionIDSize is the length of the function selector
const FunctionIDSize = 4

type FunctionID [FunctionIDSize]byte

// EventIDSize is the length of the event selector
const EventIDSize = 32

type EventID [EventIDSize]byte

func (e EventID) String() string {
	return hex.EncodeUpperToString(e[:])
}

type FunctionSpec struct {
	FunctionID FunctionID
	Constant   bool
	Inputs     []Argument
	Outputs    []Argument
}

type EventSpec struct {
	EventID   EventID
	Inputs    []Argument
	Name      string
	Anonymous bool
}

// Spec is the ABI for contract decoded.
type Spec struct {
	Constructor  FunctionSpec
	Fallback     FunctionSpec
	Functions    map[string]FunctionSpec
	EventsByName map[string]EventSpec
	EventsByID   map[EventID]EventSpec
}

type argumentJSON struct {
	Name       string
	Type       string
	Components []argumentJSON
	Indexed    bool
}

type specJSON struct {
	Name            string
	Type            string
	Inputs          []argumentJSON
	Outputs         []argumentJSON
	Constant        bool
	Payable         bool
	StateMutability string
	Anonymous       bool
}

func readArgSpec(argsJ []argumentJSON) ([]Argument, error) {
	args := make([]Argument, len(argsJ))
	var err error

	for i, a := range argsJ {
		args[i].Name = a.Name
		args[i].Indexed = a.Indexed

		baseType := a.Type
		isArray := regexp.MustCompile(`(.*)\[([0-9]+)\]`)
		m := isArray.FindStringSubmatch(a.Type)
		if m != nil {
			args[i].IsArray = true
			args[i].ArrayLength, err = strconv.ParseUint(m[2], 10, 32)
			if err != nil {
				return nil, err
			}
			baseType = m[1]
		} else if strings.HasSuffix(a.Type, "[]") {
			args[i].IsArray = true
			baseType = strings.TrimSuffix(a.Type, "[]")
		}

		isM := regexp.MustCompile("(bytes|uint|int)([0-9]+)")
		m = isM.FindStringSubmatch(baseType)
		if m != nil {
			M, err := strconv.ParseUint(m[2], 10, 32)
			if err != nil {
				return nil, err
			}
			switch m[1] {
			case "bytes":
				if M < 1 || M > 32 {
					return nil, fmt.Errorf("bytes%d is not valid type", M)
				}
				args[i].EVM = EVMBytes{M}
			case "uint":
				if M < 8 || M > 256 || (M%8) != 0 {
					return nil, fmt.Errorf("uint%d is not valid type", M)
				}
				args[i].EVM = EVMUint{M}
			case "int":
				if M < 8 || M > 256 || (M%8) != 0 {
					return nil, fmt.Errorf("uint%d is not valid type", M)
				}
				args[i].EVM = EVMInt{M}
			}
			continue
		}

		isMxN := regexp.MustCompile("(fixed|ufixed)([0-9]+)x([0-9]+)")
		m = isMxN.FindStringSubmatch(baseType)
		if m != nil {
			M, err := strconv.ParseUint(m[2], 10, 32)
			if err != nil {
				return nil, err
			}
			N, err := strconv.ParseUint(m[3], 10, 32)
			if err != nil {
				return nil, err
			}
			if M < 8 || M > 256 || (M%8) != 0 {
				return nil, fmt.Errorf("%s is not valid type", baseType)
			}
			if N == 0 || N > 80 {
				return nil, fmt.Errorf("%s is not valid type", baseType)
			}
			if m[1] == "fixed" {
				args[i].EVM = EVMFixed{N: N, M: M, signed: true}
			} else if m[1] == "ufixed" {
				args[i].EVM = EVMFixed{N: N, M: M, signed: false}
			} else {
				panic(m[1])
			}
			continue
		}
		switch baseType {
		case "uint":
			args[i].EVM = EVMUint{M: 256}
		case "int":
			args[i].EVM = EVMInt{M: 256}
		case "address":
			args[i].EVM = EVMAddress{}
		case "bool":
			args[i].EVM = EVMBool{}
		case "fixed":
			args[i].EVM = EVMFixed{M: 128, N: 8, signed: true}
		case "ufixed":
			args[i].EVM = EVMFixed{M: 128, N: 8, signed: false}
		case "bytes":
			args[i].EVM = EVMBytes{M: 0}
		case "string":
			args[i].EVM = EVMString{}
		default:
			// Assume it is a type of Contract
			args[i].EVM = EVMAddress{}
		}
	}

	return args, nil
}

// ReadSpec takes an ABI and decodes it for futher use
func ReadSpec(specBytes []byte) (*Spec, error) {
	var specJ []specJSON
	err := json.Unmarshal(specBytes, &specJ)
	if err != nil {
		// The abi spec file might a bin file, with the Abi under the Abi field in json
		var binFile struct {
			Abi []specJSON
		}
		err = json.Unmarshal(specBytes, &binFile)
		if err != nil {
			return nil, err
		}
		specJ = binFile.Abi
	}

	abiSpec := Spec{
		EventsByName: make(map[string]EventSpec),
		EventsByID:   make(map[EventID]EventSpec),
		Functions:    make(map[string]FunctionSpec),
	}

	for _, s := range specJ {
		switch s.Type {
		case "constructor":
			abiSpec.Constructor.Inputs, err = readArgSpec(s.Inputs)
			if err != nil {
				return nil, err
			}
		case "fallback":
			abiSpec.Fallback.Inputs = make([]Argument, 0)
			abiSpec.Fallback.Outputs = make([]Argument, 0)
		case "event":
			inputs, err := readArgSpec(s.Inputs)
			if err != nil {
				return nil, err
			}
			// Get signature before we deal with hashed types
			sig := Signature(s.Name, inputs)
			for i := range inputs {
				if inputs[i].Indexed && inputs[i].EVM.Dynamic() {
					// For Dynamic types, the hash is stored in stead
					inputs[i].EVM = EVMBytes{M: 32}
					inputs[i].Hashed = true
				}
			}
			ev := EventSpec{Name: s.Name, EventID: GetEventID(sig), Inputs: inputs, Anonymous: s.Anonymous}
			abiSpec.EventsByName[ev.Name] = ev
			abiSpec.EventsByID[ev.EventID] = ev
		case "function":
			inputs, err := readArgSpec(s.Inputs)
			if err != nil {
				return nil, err
			}
			outputs, err := readArgSpec(s.Outputs)
			if err != nil {
				return nil, err
			}
			fs := FunctionSpec{Inputs: inputs, Outputs: outputs, Constant: s.Constant}
			fs.SetFunctionID(s.Name)
			abiSpec.Functions[s.Name] = fs
		}
	}

	return &abiSpec, nil
}

// ReadSpecFile reads an ABI file from a file
func ReadSpecFile(filename string) (*Spec, error) {
	specBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ReadSpec(specBytes)
}

// MergeSpec takes multiple Specs and merges them into once structure. Note that
// the same function name or event name can occur in different abis, so there might be
// some information loss.
func MergeSpec(abiSpec []*Spec) *Spec {
	newSpec := Spec{
		EventsByName: make(map[string]EventSpec),
		EventsByID:   make(map[EventID]EventSpec),
		Functions:    make(map[string]FunctionSpec),
	}

	for _, s := range abiSpec {
		for n, f := range s.Functions {
			newSpec.Functions[n] = f
		}

		// Different Abis can have the Event name, but with a different signature
		// Loop over the signatures, as these are less likely to have collisions
		for _, e := range s.EventsByID {
			newSpec.EventsByName[e.Name] = e
			newSpec.EventsByID[e.EventID] = e
		}
	}

	return &newSpec
}

func typeFromReflect(v reflect.Type) Argument {
	arg := Argument{Name: v.Name()}

	if v == reflect.TypeOf(crypto.Address{}) {
		arg.EVM = EVMAddress{}
	} else if v == reflect.TypeOf(big.Int{}) {
		arg.EVM = EVMInt{M: 256}
	} else {
		if v.Kind() == reflect.Array {
			arg.IsArray = true
			arg.ArrayLength = uint64(v.Len())
			v = v.Elem()
		} else if v.Kind() == reflect.Slice {
			arg.IsArray = true
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Bool:
			arg.EVM = EVMBool{}
		case reflect.String:
			arg.EVM = EVMString{}
		case reflect.Uint64:
			arg.EVM = EVMUint{M: 64}
		case reflect.Int64:
			arg.EVM = EVMInt{M: 64}
		default:
			panic(fmt.Sprintf("no mapping for type %v", v.Kind()))
		}
	}

	return arg
}

// SpecFromStructReflect generates a FunctionSpec where the arguments and return values are
// described a struct. Both args and rets should be set to the return value of reflect.TypeOf()
// with the respective struct as an argument.
func SpecFromStructReflect(fname string, args reflect.Type, rets reflect.Type) *FunctionSpec {
	s := FunctionSpec{
		Inputs:  make([]Argument, args.NumField()),
		Outputs: make([]Argument, rets.NumField()),
	}
	for i := 0; i < args.NumField(); i++ {
		f := args.Field(i)
		a := typeFromReflect(f.Type)
		a.Name = f.Name
		s.Inputs[i] = a
	}
	for i := 0; i < rets.NumField(); i++ {
		f := rets.Field(i)
		a := typeFromReflect(f.Type)
		a.Name = f.Name
		s.Outputs[i] = a
	}
	s.SetFunctionID(fname)

	return &s
}

func SpecFromFunctionReflect(fname string, v reflect.Value, skipIn, skipOut int) *FunctionSpec {
	t := v.Type()

	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("%s is not a function", t.Name()))
	}

	s := FunctionSpec{}
	s.Inputs = make([]Argument, t.NumIn()-skipIn)
	s.Outputs = make([]Argument, t.NumOut()-skipOut)

	for i := range s.Inputs {
		s.Inputs[i] = typeFromReflect(t.In(i + skipIn))
	}

	for i := range s.Outputs {
		s.Outputs[i] = typeFromReflect(t.Out(i))
	}

	s.SetFunctionID(fname)
	return &s
}

func Signature(name string, args []Argument) (sig string) {
	sig = name + "("
	for i, a := range args {
		if i > 0 {
			sig += ","
		}
		sig += a.EVM.GetSignature()
		if a.IsArray {
			if a.ArrayLength > 0 {
				sig += fmt.Sprintf("[%d]", a.ArrayLength)
			} else {
				sig += "[]"
			}
		}
	}
	sig += ")"
	return
}

func (functionSpec *FunctionSpec) SetFunctionID(functionName string) {
	sig := Signature(functionName, functionSpec.Inputs)
	functionSpec.FunctionID = GetFunctionID(sig)
}

func (fs FunctionID) Bytes() []byte {
	return fs[:]
}

func GetFunctionID(signature string) (id FunctionID) {
	hash := sha3.NewKeccak256()
	hash.Write([]byte(signature))
	copy(id[:], hash.Sum(nil)[:4])
	return
}

func (fs EventID) Bytes() []byte {
	return fs[:]
}

func GetEventID(signature string) (id EventID) {
	hash := sha3.NewKeccak256()
	hash.Write([]byte(signature))
	copy(id[:], hash.Sum(nil))
	return
}

// UnpackRevert decodes the revert reason if a contract called revert. If no
// reason was given, message will be nil else it will point to the string
func UnpackRevert(data []byte) (message *string, err error) {
	if len(data) > 0 {
		var msg string
		err = revertAbi.UnpackWithID(data, &msg)
		message = &msg
	}
	return
}

// UnpackEvent decodes all the fields in an event (indexed topic fields or not)
func UnpackEvent(eventSpec *EventSpec, topics []burrow_binary.Word256, data []byte, args ...interface{}) error {
	// First unpack the topic fields
	topicIndex := 0
	if !eventSpec.Anonymous {
		topicIndex++
	}

	for i, a := range eventSpec.Inputs {
		if a.Indexed {
			_, err := a.EVM.unpack(topics[topicIndex].Bytes(), 0, args[i])
			if err != nil {
				return err
			}
			topicIndex++
		}
	}

	// Now unpack the other fields. unpack will step over any indexed fields
	return unpack(eventSpec.Inputs, data, func(i int) interface{} {
		return args[i]
	})
}

// Unpack decodes the return values from a function call
func (abiSpec *Spec) Unpack(data []byte, fname string, args ...interface{}) error {
	var funcSpec FunctionSpec
	var argSpec []Argument
	if fname != "" {
		if _, ok := abiSpec.Functions[fname]; ok {
			funcSpec = abiSpec.Functions[fname]
		} else {
			funcSpec = abiSpec.Fallback
		}
	} else {
		funcSpec = abiSpec.Constructor
	}

	argSpec = funcSpec.Outputs

	if argSpec == nil {
		return fmt.Errorf("Unknown function %s", fname)
	}

	return unpack(argSpec, data, func(i int) interface{} {
		return args[i]
	})
}

func (abiSpec *Spec) UnpackWithID(data []byte, args ...interface{}) error {
	var argSpec []Argument

	var id FunctionID
	copy(id[:], data)
	for _, fspec := range abiSpec.Functions {
		if id == fspec.FunctionID {
			argSpec = fspec.Outputs
		}
	}

	if argSpec == nil {
		return fmt.Errorf("Unknown function %x", id)
	}

	return unpack(argSpec, data[4:], func(i int) interface{} {
		return args[i]
	})
}

// Pack ABI encodes a function call. The fname specifies which function should called, if
// it doesn't exist exist the fallback function will be called. If fname is the empty
// string, the constructor is called. The arguments must be specified in args. The count
// must match the function being called.
// Returns the ABI encoded function call, whether the function is constant according
// to the ABI (which means it does not modified contract state)
func (abiSpec *Spec) Pack(fname string, args ...interface{}) ([]byte, *FunctionSpec, error) {
	var funcSpec FunctionSpec
	var argSpec []Argument
	if fname != "" {
		if _, ok := abiSpec.Functions[fname]; ok {
			funcSpec = abiSpec.Functions[fname]
		} else {
			funcSpec = abiSpec.Fallback
		}
	} else {
		funcSpec = abiSpec.Constructor
	}

	argSpec = funcSpec.Inputs

	if argSpec == nil {
		if fname == "" {
			return nil, nil, fmt.Errorf("Contract does not have a constructor")
		}

		return nil, nil, fmt.Errorf("Unknown function %s", fname)
	}

	packed := make([]byte, 0)

	if fname != "" {
		packed = funcSpec.FunctionID[:]
	}

	packedArgs, err := Pack(argSpec, args...)
	if err != nil {
		return nil, nil, err
	}

	return append(packed, packedArgs...), &funcSpec, nil
}

func PackIntoStruct(argSpec []Argument, st interface{}) ([]byte, error) {
	v := reflect.ValueOf(st)

	fields := v.NumField()
	if fields != len(argSpec) {
		return nil, fmt.Errorf("%d arguments expected, %d received", len(argSpec), fields)
	}

	return pack(argSpec, func(i int) interface{} {
		return v.Field(i).Interface()
	})
}

func Pack(argSpec []Argument, args ...interface{}) ([]byte, error) {
	if len(args) != len(argSpec) {
		return nil, fmt.Errorf("%d arguments expected, %d received", len(argSpec), len(args))
	}

	return pack(argSpec, func(i int) interface{} {
		return args[i]
	})
}

func pack(argSpec []Argument, getArg func(int) interface{}) ([]byte, error) {
	packed := make([]byte, 0)
	packedDynamic := []byte{}
	fixedSize := 0
	// Anything dynamic is stored after the "fixed" block. For the dynamic types, the fixed
	// block contains byte offsets to the data. We need to know the length of the fixed
	// block, so we can calcute the offsets
	for _, a := range argSpec {
		if a.IsArray {
			if a.ArrayLength > 0 {
				fixedSize += ElementSize * int(a.ArrayLength)
			} else {
				fixedSize += ElementSize
			}
		} else {
			fixedSize += ElementSize
		}
	}

	addArg := func(v interface{}, a Argument) error {
		var b []byte
		var err error
		if a.EVM.Dynamic() {
			offset := EVMUint{M: 256}
			b, _ = offset.pack(fixedSize)
			d, err := a.EVM.pack(v)
			if err != nil {
				return err
			}
			fixedSize += len(d)
			packedDynamic = append(packedDynamic, d...)
		} else {
			b, err = a.EVM.pack(v)
			if err != nil {
				return err
			}
		}
		packed = append(packed, b...)
		return nil
	}

	for i, as := range argSpec {
		a := getArg(i)
		if as.IsArray {
			s, ok := a.(string)
			if ok && s[0:1] == "[" && s[len(s)-1:] == "]" {
				a = strings.Split(s[1:len(s)-1], ",")
			}

			val := reflect.ValueOf(a)
			if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
				return nil, fmt.Errorf("argument %d should be array or slice, not %s", i, val.Kind().String())
			}

			if as.ArrayLength > 0 {
				if as.ArrayLength != uint64(val.Len()) {
					return nil, fmt.Errorf("argumment %d should be array of %d, not %d", i, as.ArrayLength, val.Len())
				}

				for n := 0; n < val.Len(); n++ {
					err := addArg(val.Index(n).Interface(), as)
					if err != nil {
						return nil, err
					}
				}
			} else {
				// dynamic array
				offset := EVMUint{M: 256}
				b, _ := offset.pack(fixedSize)
				packed = append(packed, b...)
				fixedSize += len(b)

				// store length
				b, _ = offset.pack(val.Len())
				packedDynamic = append(packedDynamic, b...)
				for n := 0; n < val.Len(); n++ {
					d, err := as.EVM.pack(val.Index(n).Interface())
					if err != nil {
						return nil, err
					}
					packedDynamic = append(packedDynamic, d...)
				}
			}
		} else {
			err := addArg(a, as)
			if err != nil {
				return nil, err
			}
		}
	}

	return append(packed, packedDynamic...), nil
}

func GetPackingTypes(args []Argument) []interface{} {
	res := make([]interface{}, len(args))

	for i, a := range args {
		if a.IsArray {
			t := reflect.TypeOf(a.EVM.getGoType())
			res[i] = reflect.MakeSlice(reflect.SliceOf(t), int(a.ArrayLength), 0).Interface()
		} else {
			res[i] = a.EVM.getGoType()
		}
	}

	return res
}

func UnpackIntoStruct(argSpec []Argument, data []byte, st interface{}) error {
	v := reflect.ValueOf(st).Elem()
	return unpack(argSpec, data, func(i int) interface{} {
		return v.Field(i).Addr().Interface()
	})
}

func Unpack(argSpec []Argument, data []byte, args ...interface{}) error {
	return unpack(argSpec, data, func(i int) interface{} {
		return args[i]
	})
}

func unpack(argSpec []Argument, data []byte, getArg func(int) interface{}) error {
	offset := 0
	offType := EVMInt{M: 64}

	getPrimitive := func(e interface{}, a Argument) error {
		if a.EVM.Dynamic() {
			var o int64
			l, err := offType.unpack(data, offset, &o)
			if err != nil {
				return err
			}
			offset += l
			_, err = a.EVM.unpack(data, int(o), e)
			if err != nil {
				return err
			}
		} else {
			l, err := a.EVM.unpack(data, offset, e)
			if err != nil {
				return err
			}
			offset += l
		}

		return nil
	}

	for i, a := range argSpec {
		if a.Indexed {
			continue
		}

		arg := getArg(i)
		if a.IsArray {
			var array *[]interface{}

			array, ok := arg.(*[]interface{})
			if !ok {
				if _, ok := arg.(*string); ok {
					// We have been asked to return the value as a string; make intermediate
					// array of strings; we will concatenate after
					intermediate := make([]interface{}, a.ArrayLength)
					for i := range intermediate {
						intermediate[i] = new(string)
					}
					array = &intermediate
				} else {
					return fmt.Errorf("argument %d should be array, slice or string", i)
				}
			}

			if a.ArrayLength > 0 {
				if int(a.ArrayLength) != len(*array) {
					return fmt.Errorf("argument %d should be array or slice of %d elements", i, a.ArrayLength)
				}

				for n := 0; n < len(*array); n++ {
					err := getPrimitive((*array)[n], a)
					if err != nil {
						return err
					}
				}
			} else {
				var o int64
				var length int64

				l, err := offType.unpack(data, offset, &o)
				if err != nil {
					return err
				}

				offset += l
				s, err := offType.unpack(data, int(o), &length)
				if err != nil {
					return err
				}
				o += int64(s)

				intermediate := make([]interface{}, length)

				if _, ok := arg.(*string); ok {
					// We have been asked to return the value as a string; make intermediate
					// array of strings; we will concatenate after
					for i := range intermediate {
						intermediate[i] = new(string)
					}
				} else {
					for i := range intermediate {
						intermediate[i] = a.EVM.getGoType()
					}
				}

				for i := 0; i < int(length); i++ {
					l, err = a.EVM.unpack(data, int(o), intermediate[i])
					if err != nil {
						return err
					}
					o += int64(l)
				}

				array = &intermediate
			}

			// If we were supposed to return a string, convert it back
			if ret, ok := arg.(*string); ok {
				s := "["
				for i, e := range *array {
					if i > 0 {
						s += ","
					}
					s += *(e.(*string))
				}
				s += "]"
				*ret = s
			}
		} else {
			err := getPrimitive(arg, a)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// quick helper padding
func pad(input []byte, size int, left bool) []byte {
	if len(input) >= size {
		return input[:size]
	}
	padded := make([]byte, size)
	if left {
		copy(padded[size-len(input):], input)
	} else {
		copy(padded, input)
	}
	return padded
}
