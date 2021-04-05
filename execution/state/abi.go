package state

import (
	"github.com/hyperledger/burrow/acm/acmstate"
)

func (s *ReadState) GetMetadata(metahash acmstate.MetadataHash) (string, error) {
	data, err := s.Plain.Get(keys.Abi.Key(metahash.Bytes()))
	return string(data), err
}

func (ws *writeState) SetMetadata(metahash acmstate.MetadataHash, abi string) error {
	return ws.plain.Set(keys.Abi.Key(metahash.Bytes()), []byte(abi))
}
