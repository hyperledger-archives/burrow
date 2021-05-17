import { EventFragment, Fragment, FunctionFragment, Interface } from 'ethers/lib/utils';
import { postDecodeResult, preEncodeResult, prefixedHexString, toBuffer } from './convert';

export type ContractCodec = {
  encodeDeploy(...args: unknown[]): Buffer;
  encodeFunctionData(signature: string, ...args: unknown[]): Buffer;
  decodeFunctionResult(signature: string, data: Uint8Array | undefined): any;
  decodeEventLog(signature: string, data: Uint8Array | undefined, topics?: Uint8Array[]): any;
};

export function getContractCodec(iface: Interface): ContractCodec {
  return {
    encodeDeploy(...args: unknown[]): Buffer {
      const frag = iface.deploy;
      try {
        return toBuffer(iface.encodeDeploy(preEncodeResult(args, frag.inputs)));
      } catch (err) {
        throwErr(err, 'encode deploy', 'constructor', args, frag);
      }
    },

    encodeFunctionData(signature: string, ...args: unknown[]): Buffer {
      let frag: FunctionFragment | undefined;
      try {
        frag = iface.getFunction(formatSignature(signature));
        return toBuffer(iface.encodeFunctionData(frag, preEncodeResult(args, frag.inputs)));
      } catch (err) {
        throwErr(err, 'encode function data', signature, args, frag);
      }
    },

    decodeFunctionResult(signature: string, data: Uint8Array = new Uint8Array()): any {
      let frag: FunctionFragment | undefined;
      try {
        frag = iface.getFunction(formatSignature(signature));
        return postDecodeResult(iface.decodeFunctionResult(frag, data), frag.outputs);
      } catch (err) {
        throwErr(err, 'decode function result', signature, { data }, frag);
      }
    },

    decodeEventLog(signature: string, data: Uint8Array = new Uint8Array(), topics?: Uint8Array[]): any {
      let frag: EventFragment | undefined;
      try {
        frag = iface.getEvent(formatSignature(signature));
        return postDecodeResult(
          iface.decodeEventLog(
            frag,
            prefixedHexString(data),
            topics?.map((topic) => prefixedHexString(topic)),
          ),
          frag.inputs,
        );
      } catch (err) {
        throwErr(err, 'decode event log', signature, { data, topics }, frag);
      }
    },
  };
}

function formatSignature(signature: string): string {
  return prefixedHexString(signature);
}

function throwErr(
  err: unknown,
  action: string,
  signature: string,
  args: Record<string, unknown> | unknown[],
  frag?: Fragment,
): never {
  const name = frag ? frag.name : `member with signature '${signature}'`;
  const inputs = frag?.inputs ? ` (inputs: ${JSON.stringify(frag.inputs)})` : '';
  throw new Error(`ContractCodec could not ${action} for ${name} with args ${JSON.stringify(args)}${inputs}: ${err}`);
}
