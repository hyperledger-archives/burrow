package rpcevents

import (
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"golang.org/x/net/context"
)

type executionEventsServer struct {
}

func NewExecutionEventsServer() pbevents.ExecutionEventsServer {
	return &executionEventsServer{}
}

func (executionEventsServer) GetEvents(context.Context, *pbevents.GetEventsRequest) (*pbevents.GetEventsResponse, error) {
	panic("implement me")
}
