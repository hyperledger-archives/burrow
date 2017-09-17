package tendermint

import (
	"testing"

	"os"

	"github.com/hyperledger/burrow/logging/lifecycle"
	"github.com/tendermint/tendermint/config"
	"github.com/stretchr/testify/assert"
)

const testDir = "./scratch"

func TestLaunchGenesisValidator(t *testing.T) {
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	os.Chdir(testDir)
	conf := config.DefaultConfig()
	logger, _ := lifecycle.NewStdErrLogger()
	LaunchGenesisValidator(conf, logger)
}
type bar struct {
	i int
}
type foo struct {
	b **bar
}

func (f *foo) integer() int {
	return (*f.b).i
}

func newFoo(b **bar) *foo {
	return &foo{b}
}

func TestLazy(t *testing.T) {
	a := new(bar)
	b := &a
	f := newFoo(b)
	assert.Equal(t, 0, f.integer())
	*b = &bar{4}
	assert.Equal(t, 4, f.integer())
	//assert.Equal(t, j, *i)
}