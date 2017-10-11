package account

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/assert"
	"encoding/binary"
)

func TestAlterPower(t *testing.T) {
	val := AsValidator(NewConcreteAccountFromSecret("seeeeecret").Account())
	valInc := val.SetPower(2442132)
	assert.Equal(t, uint64(0), val.Power())

	fmt.Println(valInc)
	var i uint64 = (1<<63) -1
	var j uint64 =  1<<63
	var k = i + j
	i = uint64(j)
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, i)
	fmt.Printf("0x%X\n", bs)
	fmt.Printf("%v + %v = %v\n",i, j, k)
	// 1 1
	// 1 1
	// &
	//
}