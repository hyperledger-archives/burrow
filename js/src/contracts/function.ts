import * as grpc from "@grpc/grpc-js";
import { Metadata } from "@grpc/grpc-js";
import { callErrorFromStatus } from "@grpc/grpc-js/build/src/call";
import * as coder from 'ethereumjs-abi';
import { Function } from 'solc';
import { Result } from '../../proto/exec_pb';
import { CallTx, ContractMeta } from "../../proto/payload_pb";
import { Envelope } from '../../proto/txs_pb';
import { TxCallback } from "../pipe";
import * as convert from "../utils/convert";
import sha3 from '../utils/sha3';
import { Address, transformToFullName } from "./abi";
import { callTx } from "./call";

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
} & DecodedResult

type DecodedResult = {
  raw?: unknown[];
  values?: Record<string, unknown>;
}

type Pipe = (payload: CallTx, callback: TxCallback) => void

// A handler may augment or alter the result
export type Handler = (result: TransactionResult) => any

function fnSignature(abi: Function): string {
  const name = transformToFullName(abi)
  return sha3(name).slice(0, 8)
}


export function encodeFunction(abi: Function, args: Array<any>): string {
  const abiInputs = abi.inputs.map(arg => arg.type);
  return fnSignature(abi) + convert.bytesTB(coder.rawEncode(abiInputs, convert.burrowToAbi(abiInputs, args)))
}

export function encodeConstructor(abi: Function, args: Array<any>, bytecode: string): string {
  const abiInputs = abi.inputs.map(arg => arg.type);
  return bytecode + convert.bytesTB(coder.rawEncode(abiInputs, args))
}

export function decode(abi: Function, output: Uint8Array): DecodedResult {
  if (!output) return

  const outputTypes = abi.outputs.map(arg => arg.type);

  // Decode raw bytes to arguments
  const raw = convert.abiToBurrow(outputTypes, coder.rawDecode(outputTypes, Buffer.from(output)));
  const result: DecodedResult = {raw: raw.slice()}

  result.values = abi.outputs.reduce(function (acc, current) {
    const value = raw.shift();
    if (current.name) {
      acc[current.name] = value;
    }
    return acc;
  }, {});

  return result;
}

function call(pipe: Pipe, payload: CallTx): Promise<TransactionResult> {
  return new Promise((resolve, reject) => {
    pipe(payload, (error, result) => {
      if (error) return reject(error)

      // Handle execution reversions
      if (result.hasException()) {
        // Decode error message if there is one otherwise default
        if (result.getResult().getReturn().length === 0) {
          return reject(callErrorFromStatus({
            code: grpc.status.ABORTED,
            metadata: new Metadata(),
            details: 'Execution Reverted',
          }))
        } else {
          // Strip first 4 bytes(function signature) the decode as a string
          return reject(callErrorFromStatus({
            code: grpc.status.ABORTED,
            metadata: new Metadata(),
            details: coder.rawDecode(['string'], Buffer.from(result.getResult().getReturn_asU8().slice(4)))[0],
          }))
        }
      }

      // Meta Data (address, caller, height, etc)
      return resolve({
        contractAddress: Buffer.from(result.getReceipt().getContractaddress_asU8()).toString('hex').toUpperCase(),
        height: result.getHeader().getHeight(),
        index: result.getHeader().getIndex(),
        hash: Buffer.from(result.getHeader().getTxhash_asU8()).toString('hex').toUpperCase(),
        type: result.getHeader().getTxtype(),
        result: result.getResult().toObject(),
        resultBytes: result.getResult().getReturn_asU8(),
        tx: result.getEnvelope().toObject(),
        caller: convert.recApply(convert.bytesTB, result.getEnvelope().getSignatoriesList().map(sig => sig.getAddress_asU8())),
      })

    })
  })
}

export async function callFunction<T>(input: Address, abi: Function, pipe: Pipe, handler: Handler<T>,
                                      callee: Address, args: unknown[]): Promise<T> {
  const result = await call(pipe, callTx(encodeFunction(abi, args), input, callee))
  // Unpack return arguments
  return handler({...result, ...decode(abi, result.resultBytes)});
}

export async function callConstructor<T>(input: Address, abi: Function, pipe: Pipe, handler: Handler<T>,
                                         contractMeta: ContractMeta[], bytecode: string, args: unknown[]): Promise<T> {
  return handler(await call(pipe, callTx(encodeConstructor(abi, args, bytecode), input, null, contractMeta)));
}
