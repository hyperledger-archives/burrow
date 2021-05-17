import { EventFragment, FormatTypes, Interface, LogDescription } from 'ethers/lib/utils';
import { Keccak } from 'sha3';
import { Client } from '../client';
import { postDecodeResult, prefixedHexString } from '../convert';
import { EndOfStream, EventStream, latestStreamingBlockRange } from '../events';

export type EventCallback = (err?: Error | EndOfStream, result?: LogDescription) => void;

export function listen(
  iface: Interface,
  frag: EventFragment,
  address: string,
  burrow: Client,
  callback: EventCallback,
  range = latestStreamingBlockRange,
): EventStream {
  const signature = sha3(frag.format(FormatTypes.sighash));

  return burrow.events.listen(range, address, signature, (err, log) => {
    if (err) {
      return callback(err);
    }
    if (!log) {
      return callback(new Error(`no LogEvent or Error provided`));
    }
    try {
      const result = iface.parseLog({
        topics: log.getTopicsList_asU8().map((topic) => prefixedHexString(topic)),
        data: prefixedHexString(log.getData_asU8()),
      });
      return callback(undefined, {
        ...result,
        args: postDecodeResult(result.args, frag.inputs),
      });
    } catch (err) {
      return callback(err);
    }
  });
}

function sha3(str: string) {
  return new Keccak(256).update(str).digest('hex').toUpperCase();
}
