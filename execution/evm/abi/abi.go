package abi

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unsafe" // just for Sizeof

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/sha3"
)

// EVM Solidity calls and return values are packed into
// pieces of 32 bytes, including a bool (wasting 255 out of 256 bits)
const ElementSize = 32

type EVMType interface {
	GetSignature() string
	getGoType() interface{}
	pack(v interface{}) ([]byte, error)
	unpack(data []byte, offset int, v interface{}) (int, error)
	isDynamic() bool
}

var _ EVMType = (*EVMBool)(nil)

type EVMBool struct {
}

func (e EVMBool) GetSignature() string {
	return "bool"
}

func (e EVMBool) getGoType() interface{} {
	return new(bool)
}

func (e EVMBool) pack(v interface{}) ([]byte, error) {
	var b bool
	arg := reflect.ValueOf(v)
	if arg.Kind() == reflect.String {
		val := arg.String()
		if strings.EqualFold(val, "true") || val == "1" {
			b = true
		} else if strings.EqualFold(val, "false") || val == "0" {
			b = false
		} else {
			return nil, fmt.Errorf("%s is not a valid value for EVM Bool type", val)
		}
	} else if arg.Kind() == reflect.Bool {
		b = arg.Bool()
	} else {
		return nil, fmt.Errorf("%s cannot be converted to EVM Bool type", arg.Kind().String())
	}
	res := make([]byte, ElementSize)
	if b {
		res[ElementSize-1] = 1
	}
	return res, nil
}

func (e EVMBool) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < 32 {
		return 0, fmt.Errorf("not enough data")
	}
	data = data[offset:]
	switch v := v.(type) {
	case *string:
		if data[ElementSize-1] == 1 {
			*v = "true"
		} else if data[ElementSize-1] == 0 {
			*v = "false"
		} else {
			return 0, fmt.Errorf("unexpected value for EVM bool")
		}
	case *int8:
		*v = int8(data[ElementSize-1])
	case *int16:
		*v = int16(data[ElementSize-1])
	case *int32:
		*v = int32(data[ElementSize-1])
	case *int64:
		*v = int64(data[ElementSize-1])
	case *int:
		*v = int(data[ElementSize-1])
	case *uint8:
		*v = uint8(data[ElementSize-1])
	case *uint16:
		*v = uint16(data[ElementSize-1])
	case *uint32:
		*v = uint32(data[ElementSize-1])
	case *uint64:
		*v = uint64(data[ElementSize-1])
	case *uint:
		*v = uint(data[ElementSize-1])
	case *bool:
		*v = data[ElementSize-1] == 1
	default:
		return 0, fmt.Errorf("cannot set type %s for EVM bool", reflect.ValueOf(v).Kind().String())
	}
	return 32, nil
}

func (e EVMBool) isDynamic() bool {
	return false
}

var _ EVMType = (*EVMUint)(nil)

type EVMUint struct {
	M uint64
}

func (e EVMUint) GetSignature() string {
	return fmt.Sprintf("uint%d", e.M)
}

func (e EVMUint) getGoType() interface{} {
	switch e.M {
	case 8:
		return new(uint8)
	case 16:
		return new(uint16)
	case 32:
		return new(uint32)
	case 64:
		return new(uint64)
	default:
		return new(big.Int)
	}
}

func (e EVMUint) pack(v interface{}) ([]byte, error) {
	n := new(big.Int)

	arg := reflect.ValueOf(v)
	switch arg.Kind() {
	case reflect.String:
		_, ok := n.SetString(arg.String(), 0)
		if !ok {
			return nil, fmt.Errorf("Failed to parse `%s", arg.String())
		}
		if n.Sign() < 0 {
			return nil, fmt.Errorf("negative value not allowed for uint%d", e.M)
		}
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uint:
		n.SetUint64(arg.Uint())
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		x := arg.Int()
		if x < 0 {
			return nil, fmt.Errorf("negative value not allowed for uint%d", e.M)
		}
		n.SetInt64(x)
	default:
		t := reflect.TypeOf(new(uint64))
		if reflect.TypeOf(v).ConvertibleTo(t) {
			n.SetUint64(reflect.ValueOf(v).Convert(t).Uint())
		} else {
			return nil, fmt.Errorf("cannot convert type %s to uint%d", arg.Kind().String(), e.M)
		}
	}

	b := n.Bytes()
	if uint64(len(b)) > e.M {
		return nil, fmt.Errorf("value to large for int%d", e.M)
	}
	return pad(b, ElementSize, true), nil
}

