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
	"fmt"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/v0/server"
)

// Used for Burrow. Implements WebSocketService.
type WebsocketService struct {
	codec           rpc.Codec
	service         *rpc.Service
	defaultHandlers map[string]RequestHandlerFunc
	logger          *logging.Logger
}

// Create a new websocket service.
func NewWebsocketService(codec rpc.Codec, service *rpc.Service, logger *logging.Logger) server.WebSocketService {
	tmwss := &WebsocketService{
		codec:   codec,
		service: service,
		logger:  logger.WithScope("NewWebsocketService"),
	}
	dhMap := GetMethods(codec, service, tmwss.logger)
	// Events
	dhMap[EVENT_SUBSCRIBE] = tmwss.EventSubscribe
	dhMap[EVENT_UNSUBSCRIBE] = tmwss.EventUnsubscribe
	tmwss.defaultHandlers = dhMap
	return tmwss
}

// Process a request.
func (ws *WebsocketService) Process(msg []byte, session *server.WSSession) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in WebsocketService.Process(): %v", r)
			ws.logger.InfoMsg("Panic in WebsocketService.Process()", structure.ErrorKey, err)
			if !session.Closed() {
				ws.writeError(err.Error(), "", rpc.INTERNAL_ERROR, session)
			}
		}
	}()
	// Create new request object and unmarshal.
	req := &rpc.RPCRequest{}
	errU := json.Unmarshal(msg, req)

	// Error when unmarshaling.
	if errU != nil {
		ws.writeError("Failed to parse request: "+errU.Error()+" . Raw: "+string(msg),
			"", rpc.PARSE_ERROR, session)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		ws.writeError("Wrong protocol version: "+req.JSONRPC, req.Id,
			rpc.INVALID_REQUEST, session)
		return
	}

	mName := req.Method

	if handler, ok := ws.defaultHandlers[mName]; ok {
		resp, errCode, err := handler(req, session)
		if err != nil {
			ws.writeError(err.Error(), req.Id, errCode, session)
		} else {
			ws.writeResponse(req.Id, resp, session)
		}
	} else {
		ws.writeError("Method not found: "+mName, req.Id,
			rpc.METHOD_NOT_FOUND, session)
	}
}

// Convenience method for writing error responses.
func (ws *WebsocketService) writeError(msg, id string, code int,
	session *server.WSSession) {
	response := rpc.NewRPCErrorResponse(id, code, msg)
	bts, err := ws.codec.EncodeBytes(response)
	// If there's an error here all bets are off.
	if err != nil {
		panic("Failed to marshal standard error response." + err.Error())
	}
	session.Write(bts)
}

// Convenience method for writing responses.
func (ws *WebsocketService) writeResponse(id string, result interface{},
	session *server.WSSession) error {
	response := rpc.NewRPCResponse(id, result)
	bts, err := ws.codec.EncodeBytes(response)
	if err != nil {
		ws.writeError("Internal error: "+err.Error(), id, rpc.INTERNAL_ERROR, session)
		return err
	}
	return session.Write(bts)
}

// *************************************** Events ************************************

func (ws *WebsocketService) EventSubscribe(request *rpc.RPCRequest,
	requester interface{}) (interface{}, int, error) {
	session, ok := requester.(*server.WSSession)
	if !ok {
		return 0, rpc.INTERNAL_ERROR,
			fmt.Errorf("Passing wrong object to websocket events")
	}
	param := &EventIdParam{}
	err := ws.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	eventId := param.EventId
	subId, err := event.GenerateSubscriptionID()
	if err != nil {
		return nil, rpc.INTERNAL_ERROR, err
	}

	err = ws.service.Subscribe(context.Background(), subId, eventId, func(resultEvent *rpc.ResultEvent) (stop bool) {
		ws.writeResponse(subId, resultEvent, session)
		return
	})
	if err != nil {
		return nil, rpc.INTERNAL_ERROR, err
	}
	return &EventSub{SubId: subId}, 0, nil
}

func (ws *WebsocketService) EventUnsubscribe(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &SubIdParam{}
	err := ws.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}

	err = ws.service.Unsubscribe(context.Background(), param.SubId)
	if err != nil {
		return nil, rpc.INTERNAL_ERROR, err
	}
	return &EventUnsub{Result: true}, 0, nil
}

func (ws *WebsocketService) EventPoll(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return nil, rpc.INTERNAL_ERROR, fmt.Errorf("Cannot poll with websockets")
}
