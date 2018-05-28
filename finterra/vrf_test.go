// MIT License
//
// Copyright (c) 2018 Finterra
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package finterra_test

import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/finterra"
	"github.com/stretchr/testify/assert"
)

func TestVRF(t *testing.T) {
	for i := 0; i < 100; i++ {
		pv, _ := acm.GeneratePrivateKey(nil)
		pa, _ := acm.GeneratePrivateAccountFromPrivateKeyBytes(pv.Bytes()[1:])
		pk := pv.PublicKey()
		m := []byte{byte(i)}

		vrf := finterra.NewVRF(pa, pa)

		var max uint64 = uint64(i + 1*1000)
		vrf.SetMax(max)
		index, proof := vrf.Evaluate(m)

		//fmt.Printf("%x\n", index)
		assert.Equal(t, index <= max, true)

		index2, result := vrf.Verify(m, pk, proof)

		assert.Equal(t, result, true)
		assert.Equal(t, index, index2)
	}
}
