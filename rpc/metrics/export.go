// Copyright 2017 Monax Industries Limited
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
package metrics

import (
	"fmt"
	"math"

	"github.com/hyperledger/burrow/rpc"
)

// gatherData - Collects the data from the API and stores into struct
func (e *Exporter) gatherData() error {
	var err error

	err = e.getStatus()
	if err != nil {
		return err
	}
	err = e.getMemPoolDepth()
	if err != nil {
		return err
	}
	err = e.getPeers()
	if err != nil {
		return err
	}
	blocks, err := e.getBlocks()
	if err != nil {
		return err
	}
	err = e.getTxBuckets(blocks)
	if err != nil {
		return err
	}
	err = e.getBlockTimeBuckets(blocks)
	if err != nil {
		return err
	}

	return nil
}

// Get status
func (e *Exporter) getStatus() error {
	res, err := e.service.Status()
	if err != nil {
		return err
	}
	e.datum.LatestBlockHeight = float64(res.SyncInfo.LatestBlockHeight)
	return nil
}

// Get unconfirmed_transactions
func (e *Exporter) getMemPoolDepth() error {
	res, err := e.service.UnconfirmedTxs(10000000000)
	if err != nil {
		return err
	}
	e.datum.UnconfirmedTxs = float64(res.NumTxs)
	return nil
}

// Get total peers
func (e *Exporter) getPeers() error {
	peers := e.service.Peers()
	e.datum.TotalPeers = float64(len(peers))
	e.datum.InboundPeers = 0
	e.datum.OutboundPeers = 0

	for _, peer := range peers {
		if peer.IsOutbound {
			e.datum.OutboundPeers += 1
		} else {
			e.datum.InboundPeers += 1
		}
	}

	return nil
}

func (e *Exporter) getBlocks() (*rpc.ResultBlocks, error) {
	var minHeight uint64
	var maxHeight uint64
	maxHeight = uint64(e.datum.LatestBlockHeight)

	if maxHeight >= e.blockSampleSize {
		minHeight = maxHeight - (e.blockSampleSize - 1)
		e.datum.BlockSampleSize = e.blockSampleSize
	} else {
		minHeight = 1
		e.datum.BlockSampleSize = maxHeight
	}

	res, err := e.service.Blocks(int64(minHeight), int64(maxHeight))
	if err != nil {
		return nil, err
	}

	if !(len(res.BlockMetas) > 0) {
		return nil, fmt.Errorf("no blocks returned")
	}

	return res, nil
}

// Get transaction buckets
func (e *Exporter) getTxBuckets(res *rpc.ResultBlocks) error {
	e.datum.TotalTxs = 0
	e.datum.TxPerBlockBuckets = map[float64]float64{}

	for _, block := range res.BlockMetas {
		txAsFloat := float64(block.Header.NumTxs)
		e.datum.TxPerBlockBuckets[txAsFloat] += 1
		e.datum.TotalTxs += txAsFloat
	}
	return nil
}

func (e *Exporter) getBlockTimeBuckets(res *rpc.ResultBlocks) error {
	// tendermint gives us the blocks in the reverse of the order we'd expect them
	//  BlockMetas[0] is the most recent block.
	timeSampleEnded := res.BlockMetas[0].Header.Time
	timeSampleBegan := res.BlockMetas[len(res.BlockMetas)-1].Header.Time
	e.datum.TotalTime = round(timeSampleEnded.Sub(timeSampleBegan).Seconds())
	e.datum.TimePerBlockBuckets = map[float64]float64{}

	for i, block := range res.BlockMetas {
		if i == 0 {
			continue
		}
		timeEnded := res.BlockMetas[i-1].Header.Time
		timeBegan := block.Header.Time
		timeDiff := round(timeEnded.Sub(timeBegan).Seconds())
		e.datum.TimePerBlockBuckets[timeDiff] += 1
	}
	return nil
}

func round(x float64) float64 {
	return math.Round(x/0.1) * 0.1
}
