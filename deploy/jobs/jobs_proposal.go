package jobs

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/proposals"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

func getAccountSequence(seq string, addressStr string, seqCache *acmstate.Cache) (string, error) {
	if seq != "" {
		return seq, nil
	}
	address, err := crypto.AddressFromHexString(addressStr)
	if err != nil {
		return "", err
	}
	acc, err := seqCache.GetAccount(address)
	if err != nil {
		return "", err
	}
	acc.Sequence++
	err = seqCache.UpdateAccount(acc)
	return fmt.Sprintf("%d", acc.Sequence), err
}

func recurseJobs(proposeBatch *payload.BatchTx, jobs []*def.Job, prop *def.Proposal, do *def.DeployArgs, parentScript *def.Playbook, client *def.Client, seqCache *acmstate.Cache, logger *logging.Logger) error {
	script := def.Playbook{Jobs: jobs, Account: FirstOf(prop.Source, parentScript.Account), Parent: parentScript, BinPath: parentScript.BinPath}

	for _, job := range script.Jobs {
		load, err := job.Payload()
		if err != nil {
			return fmt.Errorf("could not get Job payload: %v", load)
		}

		err = util.PreProcessFields(load, do, &script, client, logger)
		if err != nil {
			return err
		}
		// Revalidate with possible replacements
		err = load.Validate()
		if err != nil {
			return fmt.Errorf("error validating job %s after pre-processing variables: %v", job.Name, err)
		}

		switch load.(type) {
		case *def.Meta:
			announceProposalJob(job.Name, "Meta", logger)
			// load the package
			err = recurseJobs(proposeBatch, job.Meta.Playbook.Jobs, prop, do, &script, client, seqCache, logger)
			if err != nil {
				return err
			}

		case *def.UpdateAccount:
			announceProposalJob(job.Name, "UpdateAccount", logger)
			job.UpdateAccount.Source = FirstOf(job.UpdateAccount.Source, script.Account)
			job.UpdateAccount.Sequence, err = getAccountSequence(job.UpdateAccount.Sequence, job.UpdateAccount.Source, seqCache)
			if err != nil {
				return err
			}
			tx, _, err := FormulateUpdateAccountJob(job.UpdateAccount, script.Account, client, logger)
			if err != nil {
				return err
			}
			proposeBatch.Txs = append(proposeBatch.Txs, &payload.Any{GovTx: tx})

		case *def.RegisterName:
			announceProposalJob(job.Name, "RegisterName", logger)
			job.RegisterName.Source = FirstOf(job.RegisterName.Source, script.Account)
			job.RegisterName.Sequence, err = getAccountSequence(job.RegisterName.Sequence, job.RegisterName.Source, seqCache)
			if err != nil {
				return err
			}
			txs, err := FormulateRegisterNameJob(job.RegisterName, do, &script, client, logger)
			if err != nil {
				return err
			}
			for _, tx := range txs {
				proposeBatch.Txs = append(proposeBatch.Txs, &payload.Any{NameTx: tx})
			}
		case *def.Call:
			announceProposalJob(job.Name, "Call", logger)
			job.Call.Source = FirstOf(job.Call.Source, script.Account)
			job.Call.Sequence, err = getAccountSequence(job.Call.Sequence, job.Call.Source, seqCache)
			if err != nil {
				return err
			}
			tx, err := FormulateCallJob(job.Call, do, &script, client, logger)
			if err != nil {
				return err
			}
			proposeBatch.Txs = append(proposeBatch.Txs, &payload.Any{CallTx: tx})
		case *def.Deploy:
			announceProposalJob(job.Name, "Deploy", logger)
			job.Deploy.Source = FirstOf(job.Deploy.Source, script.Account)
			job.Deploy.Sequence, err = getAccountSequence(job.Deploy.Sequence, job.Deploy.Source, seqCache)
			if err != nil {
				return err
			}
			deployTxs, _, err := FormulateDeployJob(job.Deploy, do, &script, client, job.Intermediate, logger)
			if err != nil {
				return err
			}
			var deployAddress crypto.Address
			// Predict address
			callee, err := crypto.AddressFromHexString(job.Deploy.Source)
			if err != nil {
				return err
			}
			for _, tx := range deployTxs {
				proposeBatch.Txs = append(proposeBatch.Txs, &payload.Any{CallTx: tx})
				txEnv := txs.NewTx(tx)

				deployAddress = crypto.NewContractAddress(callee, txEnv.Hash())
			}
			job.Result = deployAddress.String()
		case *def.Permission:
			announceProposalJob(job.Name, "Permission", logger)
			job.Permission.Source = FirstOf(job.Permission.Source, script.Account)
			job.Permission.Sequence, err = getAccountSequence(job.Permission.Sequence, job.Permission.Source, seqCache)
			if err != nil {
				return err
			}
			tx, err := FormulatePermissionJob(job.Permission, script.Account, client, logger)
			if err != nil {
				return err
			}
			proposeBatch.Txs = append(proposeBatch.Txs, &payload.Any{PermsTx: tx})
		case *def.Send:
			announceProposalJob(job.Name, "Send", logger)
			job.Send.Source = FirstOf(job.Send.Source, script.Account)
			job.Send.Sequence, err = getAccountSequence(job.Send.Sequence, job.Send.Source, seqCache)
			if err != nil {
				return err
			}
			tx, err := FormulateSendJob(job.Send, script.Account, client, logger)
			if err != nil {
				return err
			}
			proposeBatch.Txs = append(proposeBatch.Txs, &payload.Any{SendTx: tx})
		case *def.QueryContract:
			announceProposalJob(job.Name, "Query Contract", logger)
			logger.InfoMsg("Query Contract jobs are IGNORED in proposals")

		case *def.Assert:
			announceProposalJob(job.Name, "Assert", logger)
			logger.InfoMsg("Assert jobs are IGNORED in proposals")
		default:
			return fmt.Errorf("jobs %s illegal job type for proposal", job.Name)
		}
	}

	return nil
}

