package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	dbm "github.com/tendermint/tendermint/libs/db"
)

func TestSave(t *testing.T) {
	db := dbm.NewMemDB()
	rwt := NewRWTree(db, 100)
	foo := bz("foo")
	gaa := bz("gaa")
	dam := bz("dam")
	rwt.Set(foo, gaa)
	rwt.Save()
	assert.Equal(t, gaa, rwt.Get(foo))
	rwt.Set(foo, dam)
	rwt.Save()
	assert.Equal(t, dam, rwt.Get(foo))
}
