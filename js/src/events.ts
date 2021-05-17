import * as grpc from '@grpc/grpc-js';
import { LogEvent } from '../proto/exec_pb';
import { IExecutionEventsClient } from '../proto/rpcevents_grpc_pb';
import { BlockRange, BlocksRequest, Bound, EventsResponse } from '../proto/rpcevents_pb';
import BoundType = Bound.BoundType;

export type EventStream = grpc.ClientReadableStream<EventsResponse>;

// Surprisingly, this seems to be as good as it gets at the time of writing (https://github.com/Microsoft/TypeScript/pull/17819)
// that is, defining various types of union here does not help on the consumer side to infer exactly one of err or log
// will be defined
export type EventCallback<T = LogEvent> = (err?: Error | EndOfStream, event?: T) => void;

export function getBlockRange(start: Bounds = 'latest', end: Bounds = 'stream'): BlockRange {
  const range = new BlockRange();
  range.setStart(boundsToBound(start));
  range.setEnd(boundsToBound(end));
  return range;
}

export const latestStreamingBlockRange = Object.freeze(getBlockRange('latest', 'stream'));

export const EndOfStream: unique symbol = Symbol('EndOfStream');
export type EndOfStream = typeof EndOfStream;

export class Events {
  constructor(private burrow: IExecutionEventsClient) {}

  stream(range: BlockRange, query: string, callback: EventCallback): EventStream {
    const arg = new BlocksRequest();
    arg.setBlockrange(range);
    arg.setQuery(query);

    const stream = this.burrow.events(arg);
    stream.on('data', (data: EventsResponse) => {
      data.getEventsList().map((event) => {
        const log = event.getLog();
        if (!log) {
          return callback(new Error(`received non-log event: ${log}`));
        }
        return callback(undefined, log);
      });
    });
    stream.on('end', () => callback(EndOfStream));
    stream.on('error', (err: grpc.ServiceError) => err.code === grpc.status.CANCELLED || callback(err));
    return stream;
  }

  listen(range: BlockRange, address: string, signature: string, callback: EventCallback): EventStream {
    const filter = "EventType = 'LogEvent' AND Address = '" + address.toUpperCase() + "'";
    // +
    // " AND Log0 = '" +
    // signature.toUpperCase() +
    // "'";
    return this.stream(range, filter, callback);
  }
}

export type Bounds = number | 'first' | 'latest' | 'stream';

export type FiniteBounds = number | 'first' | 'latest';

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

export function readEvents<T>(
  listen: (callback: EventCallback<T>, range?: BlockRange) => EventStream,
  start: FiniteBounds = 'first',
  end: FiniteBounds = 'latest',
  limit?: number,
): Promise<T[]> {
  return new Promise<T[]>((resolve, reject) => {
    const events: T[] = [];
    const stream = listen((err, event) => {
      if (err) {
        if (err === EndOfStream) {
          return resolve(events);
        }
        return reject(err);
      }
      if (!event) {
        return reject(new Error(`received empty event`));
      }
      events.push(event);
      if (limit && events.length === limit) {
        stream.cancel();
        return resolve(events);
      }
    }, getBlockRange(start, end));
  });
}
