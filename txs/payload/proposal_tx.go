package payload

import (
	"fmt"

	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func NewProposalTx(propsal *Proposal) *ProposalTx {
	return &ProposalTx{
		Proposal: propsal,
	}
}

func (tx *ProposalTx) Type() Type {
	return TypeProposal
}

func (tx *ProposalTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *ProposalTx) String() string {
	return fmt.Sprintf("ProposalTx{%v}", tx.Proposal)
}

func (tx *ProposalTx) Any() *Any {
	return &Any{
		ProposalTx: tx,
	}
}

func DecodeProposal(proposalBytes []byte) (*Proposal, error) {
	proposal := new(Proposal)
	err := cdc.UnmarshalBinary(proposalBytes, proposal)
	if err != nil {
		return nil, err
	}
	return proposal, nil
}

func (p *Proposal) Encode() ([]byte, error) {
	return cdc.MarshalBinary(p)
}

func (p *Proposal) String() string {
	return ""
}

func (v *Vote) String() string {
	return v.Address.String()
}

func DecodeBallot(ballotBytes []byte) (*Ballot, error) {
	ballot := new(Ballot)
	err := cdc.UnmarshalBinary(ballotBytes, ballot)
	if err != nil {
		return nil, err
	}
	return ballot, nil
}

func (p *Ballot) Encode() ([]byte, error) {
	return cdc.MarshalBinary(p)
}
