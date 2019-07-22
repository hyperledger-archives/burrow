package service

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/rpc/rpcquery"
)

// AbiProvider provides a method for loading ABIs from disk, and retrieving them from burrow on-demand
type AbiProvider struct {
	abiSpec *abi.Spec
	cli     rpcquery.QueryClient
}

// NewAbiProvider loads ABIs from the filesystem. A set of zero or more files or directories can be passed in the path
// argument. If an event is encountered for which no ABI is known, it is retrieved from burrow
func NewAbiProvider(paths []string, cli rpcquery.QueryClient) (provider *AbiProvider, err error) {
	abiSpec := &abi.Spec{}
	if len(paths) > 0 {
		abiSpec, err = abi.LoadPath(paths...)
		if err != nil {
			return nil, err
		}
	}

	provider = &AbiProvider{
		abiSpec,
		cli,
	}
	return
}

// GetEventAbi get the ABI for a particular eventID. If it is not known, it is retrieved from the burrow node via
// the address for the contract
func (p *AbiProvider) GetEventAbi(eventID abi.EventID, address crypto.Address, l *logging.Logger) (*abi.EventSpec, error) {
	evAbi, ok := p.abiSpec.EventsByID[eventID]
	if !ok {
		resp, err := p.cli.GetMetadata(context.Background(), &rpcquery.GetMetadataParam{Address: &address})
		if err != nil {
			l.InfoMsg("Error retrieving abi for event", "address", address.String(), "eventid", eventID.String(), "error", err)
			return nil, err
		}
		if resp == nil || resp.Metadata == "" {
			l.InfoMsg("ABI not found for contract", "address", address.String(), "eventid", eventID.String())
			return nil, fmt.Errorf("No ABI present for contract at address %v", address)
		}
		a, err := abi.ReadSpec([]byte(resp.Metadata))
		if err != nil {
			l.InfoMsg("Failed to parse abi", "address", address.String(), "eventid", eventID.String(), "abi", resp.Metadata)
			return nil, err
		}
		evAbi, ok = a.EventsByID[eventID]
		if !ok {
			l.InfoMsg("Event missing from ABI spec for contract", "address", address.String(), "eventid", eventID.String(), "abi", resp.Metadata)
			return nil, fmt.Errorf("Event missing from ABI spec for contract")
		}

		p.abiSpec = abi.MergeSpec([]*abi.Spec{p.abiSpec, a})
	}

	return &evAbi, nil
}
