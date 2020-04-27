import {IExecutionEventsClient} from '../../proto/rpcevents_grpc_pb';
import {BlocksRequest, BlockRange, Bound, EventsResponse} from '../../proto/rpcevents_pb';
import {LogEvent} from '../../proto/exec_pb';
import {Error} from './burrow';
import * as grpc from '@grpc/grpc-js';
import { LogEvent } from '../proto/exec_pb';
import { IExecutionEventsClient } from '../proto/rpcevents_grpc_pb';
import { BlockRange, BlocksRequest, Bound, EventsResponse } from '../proto/rpcevents_pb';
import { Error } from './burrow';

export type EventStream = grpc.ClientReadableStream<EventsResponse>;

export class Events {

  constructor(private burrow: IExecutionEventsClient) {
  }

  listen(query: string, callback: (err: Error, log: LogEvent) => void): EventStream {
    const start = new Bound();
    start.setType(3);
    start.setIndex(0);
    const end = new Bound();
    end.setType(4);
    end.setIndex(0);

    const range = new BlockRange();
    range.setStart(start);
    range.setEnd(end);

    const arg = new BlocksRequest();
    arg.setBlockrange(range);
    arg.setQuery(query);

    const stream = this.burrow.events(arg);
    stream.on('data', (data: EventsResponse) => {
      data.getEventsList().map(event => {
        callback(null, event.getLog());
      });
    });
    stream.on('error', (err: Error) => err.code === grpc.status.CANCELLED || callback(err, null));
    return stream;
  }

  subContractEvents(address: string, signature: string, callback: (err: Error, log: LogEvent) => void): EventStream {
    const filter = "EventType = 'LogEvent' AND Address = '" + address.toUpperCase() + "'" + " AND Log0 = '" + signature.toUpperCase() + "'";
    return this.listen(filter, callback);
  }
}
