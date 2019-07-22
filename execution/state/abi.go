package state

import (
	"github.com/hyperledger/burrow/acm/acmstate"
)

func (s *ReadState) GetMetadata(metahash acmstate.MetadataHash) (string, error) {
	return string(s.Plain.Get(keys.Abi.Key(metahash.Bytes()))), nil
}

func (ws *writeState) SetMetadata(metahash acmstate.MetadataHash, abi string) error {
	ws.plain.Set(keys.Abi.Key(metahash.Bytes()), []byte(abi))
	return nil
}
