package contexts

import (
	"crypto/sha256"
	"fmt"
	"runtime/debug"
	"unicode"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

type ProposalContext struct {
	ChainID           string
	ProposalThreshold uint64
	State             acmstate.ReaderWriter
	ValidatorSet      validator.Writer
	ProposalReg       proposal.ReaderWriter
	Logger            *logging.Logger
	tx                *payload.ProposalTx
	Contexts          map[payload.Type]Context
}

func (ctx *ProposalContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.ProposalTx)
	if !ok {
		return fmt.Errorf("payload must be ProposalTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	// Validate input
	inAcc, err := ctx.State.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	}

	if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return errors.Code.InvalidAddress
	}

	// check permission
	if !hasProposalPermission(ctx.State, inAcc, ctx.Logger) {
		return fmt.Errorf("account %s does not have Proposal permission", ctx.tx.Input.Address)
	}

	var ballot *payload.Ballot
	var proposalHash []byte

	if ctx.tx.Proposal == nil {
		// voting for existing proposal
		if ctx.tx.ProposalHash == nil || ctx.tx.ProposalHash.Size() != sha256.Size {
			return errors.Code.InvalidProposal
		}

		proposalHash = ctx.tx.ProposalHash.Bytes()
		ballot, err = ctx.ProposalReg.GetProposal(proposalHash)
		if err != nil {
			return err
		}
	} else {
		if ctx.tx.ProposalHash != nil || ctx.tx.Proposal.BatchTx == nil ||
			len(ctx.tx.Proposal.BatchTx.Txs) == 0 || len(ctx.tx.Proposal.BatchTx.GetInputs()) == 0 {
			return errors.Code.InvalidProposal
		}

		// validate the input strings
		if err := validateProposalStrings(ctx.tx.Proposal); err != nil {
			return err
		}

		proposalHash = ctx.tx.Proposal.Hash()

		ballot, err = ctx.ProposalReg.GetProposal(proposalHash)
		if err != nil {
			return err
		}

		if ballot == nil {
			ballot = &payload.Ballot{
				Proposal:      ctx.tx.Proposal,
				ProposalState: payload.Ballot_PROPOSED,
			}
		}

		// else vote for existing proposal
	}

	// Check that we have not voted this already
	for _, vote := range ballot.Votes {
		for _, i := range ctx.tx.GetInputs() {
			if i.Address == vote.Address {
				return errors.Code.AlreadyVoted
			}
		}
	}

	// count votes for proposal
	votes := make(map[crypto.Address]int64)

	if ballot.Votes == nil {
		ballot.Votes = make([]*payload.Vote, 0)
	}

	for _, v := range ballot.Votes {
		acc, err := ctx.State.GetAccount(v.Address)
		if err != nil {
			return err
		}
		// Belt and braces, should have already been checked
		if !hasProposalPermission(ctx.State, acc, ctx.Logger) {
			return fmt.Errorf("account %s does not have Proposal permission", ctx.tx.Input.Address)
		}
		votes[v.Address] = v.VotingWeight
	}

	for _, i := range ballot.Proposal.BatchTx.GetInputs() {
		// Validate input
		proposeAcc, err := ctx.State.GetAccount(i.Address)
		if err != nil {
			return err
		}

		if proposeAcc == nil {
			ctx.Logger.InfoMsg("Cannot find input account",
				"tx_input", ctx.tx.Input)
			return errors.Code.InvalidAddress
		}

		if !hasBatchPermission(ctx.State, proposeAcc, ctx.Logger) {
			return fmt.Errorf("account %s does not have batch permission", i.Address)
		}

		if proposeAcc.GetSequence()+1 != i.Sequence {
			return fmt.Errorf("proposal expired, sequence number for account %s wrong", i.Address)
		}
	}

	for _, i := range ctx.tx.GetInputs() {
		// Do we have a record of our own vote
		if _, ok := votes[i.Address]; !ok {
			votes[i.Address] = ctx.tx.VotingWeight
			ballot.Votes = append(ballot.Votes, &payload.Vote{Address: i.Address, VotingWeight: ctx.tx.VotingWeight})
		}
	}

	// Count the number of validators; ensure we have at least half the number of validators
	// This also means that when running with a single validator, a proposal will run straight away
	var power uint64
	for _, v := range votes {
		if v > 0 {
			power++
		}
	}

	stateCache := acmstate.NewCache(ctx.State)

	for i, step := range ballot.Proposal.BatchTx.Txs {
		txEnv := txs.EnvelopeFromAny(ctx.ChainID, step)

		for _, input := range txEnv.Tx.GetInputs() {
			acc, err := stateCache.GetAccount(input.Address)
			if err != nil {
				return err
			}

			acc.Sequence++

			if acc.Sequence != input.Sequence {
				return fmt.Errorf("proposal expired, sequence number %d for account %s wrong at step %d", input.Sequence, input.Address, i+1)
			}

			err = stateCache.UpdateAccount(acc)
			if err != nil {
				return err
			}
		}
	}

	if power >= ctx.ProposalThreshold {
		ballot.ProposalState = payload.Ballot_EXECUTED

		txe.TxExecutions = make([]*exec.TxExecution, 0)

		for i, step := range ballot.Proposal.BatchTx.Txs {
			txEnv := txs.EnvelopeFromAny(ctx.ChainID, step)

			containedTxe := exec.NewTxExecution(txEnv)

			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recovered from panic in executor.Execute(%s): %v\n%s", txEnv.String(), r,
						debug.Stack())
				}
			}()

			for _, input := range txEnv.Tx.GetInputs() {
				acc, err := ctx.State.GetAccount(input.Address)
				if err != nil {
					return err
				}

				acc.Sequence++

				if input.Address != acc.GetAddress() {
					return fmt.Errorf("trying to validate input from address %v but passed account %v", input.Address,
						acc.GetAddress())
				}

				if acc.Sequence != input.Sequence {
					return fmt.Errorf("proposal expired, sequence number %d for account %s wrong at step %d", input.Sequence, input.Address, i+1)
				}

				ctx.State.UpdateAccount(acc)
			}

			if txExecutor, ok := ctx.Contexts[txEnv.Tx.Type()]; ok {
				err = txExecutor.Execute(containedTxe, txEnv.Tx.Payload)

				if err != nil {
					ctx.Logger.InfoMsg("Transaction execution failed", structure.ErrorKey, err)
					return err
				}
			}

			txe.TxExecutions = append(txe.TxExecutions, containedTxe)

			if containedTxe.Exception != nil {
				ballot.ProposalState = payload.Ballot_FAILED
				break
			}
		}
	}

	return ctx.ProposalReg.UpdateProposal(proposalHash, ballot)
}

func validateProposalStrings(proposal *payload.Proposal) error {
	if len(proposal.Name) == 0 {
		return errors.Errorf(errors.Code.InvalidString, "name must not be empty")
	}

	if !validateNameRegEntryName(proposal.Name) {
		return errors.Errorf(errors.Code.InvalidString,
			"Invalid characters found in Proposal.Name (%s). Only alphanumeric, underscores, dashes, forward slashes, and @ are allowed", proposal.Name)
	}

	if !validateStringPrintable(proposal.Description) {
		return errors.Errorf(errors.Code.InvalidString,
			"Invalid characters found in Proposal.Description (%s). Only printable characters are allowed", proposal.Description)
	}

	return nil
}

func validateStringPrintable(data string) bool {
	for _, r := range data {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
