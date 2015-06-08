package server

import (
	"container/list"
)

// Simple id pool. Lets you get and release uints. Will panic
// if trying to get an id and it's empty.
type IdPool struct {
	ids *list.List
}

func NewIdPool(totNum uint) *IdPool {
	idPool := &IdPool{}
	idPool.init(totNum)
	return idPool 
}

// We start from 1, so that 0 is not used as an id.
func (idp *IdPool) init(totNum uint) {
	idp.ids = list.New()
	for i := uint(1); i <= totNum; i++ {
		idp.ids.PushBack(i)
	}
}

// Get an id from the pool.
func (idp *IdPool) GetId() uint {
	val := idp.ids.Front()
	idp.ids.Remove(val)
	num, _ := val.Value.(uint)
	return num
}

// Release an id back into the pool.
func (idp *IdPool) ReleaseId(id uint) {
	idp.ids.PushBack(id)
}
