package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPlaceholders(t *testing.T) {
	assert.True(t, PlaceholderRegex.MatchString("$foo"))
	assert.True(t, PlaceholderRegex.MatchString("   $foo"))
	assert.True(t, PlaceholderRegex.MatchString("$foo"))
	assert.True(t, PlaceholderRegex.MatchString("asdas:$foo"))
	assert.True(t, PlaceholderRegex.MatchString("Set.$AddValidator.Address.Power"))
	// Placeholder match
	assert.Equal(t, PlaceholderMatch{"$AddValidator.foobar", "AddValidator", "foobar"}, MatchPlaceholders("$AddValidator.foobar")[0])
	assert.Equal(t, PlaceholderMatch{"$AddValidator.foobar", "AddValidator", "foobar"}, MatchPlaceholders("set.$AddValidator.foobar")[0])
	// With brackets
	assert.Equal(t, PlaceholderMatch{"${AddValidator}", "AddValidator", ""}, MatchPlaceholders("set.${AddValidator}.foobar")[0])
	assert.Equal(t, PlaceholderMatch{"${AddValidator.baz}", "AddValidator", "baz"}, MatchPlaceholders("set.${AddValidator.baz}.foobar")[0])
	assert.Equal(t, PlaceholderMatch{"${Add_Validator.baz}", "Add_Validator", "baz"}, MatchPlaceholders("set.${Add_Validator.baz}.foobar")[0])
	// Non-matches
	assert.Len(t, MatchPlaceholders("set.${AddValidator.baz.foobar}"), 0)
	assert.Len(t, MatchPlaceholders("set.${}AddValidator.baz.foobar}"), 0)
	assert.Len(t, MatchPlaceholders(""), 0)
	assert.Len(t, MatchPlaceholders("set.${{foo.bar}}"), 0)
	assert.Len(t, MatchPlaceholders("set.${{foo.bar}"), 0)
	assert.Len(t, MatchPlaceholders("set.${foo,bar}"), 0)

	assert.Equal(t, PlaceholderMatch{"${foo.bar}", "foo", "bar"}, MatchPlaceholders("set.${foo.bar}}.foobar")[0])
}

func TestStripBraces(t *testing.T) {
	assert.Equal(t, `\$foo`, stripBraces(`\${foo}.bar`))
	assert.Equal(t, `\$foo.bar`, stripBraces(`\${foo.bar}`))
	assert.Equal(t, `\$foo.bar.baz`, stripBraces(`\${foo.bar.baz}`))
}
