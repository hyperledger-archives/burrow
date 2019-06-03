// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build forensics

package forensics

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/stretchr/testify/require"
)

const devStudioPath = "/home/silas/burrows/t7-dev-studio-burrow-000/t7-dev-studio-burrow-000"

func testLoadStudio(t *testing.T, i int) {
	re := newReplay(t, studioDir(i))
	bc, err := re.LatestBlockchain()
	require.NoError(t, err)
	fmt.Println(bc.LastBlockHeight())

	st, err := re.State(bc.LastBlockHeight())
	require.NoError(t, err)

	fmt.Printf("Validator %d hash: %X\n", i, st.Hash())
	//txHash := hex.MustDecodeString("DEF358F2CD8746CC2CEADE6EDF6518699FA91C512C45A3894FBB0E746E57B749")

	accum := new(exec.BlockAccumulator)
	buf := new(bytes.Buffer)
	err = st.IterateStreamEvents(nil, nil, func(ev *exec.StreamEvent) error {
		be := accum.Consume(ev)
		if be != nil {
			buf.WriteString(fmt.Sprintf("Block %d: %X\n\n", be.Height, be.Header.AppHash))
			for _, txe := range be.TxExecutions {
				if txe.Exception != nil {
					buf.WriteString(fmt.Sprintf("Tx %v: %v\n\n", txe.TxExecutions, txe.Exception))
				}
			}
		}
		return nil
	})
	require.NoError(t, err)

	err = ioutil.WriteFile(fmt.Sprintf("test-out-%d.txt", i), buf.Bytes(), 0644)
	require.NoError(t, err)
}

func TestStudioTx0(t *testing.T) {
	testLoadStudio(t, 0)
}

func TestStudioTx(t *testing.T) {
	for i := 0; i < 4; i++ {
		burrowDir := studioDir(i)
		fmt.Println(burrowDir)
		testLoadStudio(t, i)
	}
}

func studioDir(i int) string {
	return path.Join(devStudioPath, fmt.Sprintf("%03d", i))
}
