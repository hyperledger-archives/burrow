import * as grpc from '@grpc/grpc-js';
import { Event as BurrowEvent } from '../proto/exec_pb';
import { IExecutionEventsClient } from '../proto/rpcevents_grpc_pb';
import { BlockRange, BlocksRequest, Bound, EventsResponse } from '../proto/rpcevents_pb';
import { Address } from './contracts/abi';
import { Provider } from './solts/interface.gd';
import BoundType = Bound.BoundType;

export type EventStream = grpc.ClientReadableStream<EventsResponse>;

export type Event = {
  header: Header;
  log: Log;
};

export type Header = {
  height: number;
  index: number;
  txHash: string;
  eventId: string;
};

export type Log = {
  data: Uint8Array;
  topics: Uint8Array[];
};

// We opt for a duck-type brand rather than a unique symbol so that dependencies using different versions of Burrow
const SignalCodes = {
  cancelStream: 'cancelStream',
  endOfStream: 'endOfStream',
} as const;

type SignalCodes = keyof typeof SignalCodes;

// can be used together when compatible (since Signal is exported by the Provider interface)
export interface Signal<T extends SignalCodes> {
  __isBurrowSignal__: '__isBurrowSignal__';
  signal: T;
}

export type CancelStreamSignal = Signal<'cancelStream'>;

const cancelStream: CancelStreamSignal = {
  __isBurrowSignal__: '__isBurrowSignal__',
  signal: 'cancelStream',
} as const;

export const Signal = Object.freeze({
  cancelStream,
} as const);

// Surprisingly, this seems to be as good as it gets at the time of writing (https://github.com/Microsoft/TypeScript/pull/17819)
// that is, defining various types of union here does not help on the consumer side to infer exactly one of err or log
// will be defined
export type EventCallback<T> = (err?: Error, event?: T) => CancelStreamSignal | void;

export type Bounds = number | 'first' | 'latest' | 'stream';

export type FiniteBounds = number | 'first' | 'latest';

type EventRegistry<T extends string> = Record<T, { signature: string }>;

// Note: typescript breaks instanceof for Error (https://github.com/microsoft/TypeScript/issues/13965)
const burrowSignalToken = '__isBurrowSignal__' as const;

class EndOfStreamError extends Error implements Signal<'endOfStream'> {
  __isBurrowSignal__ = burrowSignalToken;

  public readonly signal = 'endOfStream';

  constructor() {
    super('End of stream, no more data will be sent - use isEndOfStream(err) to check for this signal');
  }
}

const endOfStreamError = Object.freeze(new EndOfStreamError());

export function isBurrowSignal(value: unknown): value is Signal<SignalCodes> {
  const v = value as Signal<SignalCodes>;
  return v && v.__isBurrowSignal__ === burrowSignalToken && (v.signal === 'cancelStream' || v.signal === 'endOfStream');
}

export function isEndOfStream(value: unknown): value is Signal<'endOfStream'> {
  return isBurrowSignal(value) && value.signal === 'endOfStream';
}

export function isCancelStream(value: unknown): value is CancelStreamSignal {
  return isBurrowSignal(value) && value.signal === 'cancelStream';
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
      .map((event) => {
        try {
          return callback(undefined, burrowEventToInterfaceEvent(event));
        } catch (err) {
          stream.cancel();
          throw err;
        }
      })
      .find(isCancelStream);
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
  listener: (callback: EventCallback<T>, start?: Bounds, end?: Bounds) => unknown,
  start: FiniteBounds = 'first',
  end: FiniteBounds = 'latest',
  limit?: number,
): Promise<T[]> {
  return reduceEvents(
    listener,
    (events, event) => {
      if (limit && events.length === limit) {
        return Signal.cancelStream;
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
  listener: (callback: EventCallback<T>, start?: Bounds, end?: Bounds) => unknown,
  reducer: (event: T) => CancelStreamSignal | void,
  start: FiniteBounds = 'first',
  end: FiniteBounds = 'latest',
): Promise<void> {
  return reduceEvents(listener, (acc, event) => reducer(event), undefined as void, start, end);
}

export function reduceEvents<T, U>(
  listener: (callback: EventCallback<T>, start?: Bounds, end?: Bounds) => unknown,
  reducer: (accumulator: U, event: T) => U | CancelStreamSignal,
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
          return Signal.cancelStream;
        }
        const reduced = reducer(accumulator, event);
        if (isCancelStream(reduced)) {
          resolve(accumulator);
          return Signal.cancelStream;
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
) => (callback: EventCallback<{ name: T; payload: unknown; event: Event }>, start?: Bounds, end?: Bounds): unknown => {
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
      const log0 = event.log.topics[0];
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
      const payload = decode(client, event.log.data, event.log.topics)[name]();
      return callback(undefined, { name, payload, event });
    },
    start,
    end,
  );
};

export function burrowEventToInterfaceEvent(event: BurrowEvent): Event {
  const log = event.getLog();
  const header = event.getHeader();
  return {
    log: {
      data: log?.getData_asU8() ?? new Uint8Array(),
      topics: log?.getTopicsList_asU8() || [],
    },
    header: {
      height: header?.getHeight() ?? 0,
      index: header?.getIndex() ?? 0,
      eventId: header?.getEventid() ?? '',

      txHash: Buffer.from(header?.getTxhash_asU8() ?? []).toString('hex'),
    },
  };
}
