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

package txs

import (
	"fmt"

	"github.com/hyperledger/burrow/txs/payload"
	amino "github.com/tendermint/go-amino"
)

type aminoCodec struct {
	*amino.Codec
}

func NewAminoCodec() *aminoCodec {
	cdc := amino.NewCodec()
	cdc.RegisterInterface((*payload.Payload)(nil), nil)
	registerTx(cdc, &payload.SendTx{})
	registerTx(cdc, &payload.CallTx{})
	registerTx(cdc, &payload.BondTx{})
	registerTx(cdc, &payload.UnbondTx{})
	registerTx(cdc, &payload.PermsTx{})
	registerTx(cdc, &payload.NameTx{})
	registerTx(cdc, &payload.GovTx{})
	registerTx(cdc, &payload.ProposalTx{})
	return &aminoCodec{cdc}
}

func (gwc *aminoCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return gwc.MarshalBinaryBare(env)
}

func (gwc *aminoCodec) DecodeTx(txBytes []byte) (*Envelope, error) {
	env := new(Envelope)
	err := gwc.UnmarshalBinaryBare(txBytes, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func registerTx(cdc *amino.Codec, tx payload.Payload) {
	cdc.RegisterConcrete(tx, fmt.Sprintf("burrow/txs/payload/%v", tx.Type()), nil)
}
