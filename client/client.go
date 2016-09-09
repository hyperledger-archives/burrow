// Copyright 2015, 2016 Eris Industries (UK) Ltd.
// This file is part of Eris-RT

// Eris-RT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Eris-RT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Eris-RT.  If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"github.com/tendermint/go-rpc/client"

	tendermint_client "github.com/eris-ltd/eris-db/rpc/tendermint/client"
	"github.com/eris-ltd/eris-db/txs"
)

type Client interface{
	Broadcast(transaction txs.Tx) (*txs.Receipt, error)
}

// NOTE [ben] Compiler check to ensure ErisClient successfully implements
// eris-db/client.Client
var _ Client = (*ErisClient)(nil)

// Eris-Client is 
type ErisClient struct {
	broadcastRPC string
}

//------------------------------------------------------------------------------------
// broadcast to blockchain node
// NOTE: [ben] Eris Client first continues from tendermint rpc, but will have handshake to negotiate
// protocol version for moving towards rpc/v1 

func (erisClient *ErisClient) Broadcast(tx txs.Tx) (*txs.Receipt, error) {
	client := rpcclient.NewClientURI(erisClient.broadcastRPC)
	receipt, err := tendermint_client.BroadcastTx(client, tx)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}