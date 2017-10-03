package tendermint

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"
	tm_types "github.com/tendermint/tendermint/types"
)

const (
	StepError     = -1
	StepNone      = 0 // Used to distinguish the initial state
	StepPropose   = 1
	StepPrevote   = 2
	StepPrecommit = 3
)

func VoteToStep(vote *tm_types.Vote) int8 {
	switch vote.Type {
	case tm_types.VoteTypePrevote:
		return StepPrevote
	case tm_types.VoteTypePrecommit:
		return StepPrecommit
	default:
		return StepError
	}
}

type privValidatorMemory struct {
	privateAccount acm.PrivateAccount
	lastSignedInfo *LastSignedInfo
}

var _ tm_types.PrivValidator = &privValidatorMemory{}

func NewPrivValidatorMemory(privateAccount acm.PrivateAccount) *privValidatorMemory {
	return &privValidatorMemory{
		privateAccount: privateAccount,
		lastSignedInfo: new(LastSignedInfo),
	}
}

func (pvm *privValidatorMemory) GetAddress() data.Bytes {
	return pvm.privateAccount.Address().Bytes()
}

func (pvm *privValidatorMemory) GetPubKey() crypto.PubKey {
	return pvm.privateAccount.PubKey()

}

func (pvm *privValidatorMemory) SignVote(chainID string, vote *tm_types.Vote) error {
	return pvm.lastSignedInfo.SignVote(pvm.privateAccount, chainID, vote)
}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tm_types.Proposal) error {
	return pvm.lastSignedInfo.SignProposal(pvm.privateAccount, chainID, proposal)

}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tm_types.Heartbeat) error {
	return pvm.lastSignedInfo.SignHeartbeat(pvm.privateAccount, chainID, heartbeat)
}

type Verifier interface {
	SignVote(signer tm_types.Signer, chainID string, vote *tm_types.Vote) error
	SignProposal(signer tm_types.Signer, chainID string, proposal *tm_types.Proposal) error
	SignHeartbeat(signer tm_types.Signer, chainID string, heartbeat *tm_types.Heartbeat) error
}

type LastSignedInfo struct {
	mtx           sync.Mutex
	LastHeight    int              `json:"last_height"`
	LastRound     int              `json:"last_round"`
	LastStep      int8             `json:"last_step"`
	LastSignature crypto.Signature `json:"last_signature,omitempty"` // so we dont lose signatures
	LastSignBytes data.Bytes       `json:"last_signbytes,omitempty"` // so we dont lose signatures
}

var _ Verifier = &LastSignedInfo{}

// SignVote signs a canonical representation of the vote, along with the chainID.
// Implements PrivValidator.
func (lsi *LastSignedInfo) SignVote(signer tm_types.Signer, chainID string, vote *tm_types.Vote) error {
	lsi.mtx.Lock()
	defer lsi.mtx.Unlock()
	signature, err := lsi.signBytesHRS(signer, vote.Height, vote.Round, VoteToStep(vote), tm_types.SignBytes(chainID, vote))
	if err != nil {
		return fmt.Errorf("error signing vote: %v", err)
	}
	vote.Signature = signature
	return nil
}

// SignProposal signs a canonical representation of the proposal, along with the chainID.
// Implements PrivValidator.
func (lsi *LastSignedInfo) SignProposal(signer tm_types.Signer, chainID string, proposal *tm_types.Proposal) error {
	lsi.mtx.Lock()
	defer lsi.mtx.Unlock()
	signature, err := lsi.signBytesHRS(signer, proposal.Height, proposal.Round, StepPropose, tm_types.SignBytes(chainID, proposal))
	if err != nil {
		return fmt.Errorf("error signing proposal: %v", err)
	}
	proposal.Signature = signature
	return nil
}

// SignHeartbeat signs a canonical representation of the heartbeat, along with the chainID.
// Implements PrivValidator.
func (lsi *LastSignedInfo) SignHeartbeat(signer tm_types.Signer, chainID string, heartbeat *tm_types.Heartbeat) error {
	lsi.mtx.Lock()
	defer lsi.mtx.Unlock()
	var err error
	heartbeat.Signature, err = signer.Sign(tm_types.SignBytes(chainID, heartbeat))
	return err
}

// signBytesHRS signs the given signBytes if the height/round/step (HRS)
// are greater than the latest state. If the HRS are equal,
// it returns the privValidator.LastSignature.
func (lsi *LastSignedInfo) signBytesHRS(signer tm_types.Signer, height, round int, step int8, signBytes []byte) (crypto.Signature, error) {

	sig := crypto.Signature{}
	// If height regression, err
	if lsi.LastHeight > height {
		return sig, errors.New("height regression")
	}
	// More cases for when the height matches
	if lsi.LastHeight == height {
		// If round regression, err
		if lsi.LastRound > round {
			return sig, errors.New("round regression")
		}
		// If step regression, err
		if lsi.LastRound == round {
			if lsi.LastStep > step {
				return sig, errors.New("step regression")
			} else if lsi.LastStep == step {
				if lsi.LastSignBytes != nil {
					if lsi.LastSignature.Empty() {
						return crypto.Signature{}, errors.New("lsi: LastSignature is nil but LastSignBytes is not")
					}
					// so we dont sign a conflicting vote or proposal
					// NOTE: proposals are non-deterministic (include time),
					// so we can actually lose them, but will still never sign conflicting ones
					if bytes.Equal(lsi.LastSignBytes, signBytes) {
						// log.Notice("Using lsi.LastSignature", "sig", lsi.LastSignature)
						return lsi.LastSignature, nil
					}
				}
				return sig, errors.New("step regression")
			}
		}
	}

	// Sign
	sig, err := signer.Sign(signBytes)
	if err != nil {
		return sig, err
	}

	// Persist height/round/step
	lsi.LastHeight = height
	lsi.LastRound = round
	lsi.LastStep = step
	lsi.LastSignature = sig
	lsi.LastSignBytes = signBytes

	return sig, nil
}

func (lsi *LastSignedInfo) String() string {
	return fmt.Sprintf("LastSignedInfo{LastHeight:%v, LastRound:%v, LastStep:%v}",
		lsi.LastHeight, lsi.LastRound, lsi.LastStep)
}
