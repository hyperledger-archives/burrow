package validator

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"
	tm_types "github.com/tendermint/tendermint/types"
)

type privValidatorMemory struct {
	acm.Addressable
	tm_types.Signer
	lastSignedInfo *LastSignedInfo
}

var _ tm_types.PrivValidator = &privValidatorMemory{}

// Create a PrivValidator with in-memory state that takes an addressable representing the validator identity
// and a signer providing private signing for that identity.
func NewPrivValidatorMemory(addressable acm.Addressable, signer acm.Signer) *privValidatorMemory {
	return &privValidatorMemory{
		Addressable:    addressable,
		Signer:         asTendermintSigner(signer),
		lastSignedInfo: NewLastSignedInfo(),
	}
}

func (pvm *privValidatorMemory) GetAddress() data.Bytes {
	return pvm.Address().Bytes()
}

func (pvm *privValidatorMemory) GetPubKey() crypto.PubKey {
	return pvm.PublicKey().PubKey
}

// TODO: consider persistence to disk/database to avoid double signing after a crash
func (pvm *privValidatorMemory) SignVote(chainID string, vote *tm_types.Vote) error {
	return pvm.lastSignedInfo.SignVote(pvm.Signer, chainID, vote)
}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tm_types.Proposal) error {
	return pvm.lastSignedInfo.SignProposal(pvm.Signer, chainID, proposal)
}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tm_types.Heartbeat) error {
	var err error
	heartbeat.Signature, err = pvm.Signer.Sign(tm_types.SignBytes(chainID, heartbeat))
	return err
}

func asTendermintSigner(signer acm.Signer) tm_types.Signer {
	return tendermintSigner{Signer: signer}
}

type tendermintSigner struct {
	acm.Signer
}

func (tms tendermintSigner) Sign(msg []byte) (crypto.Signature, error) {
	sig, err := tms.Signer.Sign(msg)
	if err != nil {
		return crypto.Signature{}, err
	}
	return sig.GoCryptoSignature(), nil
}
