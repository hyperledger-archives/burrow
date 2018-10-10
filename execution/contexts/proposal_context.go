package contexts

import (
	"fmt"
	"unicode"

	"github.com/hyperledger/burrow/acm/state"
	"github.com/hyperledger/burrow/acm/validator"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm/sha3"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/execution/proposal"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
)

type ProposalContext struct {
	Tip          bcm.BlockchainInfo
	StateWriter  state.ReaderWriter
	ValidatorSet validator.Writer
	ProposalReg  proposal.ReaderWriter
	Logger       *logging.Logger
	tx           *payload.ProposalTx
	majorityVote bool
	proposal     *payload.Proposal
}

func (ctx *ProposalContext) Execute(txe *exec.TxExecution) error {
	var ok bool
	ctx.tx, ok = txe.Envelope.Tx.Payload.(*payload.ProposalTx)
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
	if ctx.tx.Input.Amount < ctx.tx.Fee {
		ctx.Logger.InfoMsg("Sender did not send enough to cover the fee",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeInsufficientFunds
	}

	proposal := ctx.tx.Proposal
	var proposalHash []byte

	if proposal == nil {
		// voting for existing proposal
		if ctx.tx.ProposalHash == nil {
			return errors.ErrorCodeInvalidProposal
		}
		proposal, err = ctx.ProposalReg.GetProposal(ctx.tx.ProposalHash.Bytes())
		if err != nil {
			return err
		}
	} else {
		if ctx.tx.ProposalHash != nil || proposal.Votes != nil || len(proposal.BatchTx.Txs) == 0 {
			return errors.ErrorCodeInvalidProposal
		}

		// validate the input strings
		if err := validateProposalStrings(proposal); err != nil {
			return err
		}

		bs, err := proposal.Encode()
		if err != nil {
			return err
		}

		proposalHash = sha3.Sha3(bs)

		prop, err := ctx.ProposalReg.GetProposal(proposalHash)
		if err == nil && prop != nil {
			// vote for prop
			proposal = prop
		}
	}

	if proposal.Executed == true {
		return errors.ErrorCodeProposalExecuted
	}

	// Validate input
	proposeAcc, err := ctx.StateWriter.GetAccount(proposal.Input.Address)
	if err != nil {
		return err
	}

	if proposeAcc == nil {
		ctx.Logger.InfoMsg("Cannot find input account",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeInvalidAddress
	}

	if proposeAcc.GetSequence()+1 != proposal.Input.Sequence {
		ctx.Logger.InfoMsg("Expired sequence number",
			"tx_input", ctx.tx.Input)
		return errors.ErrorCodeExpiredProposal
	}

	// vote for thing
	votes := make(map[crypto.Address]int64)
	seenOurVote := false
	for _, v := range proposal.Votes {
		if v.Address == ctx.tx.Input.Address {
			seenOurVote = true
		}
		acc, err := ctx.StateWriter.GetAccount(v.Address)
		if err != nil {
			return err
		}
		if !hasProposalPermission(ctx.StateWriter, acc, ctx.Logger) {
			return fmt.Errorf("account %s does not have Proposal permission", ctx.tx.Input.Address)
		}

		votes[v.Address] = v.VotingWeight
	}
	if !seenOurVote {
		proposal.Votes = append(proposal.Votes, &payload.Vote{Address: ctx.tx.Input.Address, VotingWeight: ctx.tx.VotingWeight})
		votes[ctx.tx.Input.Address] = ctx.tx.VotingWeight
	}

	// Count the number of validators; ensure we have at least half the number of validators
	// This also means that when running with a single validator, a proposal will run straight away
	power := 0
	for _, v := range votes {
		if v > 0 {
			power++
		}
	}
	proposal.Executed = power*2 > ctx.Tip.NumValidators()

	// The CheckTx happens in BatchExecuter; do minimal checking here
	for _, tx := range proposal.BatchTx.Txs {
		if tx.CallTx != nil {
			if tx.CallTx.Input.Address != proposal.Input.Address {
				return errors.ErrorCodeInvalidAddress
			}
		}
	}

	ctx.proposal = proposal
	return ctx.ProposalReg.UpdateProposal(proposalHash, proposal)
}

func (ctx *ProposalContext) GetProposal() (*payload.Proposal, bool) {
	return ctx.proposal, ctx.majorityVote
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
