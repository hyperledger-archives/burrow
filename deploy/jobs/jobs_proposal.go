package jobs

import (
	"fmt"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/txs/payload"
)

func ProposalJob(prop *def.Proposal, do *def.DeployArgs, client *def.Client, jobs chan *intermediateJob) (string, error) {
	var ProposeBatch payload.BatchTx

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

		item := payload.BatchTxItem{}

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

	input, err := client.TxInput(prop.Source, "", prop.Sequence)
	if err != nil {
		return "", err
	}
	proposal.Input = input

	txe, err := client.SignAndBroadcast(&payload.ProposalTx{VotingWeight: 1, Input: input, Proposal: &proposal})
	if err != nil {
		var err = util.ChainErrorHandler(do.Package.Account, err)
		return "", err
	}

	result := fmt.Sprintf("%X", txe.Receipt.TxHash)

	return result, nil
}
