package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
)

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"` // must be map[string]interface{} or []interface{}
}

func NewRPCRequest(id string, method string, params json.RawMessage) RPCRequest {
	return RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

func (req RPCRequest) String() string {
	return fmt.Sprintf("[%s %s]", req.ID, req.Method)
}

func NewRequest(id string, method string, params interface{}) (RPCRequest, error) {
	var payload json.RawMessage
	var err error
	if params != nil {
		payload, err = json.Marshal(params)
		if err != nil {
			return RPCRequest{}, err
		}
	}
	request := NewRPCRequest(id, method, payload)
	return request, nil
}

type RPCError struct {
	Code    RPCErrorCode    `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (err *RPCError) IsServerError() bool {
	return err.Code.IsServerError()
}

func (err *RPCError) Error() string {
	const baseFormat = "%v - %s"
	if len(err.Data) > 0 {
		return fmt.Sprintf(baseFormat+": %s", err.Code, err.Message, err.Data)
	}
	return fmt.Sprintf(baseFormat, err.Code, err.Message)
}

func (err *RPCError) HTTPStatusCode() int {
	if err == nil {
		return 200
	}
	return err.Code.HTTPStatusCode()
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

func NewRPCSuccessResponse(id string, res interface{}) RPCResponse {
	var rawMsg json.RawMessage

	if res != nil {
		var err error
		rawMsg, err = json.Marshal(res)
		if err != nil {
			return RPCInternalError(id, errors.Wrap(err, "Error marshalling response"))
		}
	}

	return RPCResponse{JSONRPC: "2.0", ID: id, Result: rawMsg}
}

func NewRPCErrorResponse(id string, code RPCErrorCode, data string) RPCResponse {
	var bs []byte
	if data != "" {
		var err error
		bs, err = json.Marshal(data)
		if err != nil {
			panic(fmt.Errorf("unexpected error JSON marshalling string: %w", err))
		}
	}
	return RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: code.String(), Data: bs},
	}
}

func (resp RPCResponse) String() string {
	if resp.Error == nil {
		return fmt.Sprintf("[%s %v]", resp.ID, resp.Result)
	}
	return fmt.Sprintf("[%s %s]", resp.ID, resp.Error)
}

func RPCParseError(id string, err error) RPCResponse {
	return NewRPCErrorResponse(id, RPCErrorCodeParseError, err.Error())
}

func RPCInvalidRequestError(id string, err error) RPCResponse {
	return NewRPCErrorResponse(id, RPCErrorCodeInvalidRequest, err.Error())
}

func RPCMethodNotFoundError(id string) RPCResponse {
	return NewRPCErrorResponse(id, RPCErrorCodeMethodNotFound, "")
}

func RPCInvalidParamsError(id string, err error) RPCResponse {
	return NewRPCErrorResponse(id, RPCErrorCodeInvalidParams, err.Error())
}

func RPCInternalError(id string, err error) RPCResponse {
	return NewRPCErrorResponse(id, RPCErrorCodeInternalError, err.Error())
}

// EventSubscriber mirros tendermint/tendermint/types.EventBusSubscriber
type EventSubscriber interface {
	Subscribe(ctx context.Context, subscriber string, query tmpubsub.Query, out chan<- interface{}) error
	Unsubscribe(ctx context.Context, subscriber string, query tmpubsub.Query) error
	UnsubscribeAll(ctx context.Context, subscriber string) error
}
