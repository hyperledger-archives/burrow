import * as grpc from '@grpc/grpc-js';
import { LogEvent } from '../proto/exec_pb';
import { IExecutionEventsClient } from '../proto/rpcevents_grpc_pb';
import { BlockRange, BlocksRequest, Bound, EventsResponse } from '../proto/rpcevents_pb';
import BoundType = Bound.BoundType;

export type EventStream = grpc.ClientReadableStream<EventsResponse>;

export type EventCallback = (err?: Error, log?: LogEvent) => void;

export function getBlockRange(start: Bounds = 'latest', end: Bounds = 'stream'): BlockRange {
  const range = new BlockRange();
  range.setStart(boundsToBound(start));
  range.setEnd(boundsToBound(end));
  return range;
}

export const latestStreamingBlockRange = Object.freeze(getBlockRange('latest', 'stream'));

export class Events {
  constructor(private burrow: IExecutionEventsClient) {}

  stream(range: BlockRange, query: string, callback: EventCallback): EventStream {
    const arg = new BlocksRequest();
    arg.setBlockrange(range);
    arg.setQuery(query);

    const stream = this.burrow.events(arg);
    stream.on('data', (data: EventsResponse) => {
      data.getEventsList().map((event) => {
        return callback(undefined, event.getLog());
      });
    });
    stream.on('error', (err: grpc.ServiceError) => err.code === grpc.status.CANCELLED || callback(err));
    return stream;
  }

  listen(range: BlockRange, address: string, signature: string, callback: EventCallback): EventStream {
    const filter =
      "EventType = 'LogEvent' AND Address = '" +
      address.toUpperCase() +
      "'" +
      " AND Log0 = '" +
      signature.toUpperCase() +
      "'";
    return this.stream(range, filter, callback);
  }
}

export type Bounds = 'first' | 'latest' | 'stream' | number;

function boundsToBound(bounds: Bounds): Bound {
  const bound = new Bound();
  bound.setIndex(0);

  switch (bounds) {
    case 'first':
      bound.setType(BoundType.FIRST);
      break;
    case 'latest':
      bound.setType(BoundType.LATEST);
      break;
    case 'stream':
      bound.setType(BoundType.STREAM);
      break;
    default:
      bound.setType(BoundType.ABSOLUTE);
      bound.setIndex(bounds);
  }

  return bound;
}
