package tendermint

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/burrow/binary"
	"github.com/tendermint/tendermint/types"
)

// TODO: type ?
const (
	stepNone      int8 = 0 // Used to distinguish the initial state
	stepPropose   int8 = 1
	stepPrevote   int8 = 2
	stepPrecommit int8 = 3
)

func voteToStep(vote *types.Vote) int8 {
	switch vote.Type {
	case types.VoteTypePrevote:
		return stepPrevote
	case types.VoteTypePrecommit:
		return stepPrecommit
	default:
		panic("Unknown vote type")
		return 0
	}
}

// LastSignedInfo contains information about the latest
// data signed by a validator to help prevent double signing.
type LastSignedInfo struct {
	sync.Mutex
	Height    int64           `json:"height"`
	Round     int             `json:"round"`
	Step      int8            `json:"step"`
	Signature []byte          `json:"signature,omitempty"` // so we dont lose signatures
	SignBytes binary.HexBytes `json:"signbytes,omitempty"` // so we dont lose signatures
}

func NewLastSignedInfo() *LastSignedInfo {
	return &LastSignedInfo{
		Step: stepNone,
	}
}

type tmCryptoSigner func(msg []byte) []byte

// SignVote signs a canonical representation of the vote, along with the
// chainID. Implements PrivValidator.
func (lsi *LastSignedInfo) SignVote(sign tmCryptoSigner, chainID string, vote *types.Vote) error {
	lsi.Lock()
	defer lsi.Unlock()
	if err := lsi.signVote(sign, chainID, vote); err != nil {
		return fmt.Errorf("error signing vote: %v", err)
	}
	return nil
}

// SignProposal signs a canonical representation of the proposal, along with
// the chainID. Implements PrivValidator.
func (lsi *LastSignedInfo) SignProposal(sign tmCryptoSigner, chainID string, proposal *types.Proposal) error {
	lsi.Lock()
	defer lsi.Unlock()
	if err := lsi.signProposal(sign, chainID, proposal); err != nil {
		return fmt.Errorf("error signing proposal: %v", err)
	}
	return nil
}

// returns error if HRS regression or no SignBytes. returns true if HRS is unchanged
func (lsi *LastSignedInfo) checkHRS(height int64, round int, step int8) (bool, error) {
	if lsi.Height > height {
		return false, errors.New("Height regression")
	}

	if lsi.Height == height {
		if lsi.Round > round {
			return false, errors.New("Round regression")
		}

		if lsi.Round == round {
			if lsi.Step > step {
				return false, errors.New("Step regression")
			} else if lsi.Step == step {
				if lsi.SignBytes != nil {
					if lsi.Signature == nil {
						panic("pv: Signature is nil but SignBytes is not!")
					}
					return true, nil
				}
				return false, errors.New("No Signature found")
			}
		}
	}
	return false, nil
}

// signVote checks if the vote is good to sign and sets the vote signature.
// It may need to set the timestamp as well if the vote is otherwise the same as
// a previously signed vote (ie. we crashed after signing but before the vote hit the WAL).
func (lsi *LastSignedInfo) signVote(sign tmCryptoSigner, chainID string, vote *types.Vote) error {
	height, round, step := vote.Height, vote.Round, voteToStep(vote)
	signBytes := vote.SignBytes(chainID)

	sameHRS, err := lsi.checkHRS(height, round, step)
	if err != nil {
		return err
	}

	// We might crash before writing to the wal,
	// causing us to try to re-sign for the same HRS.
	// If signbytes are the same, use the last signature.
	// If they only differ by timestamp, use last timestamp and signature
	// Otherwise, return error
	if sameHRS {
		if bytes.Equal(signBytes, lsi.SignBytes) {
			vote.Signature = lsi.Signature
		} else if timestamp, ok := checkVotesOnlyDifferByTimestamp(lsi.SignBytes, signBytes); ok {
			vote.Timestamp = timestamp
			vote.Signature = lsi.Signature
		} else {
			err = fmt.Errorf("Conflicting data")
		}
		return err
	}

	// It passed the checks. Sign the vote
	sig := sign(signBytes)
	lsi.saveSigned(height, round, step, signBytes, sig)
	vote.Signature = sig
	return nil
}

