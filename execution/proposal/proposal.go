package proposal

import (
	"github.com/hyperledger/burrow/txs/payload"
)

type Reader interface {
	GetProposal(proposalHash []byte) (*payload.Ballot, error)
}

type Writer interface {
	// Updates the proposal, creating it if it does not exist
	UpdateProposal(proposalHash []byte, proposal *payload.Ballot) error
	// Remove the proposal by hash
	RemoveProposal(proposalHash []byte) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateProposals(consumer func(proposalHash []byte, proposal *payload.Ballot) error) (err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}
