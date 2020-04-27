import * as coder from 'ethereumjs-abi';
import * as convert from '../utils/convert';
import sha3 from '../utils/sha3';
import { Burrow, Error } from '../burrow';
import { LogEvent } from '../../proto/exec_pb'
import { Event, EventInput } from 'solc';
import { extractDisplayName, extractTypeName, transformToFullName } from "./abi";

const types = (abi: Array<EventInput>, indexed: boolean) =>
  abi.filter(i => i.indexed === indexed).map(i => i.type);

type EventResult = {
  event?: string;
  address?: Uint8Array;
  args?: Record<string, any>;
}

const decode = function (abi: Event, data: LogEvent): EventResult {
  if (!data) return

  const argTopics = abi.anonymous ? data.getTopicsList_asU8() : data.getTopicsList_asU8().slice(1)
  const indexedParamsABI = types(abi.inputs, true)
  const nonIndexedParamsABI = types(abi.inputs, false)
  const indexedData = Buffer.concat(argTopics)
  const indexedParams = convert.abiToBurrow(indexedParamsABI, coder.rawDecode(indexedParamsABI, indexedData))
  const nonIndexedParams = convert.abiToBurrow(nonIndexedParamsABI, coder.rawDecode(nonIndexedParamsABI, Buffer.from(data.getData_asU8())))

  return {
    event: transformToFullName(abi),
    address: data.getAddress_asU8(),
    args: abi.inputs.reduce(function (acc, current) {
      acc[current.name] = current.indexed ? indexedParams.shift() : nonIndexedParams.shift()
      return acc
    }, {}),
  };
}

export const SolidityEvent = function (abi: Event, burrow: Burrow) {
  const name = transformToFullName(abi);
  const displayName = extractDisplayName(name);
  const typeName = extractTypeName(name);
  const signature = sha3(name);

  const call = function (address: string, callback: (err: Error, decoded: EventResult) => void) {
    address = address || this.address;
    if (!callback) throw new Error('Can not subscribe to an event without a callback.');

    return burrow.pipe.eventSub(address, signature, (err, event) => {
      if (err) callback(err, null);
      let decoded: EventResult;
      try {
        decoded = decode(abi, event);
      } catch (err) {
        return callback(err, null);
      }
      callback(null, decoded);
    });
  }

  return {displayName, typeName, call};
}
