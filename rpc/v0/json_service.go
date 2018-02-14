// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v0

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/v0/server"
)

// EventSubscribe
type EventSub struct {
	SubId string `json:"sub_id"`
}

// EventUnsubscribe
type EventUnsub struct {
	Result bool `json:"result"`
}

// EventPoll
type PollResponse struct {
	Events []interface{} `json:"events"`
}

// Server used to handle JSON-RPC 2.0 requests. Implements server.Server
type JsonRpcServer struct {
	service server.HttpService
	running bool
}

// Create a new JsonRpcServer
func NewJSONServer(service server.HttpService) *JsonRpcServer {
	return &JsonRpcServer{service: service}
}

// Start adds the rpc path to the router.
func (jrs *JsonRpcServer) Start(config *server.ServerConfig,
	router *gin.Engine) {
	router.POST(config.HTTP.JsonRpcEndpoint, jrs.handleFunc)
	jrs.running = true
}

// Is the server currently running?
func (jrs *JsonRpcServer) Running() bool {
	return jrs.running
}

// Shut the server down. Does nothing.
func (jrs *JsonRpcServer) Shutdown(ctx context.Context) error {
	jrs.running = false
	return nil
}

// Handler passes message on directly to the service which expects
// a normal http request and response writer.
func (jrs *JsonRpcServer) handleFunc(c *gin.Context) {
	r := c.Request
	w := c.Writer

	jrs.service.Process(r, w)
}

// Used for Burrow. Implements server.HttpService
type JSONService struct {
	codec           rpc.Codec
	service         rpc.Service
	eventSubs       *Subscriptions
	defaultHandlers map[string]RequestHandlerFunc
}

// Create a new JSON-RPC 2.0 service for burrow (tendermint).
func NewJSONService(codec rpc.Codec, service rpc.Service) server.HttpService {

	tmhttps := &JSONService{
		codec:     codec,
		service:   service,
		eventSubs: NewSubscriptions(service),
	}

	dhMap := GetMethods(codec, service)
	// Events
	dhMap[EVENT_SUBSCRIBE] = tmhttps.EventSubscribe
	dhMap[EVENT_UNSUBSCRIBE] = tmhttps.EventUnsubscribe
	dhMap[EVENT_POLL] = tmhttps.EventPoll
	tmhttps.defaultHandlers = dhMap
	return tmhttps
}

// Process a request.
func (js *JSONService) Process(r *http.Request, w http.ResponseWriter) {

	// Create new request object and unmarshal.
	req := &rpc.RPCRequest{}
	decoder := json.NewDecoder(r.Body)
	errU := decoder.Decode(req)

	// Error when decoding.
	if errU != nil {
		js.writeError("Failed to parse request: "+errU.Error(), "",
			rpc.PARSE_ERROR, w)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		js.writeError("Wrong protocol version: "+req.JSONRPC, req.Id,
			rpc.INVALID_REQUEST, w)
		return
	}

	mName := req.Method

	if handler, ok := js.defaultHandlers[mName]; ok {
		resp, errCode, err := handler(req, w)
		if err != nil {
			js.writeError(err.Error(), req.Id, errCode, w)
		} else {
			js.writeResponse(req.Id, resp, w)
		}
	} else {
		js.writeError("Method not found: "+mName, req.Id, rpc.METHOD_NOT_FOUND, w)
	}
}

// Helper for writing error responses.
func (js *JSONService) writeError(msg, id string, code int, w http.ResponseWriter) {
	response := rpc.NewRPCErrorResponse(id, code, msg)
	err := js.codec.Encode(response, w)
	// If there's an error here all bets are off.
	if err != nil {
		http.Error(w, "Failed to marshal standard error response: "+err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

// Helper for writing responses.
func (js *JSONService) writeResponse(id string, result interface{}, w http.ResponseWriter) {
	response := rpc.NewRPCResponse(id, result)
	err := js.codec.Encode(response, w)
	if err != nil {
		js.writeError("Internal error: "+err.Error(), id, rpc.INTERNAL_ERROR, w)
		return
	}
	w.WriteHeader(200)
}

// *************************************** Events ************************************

// Subscribe to an
func (js *JSONService) EventSubscribe(request *rpc.RPCRequest,
	requester interface{}) (interface{}, int, error) {
	param := &EventIdParam{}
	err := json.Unmarshal(request.Params, param)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	eventId := param.EventId
	subId, errC := js.eventSubs.Add(eventId)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &EventSub{subId}, 0, nil
}

// Un-subscribe from an
func (js *JSONService) EventUnsubscribe(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SubIdParam{}
	err := json.Unmarshal(request.Params, param)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	subId := param.SubId

	err = js.service.Unsubscribe(context.Background(), subId)
	if err != nil {
		return nil, rpc.INTERNAL_ERROR, err
	}
	return &EventUnsub{true}, 0, nil
}

// Check subscription event cache for new data.
func (js *JSONService) EventPoll(request *rpc.RPCRequest,
	requester interface{}) (interface{}, int, error) {
	param := &SubIdParam{}
	err := json.Unmarshal(request.Params, param)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	subId := param.SubId

	result, errC := js.eventSubs.Poll(subId)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &PollResponse{result}, 0, nil
}
