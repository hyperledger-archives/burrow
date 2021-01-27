package tendermint

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/burrow/binary"
	"github.com/tendermint/tendermint/libs/protoio"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

// TODO: type ?
const (
	stepNone      int8 = 0 // Used to distinguish the initial state
	stepPropose   int8 = 1
	stepPrevote   int8 = 2
	stepPrecommit int8 = 3
)

// A vote is either stepPrevote or stepPrecommit.
func voteToStep(vote *tmproto.Vote) int8 {
	switch vote.Type {
	case tmproto.PrevoteType:
		return stepPrevote
	case tmproto.PrecommitType:
		return stepPrecommit
	default:
		panic(fmt.Sprintf("Unknown vote type: %v", vote.Type))
	}
}

// LastSignedInfo contains information about the latest
// data signed by a validator to help prevent double signing.
type LastSignedInfo struct {
	sync.Mutex
	Height    int64           `json:"height"`
	Round     int32           `json:"round"`
	Step      int8            `json:"step"`
	Signature []byte          `json:"signature,omitempty"` // so we don't lose signatures
	SignBytes binary.HexBytes `json:"signbytes,omitempty"` // so we don't lose signatures
}

func NewLastSignedInfo() *LastSignedInfo {
	return &LastSignedInfo{
		Step: stepNone,
	}
}

type tmCryptoSigner func(msg []byte) []byte

// SignVote signs a canonical representation of the vote, along with the
// chainID. Implements PrivValidator.
func (lsi *LastSignedInfo) SignVote(sign tmCryptoSigner, chainID string, vote *tmproto.Vote) error {
	lsi.Lock()
	defer lsi.Unlock()
	if err := lsi.signVote(sign, chainID, vote); err != nil {
		return fmt.Errorf("error signing vote: %v", err)
	}
	return nil
}

// SignProposal signs a canonical representation of the proposal, along with
// the chainID. Implements PrivValidator.
func (lsi *LastSignedInfo) SignProposal(sign tmCryptoSigner, chainID string, proposal *tmproto.Proposal) error {
	lsi.Lock()
	defer lsi.Unlock()
	if err := lsi.signProposal(sign, chainID, proposal); err != nil {
		return fmt.Errorf("error signing proposal: %v", err)
	}
	return nil
}

// returns error if HRS regression or no SignBytes. returns true if HRS is unchanged
func (lsi *LastSignedInfo) checkHRS(height int64, round int32, step int8) (bool, error) {
	if lsi.Height > height {
		return false, errors.New("height regression")
	}

	if lsi.Height == height {
		if lsi.Round > round {
			return false, errors.New("round regression")
		}

		if lsi.Round == round {
			if lsi.Step > step {
				return false, errors.New("step regression")
			} else if lsi.Step == step {
				if lsi.SignBytes != nil {
					if lsi.Signature == nil {
						panic("pv: Signature is nil but SignBytes is not!")
					}
					return true, nil
				}
				return false, errors.New("no Signature found")
			}
		}
	}
	return false, nil
}

// signVote checks if the vote is good to sign and sets the vote signature.
// It may need to set the timestamp as well if the vote is otherwise the same as
// a previously signed vote (ie. we crashed after signing but before the vote hit the WAL).
func (lsi *LastSignedInfo) signVote(sign tmCryptoSigner, chainID string, vote *tmproto.Vote) error {
	height, round, step := vote.Height, vote.Round, voteToStep(vote)

	sameHRS, err := lsi.checkHRS(height, round, step)
	if err != nil {
		return err
	}

	signBytes := types.VoteSignBytes(chainID, vote)
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
			err = fmt.Errorf("conflicting data")
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
func (lsi *LastSignedInfo) signProposal(sign tmCryptoSigner, chainID string, proposal *tmproto.Proposal) error {
	height, round, step := proposal.Height, proposal.Round, stepPropose

	sameHRS, err := lsi.checkHRS(height, round, step)
	if err != nil {
		return err
	}

	signBytes := types.ProposalSignBytes(chainID, proposal)

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
			err = fmt.Errorf("conflicting data")
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
func (lsi *LastSignedInfo) saveSigned(height int64, round int32, step int8,
	signBytes []byte, sig []byte) {

	lsi.Height = height
	lsi.Round = round
	lsi.Step = step
	lsi.Signature = sig
	lsi.SignBytes = signBytes
}

// String returns a string representation of the LastSignedInfo.
func (lsi *LastSignedInfo) String() string {
	return fmt.Sprintf("PrivValidator{LH:%v, LR:%v, LS:%v}", lsi.Height, lsi.Round, lsi.Step)
}

//-------------------------------------

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the votes is their timestamp.
func checkVotesOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastVote, newVote tmproto.CanonicalVote
	if err := protoio.UnmarshalDelimited(lastSignBytes, &lastVote); err != nil {
		panic(fmt.Sprintf("LastSignBytes cannot be unmarshalled into vote: %v", err))
	}
	if err := protoio.UnmarshalDelimited(newSignBytes, &newVote); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into vote: %v", err))
	}

	lastTime := lastVote.Timestamp

	// set the times to the same value and check equality
	now := tmtime.Now()
	lastVote.Timestamp = now
	newVote.Timestamp = now

	return lastTime, proto.Equal(&newVote, &lastVote)
}

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the proposals is their timestamp
func checkProposalsOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastProposal, newProposal tmproto.CanonicalProposal
	if err := protoio.UnmarshalDelimited(lastSignBytes, &lastProposal); err != nil {
		panic(fmt.Sprintf("LastSignBytes cannot be unmarshalled into proposal: %v", err))
	}
	if err := protoio.UnmarshalDelimited(newSignBytes, &newProposal); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into proposal: %v", err))
	}

	lastTime := lastProposal.Timestamp
	// set the times to the same value and check equality
	now := tmtime.Now()
	lastProposal.Timestamp = now
	newProposal.Timestamp = now

	return lastTime, proto.Equal(&newProposal, &lastProposal)
}
