package rpcevents

import (
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/execution/events"
	"github.com/hyperledger/burrow/execution/events/pbevents"
)

type executionEventsServer struct {
	eventsProvider events.Provider
}

func NewExecutionEventsServer(eventsProvider events.Provider) pbevents.ExecutionEventsServer {
	return &executionEventsServer{eventsProvider: eventsProvider}
}

func (ees executionEventsServer) GetEvents(request *pbevents.GetEventsRequest,
	stream pbevents.ExecutionEvents_GetEventsServer) error {

	buf := new(pbevents.GetEventsResponse)

	blockRange := request.GetBlockRange()
	batchSize := request.GetBatchSize()
	start, end := blockRange.Bounds()
	ch, err := ees.eventsProvider.GetEvents(start, end, query.NewBuilder(request.Query))
	if err != nil {
		return err
	}
	for ev := range ch {
		buf.Events = append(buf.Events, pbevents.GetEvent(ev))
		if batchSize == 0 || uint32(len(buf.Events))%batchSize == 0 {
			err := stream.Send(buf)
			if err != nil {
				return err
			}
			buf = new(pbevents.GetEventsResponse)
		}
	}

	if len(buf.Events) > 0 {
		err := stream.Send(buf)
		if err != nil {
			return err
		}
	}
	return nil
}
