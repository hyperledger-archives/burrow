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
	"encoding/json"
	"net/http"

	definitions "github.com/hyperledger/burrow/definitions"
	event "github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/rpc"
	server "github.com/hyperledger/burrow/server"

	"github.com/gin-gonic/gin"
)

// Server used to handle JSON-RPC 2.0 requests. Implements server.Server
type JsonRpcServer struct {
	service server.HttpService
	running bool
}

// Create a new JsonRpcServer
func NewJsonRpcServer(service server.HttpService) *JsonRpcServer {
	return &JsonRpcServer{service: service}
}

// Start adds the rpc path to the router.
func (this *JsonRpcServer) Start(config *server.ServerConfig,
	router *gin.Engine) {
	router.POST(config.HTTP.JsonRpcEndpoint, this.handleFunc)
	this.running = true
}

// Is the server currently running?
func (this *JsonRpcServer) Running() bool {
	return this.running
}

// Shut the server down. Does nothing.
func (this *JsonRpcServer) ShutDown() {
	this.running = false
}

// Handler passes message on directly to the service which expects
// a normal http request and response writer.
func (this *JsonRpcServer) handleFunc(c *gin.Context) {
	r := c.Request
	w := c.Writer

	this.service.Process(r, w)
}

// Used for Burrow. Implements server.HttpService
type BurrowJsonService struct {
	codec           rpc.Codec
	pipe            definitions.Pipe
	eventSubs       *event.Subscriptions
	defaultHandlers map[string]RequestHandlerFunc
}

// Create a new JSON-RPC 2.0 service for burrow (tendermint).
func NewBurrowJsonService(codec rpc.Codec, pipe definitions.Pipe,
	eventSubs *event.Subscriptions) server.HttpService {

	tmhttps := &BurrowJsonService{codec: codec, pipe: pipe, eventSubs: eventSubs}
	mtds := NewBurrowMethods(codec, pipe)

	dhMap := mtds.getMethods()
	// Events
	dhMap[EVENT_SUBSCRIBE] = tmhttps.EventSubscribe
	dhMap[EVENT_UNSUBSCRIBE] = tmhttps.EventUnsubscribe
	dhMap[EVENT_POLL] = tmhttps.EventPoll
	tmhttps.defaultHandlers = dhMap
	return tmhttps
}

// Process a request.
func (this *BurrowJsonService) Process(r *http.Request, w http.ResponseWriter) {

	// Create new request object and unmarshal.
	req := &rpc.RPCRequest{}
	decoder := json.NewDecoder(r.Body)
	errU := decoder.Decode(req)

	// Error when decoding.
	if errU != nil {
		this.writeError("Failed to parse request: "+errU.Error(), "",
			rpc.PARSE_ERROR, w)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		this.writeError("Wrong protocol version: "+req.JSONRPC, req.Id,
			rpc.INVALID_REQUEST, w)
		return
	}

	mName := req.Method

	if handler, ok := this.defaultHandlers[mName]; ok {
		resp, errCode, err := handler(req, w)
		if err != nil {
			this.writeError(err.Error(), req.Id, errCode, w)
		} else {
			this.writeResponse(req.Id, resp, w)
		}
	} else {
		this.writeError("Method not found: "+mName, req.Id, rpc.METHOD_NOT_FOUND, w)
	}
}

// Helper for writing error responses.
func (this *BurrowJsonService) writeError(msg, id string, code int, w http.ResponseWriter) {
	response := rpc.NewRPCErrorResponse(id, code, msg)
	err := this.codec.Encode(response, w)
	// If there's an error here all bets are off.
	if err != nil {
		http.Error(w, "Failed to marshal standard error response: "+err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

// Helper for writing responses.
func (this *BurrowJsonService) writeResponse(id string, result interface{}, w http.ResponseWriter) {
	response := rpc.NewRPCResponse(id, result)
	err := this.codec.Encode(response, w)
	if err != nil {
		this.writeError("Internal error: "+err.Error(), id, rpc.INTERNAL_ERROR, w)
		return
	}
	w.WriteHeader(200)
}

// *************************************** Events ************************************

// Subscribe to an event.
func (this *BurrowJsonService) EventSubscribe(request *rpc.RPCRequest,
	requester interface{}) (interface{}, int, error) {
	param := &EventIdParam{}
	err := json.Unmarshal(request.Params, param)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	eventId := param.EventId
	subId, errC := this.eventSubs.Add(eventId)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &event.EventSub{subId}, 0, nil
}

// Un-subscribe from an event.
func (this *BurrowJsonService) EventUnsubscribe(request *rpc.RPCRequest,
	requester interface{}) (interface{}, int, error) {
	param := &SubIdParam{}
	err := json.Unmarshal(request.Params, param)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	subId := param.SubId

	errC := this.pipe.Events().Unsubscribe(subId)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &event.EventUnsub{true}, 0, nil
}

// Check subscription event cache for new data.
func (this *BurrowJsonService) EventPoll(request *rpc.RPCRequest,
	requester interface{}) (interface{}, int, error) {
	param := &SubIdParam{}
	err := json.Unmarshal(request.Params, param)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	subId := param.SubId

	result, errC := this.eventSubs.Poll(subId)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &event.PollResponse{result}, 0, nil
}
