package txs

import (
	"encoding/json"
)

type jsonCodec struct{}

func NewJSONCodec() *jsonCodec {
	return &jsonCodec{}
}

func (gwc *jsonCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return json.Marshal(env)
}

func (gwc *jsonCodec) DecodeTx(txBytes []byte) (*Envelope, error) {
	env := new(Envelope)
	err := json.Unmarshal(txBytes, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}
