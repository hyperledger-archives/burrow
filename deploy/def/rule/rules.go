package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"strings"

	"reflect"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

var PlaceholderRegex = regexp.MustCompile(`\$(?P<bracket>{?)(?P<job>[[:word:]]+)(\.(?P<variable>[[:word:]]+))?}?`)

func MatchPlaceholders(str string) []PlaceholderMatch {
	matches := PlaceholderRegex.FindAllStringSubmatch(str, -1)
	pms := make([]PlaceholderMatch, 0, len(matches))
	for _, match := range matches {
		pm := NewPlaceholderMatch(match)
		if pm.IsMatch() {
			pms = append(pms, pm)
		}
	}
	return pms
}

type PlaceholderMatch struct {
	Match        string
	JobName      string
	VariableName string
}

func (pm PlaceholderMatch) String() string {
	var varStr string
	if pm.VariableName != "" {
		varStr = fmt.Sprintf(", Variable: %s", pm.VariableName)
	}
	return fmt.Sprintf("PlaceholderMatch{'%s': Job: %s%s}", pm.Match, pm.JobName, varStr)
}

func (pm PlaceholderMatch) IsMatch() bool {
	return pm.Match != ""
}

func NewPlaceholderMatch(match []string) (pm PlaceholderMatch) {
	pm.Match = match[0]
	for i, name := range PlaceholderRegex.SubexpNames() {
		switch name {
		case "bracket":
			if match[i] == "{" {
				stripMatch := PlaceholderRegex.FindStringSubmatch(stripBraces(pm.Match))
				if len(stripMatch) == 0 {
					return PlaceholderMatch{}
				}
				// Match stripped but keep the outer match with brackets for use as replacement string
				pmStripped := NewPlaceholderMatch(stripMatch)
				pm.JobName = pmStripped.JobName
				pm.VariableName = pmStripped.VariableName
			}
		case "job":
			pm.JobName = match[i]
		case "variable":
			pm.VariableName = match[i]
		}
	}
	return
}

// Strips braces and return simple variable confined between braces
func stripBraces(str string) string {
	bs := []byte(str)
	const lb = byte('{')
	const rb = byte('}')
	start := 0
	for i := 0; i < len(bs); i++ {
		switch bs[i] {
		case lb:
			start = i + 1
		case rb:
			return `\$` + str[start:i]
		}
	}
	return str[start:]
}

var exampleAddress = acm.GeneratePrivateAccountFromSecret("marmot").Address()

// Rules
var (
	Placeholder = validation.Match(PlaceholderRegex).Error("must be a variable placeholder like $marmotVariable")

	Address = validation.NewStringRule(IsAddress,
		fmt.Sprintf("must be valid 20 byte hex-encoded string like '%v'", exampleAddress))

	AddressOrPlaceholder = Or(Placeholder, Address)

	Relation = validation.In("eq", "ne", "ge", "gt", "le", "lt", "==", "!=", ">=", ">", "<=", "<")

	PermissionOrPlaceholder = Or(Placeholder, Permission)

	Permission = validation.By(func(value interface{}) error {
		str, err := validation.EnsureString(value)
		if err != nil {
			return fmt.Errorf("must be a permission name, but %v is not a string", value)
		}
		_, err = permission.PermStringToFlag(str)
		if err != nil {
			return err
		}
		return nil
	})

	Uint64OrPlaceholder = Or(Placeholder, Uint64)

	Uint64 = validation.By(func(value interface{}) error {
		str, err := validation.EnsureString(value)
		if err != nil {
			return fmt.Errorf("should be a numeric string but '%v' is not a string", value)
		}
		_, err = strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("should be a 64 bit unsigned integer: ")
		}
		return nil
	})
)

func Exactly(identity interface{}) validation.Rule {
	return validation.By(func(value interface{}) error {
		if !reflect.DeepEqual(identity, value) {
			return fmt.Errorf("value %v does not exactly match %v", value, identity)
		}
		return nil
	})
}

func Or(rules ...validation.Rule) *orRule {
	return &orRule{
		rules: rules,
	}
}

type orRule struct {
	rules []validation.Rule
}

func (orr *orRule) Validate(value interface{}) error {
	errs := make([]string, len(orr.rules))
	for i, r := range orr.rules {
		err := r.Validate(value)
		if err == nil {
			return nil
		}
		errs[i] = err.Error()
	}
	return fmt.Errorf("did not validate any requirements: %s", strings.Join(errs, ", "))
}

func IsAddress(value string) bool {
	_, err := crypto.AddressFromHexString(value)
	return err == nil
}

// Returns true IFF value is zero value or has length 0
func IsOmitted(value interface{}) bool {
	value, isNil := validation.Indirect(value)
	if isNil || validation.IsEmpty(value) {
		return true
	}
	// Accept and empty slice or map
	length, err := validation.LengthOfValue(value)
	if err == nil && length == 0 {
		return true
	}
	return false
}

type boolRule struct {
	isValid func(value interface{}) bool
	message string
}

func New(isValid func(value interface{}) bool, message string, args ...interface{}) *boolRule {
	return &boolRule{
		isValid: isValid,
		message: fmt.Sprintf(message, args...),
	}
}

func (r *boolRule) Validate(value interface{}) error {
	if r.isValid(value) {
		return nil
	}
	return errors.New(r.message)
}

func (r *boolRule) Error(message string, args ...interface{}) *boolRule {
	r.message = fmt.Sprintf(message, args...)
	return r
}