// signProposal checks if the proposal is good to sign and sets the proposal signature.
// It may need to set the timestamp as well if the proposal is otherwise the same as
// a previously signed proposal ie. we crashed after signing but before the proposal hit the WAL).
func (lsi *LastSignedInfo) signProposal(sign tmCryptoSigner, chainID string, proposal *types.Proposal) error {
	height, round, step := proposal.Height, proposal.Round, stepPropose
	signBytes := proposal.SignBytes(chainID)

	sameHRS, err := lsi.checkHRS(height, round, step)
	if err != nil {
		return err
	}

	// We might crash before writing to the wal,
	// causing us to try to re-sign for the same HRS.
	// If signbytes are the same, use the last signature.
	// If they only differ by timestamp, use last timestamp and signature
	// Otherwise, return error
	if sameHRS {
		if bytes.Equal(signBytes, lsi.SignBytes) {
			proposal.Signature = lsi.Signature
		} else if timestamp, ok := checkProposalsOnlyDifferByTimestamp(lsi.SignBytes, signBytes); ok {
			proposal.Timestamp = timestamp
			proposal.Signature = lsi.Signature
		} else {
			err = fmt.Errorf("Conflicting data")
		}
		return err
	}

	// It passed the checks. Sign the proposal
	sig := sign(signBytes)
	lsi.saveSigned(height, round, step, signBytes, sig)
	proposal.Signature = sig
	return nil
}

// Persist height/round/step and signature
func (lsi *LastSignedInfo) saveSigned(height int64, round int, step int8,
	signBytes []byte, sig []byte) {

	lsi.Height = height
	lsi.Round = round
	lsi.Step = step
	lsi.Signature = sig
	lsi.SignBytes = signBytes
}

// SignHeartbeat signs a canonical representation of the heartbeat, along with the chainID.
// Implements PrivValidator.
func (lsi *LastSignedInfo) SignHeartbeat(sign tmCryptoSigner, chainID string, heartbeat *types.Heartbeat) error {
	lsi.Lock()
	defer lsi.Unlock()
	heartbeat.Signature = sign(heartbeat.SignBytes(chainID))
	return nil
}

// String returns a string representation of the LastSignedInfo.
func (lsi *LastSignedInfo) String() string {
	return fmt.Sprintf("PrivValidator{LH:%v, LR:%v, LS:%v}", lsi.Height, lsi.Round, lsi.Step)
}

//-------------------------------------

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the votes is their timestamp.
func checkVotesOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastVote, newVote types.CanonicalJSONVote
	if err := cdc.UnmarshalJSON(lastSignBytes, &lastVote); err != nil {
		panic(fmt.Sprintf("SignBytes cannot be unmarshalled into vote: %v", err))
	}
	if err := cdc.UnmarshalJSON(newSignBytes, &newVote); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into vote: %v", err))
	}

	lastTime, err := time.Parse(types.TimeFormat, lastVote.Timestamp)
	if err != nil {
		panic(err)
	}

	// set the times to the same value and check equality
	now := types.CanonicalTime(time.Now())
	lastVote.Timestamp = now
	newVote.Timestamp = now
	lastVoteBytes, _ := cdc.MarshalJSON(lastVote)
	newVoteBytes, _ := cdc.MarshalJSON(newVote)

	return lastTime, bytes.Equal(newVoteBytes, lastVoteBytes)
}

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the proposals is their timestamp
func checkProposalsOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastProposal, newProposal types.CanonicalJSONProposal
	if err := cdc.UnmarshalJSON(lastSignBytes, &lastProposal); err != nil {
		panic(fmt.Sprintf("SignBytes cannot be unmarshalled into proposal: %v", err))
	}
	if err := cdc.UnmarshalJSON(newSignBytes, &newProposal); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into proposal: %v", err))
	}

	lastTime, err := time.Parse(types.TimeFormat, lastProposal.Timestamp)
	if err != nil {
		panic(err)
	}

	// set the times to the same value and check equality
	now := types.CanonicalTime(time.Now())
	lastProposal.Timestamp = now
	newProposal.Timestamp = now
	lastProposalBytes, _ := cdc.MarshalJSON(lastProposal)
	newProposalBytes, _ := cdc.MarshalJSON(newProposal)

	return lastTime, bytes.Equal(newProposalBytes, lastProposalBytes)
}
