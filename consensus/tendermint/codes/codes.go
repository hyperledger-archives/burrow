package codes

import (
	abci_types "github.com/tendermint/tendermint/abci/types"
)

const (
	// Success
	TxExecutionSuccessCode uint32 = abci_types.CodeTypeOK

	// Informational
	UnsupportedRequestCode uint32 = 400

	// Internal errors
	EncodingErrorCode    uint32 = 500
	TxExecutionErrorCode uint32 = 501
	CommitErrorCode      uint32 = 502
)
