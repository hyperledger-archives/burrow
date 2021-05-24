import * as grpc from '@grpc/grpc-js';
import { Event } from '../proto/exec_pb';
import { IExecutionEventsClient } from '../proto/rpcevents_grpc_pb';
import { BlockRange, BlocksRequest, Bound, EventsResponse } from '../proto/rpcevents_pb';
import { Address } from './contracts/abi';
import { Provider } from './solts/interface.gd';
import BoundType = Bound.BoundType;

export type EventStream = grpc.ClientReadableStream<EventsResponse>;

// Surprisingly, this seems to be as good as it gets at the time of writing (https://github.com/Microsoft/TypeScript/pull/17819)
// that is, defining various types of union here does not help on the consumer side to infer exactly one of err or log
// will be defined
export type EventCallback<T> = (err?: Error, event?: T) => Signal | void;

export type Bounds = number | 'first' | 'latest' | 'stream';

export type FiniteBounds = number | 'first' | 'latest';

type EventRegistry<T extends string> = Record<T, { signature: string }>;

// Emitted for consumers when stream ends
const endOfStream: unique symbol = Symbol('EndOfStream');
// Emitted by consumers to signal the stream should end
export const cancelStream: unique symbol = Symbol('CancelStream');

export type Signal = typeof cancelStream;

// Note: typescript breaks instanceof for Error (https://github.com/microsoft/TypeScript/issues/13965)
class EndOfStreamError extends Error {
  public readonly endOfStream = endOfStream;

  constructor() {
    super('End of stream, no more data will be sent - use isEndOfStream(err) to check for this signal');
  }
}

const endOfStreamError = Object.freeze(new EndOfStreamError());

export function isEndOfStream(err: Error): boolean {
  return (err as EndOfStreamError).endOfStream === endOfStream;
}

export function getBlockRange(start: Bounds = 'latest', end: Bounds = 'stream'): BlockRange {
  const range = new BlockRange();
  range.setStart(boundsToBound(start));
  range.setEnd(boundsToBound(end));
  return range;
}

export function stream(
  client: IExecutionEventsClient,
  range: BlockRange,
  query: string,
  callback: EventCallback<Event>,
): EventStream {
  const arg = new BlocksRequest();
  arg.setBlockrange(range);
  arg.setQuery(query);

  const stream = client.events(arg);
  stream.on('data', (data: EventsResponse) => {
    const cancel = data
      .getEventsList()
      .map((event) => callback(undefined, event))
      .find((s) => s === cancelStream);
    if (cancel) {
      stream.cancel();
    }
  });
  stream.on('end', () => callback(endOfStreamError));
  stream.on('error', (err: grpc.ServiceError) => err.code === grpc.status.CANCELLED || callback(err));
  return stream;
}

export type QueryOptions = {
  signatures: string[];
  address?: string;
};

export function queryFor({ signatures, address }: QueryOptions): string {
  return and(
    equals('EventType', 'LogEvent'),
    equals('Address', address),
    or(...signatures.map((s) => equals('Log0', s))),
  );
}

function and(...predicates: string[]): string {
  return predicates.filter((p) => p).join(' AND ');
}

function or(...predicates: string[]): string {
  const query = predicates.filter((p) => p).join(' OR ');
  if (!query) {
    return '';
  }
  return '(' + query + ')';
}

function equals(key: string, value?: string): string {
  if (!value) {
    return '';
  }
  return key + " = '" + value + "'";
}

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
  listener: (callback: EventCallback<T>, start?: Bounds, end?: Bounds) => EventStream,
  start: FiniteBounds = 'first',
  end: FiniteBounds = 'latest',
  limit?: number,
): Promise<T[]> {
  return reduceEvents(
    listener,
    (events, event) => {
      if (limit && events.length === limit) {
        return cancelStream;
      }
      events.push(event);
      return events;
    },
    [] as T[],
    start,
    end,
  );
}

export function iterateEvents<T>(
  listener: (callback: EventCallback<T>, start?: Bounds, end?: Bounds) => EventStream,
  reducer: (event: T) => Signal | void,
  start: FiniteBounds = 'first',
  end: FiniteBounds = 'latest',
): Promise<void> {
  return reduceEvents(listener, (acc, event) => reducer(event), undefined as void, start, end);
}

export function reduceEvents<T, U>(
  listener: (callback: EventCallback<T>, start?: Bounds, end?: Bounds) => EventStream,
  reducer: (accumulator: U, event: T) => U | Signal,
  initialValue: U,
  start: FiniteBounds = 'first',
  end: FiniteBounds = 'latest',
): Promise<U> {
  let accumulator = initialValue;
  return new Promise<U>((resolve, reject) =>
    listener(
      (err, event) => {
        if (err) {
          if (isEndOfStream(err)) {
            return resolve(accumulator);
          }
          return reject(err);
        }
        if (!event) {
          reject(new Error(`received empty event`));
          return cancelStream;
        }
        const reduced = reducer(accumulator, event);
        if (reduced === cancelStream) {
          resolve(accumulator);
          return cancelStream;
        }
        accumulator = reduced;
      },
      start,
      end,
    ),
  );
}

export const listenerFor = <T extends string>(
  client: Provider,
  address: Address,
  eventRegistry: EventRegistry<T>,
  decode: (client: Provider, data?: Uint8Array, topics?: Uint8Array[]) => Record<T, () => unknown>,
  eventNames: T[],
) => (
  callback: EventCallback<{ name: T; payload: unknown; event: Event }>,
  start?: Bounds,
  end?: Bounds,
): EventStream => {
  const signatureToName = eventNames.reduce((acc, n) => acc.set(eventRegistry[n].signature, n), new Map<string, T>());

  return client.listen(
    Array.from(signatureToName.keys()),
    address,
    (err, event) => {
      if (err) {
        return callback(err);
      }
      if (!event) {
        return callback(new Error(`Empty event received`));
      }
      const topics = event?.getLog()?.getTopicsList_asU8();
      const log0 = topics?.[0];
      if (!log0) {
        return callback(new Error(`Event has no Log0: ${event?.toString()}`));
      }
      const signature = Buffer.from(log0).toString('hex').toUpperCase();
      const name = signatureToName.get(signature);
      if (!name) {
        return callback(
          new Error(`Could not find event with signature ${signature} in registry: ${JSON.stringify(eventRegistry)}`),
        );
      }
      const data = event?.getLog()?.getData_asU8();
      const payload = decode(client, data, topics)[name]();
      return callback(undefined, { name, payload, event });
    },
    start,
    end,
  );
};
