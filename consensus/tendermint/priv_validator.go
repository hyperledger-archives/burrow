package tendermint

import (
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"
	tm_types "github.com/tendermint/tendermint/types"
	acm "github.com/hyperledger/burrow/account"
)

type privValidatorMemory struct {
	privAccount acm.PrivateAccount
}

func NewPrivValidatorMemory(privAccount acm.PrivateAccount) *privValidatorMemory {
	return &privValidatorMemory{}
}

func (pvm *privValidatorMemory) GetAddress() data.Bytes {

}

func (pvm *privValidatorMemory) GetPubKey() crypto.PubKey {

}

func (pvm *privValidatorMemory) SignVote(chainID string, vote *tm_types.Vote) error {

}

func (pvm *privValidatorMemory) SignProposal(chainID string, proposal *tm_types.Proposal) error {

}

func (pvm *privValidatorMemory) SignHeartbeat(chainID string, heartbeat *tm_types.Heartbeat) error {

}
