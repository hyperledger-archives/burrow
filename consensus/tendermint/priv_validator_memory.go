package tendermint

import (
	"github.com/hyperledger/burrow/crypto"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmTypes "github.com/tendermint/tendermint/types"
)

type privValidatorMemory struct {
	crypto.Addressable
	signer         func(msg []byte) []byte
	lastSignedInfo *LastSignedInfo
}

var _ tmTypes.PrivValidator = &privValidatorMemory{}

// Create a PrivValidator with in-memory state that takes an addressable representing the validator identity
// and a signer providing private signing for that identity.
func NewPrivValidatorMemory(addressable crypto.Addressable, signer crypto.Signer) *privValidatorMemory {
	return &privValidatorMemory{
		Addressable:    addressable,
		signer:         asTendermintSigner(signer),
		lastSignedInfo: NewLastSignedInfo(),
	}
}

func asTendermintSigner(signer crypto.Signer) func(msg []byte) []byte {
	return func(msg []byte) []byte {
		sig, err := signer.Sign(msg)
		if err != nil {
			return nil
		}
		return sig.TendermintSignature()
	}
}

func (pvm *privValidatorMemory) GetAddress() tmTypes.Address {
	return pvm.Address().Bytes()
}

func (pvm *privValidatorMemory) GetPubKey() tmCrypto.PubKey {
	return pvm.PublicKey().TendermintPubKey()
}

// TODO: consider persistence to disk/database to avoid double signing after a crash
func (pvm *privValidatorMemory) SignVote(chainID string, vote *tmTypes.Vote) error {
	return pvm.lastSignedInfo.SignVote(pvm.signer, chainID, vote)
}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tmTypes.Proposal) error {
	return pvm.lastSignedInfo.SignProposal(pvm.signer, chainID, proposal)
}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tmTypes.Heartbeat) error {
	return pvm.lastSignedInfo.SignHeartbeat(pvm.signer, chainID, heartbeat)
}