func (e EVMUint) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < ElementSize {
		return 0, fmt.Errorf("not enough data")
	}

	data = data[offset:]
	empty := 0
	for empty = 0; empty < ElementSize; empty++ {
		if data[empty] != 0 {
			break
		}
	}

	length := ElementSize - empty

	switch v := v.(type) {
	case *string:
		b := new(big.Int)
		b.SetBytes(data[empty:ElementSize])
		*v = b.String()
	case *big.Int:
		b := new(big.Int)
		b.SetBytes(data[0:ElementSize])
		v = b
	case *uint64:
		maxLen := int(unsafe.Sizeof(new(uint64)))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize])
	case *uint32:
		maxLen := int(unsafe.Sizeof(new(uint32)))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize])
	case *uint16:
		maxLen := int(unsafe.Sizeof(new(uint16)))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint16")
		}
		*v = binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize])
	case *uint8:
		maxLen := 1
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint8")
		}
		*v = uint8(data[31])
	case *int64:
		maxLen := int(unsafe.Sizeof(new(int64)))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int64")
		}
		*v = int64(binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize]))
	case *int32:
		maxLen := int(unsafe.Sizeof(new(int32)))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int64")
		}
		*v = int32(binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize]))
	case *int16:
		maxLen := int(unsafe.Sizeof(new(uint16)))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int16")
		}
		*v = int16(binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize]))
	case *int8:
		maxLen := 1
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for int8")
		}
		*v = int8(data[ElementSize-1])
	default:
		return 0, fmt.Errorf("unable to convert %s to %s", e.GetSignature(), reflect.ValueOf(v).Kind().String())
	}

	return 32, nil
}

func (e EVMUint) isDynamic() bool {
	return false
}

var _ EVMType = (*EVMInt)(nil)

type EVMInt struct {
	M uint64
}

func (e EVMInt) getGoType() interface{} {
	switch e.M {
	case 8:
		return new(int8)
	case 16:
		return new(int16)
	case 32:
		return new(int32)
	case 64:
		return new(int64)
	default:
		return new(big.Int)
	}
}

func (e EVMInt) GetSignature() string {
	return fmt.Sprintf("int%d", e.M)
}

func (e EVMInt) pack(v interface{}) ([]byte, error) {
	n := new(big.Int)

	arg := reflect.ValueOf(v)
	switch arg.Kind() {
	case reflect.String:
		_, ok := n.SetString(arg.String(), 0)
		if !ok {
			return nil, fmt.Errorf("Failed to parse `%s", arg.String())
		}
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uint:
		n.SetUint64(arg.Uint())
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		n.SetInt64(arg.Int())
	default:
		t := reflect.TypeOf(new(int64))
		if reflect.TypeOf(v).ConvertibleTo(t) {
			n.SetInt64(reflect.ValueOf(v).Convert(t).Int())
		} else {
			return nil, fmt.Errorf("cannot convert type %s to int%d", arg.Kind().String(), e.M)
		}
	}

	b := n.Bytes()
	if uint64(len(b)) > e.M {
		return nil, fmt.Errorf("value to large for int%d", e.M)
	}
	res := pad(b, ElementSize, true)
	if (res[0] & 0x80) != 0 {
		return nil, fmt.Errorf("value to large for int%d", e.M)
	}
	if n.Sign() < 0 {
		// One's complement; i.e. 0xffff is -1, not 0.
		n.Add(n, big.NewInt(1))
		b := n.Bytes()
		res = pad(b, ElementSize, true)
		for i := 0; i < len(res); i++ {
			res[i] = ^res[i]
		}
	}
	return res, nil
}

