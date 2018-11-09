package project

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullVersion(t *testing.T) {
	assert.Equal(t, History.CurrentVersion().String(), FullVersion())
	commit = "0e90ed60"
	date = "2018-11-07T14:29:28Z"
	assert.Equal(t, fmt.Sprintf("%v+commit.%v+%v", History.CurrentVersion().String(), commit, date), FullVersion())
}
