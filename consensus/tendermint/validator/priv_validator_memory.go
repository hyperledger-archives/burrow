package validator

import (
	"github.com/hyperledger/burrow/crypto"
	tm_crypto "github.com/tendermint/go-crypto"
	tm_types "github.com/tendermint/tendermint/types"
)

type privValidatorMemory struct {
	crypto.Addressable
	signer         goCryptoSigner
	lastSignedInfo *LastSignedInfo
}

var _ tm_types.PrivValidator = &privValidatorMemory{}

// Create a PrivValidator with in-memory state that takes an addressable representing the validator identity
// and a signer providing private signing for that identity.
func NewPrivValidatorMemory(addressable crypto.Addressable, signer crypto.Signer) *privValidatorMemory {
	return &privValidatorMemory{
		Addressable:    addressable,
		signer:         asTendermintSigner(signer),
		lastSignedInfo: NewLastSignedInfo(),
	}
}

func (pvm *privValidatorMemory) GetAddress() tm_types.Address {
	return pvm.Address().Bytes()
}

func (pvm *privValidatorMemory) GetPubKey() tm_crypto.PubKey {
	tm := tm_crypto.PubKeyEd25519{}
	copy(tm[:], pvm.PublicKey().RawBytes())
	return tm
}

// TODO: consider persistence to disk/database to avoid double signing after a crash
func (pvm *privValidatorMemory) SignVote(chainID string, vote *tm_types.Vote) error {
	return pvm.lastSignedInfo.SignVote(pvm.signer, chainID, vote)
}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tm_types.Proposal) error {
	return pvm.lastSignedInfo.SignProposal(pvm.signer, chainID, proposal)
}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tm_types.Heartbeat) error {
	return pvm.lastSignedInfo.SignHeartbeat(pvm.signer, chainID, heartbeat)
}

func asTendermintSigner(signer crypto.Signer) goCryptoSigner {
	return func(msg []byte) tm_crypto.Signature {
		sig, err := signer.Sign(msg)
		if err != nil {
			return nil
		}
		tmSig := tm_crypto.SignatureEd25519{}
		copy(tmSig[:], sig.RawBytes())
		return tmSig

	}
}
