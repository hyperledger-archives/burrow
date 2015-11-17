package erisdb

import (
	"encoding/json"
	"fmt"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	rpc "github.com/eris-ltd/eris-db/rpc"
	"github.com/eris-ltd/eris-db/server"

	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

// Used for ErisDb. Implements WebSocketService.
type ErisDbWsService struct {
	codec           rpc.Codec
	pipe            ep.Pipe
	defaultHandlers map[string]RequestHandlerFunc
}

// Create a new websocket service.
func NewErisDbWsService(codec rpc.Codec, pipe ep.Pipe) server.WebSocketService {
	tmwss := &ErisDbWsService{codec: codec, pipe: pipe}
	mtds := &ErisDbMethods{codec, pipe}

	dhMap := mtds.getMethods()
	// Events
	dhMap[EVENT_SUBSCRIBE] = tmwss.EventSubscribe
	dhMap[EVENT_UNSUBSCRIBE] = tmwss.EventUnsubscribe
	tmwss.defaultHandlers = dhMap
	return tmwss
}

// Process a request.
func (this *ErisDbWsService) Process(msg []byte, session *server.WSSession) {
	log.Debug("REQUEST: %s\n", string(msg))
	// Create new request object and unmarshal.
	req := &rpc.RPCRequest{}
	errU := json.Unmarshal(msg, req)

	// Error when unmarshaling.
	if errU != nil {
		this.writeError("Failed to parse request: "+errU.Error()+" . Raw: "+string(msg), "", rpc.PARSE_ERROR, session)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		this.writeError("Wrong protocol version: "+req.JSONRPC, req.Id, rpc.INVALID_REQUEST, session)
		return
	}

	mName := req.Method

	if handler, ok := this.defaultHandlers[mName]; ok {
		resp, errCode, err := handler(req, session)
		if err != nil {
			this.writeError(err.Error(), req.Id, errCode, session)
		} else {
			this.writeResponse(req.Id, resp, session)
		}
	} else {
		this.writeError("Method not found: "+mName, req.Id, rpc.METHOD_NOT_FOUND, session)
	}
}

// Convenience method for writing error responses.
func (this *ErisDbWsService) writeError(msg, id string, code int, session *server.WSSession) {
	response := rpc.NewRPCErrorResponse(id, code, msg)
	bts, err := this.codec.EncodeBytes(response)
	// If there's an error here all bets are off.
	if err != nil {
		panic("Failed to marshal standard error response." + err.Error())
	}
	session.Write(bts)
}

// Convenience method for writing responses.
func (this *ErisDbWsService) writeResponse(id string, result interface{}, session *server.WSSession) error {
	response := rpc.NewRPCResponse(id, result)
	bts, err := this.codec.EncodeBytes(response)
	log.Debug("RESPONSE: %v\n", response)
	if err != nil {
		this.writeError("Internal error: "+err.Error(), id, rpc.INTERNAL_ERROR, session)
		return err
	}
	return session.Write(bts)
}

// *************************************** Events ************************************

func (this *ErisDbWsService) EventSubscribe(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	session, ok := requester.(*server.WSSession)
	if !ok {
		return 0, rpc.INTERNAL_ERROR, fmt.Errorf("Passing wrong object to websocket events")
	}
	param := &EventIdParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	eventId := param.EventId
	subId, errSID := generateSubId()
	if errSID != nil {
		return nil, rpc.INTERNAL_ERROR, errSID
	}
	callback := func(ret types.EventData) {
		writeErr := this.writeResponse(subId, ret, session)
		if writeErr != nil {
			this.pipe.Events().Unsubscribe(subId)
		}
	}
	_, errC := this.pipe.Events().Subscribe(subId, eventId, callback)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.EventSub{subId}, 0, nil
}

func (this *ErisDbWsService) EventUnsubscribe(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	param := &EventIdParam{}
	err := this.codec.DecodeBytes(param, request.Params)
	if err != nil {
		return nil, rpc.INVALID_PARAMS, err
	}
	eventId := param.EventId

	result, errC := this.pipe.Events().Unsubscribe(eventId)
	if errC != nil {
		return nil, rpc.INTERNAL_ERROR, errC
	}
	return &ep.EventUnsub{result}, 0, nil
}

func (this *ErisDbWsService) EventPoll(request *rpc.RPCRequest, requester interface{}) (interface{}, int, error) {
	return nil, rpc.INTERNAL_ERROR, fmt.Errorf("Cannot poll with websockets")
}
