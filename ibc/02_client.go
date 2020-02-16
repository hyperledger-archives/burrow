package ibc

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	tmtypes "github.com/tendermint/tendermint/types"

	tm "github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
)

// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics

func NewClientCreate(state exported.ConsensusState, clientID string, address []byte) {
	types.NewMsgCreateClient(
		clientID,
		state.ClientType().String(),
		state,
		address,
	)
}

func NewClientUpdate(header exported.Header, clientID string, address []byte) {
	types.NewMsgUpdateClient(clientID, header, address)
}

// func SubmitMisbehaviour(ev evidenceexported.Evidence, address []byte) {
// 	evidence.NewMsgSubmitEvidence(ev, address)
// 	// return msg, msg.ValidateBasic()
// }

func QueryNodeConsensusState(view tm.NodeView) (tendermint.ConsensusState, int64, error) {
	state := tendermint.ConsensusState{
		Root:             commitment.NewRoot(view.RoundState().LockedBlock.AppHash.Bytes()),
		ValidatorSetHash: tmtypes.NewValidatorSet(view.RoundState().Validators.Validators).Hash(),
	}

	return state, view.RoundState().Height, nil
}

type CommitmentRoot = []byte
type Signature []byte

type ClientState struct {
	frozen         bool
	pastPublicKeys []*crypto.PublicKey
	verifiedRoots  map[uint64]CommitmentRoot
}

type ConsensusState struct {
	sequence  uint64
	publicKey *crypto.PublicKey
}

type Header struct {
	sequence       uint64
	commitmentRoot CommitmentRoot
	signature      Signature
	newPublicKey   *crypto.PublicKey
}

func (state *ConsensusState) CheckValidityAndUpdateState(client ClientState, header Header) {
	if state.sequence+1 != header.sequence {
		return
	} else if err := state.publicKey.Verify(nil, &crypto.Signature{}); err != nil {
		return
	}
	if header.newPublicKey != nil {
		state.publicKey = header.newPublicKey
		client.pastPublicKeys = append(client.pastPublicKeys, header.newPublicKey)
	}
	state.sequence = header.sequence
	client.verifiedRoots[header.sequence] = header.commitmentRoot
}
