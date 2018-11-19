package contexts

import (
	"crypto/sha256"
	"fmt"
	"unicode"

	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/crypto/sha3"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/burrow/txs/payload"
)

type ProposalContext struct {
	Tip          bcm.BlockchainInfo
	StateWriter  state.ReaderWriter
	ValidatorSet validator.Writer
	ProposalReg  proposal.ReaderWriter
	Logger       *logging.Logger
	tx           *payload.ProposalTx
	Contexts     map[payload.Type]Context
}

func (ctx *ProposalContext) Execute(txe *exec.TxExecution, p payload.Payload) error {
	var ok bool
	ctx.tx, ok = p.(*payload.ProposalTx)
	if !ok {
		return fmt.Errorf("payload must be ProposalTx, but is: %v", txe.Envelope.Tx.Payload)
	}
	// Validate input
	inAcc, err := ctx.StateWriter.GetAccount(ctx.tx.Input.Address)
	if err != nil {
		return err
	}
	if inAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeInvalidAddress
	}

	// check permission
	if !hasProposalPermission(ctx.StateWriter, inAcc, ctx.Logger) {
		return fmt.Errorf("account %s does not have Proposal permission", ctx.tx.Input.Address)
	}

	var ballot *payload.Ballot
	var proposalHash []byte

	if ctx.tx.Proposal == nil {
		// voting for existing proposal
		if ctx.tx.ProposalHash == nil || ctx.tx.ProposalHash.Size() != sha256.Size {
			return errors.ErrorCodeInvalidProposal
		}
		ballot, err = ctx.ProposalReg.GetProposal(ctx.tx.ProposalHash.Bytes())
		if err != nil {
			return err
		}
		proposalHash = ctx.tx.ProposalHash.Bytes()
	} else {
		if ctx.tx.ProposalHash != nil || ctx.tx.Proposal.BatchTx == nil ||
			len(ctx.tx.Proposal.BatchTx.Txs) == 0 || len(ctx.tx.Proposal.BatchTx.GetInputs()) == 0 {
			return errors.ErrorCodeInvalidProposal
		}

		// validate the input strings
		if err := validateProposalStrings(ctx.tx.Proposal); err != nil {
			return err
		}

		proposalHash = ctx.tx.Proposal.Hash()

		ballot, err = ctx.ProposalReg.GetProposal(proposalHash)
		if ballot == nil && err == nil {
			ballot = &payload.Ballot{
				Proposal:      ctx.tx.Proposal,
				ProposalState: payload.Ballot_PROPOSED,
			}
		}
		if err != nil {
			return err
		}
	}

	// count votes for proposal
	votes := make(map[crypto.Address]int64)

	if ballot.Votes == nil {
		ballot.Votes = make([]*payload.Vote, 0)
	}

	for _, v := range ballot.Votes {
		acc, err := ctx.StateWriter.GetAccount(v.Address)
		if err != nil {
			return err
		}
		// Belt and braces, should have already been checked
		if !hasProposalPermission(ctx.StateWriter, acc, ctx.Logger) {
			return fmt.Errorf("account %s does not have Proposal permission", ctx.tx.Input.Address)
		}
		votes[v.Address] = v.VotingWeight
	}

	for _, i := range ballot.Proposal.BatchTx.GetInputs() {
		// Validate input
		proposeAcc, err := ctx.StateWriter.GetAccount(i.Address)
		if err != nil {
			return err
		}

		if proposeAcc == nil {
			ctx.Logger.InfoMsg("Cannot find input account",
				"tx_input", ctx.tx.Input)
			return errors.ErrorCodeInvalidAddress
		}

		if !hasBatchPermission(ctx.StateWriter, proposeAcc, ctx.Logger) {
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

	for _, step := range ballot.Proposal.BatchTx.Txs {
		txE := txs.EnvelopeFromAny("", step)

		for _, i := range txE.Tx.GetInputs() {
			_, err := ctx.StateWriter.GetAccount(i.Address)
			if err != nil {
				return err
			}

			// Do not check sequence numbers of inputs
		}
	}

	if power >= ctx.Tip.GenesisDoc().Params.ProposalThreshold {
		ballot.ProposalState = payload.Ballot_EXECUTED

		for i, step := range ballot.Proposal.BatchTx.Txs {
			txE := txs.EnvelopeFromAny("", step)

			txe.PayloadEvent(&exec.PayloadEvent{TxType: txE.Tx.Type(), Index: uint32(i)})

			if txExecutor, ok := ctx.Contexts[txE.Tx.Type()]; ok {
				err = txExecutor.Execute(txe, txE.Tx.Payload)
				if err != nil {
					ctx.Logger.InfoMsg("Transaction execution failed", structure.ErrorKey, err)
					return err
				}
			}

			if txe.Exception != nil {
				ballot.ProposalState = payload.Ballot_FAILED
				break
			}
		}
	}

	return ctx.ProposalReg.UpdateProposal(proposalHash, ballot)
}

func validateProposalStrings(proposal *payload.Proposal) error {
	if len(proposal.Name) == 0 {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString, "name must not be empty")
	}

	if !validateNameRegEntryName(proposal.Name) {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString,
			"Invalid characters found in Proposal.Name (%s). Only alphanumeric, underscores, dashes, forward slashes, and @ are allowed", proposal.Name)
	}

	if !validateStringPrintable(proposal.Description) {
		return errors.ErrorCodef(errors.ErrorCodeInvalidString,
			"Invalid characters found in Proposal.Description (%s). Only printable characters are allowed", proposal.Description)
	}

	return nil
}

func validateStringPrintable(data string) bool {
	for _, r := range []rune(data) {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
