package server

import (
	//"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Unit tests for server components goes here. Full-on client-server tests
// can be found in the test folder. TODO change that?

func TestIdGet(t *testing.T) {
	idPool := NewIdPool(100)
	idparr := make([]uint, 100)
	arr := make([]uint, 100)
	for i := 0; i < 100; i++ {
		idparr[i] = uint(i + 1)
		arr[i], _ = idPool.GetId()
	}
	assert.Equal(t, idparr, arr, "Array of gotten id's is not [1, 2, ..., 101] as expected")
}

func TestIdPut(t *testing.T) {
	idPool := NewIdPool(10)
	for i := 0; i < 10; i++ {
		idPool.GetId()
	}
	idPool.ReleaseId(5)
	id, _ := idPool.GetId()
	assert.Equal(t, id, uint(5), "Id gotten is not 5.")
}

func TestIdFull(t *testing.T) {
	idPool := NewIdPool(10)
	for i := 0; i < 10; i++ {
		idPool.GetId()
	}
	_, err := idPool.GetId()
	assert.Error(t, err)
}
