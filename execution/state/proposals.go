package state

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/txs/payload"
)

var _ proposal.IterableReader = &State{}

func (s *ReadState) GetProposal(proposalHash []byte) (*payload.Ballot, error) {
	tree, err := s.forest.Reader(proposalKeyFormat.Prefix())
	if err != nil {
		return nil, err
	}
	bs := tree.Get(proposalKeyFormat.KeyNoPrefix(proposalHash))
	if len(bs) == 0 {
		return nil, nil
	}

	return payload.DecodeBallot(bs)
}

func (ws *writeState) UpdateProposal(proposalHash []byte, p *payload.Ballot) error {
	tree, err := ws.forest.Writer(proposalKeyFormat.Prefix())
	if err != nil {
		return err
	}
	bs, err := p.Encode()
	if err != nil {
		return err
	}

	tree.Set(proposalKeyFormat.KeyNoPrefix(proposalHash), bs)
	return nil
}

func (ws *writeState) RemoveProposal(proposalHash []byte) error {
	tree, err := ws.forest.Writer(proposalKeyFormat.Prefix())
	if err != nil {
		return err
	}
	tree.Delete(proposalKeyFormat.KeyNoPrefix(proposalHash))
	return nil
}

func (s *ReadState) IterateProposals(consumer func(proposalHash []byte, proposal *payload.Ballot) error) error {
	tree, err := s.forest.Reader(proposalKeyFormat.Prefix())
	if err != nil {
		return err
	}
	return tree.Iterate(nil, nil, true, func(key []byte, value []byte) error {
		entry, err := payload.DecodeBallot(value)
		if err != nil {
			return fmt.Errorf("State.IterateProposal() could not iterate over proposals: %v", err)
		}
		return consumer(key, entry)
	})
}