func (e EVMInt) unpack(data []byte, offset int, v interface{}) (int, error) {
	if len(data)-offset < ElementSize {
		return 0, fmt.Errorf("not enough data")
	}

	data = data[offset:]
	sign := (data[0] & 0x80) != 0

	empty := 0
	for empty = 0; empty < ElementSize; empty++ {
		if (sign && data[empty] != 255) || (!sign && data[empty] != 0) {
			break
		}
	}

	length := ElementSize - empty
	inv := make([]byte, ElementSize)
	for i := 0; i < ElementSize; i++ {
		if sign {
			inv[i] = ^data[i]
		} else {
			inv[i] = data[i]
		}
	}
	toType := reflect.ValueOf(v).Kind().String()

	switch v := v.(type) {
	case *string:
		b := new(big.Int)
		b.SetBytes(inv[empty:ElementSize])
		if sign {
			*v = b.Sub(big.NewInt(-1), b).String()
		} else {
			*v = b.String()
		}
	case *big.Int:
		b := new(big.Int)
		b.SetBytes(data[0:ElementSize])
		if sign {
			v = b.Neg(b)
		} else {
			v = b
		}
	case *uint64:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %s", toType)
		}
		maxLen := int(unsafe.Sizeof(new(uint64)))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize])
	case *uint32:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %s", toType)
		}
		maxLen := int(unsafe.Sizeof(new(uint32)))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize])
	case *uint16:
		if sign {
			return 0, fmt.Errorf("cannot convert negative EVM int to %s", toType)
		}
		maxLen := int(unsafe.Sizeof(new(uint16)))
		if length > maxLen || (data[ElementSize-maxLen]&0x80) != 0 {
			return 0, fmt.Errorf("value to large for uint16")
		}
		*v = binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize])
	case *int64:
		maxLen := int(unsafe.Sizeof(new(int64)))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = int64(binary.BigEndian.Uint64(data[ElementSize-maxLen : ElementSize]))
	case *int32:
		maxLen := int(unsafe.Sizeof(new(int32)))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint64")
		}
		*v = int32(binary.BigEndian.Uint32(data[ElementSize-maxLen : ElementSize]))
	case *int16:
		maxLen := int(unsafe.Sizeof(new(uint16)))
		if length > maxLen {
			return 0, fmt.Errorf("value to large for uint16")
		}
		*v = int16(binary.BigEndian.Uint16(data[ElementSize-maxLen : ElementSize]))
	default:
		return 0, fmt.Errorf("unable to convert %s to %s", e.GetSignature(), toType)
	}

	return ElementSize, nil
}

func (e EVMInt) fixedSize() int {
	return ElementSize
}

func (e EVMInt) isDynamic() bool {
	return false
}

var _ EVMType = (*EVMAddress)(nil)

type EVMAddress struct {
}

func (e EVMAddress) getGoType() interface{} {
	return new(crypto.Address)
}

func (e EVMAddress) GetSignature() string {
	return "address"
}

func (e EVMAddress) pack(v interface{}) ([]byte, error) {
	var err error
	a, ok := v.(crypto.Address)
	if !ok {
		s, ok := v.(string)
		if ok {
			a, err = crypto.AddressFromHexString(s)
			if err != nil {
				return nil, err
			}
		}
	} else {
		b, ok := v.([]byte)
		if !ok {
			return nil, fmt.Errorf("cannot map to %s to EVM address", reflect.ValueOf(v).Kind().String())
		}

		a, err = crypto.AddressFromBytes(b)
		if err != nil {
			return nil, err
		}
	}

	return pad(a[:], ElementSize, true), nil
}

func (e EVMAddress) unpack(data []byte, offset int, v interface{}) (int, error) {
	addr, err := crypto.AddressFromBytes(data[offset+ElementSize-crypto.AddressLength : offset+ElementSize])
	if err != nil {
		return 0, err
	}
	switch v := v.(type) {
	case *string:
		*v = addr.String()
	case *crypto.Address:
		*v = addr
	case *([]byte):
		*v = data[offset+ElementSize-crypto.AddressLength : offset+ElementSize]
	default:
		return 0, fmt.Errorf("cannot map EVM address to %s", reflect.ValueOf(v).Kind().String())
	}

	return ElementSize, nil
}

func (e EVMAddress) isDynamic() bool {
	return false
}

var _ EVMType = (*EVMBytes)(nil)

type EVMBytes struct {
	M uint64
}

func (e EVMBytes) getGoType() interface{} {
	return make([]byte, e.M)
}

