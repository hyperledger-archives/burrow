import * as grpc from '@grpc/grpc-js';
import { Metadata } from '@grpc/grpc-js';
import { callErrorFromStatus } from '@grpc/grpc-js/build/src/call';
import { AbiCoder, FunctionFragment, Interface, Result as DecodedResult } from 'ethers/lib/utils';
import { SolidityFunction } from 'solc';
import { Result } from '../../proto/exec_pb';
import { CallTx, ContractMeta, TxInput } from '../../proto/payload_pb';
import { Envelope } from '../../proto/txs_pb';
import { Pipe } from '../client';
import { postDecodeResult, preEncodeResult, toBuffer } from '../convert';
import { Address } from './abi';
import { CallOptions } from './contract';

export const DEFAULT_GAS = 1111111111;

export { Result as DecodeResult } from 'ethers/lib/utils';

const WasmMagic = Buffer.from('\0asm');

const coder = new AbiCoder();

export function makeCallTx(
  data: Uint8Array,
  inputAddress: Address,
  contractAddress?: Address,
  contractMeta: ContractMeta[] = [],
): CallTx {
  const input = new TxInput();
  input.setAddress(Buffer.from(inputAddress, 'hex'));
  input.setAmount(0);

  const payload = new CallTx();
  payload.setInput(input);
  if (contractAddress) {
    payload.setAddress(Buffer.from(contractAddress, 'hex'));
  }
  payload.setGaslimit(DEFAULT_GAS);
  payload.setFee(0);
  // if we are deploying and it looks like wasm, it must be wasm. Note that
  // evm opcode 0 is stop, so this would not make any sense.
  if (!contractAddress && !Buffer.compare(data.slice(0, 4), WasmMagic)) {
    payload.setWasm(data);
  } else {
    payload.setData(data);
  }
  payload.setContractmetaList(contractMeta);

  return payload;
}

export type TransactionResult = {
  contractAddress: string;
  height: number;
  index: number;
  hash: string;
  type: number;
  result: Result.AsObject;
  resultBytes: Uint8Array;
  tx: Envelope.AsObject;
  caller: string | string[];
};

// export function encodeFunction(abi: SolidityFunction, ...args: unknown[]): string {
//   const abiInputs = abi.inputs.map((arg) => arg.type);
//   return functionSignature(abi) + coder.encode(abiInputs, convert.burrowToAbi(abiInputs, args));
// }

export function encodeConstructor(abi: SolidityFunction | void, bytecode: string, ...args: unknown[]): string {
  const abiInputs = abi ? abi.inputs.map((arg) => arg.type) : [];
  return bytecode + coder.encode(abiInputs, args);
}

export function decode(abi: SolidityFunction, output: Uint8Array): DecodedResult {
  if (!abi.outputs) {
    return [];
  }
  return coder.decode(abi.outputs as any, Buffer.from(output));
}

export function call(pipe: Pipe, payload: CallTx): Promise<TransactionResult> {
  return new Promise((resolve, reject) => {
    pipe(payload, (err, txe) => {
      if (err) {
        return reject(err);
      }

      if (!txe) {
        return reject(new Error(`call received no result after passing tx ${JSON.stringify(payload.toObject())}`));
      }

      // Handle execution reversions
      const result = txe.getResult();
      const header = txe.getHeader();
      if (!result) {
        return reject(new Error(`tx ${header?.getTxhash_asU8().toString()} has no result`));
      }
      if (txe.hasException()) {
        // Decode error message if there is one otherwise default
        if (result.getReturn().length === 0) {
          return reject(
            callErrorFromStatus({
              code: grpc.status.ABORTED,
              metadata: new Metadata(),
              details: 'Execution Reverted',
            }),
          );
        } else {
          // Strip first 4 bytes(function signature) the decode as a string
          return reject(
            callErrorFromStatus({
              code: grpc.status.ABORTED,
              metadata: new Metadata(),
              details: coder.decode(['string'], Buffer.from(result.getReturn_asU8().slice(4)))[0],
            }),
          );
        }
      }

      // Meta Data (address, caller, height, etc)
      const envelope = txe.getEnvelope();
      const receipt = txe.getReceipt();
      if (!header || !envelope || !receipt) {
        return reject(new Error(``));
      }
      let caller: string | string[] = envelope.getSignatoriesList().map((sig) => sig.getAddress_asB64().toUpperCase());
      if (caller.length === 1) {
        caller = caller[0];
      }
      return resolve({
        contractAddress: Buffer.from(receipt.getContractaddress_asU8()).toString('hex').toUpperCase(),
        height: header.getHeight(),
        index: header.getIndex(),
        hash: Buffer.from(header.getTxhash_asU8()).toString('hex').toUpperCase(),
        type: header.getTxtype(),
        result: result.toObject(),
        resultBytes: result.getReturn_asU8(),
        tx: envelope.toObject(),
        caller: caller,
      });
    });
  });
}

export type CallResult = {
  transactionResult: TransactionResult;
  result: DecodedResult;
};

export async function callFunction(
  iface: Interface,
  frag: FunctionFragment,
  { handler, middleware }: CallOptions,
  input: Address,
  pipe: Pipe,
  callee: Address,
  args: unknown[],
): Promise<unknown> {
  const data = toBuffer(iface.encodeFunctionData(frag, preEncodeResult(args, frag.inputs)));
  const transactionResult = await call(pipe, middleware(makeCallTx(data, input, callee)));
  // Unpack return arguments
  const result = postDecodeResult(iface.decodeFunctionResult(frag, transactionResult.resultBytes), frag.outputs);
  return handler({ transactionResult, result });
}
