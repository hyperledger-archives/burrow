package types

import (
	"fmt"
	"net/http"
)

// From JSONRPC 2.0 spec
type RPCErrorCode int

const (
	RPCErrorCodeParseError     RPCErrorCode = -32700
	RPCErrorCodeInvalidRequest RPCErrorCode = -32600
	RPCErrorCodeMethodNotFound RPCErrorCode = -32601
	RPCErrorCodeInvalidParams  RPCErrorCode = -32602
	RPCErrorCodeInternalError  RPCErrorCode = -32603
	// Available for custom server-defined errors
	RPCErrorCodeServerErrorStart RPCErrorCode = -32000
	RPCErrorCodeServerErrorEnd   RPCErrorCode = -32099
)

func (code RPCErrorCode) String() string {
	switch code {
	case RPCErrorCodeParseError:
		return "Parse Error"
	case RPCErrorCodeInvalidRequest:
		return "Parse Error"
	case RPCErrorCodeMethodNotFound:
		return "Method Not Found"
	case RPCErrorCodeInvalidParams:
		return "Invalid Params"
	case RPCErrorCodeInternalError:
		return "Internal Error"
	default:
		if code.IsServerError() {
			return fmt.Sprintf("Server Error %d", code)
		}
		return fmt.Sprintf("Unknown Error %d", code)
	}
}

func (code RPCErrorCode) IsServerError() bool {
	return code >= RPCErrorCodeServerErrorStart && code <= RPCErrorCodeServerErrorEnd
}

func (code RPCErrorCode) HTTPStatusCode() int {
	switch code {
	case RPCErrorCodeInvalidRequest:
		return http.StatusBadRequest
	case RPCErrorCodeMethodNotFound:
		return http.StatusMethodNotAllowed
	default:
		return http.StatusInternalServerError
	}
}

func (code RPCErrorCode) Error() string {
	return code.String()
}
