import { EventFragment, FormatTypes, Interface, LogDescription } from '@ethersproject/abi';
import { Keccak } from 'sha3';
import { Client } from '../client';
import { postDecodeResult, prefixedHexString } from '../convert';
import { Bounds, EventStream, isEndOfStream } from '../events';

export type EventCallback = (err?: Error, result?: LogDescription) => void;

export function listen(
  iface: Interface,
  frag: EventFragment,
  address: string,
  burrow: Client,
  callback: EventCallback,
  start: Bounds = 'latest',
  end: Bounds = 'stream',
): EventStream {
  const signature = sha3(frag.format(FormatTypes.sighash));

  return burrow.listen(
    [signature],
    address,
    (err, event) => {
      if (err) {
        return isEndOfStream(err) ? undefined : callback(err);
      }
      const log = event?.log;
      if (!log) {
        return callback(new Error(`no LogEvent or Error provided`));
      }
      try {
        const result = iface.parseLog({
          topics: log.topics.map((topic) => prefixedHexString(topic)),
          data: prefixedHexString(log.data),
        });
        return callback(undefined, {
          ...result,
          args: postDecodeResult(result.args, frag.inputs),
        });
      } catch (err) {
        return callback(err);
      }
    },
    start,
    end,
  );
}

function sha3(str: string) {
  return new Keccak(256).update(str).digest('hex').toUpperCase();
}
