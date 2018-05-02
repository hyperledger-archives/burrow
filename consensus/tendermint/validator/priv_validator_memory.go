package validator

import (
	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	tm_crypto "github.com/tendermint/go-crypto"
	tm_types "github.com/tendermint/tendermint/types"
	val_types "github.com/tendermint/tendermint/types/priv_validator"
)

type privValidatorMemory struct {
	acm.Addressable
	tm_types.Signer
	lastSignedInfo *val_types.LastSignedInfo
}

var _ tm_types.PrivValidator = &privValidatorMemory{}

// Create a PrivValidator with in-memory state that takes an addressable representing the validator identity
// and a signer providing private signing for that identity.
func NewPrivValidatorMemory(addressable acm.Addressable, signer crypto.Signer) *privValidatorMemory {
	return &privValidatorMemory{
		Addressable:    addressable,
		Signer:         asTendermintSigner(signer),
		lastSignedInfo: val_types.NewLastSignedInfo(),
	}
}

func (pvm *privValidatorMemory) GetAddress() tm_types.Address {
	return pvm.Address().Bytes()
}

func (pvm *privValidatorMemory) GetPubKey() tm_crypto.PubKey {
	tm := tm_crypto.PubKeyEd25519{}
	copy(tm[:], pvm.PublicKey().RawBytes())
	return tm_crypto.PubKey{tm}
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
	heartbeat.Signature, err = pvm.Signer.Sign(heartbeat.SignBytes(chainID))
	return err
}

func asTendermintSigner(signer crypto.Signer) tm_types.Signer {
	return tendermintSigner{Signer: signer}
}

type tendermintSigner struct {
	crypto.Signer
}

func (tms tendermintSigner) Sign(msg []byte) (tm_crypto.Signature, error) {
	sig, err := tms.Signer.Sign(msg)
	if err != nil {
		return tm_crypto.Signature{}, err
	}
	tm_sig := tm_crypto.SignatureEd25519{}
	copy(tm_sig[:], sig.RawBytes())
	return tm_crypto.Signature{tm_sig}, nil
}
