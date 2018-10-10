package proposal

import (
	"github.com/hyperledger/burrow/txs/payload"
)

type Reader interface {
	GetProposal(proposalHash []byte) (*payload.Proposal, error)
}

type Writer interface {
	// Updates the name entry creating it if it does not exist
	UpdateProposal(proposalHash []byte, proposal *payload.Proposal) error
	// Remove the name entry
	RemoveProposal(proposalHash []byte) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateProposals(consumer func(proposalHash []byte, proposal *payload.Proposal) (stop bool)) (stopped bool, err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}
