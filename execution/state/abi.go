package state

import (
	"github.com/hyperledger/burrow/acm/acmstate"
)

func (s *ReadState) GetAbi(abihash acmstate.AbiHash) (string, error) {
	return string(s.Plain.Get(keys.Abi.Key(abihash.Bytes()))), nil
}

func (ws *writeState) SetAbi(abihash acmstate.AbiHash, abi string) error {
	ws.plain.Set(keys.Abi.Key(abihash.Bytes()), []byte(abi))
	return nil
}
