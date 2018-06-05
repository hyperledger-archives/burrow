package amino

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// cdc.decodeReflectJSON

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}
	if printLog {
		spew.Printf("(D) decodeReflectJSON(bz: %s, info: %v, rv: %#v (%v), opts: %v)\n",
			bz, info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(D) -> err: %v\n", err)
		}()
	}

	// Read disfix bytes if registered.
	if info.Registered {
		// Strip the disfix bytes after checking it.
		var disfix DisfixBytes
		disfix, bz, err = decodeDisfixJSON(bz)
		if err != nil {
			return
		}
		if !info.GetDisfix().EqualBytes(disfix[:]) {
			err = fmt.Errorf("Expected disfix bytes %X but got %X", info.GetDisfix(), disfix)
			return
		}
	}

	err = cdc._decodeReflectJSON(bz, info, rv, opts)
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) _decodeReflectJSON(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if info.Type.Kind() == reflect.Interface && rv.Kind() == reflect.Ptr {
		panic("should not happen")
	}
	if printLog {
		spew.Printf("(_) _decodeReflectJSON(bz: %s, info: %v, rv: %#v (%v), opts: %v)\n",
			bz, info, rv.Interface(), rv.Type(), opts)
		defer func() {
			fmt.Printf("(_) -> err: %v\n", err)
		}()
	}

	// Special case for null for either interface, pointer, slice
	// NOTE: This doesn't match the binary implementation completely.
	if nullBytes(bz) {
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Array:
			rv.Set(reflect.Zero(rv.Type()))
			return
		}
	}

	// Dereference-and-construct pointers all the way.
	// This works for pointer-pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			newPtr := reflect.New(rv.Type().Elem())
			rv.Set(newPtr)
		}
		rv = rv.Elem()
	}

	// Handle override if a pointer to rv implements json.Unmarshaler.
	if rv.Addr().Type().Implements(jsonUnmarshalerType) {
		err = rv.Addr().Interface().(json.Unmarshaler).UnmarshalJSON(bz)
		return
	}

	// Handle override if a pointer to rv implements UnmarshalAmino.
	if info.IsAminoUnmarshaler {
		// First, decode repr instance from bytes.
		rrv, rinfo := reflect.New(info.AminoUnmarshalReprType).Elem(), (*TypeInfo)(nil)
		rinfo, err = cdc.getTypeInfo_wlock(info.AminoUnmarshalReprType)
		if err != nil {
			return
		}
		err = cdc._decodeReflectJSON(bz, rinfo, rrv, opts)
		if err != nil {
			return
		}
		// Then, decode from repr instance.
		uwrm := rv.Addr().MethodByName("UnmarshalAmino")
		uwouts := uwrm.Call([]reflect.Value{rrv})
		erri := uwouts[0].Interface()
		if erri != nil {
			err = erri.(error)
		}
		return
	}

	switch ikind := info.Type.Kind(); ikind {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		err = cdc.decodeReflectJSONInterface(bz, info, rv, opts)

	case reflect.Array:
		err = cdc.decodeReflectJSONArray(bz, info, rv, opts)

	case reflect.Slice:
		err = cdc.decodeReflectJSONSlice(bz, info, rv, opts)

	case reflect.Struct:
		err = cdc.decodeReflectJSONStruct(bz, info, rv, opts)

	case reflect.Map:
		err = cdc.decodeReflectJSONMap(bz, info, rv, opts)

	//----------------------------------------
	// Signed, Unsigned

	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int,
		reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		err = invokeStdlibJSONUnmarshal(bz, rv, opts)

	//----------------------------------------
	// Misc

	case reflect.Float32, reflect.Float64:
		if !opts.Unsafe {
			return errors.New("Amino.JSON float* support requires `amino:\"unsafe\"`.")
		}
		fallthrough
	case reflect.Bool, reflect.String:
		err = invokeStdlibJSONUnmarshal(bz, rv, opts)

	//----------------------------------------
	// Default

	default:
		panic(fmt.Sprintf("unsupported type %v", info.Type.Kind()))
	}

	return
}

