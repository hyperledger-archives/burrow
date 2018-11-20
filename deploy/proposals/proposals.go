package proposals

import (
	"fmt"
	"strings"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/txs/payload"
	log "github.com/sirupsen/logrus"
)

func ListProposals(client *def.Client, reqState string) {
	props, err := client.ListProposals(strings.EqualFold(reqState, "proposed"))
	if err != nil {
		log.Warnf("Failed to list proposals: %v", err)
		return
	}

	for i, prop := range props {
		var state string
		switch prop.Ballot.ProposalState {
		case payload.Ballot_FAILED:
			state = "FAILED"
		case payload.Ballot_EXECUTED:
			state = "EXECUTED"
		case payload.Ballot_PROPOSED:
			if ProposalExpired(prop.Ballot.Proposal, client) != nil {
				state = "EXPIRED"
			} else {
				state = "PROPOSED"
			}
		}

		if !strings.EqualFold(state, reqState) && !strings.EqualFold(reqState, "all") {
			continue
		}

		log.WithFields(log.Fields{
			"ProposalHash": fmt.Sprintf("%x", prop.Hash),
			"Name":         prop.Ballot.Proposal.Name,
			"Description":  prop.Ballot.Proposal.Description,
			"State":        state,
			"Votes":        len(prop.Ballot.GetVotes()),
		}).Infof("Proposal %d", i)
	}
}

func ProposalExpired(proposal *payload.Proposal, client *def.Client) error {
	for _, input := range proposal.BatchTx.Inputs {
		acc, err := client.GetAccount(input.Address)
		if err != nil {
			return err
		}

		if input.Sequence != acc.Sequence+1 {
			return fmt.Errorf("Proposal has expired")
		}
	}

	return nil
}
