package jobs

import (
	"fmt"

	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/proposals"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/txs/payload"
	log "github.com/sirupsen/logrus"
)

func ProposalJob(prop *def.Proposal, do *def.DeployArgs, client *def.Client) (string, error) {
	var ProposeBatch payload.BatchTx

	prop.Source = useDefault(prop.Source, do.Package.Account)

	for _, job := range prop.Jobs {
		load, err := job.Payload()
		if err != nil {
			return "", fmt.Errorf("could not get Job payload: %v", load)
		}

		err = util.PreProcessFields(load, do, client)
		if err != nil {
			return "", err
		}
		// Revalidate with possible replacements
		err = load.Validate()
		if err != nil {
			return "", fmt.Errorf("error validating job %s after pre-processing variables: %v", job.Name, err)
		}

		item := payload.Any{}

		switch load.(type) {
		case *def.Call:
			announceProposalJob(job.Name, "Call")
			CallTx, ferr := FormulateCallJob(job.Call, do, client)
			if ferr != nil {
				return "", ferr
			}
			item.CallTx = CallTx
			break
		default:
			return "", fmt.Errorf("jobs %s illegal job type for proposal", job.Name)
		}

		ProposeBatch.Txs = append(ProposeBatch.Txs, &item)
	}

	proposal := payload.Proposal{Name: prop.Name, Description: prop.Description, BatchTx: &ProposeBatch}

	proposalInput, err := client.TxInput(prop.ProposalAddress, "", prop.ProposalSequence, false)
	if err != nil {
		return "", err
	}
	proposal.BatchTx.Inputs = []*payload.TxInput{proposalInput}
	proposalHash := proposal.Hash()

	var proposalTx *payload.ProposalTx
	if do.ProposeVerify {
		ballot, err := client.GetProposal(proposalHash)
		if err != nil {
			log.Warnf("Proposal could NOT be verified, error %v", err)
			return "", err
		}

		err = proposals.ProposalExpired(ballot.Proposal, client)
		if err != nil {
			log.Warnf("Proposal verify FAILED: %v", err)
			return "", err
		}

		log.Warnf("Proposal VERIFY SUCCESSFUL")
		log.Warnf("Proposal has %d votes:", len(ballot.Votes))
		for _, v := range ballot.Votes {
			log.Warnf("\t%s\n", v.Address)
		}

		return "", err
	} else if do.ProposeVote {
		ballot, err := client.GetProposal(proposalHash)
		if err != nil {
			log.Warnf("Proposal could not be found: %v", err)
			return "", err
		}

		err = proposals.ProposalExpired(ballot.Proposal, client)
		if err != nil {
			log.Warnf("Proposal error: %v", err)
			return "", err
		}

		// proposal is there and current, let's vote for it
		input, err := client.TxInput(prop.Source, "", prop.Sequence, true)
		if err != nil {
			return "", err
		}

		h := binary.HexBytes(proposalHash)
		proposalTx = &payload.ProposalTx{ProposalHash: &h, VotingWeight: 1, Input: input}
	} else if do.ProposeCreate {
		input, err := client.TxInput(prop.Source, "", prop.Sequence, true)
		if err != nil {
			return "", err
		}
		log.Warnf("Creating Proposal with hash: %x\n", proposalHash)

		proposalTx = &payload.ProposalTx{VotingWeight: 1, Input: input, Proposal: &proposal}
	} else {
		log.Errorf("please specify one of --propose-create, --propose-vote, --propose-verify")
		return "", nil
	}

	txe, err := client.SignAndBroadcast(proposalTx)
	if err != nil {
		var err = util.ChainErrorHandler(do.Package.Account, err)
		return "", err
	}

	result := fmt.Sprintf("%X", txe.Receipt.TxHash)

	return result, nil
}
