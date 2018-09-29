package codes

import (
	abci_types "github.com/tendermint/tendermint/abci/types"
)

const (
	// Success
	TxExecutionSuccessCode   uint32 = abci_types.CodeTypeOK
	PeerFilterAuthorizedCode uint32 = abci_types.CodeTypeOK

	// Informational
	UnsupportedRequestCode  uint32 = 400
	PeerFilterForbiddenCode uint32 = 403

	// Internal errors
	EncodingErrorCode    uint32 = 500
	TxExecutionErrorCode uint32 = 501
	CommitErrorCode      uint32 = 502
)
