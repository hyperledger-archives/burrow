package pipe

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/binary"
	cm "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/consensus"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/p2p"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
)

// The consensus struct.
type consensus struct {
	consensusState *cm.ConsensusState
	p2pSwitch      *p2p.Switch
}

func newConsensus(consensusState *cm.ConsensusState, p2pSwitch *p2p.Switch) *consensus {
	return &consensus{consensusState, p2pSwitch}
}

// Get the current consensus state.
func (this *consensus) State() (*ConsensusState, error) {
	roundState := this.consensusState.GetRoundState()
	peerRoundStates := []string{}
	for _, peer := range this.p2pSwitch.Peers().List() {
		// TODO: clean this up?
		peerState := peer.Data.Get(cm.PeerStateKey).(*cm.PeerState)
		peerRoundState := peerState.GetRoundState()
		peerRoundStateStr := peer.Key + ":" + string(binary.JSONBytes(peerRoundState))
		peerRoundStates = append(peerRoundStates, peerRoundStateStr)
	}
	return FromRoundState(roundState), nil
}

// Get all validators.
func (this *consensus) Validators() (*ValidatorList, error) {
	var blockHeight uint
	bondedValidators := make([]*state.Validator, 0)
	unbondingValidators := make([]*state.Validator, 0)

	s := this.consensusState.GetState()
	blockHeight = s.LastBlockHeight
	s.BondedValidators.Iterate(func(index uint, val *state.Validator) bool {
		bondedValidators = append(bondedValidators, val)
		return false
	})
	s.UnbondingValidators.Iterate(func(index uint, val *state.Validator) bool {
		unbondingValidators = append(unbondingValidators, val)
		return false
	})

	return &ValidatorList{blockHeight, bondedValidators, unbondingValidators}, nil
}