func invokeStdlibJSONUnmarshal(bz []byte, rv reflect.Value, opts FieldOptions) error {
	if !rv.CanAddr() && rv.Kind() != reflect.Ptr {
		panic("rv not addressable nor pointer")
	}

	var rrv reflect.Value = rv
	if rv.Kind() != reflect.Ptr {
		rrv = reflect.New(rv.Type())
	}

	if err := json.Unmarshal(bz, rrv.Interface()); err != nil {
		return err
	}
	rv.Set(rrv.Elem())
	return nil
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONInterface(bz []byte, iinfo *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectJSONInterface")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}

	/*
		We don't make use of user-provided interface values because there are a
		lot of edge cases.

		* What if the type is mismatched?
		* What if the JSON field entry is missing?
		* Circular references?
	*/
	if !rv.IsNil() {
		// We don't strictly need to set it nil, but lets keep it here for a
		// while in case we forget, for defensive purposes.
		rv.Set(iinfo.ZeroValue)
	}

	// Consume disambiguation / prefix info.
	disfix, bz, err := decodeDisfixJSON(bz)
	if err != nil {
		return
	}

	// XXX: Check disfix against interface to make sure that it actually
	// matches, and return an error if it doesn't.

	// NOTE: Unlike decodeReflectBinaryInterface, we already dealt with nil in _decodeReflectJSON.
	// NOTE: We also "consumed" the disfix wrapper by replacing `bz` above.

	// Get concrete type info.
	// NOTE: Unlike decodeReflectBinaryInterface, always disfix.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfoFromDisfix_rlock(disfix)
	if err != nil {
		return
	}

	// Construct the concrete type.
	var crv, irvSet = constructConcreteType(cinfo)

	// Decode into the concrete type.
	err = cdc._decodeReflectJSON(bz, cinfo, crv, opts)
	if err != nil {
		rv.Set(irvSet) // Helps with debugging
		return
	}

	// We need to set here, for when !PointerPreferred and the type
	// is say, an array of bytes (e.g. [32]byte), then we must call
	// rv.Set() *after* the value was acquired.
	rv.Set(irvSet)
	return
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONArray(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectJSONArray")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	length := info.Type.Len()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte array
		var buf []byte
		err = json.Unmarshal(bz, &buf)
		if err != nil {
			return
		}
		if len(buf) != length {
			err = fmt.Errorf("decodeReflectJSONArray: byte-length mismatch, got %v want %v",
				len(buf), length)
		}
		reflect.Copy(rv, reflect.ValueOf(buf))
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}

		// Read into rawSlice.
		var rawSlice []json.RawMessage
		if err = json.Unmarshal(bz, &rawSlice); err != nil {
			return
		}
		if len(rawSlice) != length {
			err = fmt.Errorf("decodeReflectJSONArray: length mismatch, got %v want %v", len(rawSlice), length)
			return
		}

		// Decode each item in rawSlice.
		for i := 0; i < length; i++ {
			erv := rv.Index(i)
			ebz := rawSlice[i]
			err = cdc.decodeReflectJSON(ebz, einfo, erv, opts)
			if err != nil {
				return
			}
		}
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONSlice(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectJSONSlice")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}

	var ert = info.Type.Elem()

	switch ert.Kind() {

	case reflect.Uint8: // Special case: byte slice
		err = json.Unmarshal(bz, rv.Addr().Interface())
		if err != nil {
			return
		}
		if rv.Len() == 0 {
			// Special case when length is 0.
			// NOTE: We prefer nil slices.
			rv.Set(info.ZeroValue)
		} else {
			// NOTE: Already set via json.Unmarshal() above.
		}
		return

	default: // General case.
		var einfo *TypeInfo
		einfo, err = cdc.getTypeInfo_wlock(ert)
		if err != nil {
			return
		}

		// Read into rawSlice.
		var rawSlice []json.RawMessage
		if err = json.Unmarshal(bz, &rawSlice); err != nil {
			return
		}

		// Special case when length is 0.
		// NOTE: We prefer nil slices.
		var length = len(rawSlice)
		if length == 0 {
			rv.Set(info.ZeroValue)
			return
		}

		// Read into a new slice.
		var esrt = reflect.SliceOf(ert) // TODO could be optimized.
		var srv = reflect.MakeSlice(esrt, length, length)
		for i := 0; i < length; i++ {
			erv := srv.Index(i)
			ebz := rawSlice[i]
			err = cdc.decodeReflectJSON(ebz, einfo, erv, opts)
			if err != nil {
				return
			}
		}

		// TODO do we need this extra step?
		rv.Set(srv)
		return
	}
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONStruct(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectJSONStruct")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}

	// Map all the fields(keys) to their blobs/bytes.
	// NOTE: In decodeReflectBinaryStruct, we don't need to do this,
	// since fields are encoded in order.
	var rawMap = make(map[string]json.RawMessage)
	err = json.Unmarshal(bz, &rawMap)
	if err != nil {
		return
	}

	for _, field := range info.Fields {

		// Get field rv and info.
		var frv = rv.Field(field.Index)
		var finfo *TypeInfo
		finfo, err = cdc.getTypeInfo_wlock(field.Type)
		if err != nil {
			return
		}

		// Get value from rawMap.
		var valueBytes = rawMap[field.JSONName]
		if len(valueBytes) == 0 {
			// TODO: Since the Go stdlib's JSON codec allows case-insensitive
			// keys perhaps we need to also do case-insensitive lookups here.
			// So "Vanilla" and "vanilla" would both match to the same field.
			// It is actually a security flaw with encoding/json library
			// - See https://github.com/golang/go/issues/14750
			// but perhaps we are aiming for as much compatibility here.
			// JAE: I vote we depart from encoding/json, than carry a vuln.

			// Set nil/zero on frv.
			frv.Set(reflect.Zero(frv.Type()))
			continue
		}

		// Decode into field rv.
		err = cdc.decodeReflectJSON(valueBytes, finfo, frv, opts)
		if err != nil {
			return
		}
	}

	return nil
}

// CONTRACT: rv.CanAddr() is true.
func (cdc *Codec) decodeReflectJSONMap(bz []byte, info *TypeInfo, rv reflect.Value, opts FieldOptions) (err error) {
	if !rv.CanAddr() {
		panic("rv not addressable")
	}
	if printLog {
		fmt.Println("(d) decodeReflectJSONMap")
		defer func() {
			fmt.Printf("(d) -> err: %v\n", err)
		}()
	}

	// Map all the fields(keys) to their blobs/bytes.
	// NOTE: In decodeReflectBinaryMap, we don't need to do this,
	// since fields are encoded in order.
	var rawMap = make(map[string]json.RawMessage)
	err = json.Unmarshal(bz, &rawMap)
	if err != nil {
		return
	}

	var krt = rv.Type().Key()
	if krt.Kind() != reflect.String {
		err = fmt.Errorf("decodeReflectJSONMap: key type must be string") // TODO also support []byte and maybe others
		return
	}
	var vinfo *TypeInfo
	vinfo, err = cdc.getTypeInfo_wlock(rv.Type().Elem())
	if err != nil {
		return
	}

	var mrv = reflect.MakeMapWithSize(rv.Type(), len(rawMap))
	for key, valueBytes := range rawMap {

		// Get map value rv.
		vrv := reflect.New(mrv.Type().Elem()).Elem()

		// Decode valueBytes into vrv.
		err = cdc.decodeReflectJSON(valueBytes, vinfo, vrv, opts)
		if err != nil {
			return
		}

		// And set.
		krv := reflect.New(reflect.TypeOf("")).Elem()
		krv.SetString(key)
		mrv.SetMapIndex(krv, vrv)
	}
	rv.Set(mrv)

	return nil
}

//----------------------------------------
// Misc.

type disfixWrapper struct {
	Disfix string          `json:"type"`
	Data   json.RawMessage `json:"value"`
}

// decodeDisfixJSON helps unravel the disfix and
// the stored data, which are expected in the form:
// {
//    "type": "XXXXXXXXXXXXXXXXX",
//    "value":  {}
// }
func decodeDisfixJSON(bz []byte) (df DisfixBytes, data []byte, err error) {
	if string(bz) == "null" {
		panic("yay")
	}
	dfw := new(disfixWrapper)
	err = json.Unmarshal(bz, dfw)
	if err != nil {
		err = fmt.Errorf("Cannot parse disfix JSON wrapper: %v", err)
		return
	}
	dfBytes, err := hex.DecodeString(dfw.Disfix)
	if err != nil {
		return
	}

	// Get disfix.
	if g, w := len(dfBytes), DisfixBytesLen; g != w {
		err = fmt.Errorf("Disfix length got=%d want=%d data=%s", g, w, bz)
		return
	}
	copy(df[:], dfBytes)
	if (DisfixBytes{}).EqualBytes(df[:]) {
		err = errors.New("Unexpected zero disfix in JSON")
		return
	}

	// Get data.
	if len(dfw.Data) == 0 {
		err = errors.New("Disfix JSON wrapper should have non-empty value field")
		return
	}
	data = dfw.Data
	return
}

func nullBytes(b []byte) bool {
	return bytes.Equal(b, []byte(`null`))
}