func (e EVMBytes) pack(v interface{}) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		s, ok := v.(string)
		if ok {
			b = []byte(s)
		} else {
			return nil, fmt.Errorf("cannot map to %s to EVM bytes", reflect.ValueOf(v).Kind().String())
		}
	}

	if e.M > 0 {
		if uint64(len(b)) > e.M {
			return nil, fmt.Errorf("[%d]byte to long for %s", len(b), e.GetSignature())
		}
		return pad(b, ElementSize, false), nil
	} else {
		length := EVMUint{M: 256}
		p, err := length.pack(len(b))
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(b); i += ElementSize {
			a := b[i:]
			if len(a) == 0 {
				break
			}
			p = append(p, pad(a, ElementSize, false)...)
		}

		return p, nil
	}
}

func (e EVMBytes) unpack(data []byte, offset int, v interface{}) (int, error) {
	if e.M == 0 {
		s := EVMString{}

		return s.unpack(data, offset, v)
	}

	switch v := v.(type) {
	case *string:
		start := 0
		end := int(e.M)

		for start < ElementSize-1 && data[offset+start] == 0 && start < end {
			start++
		}
		for end > start && data[offset+end-1] == 0 {
			end--
		}
		*v = string(data[offset+start : offset+end])
	case *[]byte:
		*v = data[offset : offset+int(e.M)]
	default:
		return 0, fmt.Errorf("cannot map EVM %s to %s", e.GetSignature(), reflect.ValueOf(v).Kind().String())
	}

	return ElementSize, nil
}

func (e EVMBytes) fixedSize() int {
	return ElementSize
}

func (e EVMBytes) isDynamic() bool {
	return e.M == 0
}

func (e EVMBytes) GetSignature() string {
	if e.M > 0 {
		return fmt.Sprintf("bytes%d", e.M)
	} else {
		return "bytes"
	}
}

var _ EVMType = (*EVMString)(nil)

type EVMString struct {
}

func (e EVMString) GetSignature() string {
	return "string"
}

func (e EVMString) getGoType() interface{} {
	return new(string)
}

func (e EVMString) pack(v interface{}) ([]byte, error) {
	b := EVMBytes{M: 0}

	return b.pack(v)
}

func (e EVMString) unpack(data []byte, offset int, v interface{}) (int, error) {
	lenType := EVMInt{M: 64}
	var len int64
	l, err := lenType.unpack(data, offset, &len)
	if err != nil {
		return 0, err
	}
	offset += l

	switch v := v.(type) {
	case *string:
		*v = string(data[offset : offset+int(len)])
	case *[]byte:
		*v = data[offset : offset+int(len)]
	default:
		return 0, fmt.Errorf("cannot map EVM string to %s", reflect.ValueOf(v).Kind().String())
	}

	return ElementSize, nil
}

func (e EVMString) isDynamic() bool {
	return true
}

var _ EVMType = (*EVMFixed)(nil)

type EVMFixed struct {
	N, M   uint64
	signed bool
}

func (e EVMFixed) getGoType() interface{} {
	// This is not right, obviously
	return new(big.Float)
}

func (e EVMFixed) GetSignature() string {
	if e.signed {
		return fmt.Sprintf("fixed%dx%d", e.M, e.N)
	} else {
		return fmt.Sprintf("ufixed%dx%d", e.M, e.N)
	}
}

func (e EVMFixed) pack(v interface{}) ([]byte, error) {
	// The ABI spec does not describe how this should be packed; go-ethereum abi does not implement this
	// need to dig in solidity to find out how this is packed
	return nil, fmt.Errorf("packing of %s not implemented, patches welcome", e.GetSignature())
}

func (e EVMFixed) unpack(data []byte, offset int, v interface{}) (int, error) {
	// The ABI spec does not describe how this should be packed; go-ethereum abi does not implement this
	// need to dig in solidity to find out how this is packed
	return 0, fmt.Errorf("unpacking of %s not implemented, patches welcome", e.GetSignature())
}

func (e EVMFixed) fixedSize() int {
	return ElementSize
}

func (e EVMFixed) isDynamic() bool {
	return false
}

type Argument struct {
	Name        string
	EVM         EVMType
	IsArray     bool
	Indexed     bool
	ArrayLength uint64
}