func ProposalJob(prop *def.Proposal, do *def.DeployArgs, parentScript *def.Playbook, client *def.Client, logger *logging.Logger) (string, error) {
	var proposeBatch payload.BatchTx

	seqCache := acmstate.NewCache(client)

	err := recurseJobs(&proposeBatch, prop.Jobs, prop, do, parentScript, client, seqCache, logger)
	if err != nil {
		return "", err
	}

	proposal := payload.Proposal{Name: prop.Name, Description: prop.Description, BatchTx: &proposeBatch}

	proposalInput, err := client.TxInput(prop.ProposalAddress, "", prop.ProposalSequence, false, logger)
	if err != nil {
		return "", err
	}
	proposal.BatchTx.Inputs = []*payload.TxInput{proposalInput}
	proposalHash := proposal.Hash()

	var proposalTx *payload.ProposalTx
	if do.ProposeVerify {
		ballot, err := client.GetProposal(proposalHash, logger)
		if err != nil {
			logger.InfoMsg("Proposal could NOT be verified", "error", err)
			return "", err
		}

		err = proposals.ProposalExpired(ballot.Proposal, client, logger)
		if err != nil {
			logger.InfoMsg("Proposal verify FAILED", "error", err)
			return "", err
		}

		voteAddresses := make([]string, 0, len(ballot.Votes))
		for _, v := range ballot.Votes {
			voteAddresses = append(voteAddresses, v.Address.String())
		}

		logger.InfoMsg("Proposal VERIFY SUCCESSFUL",
			"votes count", len(ballot.Votes),
			"votes", voteAddresses)

		return "", err
	} else if do.ProposeVote {
		ballot, err := client.GetProposal(proposalHash, logger)
		if err != nil {
			logger.InfoMsg("Proposal could not be found", "error", err)
			return "", err
		}

		err = proposals.ProposalExpired(ballot.Proposal, client, logger)
		if err != nil {
			logger.InfoMsg("Proposal error", "error", err)
			return "", err
		}

		// proposal is there and current, let's vote for it
		input, err := client.TxInput(parentScript.Account, "", prop.Sequence, true, logger)
		if err != nil {
			return "", err
		}

		logger.InfoMsg("Voting for proposal", "hash", proposalHash)

		h := binary.HexBytes(proposalHash)
		proposalTx = &payload.ProposalTx{ProposalHash: &h, VotingWeight: 1, Input: input}
	} else if do.ProposeCreate {
		input, err := client.TxInput(FirstOf(prop.Source, parentScript.Account), "", prop.Sequence, true, logger)
		if err != nil {
			return "", err
		}
		logger.InfoMsg("Creating Proposal", "hash", proposalHash)

		bs, _ := json.Marshal(proposal)
		logger.TraceMsg("Proposal json", "json", string(bs))
		proposalTx = &payload.ProposalTx{VotingWeight: 1, Input: input, Proposal: &proposal}
	} else {
		logger.InfoMsg("please specify one of --proposal-create, --proposal-vote, --proposal-verify")
		return "", nil
	}

	txe, err := client.SignAndBroadcast(proposalTx, logger)
	if err != nil {
		var err = util.ChainErrorHandler(proposalTx.Input.Address.String(), err, logger)
		return "", err
	}

	result := fmt.Sprintf("%X", txe.Receipt.TxHash)

	return result, nil
}
