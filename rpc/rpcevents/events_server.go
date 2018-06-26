package rpcevents

import (
	"github.com/hyperledger/burrow/execution/events/pbevents"
	"github.com/hyperledger/burrow/rpc"
	"golang.org/x/net/context"
)

type eventServer struct {
	subscriptions *rpc.Subscriptions
}

func NewEventsServer(subscriptions *rpc.Subscriptions) pbevents.EventsServer {
	return &eventServer{
		subscriptions: subscriptions,
	}
}

func (es *eventServer) EventPoll(ctx context.Context, param *pbevents.SubIdParam) (*pbevents.PollResponse, error) {
	msgs, err := es.subscriptions.Poll(param.GetSubId())
	if err != nil {
		return nil, err
	}
	resp := &pbevents.PollResponse{
		Events: make([]*pbevents.Event, 0, len(msgs)),
	}
	for _, msg := range msgs {
		if resultEvent, ok := msg.(*rpc.ResultEvent); ok {
			ev, err := resultEvent.GetEvent()
			if err != nil {
				return nil, err
			}
			resp.Events = append(resp.Events, ev)
		}
	}
	return resp, nil
}

func (es *eventServer) EventSubscribe(ctx context.Context, param *pbevents.EventIdParam) (*pbevents.SubIdParam, error) {
	subID, err := es.subscriptions.Add(param.GetEventId())
	if err != nil {
		return nil, err
	}
	return &pbevents.SubIdParam{
		SubId: subID,
	}, nil
}

func (es *eventServer) EventUnsubscribe(ctx context.Context, param *pbevents.SubIdParam) (*pbevents.EventUnSub, error) {
	err := es.subscriptions.Remove(param.GetSubId())
	if err != nil {
		return nil, err
	}
	return &pbevents.EventUnSub{
		Result: true,
	}, nil
}
