package txs

import (
	"fmt"

	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tendermint/go-amino"
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
	registerTx(cdc, &payload.PermissionsTx{})
	registerTx(cdc, &payload.NameTx{})
	registerTx(cdc, &payload.GovernanceTx{})
	return &aminoCodec{cdc}
}

func (gwc *aminoCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return gwc.MarshalBinary(env)
}

func (gwc *aminoCodec) DecodeTx(txBytes []byte) (*Envelope, error) {
	env := new(Envelope)
	err := gwc.UnmarshalBinary(txBytes, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}

func registerTx(cdc *amino.Codec, tx payload.Payload) {
	cdc.RegisterConcrete(tx, fmt.Sprintf("burrow/txs/payload/%v", tx.Type()), nil)
}
