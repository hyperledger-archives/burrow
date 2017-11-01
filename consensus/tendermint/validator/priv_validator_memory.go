package validator

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"
	tm_types "github.com/tendermint/tendermint/types"
)

type privValidatorMemory struct {
	acm.Addressable
	acm.Signer
	lastSignedInfo *LastSignedInfo
}

var _ tm_types.PrivValidator = &privValidatorMemory{}

// Create a PrivValidator with in-memory state that takes an addressable representing the validator identity
// and a signer providing private signing for that identity.
func NewPrivValidatorMemory(addressable acm.Addressable, signer acm.Signer) *privValidatorMemory {
	return &privValidatorMemory{
		Addressable:    addressable,
		Signer:         signer,
		lastSignedInfo: new(LastSignedInfo),
	}
}

func (pvm *privValidatorMemory) GetAddress() data.Bytes {
	return pvm.Address().Bytes()
}

func (pvm *privValidatorMemory) GetPubKey() crypto.PubKey {
	return pvm.PubKey()
}

func (pvm *privValidatorMemory) SignVote(chainID string, vote *tm_types.Vote) error {
	return pvm.lastSignedInfo.SignVote(pvm, chainID, vote)
}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tm_types.Proposal) error {
	return pvm.lastSignedInfo.SignProposal(pvm, chainID, proposal)
}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tm_types.Heartbeat) error {
	return pvm.lastSignedInfo.SignHeartbeat(pvm, chainID, heartbeat)
}
