package service

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/vent/chain"
)

type EventSpecGetter func(abi.EventID, crypto.Address) (*abi.EventSpec, error)

// AbiProvider provides a method for loading ABIs from disk, and retrieving them from burrow on-demand
type AbiProvider struct {
	abiSpec *abi.Spec
	chain   chain.Chain
	logger  *logging.Logger
}

// NewAbiProvider loads ABIs from the filesystem. A set of zero or more files or directories can be passed in the path
// argument. If an event is encountered for which no ABI is known, it is retrieved from burrow
func NewAbiProvider(paths []string, chain chain.Chain, logger *logging.Logger) (provider *AbiProvider, err error) {
	abiSpec := abi.NewSpec()
	if len(paths) > 0 {
		abiSpec, err = abi.LoadPath(paths...)
		if err != nil {
			return nil, err
		}
	}

	provider = &AbiProvider{
		abiSpec: abiSpec,
		chain:   chain,
		logger:  logger.WithScope("NewAbiProvider"),
	}
	return
}

// GetEventAbi get the ABI for a particular eventID. If it is not known, it is retrieved from the burrow node via
// the address for the contract
func (p *AbiProvider) GetEventAbi(eventID abi.EventID, address crypto.Address) (*abi.EventSpec, error) {
	evAbi, ok := p.abiSpec.EventsByID[eventID]
	if !ok {
		metadata, err := p.chain.GetABI(context.Background(), address)
		if err != nil {
			p.logger.InfoMsg("Error retrieving abi for event", "address", address.String(), "eventid", eventID.String(), "error", err)
			return nil, err
		}
		if metadata == "" {
			p.logger.InfoMsg("ABI not found for contract", "address", address.String(), "eventid", eventID.String())
			return nil, fmt.Errorf("No ABI present for contract at address %v", address)
		}
		a, err := abi.ReadSpec([]byte(metadata))
		if err != nil {
			p.logger.InfoMsg("Failed to parse abi", "address", address.String(), "eventid", eventID.String(), "abi", metadata)
			return nil, err
		}
		evAbi, ok = a.EventsByID[eventID]
		if !ok {
			p.logger.InfoMsg("Event missing from ABI spec for contract", "address", address.String(), "eventid", eventID.String(), "abi", metadata)
			return nil, fmt.Errorf("Event missing from ABI spec for contract")
		}

		p.abiSpec = abi.MergeSpec([]*abi.Spec{p.abiSpec, a})
	}

	return evAbi, nil
}