const FunctionIDSize = 4

type FunctionID [FunctionIDSize]byte

type FunctionSpec struct {
	FunctionID FunctionID
	Inputs     []Argument
	Outputs    []Argument
}

type Event struct {
	Inputs    []Argument
	Anonymous bool
}

type AbiSpec struct {
	Constructor FunctionSpec
	Fallback    FunctionSpec
	Functions   map[string]FunctionSpec
	Events      map[string]Event
}

type ArgumentJSON struct {
	Name       string
	Type       string
	Components []ArgumentJSON
	Indexed    bool
}

type AbiSpecJSON struct {
	Name            string
	Type            string
	Inputs          []ArgumentJSON
	Outputs         []ArgumentJSON
	Constant        bool
	Payable         bool
	StateMutability string
	Anonymous       bool
}

func readArgSpec(argsJ []ArgumentJSON) ([]Argument, error) {
	args := make([]Argument, len(argsJ))
	var err error

	for i, a := range argsJ {
		args[i].Name = a.Name
		args[i].Indexed = a.Indexed

		baseType := a.Type
		isArray := regexp.MustCompile("(.*)\\[([0-9]+)\\]")
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
			if N <= 0 || N > 80 {
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
		}
	}

	return args, nil
}

func ReadAbiSpec(specBytes []byte) (*AbiSpec, error) {
	var specJ []AbiSpecJSON
	err := json.Unmarshal(specBytes, &specJ)
	if err != nil {
		return nil, err
	}

	abiSpec := AbiSpec{
		Events:    make(map[string]Event),
		Functions: make(map[string]FunctionSpec),
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
			abiSpec.Events[s.Name] = Event{Inputs: inputs}
		case "function":
			inputs, err := readArgSpec(s.Inputs)
			if err != nil {
				return nil, err
			}
			outputs, err := readArgSpec(s.Outputs)
			if err != nil {
				return nil, err
			}
			fs := FunctionSpec{Inputs: inputs, Outputs: outputs}
			fs.SetFunctionID(s.Name)
			abiSpec.Functions[s.Name] = fs
		}
	}

	return &abiSpec, nil
}

func ReadAbiSpecFile(filename string) (*AbiSpec, error) {
	specBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ReadAbiSpec(specBytes)
}

func EVMTypeFromReflect(v reflect.Type) Argument {
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
		a := EVMTypeFromReflect(f.Type)
		a.Name = f.Name
		s.Inputs[i] = a
	}
	for i := 0; i < rets.NumField(); i++ {
		f := rets.Field(i)
		a := EVMTypeFromReflect(f.Type)
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
		s.Inputs[i] = EVMTypeFromReflect(t.In(i + skipIn))
	}

	for i := range s.Outputs {
		s.Outputs[i] = EVMTypeFromReflect(t.Out(i))
	}

	s.SetFunctionID(fname)
	return &s
}

func (functionSpec *FunctionSpec) SetFunctionID(functionName string) {
	sig := functionName + "("
	for i, a := range functionSpec.Inputs {
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

func (abiSpec *AbiSpec) Pack(fname string, args ...interface{}) ([]byte, error) {
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
		return nil, fmt.Errorf("Unknown function %s", fname)
	}

	packed := make([]byte, 0)

	if fname != "" {
		packed = funcSpec.FunctionID[:]
	}

	packedArgs, err := Pack(argSpec, args...)
	if err != nil {
		return nil, err
	}

	return append(packed, packedArgs...), nil
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
		if a.EVM.isDynamic() {
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
		if a.EVM.isDynamic() {
			var o int64
			l, err := offType.unpack(data, offset, &o)
			if err != nil {
				return err
			}
			offset += l
			l, err = a.EVM.unpack(data, int(o), e)
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
		arg := getArg(i)
		if a.IsArray {
			var array *[]interface{}

			array, ok := arg.(*[]interface{})
			if !ok {
				if _, ok := arg.(*string); ok {
					// We have been asked to return the value as a string; make intermediate
					// array of strings; we will concatenate after
					intermediate := make([]interface{}, a.ArrayLength)
					for i, _ := range intermediate {
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
					for i, _ := range intermediate {
						intermediate[i] = new(string)
					}
				} else {
					for i, _ := range intermediate {
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
