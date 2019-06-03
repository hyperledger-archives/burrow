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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"google.golang.org/grpc"
)

func TestSpin(t *testing.T) {
	const listenAddress = "localhost:10997"
	wg := new(sync.WaitGroup)
	wg.Add(1)
	err := consume("spin", listenAddress, wg)
	wg.Wait()
	require.NoError(t, err)
}

func TestSpinAll(t *testing.T) {
	const listenAddress = "localhost:10997"
	numConsumer := 100
	wg := new(sync.WaitGroup)
	wg.Add(numConsumer)
	for i := 0; i < numConsumer; i++ {
		go consume(fmt.Sprintf("consumer %d", i), listenAddress, wg)
	}
	wg.Wait()
}

func consume(name, listenAddress string, wg *sync.WaitGroup) error {
	defer wg.Done()
	conn, err := grpc.Dial(listenAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	cli := rpcevents.NewExecutionEventsClient(conn)
	stream, err := cli.Stream(context.Background(), &rpcevents.BlocksRequest{
		BlockRange: rpcevents.NewBlockRange(rpcevents.AbsoluteBound(0), rpcevents.StreamBound()),
	})
	startTime := time.Now()
	timer := time.NewTicker(5 * time.Second)
	var blocks, height uint64
	defer timer.Stop()
	go func() {
		for t := range timer.C {
			dur := t.Sub(startTime)
			blocksSec := float64(blocks*uint64(time.Second)) / float64(dur)
			fmt.Printf("%s height %d: blocks per second: %f\n", name, height, blocksSec)
		}
		fmt.Printf("BYE")
	}()
	return rpcevents.ConsumeBlockExecutions(stream, func(be *exec.BlockExecution) error {
		blocks++
		height = be.Height
		return nil
	})

}
