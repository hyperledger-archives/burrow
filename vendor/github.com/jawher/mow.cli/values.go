package cli

import (
	"flag"
	"fmt"
	"strconv"
)

type boolValued interface {
	flag.Value
	IsBoolFlag() bool
}

type multiValued interface {
	flag.Value
	Clear()
}

type defaultValued interface {
	IsDefault() bool
}

/******************************************************************************/
/* BOOL                                                                        */
/******************************************************************************/

type boolValue bool

var (
	_ flag.Value    = newBoolValue(new(bool), false)
	_ boolValued    = newBoolValue(new(bool), false)
	_ defaultValued = newBoolValue(new(bool), false)
)

func newBoolValue(into *bool, v bool) *boolValue {
	*into = v
	return (*boolValue)(into)
}

func (bo *boolValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	*bo = boolValue(b)
	return nil
}

func (bo *boolValue) IsBoolFlag() bool {
	return true
}

func (bo *boolValue) String() string {
	return fmt.Sprintf("%v", *bo)
}

func (bo *boolValue) IsDefault() bool {
	return !bool(*bo)
}

/******************************************************************************/
/* STRING                                                                        */
/******************************************************************************/

type stringValue string

var (
	_ flag.Value    = newStringValue(new(string), "")
	_ defaultValued = newStringValue(new(string), "")
)

func newStringValue(into *string, v string) *stringValue {
	*into = v
	return (*stringValue)(into)
}

func (sa *stringValue) Set(s string) error {
	*sa = stringValue(s)
	return nil
}

func (sa *stringValue) String() string {
	return fmt.Sprintf("%#v", *sa)
}

func (sa *stringValue) IsDefault() bool {
	return string(*sa) == ""
}

/******************************************************************************/
/* INT                                                                        */
/******************************************************************************/

type intValue int

var (
	_ flag.Value = newIntValue(new(int), 0)
)

func newIntValue(into *int, v int) *intValue {
	*into = v
	return (*intValue)(into)
}

func (ia *intValue) Set(s string) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*ia = intValue(int(i))
	return nil
}

func (ia *intValue) String() string {
	return fmt.Sprintf("%v", *ia)
}

/******************************************************************************/
/* STRINGS                                                                    */
/******************************************************************************/

// Strings describes a string slice argument
type stringsValue []string

var (
	_ flag.Value    = newStringsValue(new([]string), nil)
	_ multiValued   = newStringsValue(new([]string), nil)
	_ defaultValued = newStringsValue(new([]string), nil)
)

func newStringsValue(into *[]string, v []string) *stringsValue {
	*into = v
	return (*stringsValue)(into)
}

func (sa *stringsValue) Set(s string) error {
	*sa = append(*sa, s)
	return nil
}

func (sa *stringsValue) String() string {
	res := "["
	for idx, s := range *sa {
		if idx > 0 {
			res += ", "
		}
		res += fmt.Sprintf("%#v", s)
	}
	return res + "]"
}

func (sa *stringsValue) Clear() {
	*sa = nil
}

func (sa *stringsValue) IsDefault() bool {
	return len(*sa) == 0
}

/******************************************************************************/
/* INTS                                                                       */
/******************************************************************************/

type intsValue []int

var (
	_ flag.Value    = newIntsValue(new([]int), nil)
	_ multiValued   = newIntsValue(new([]int), nil)
	_ defaultValued = newIntsValue(new([]int), nil)
)

func newIntsValue(into *[]int, v []int) *intsValue {
	*into = v
	return (*intsValue)(into)
}

func (ia *intsValue) Set(s string) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*ia = append(*ia, int(i))
	return nil
}

func (ia *intsValue) String() string {
	res := "["
	for idx, s := range *ia {
		if idx > 0 {
			res += ", "
		}
		res += fmt.Sprintf("%v", s)
	}
	return res + "]"
}

func (ia *intsValue) Clear() {
	*ia = nil
}

func (ia *intsValue) IsDefault() bool {
	return len(*ia) == 0
}
