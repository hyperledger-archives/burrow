import { Metadata } from "@grpc/grpc-js";
import { callErrorFromStatus } from "@grpc/grpc-js/build/src/call";
import { Keccak } from "sha3";
import * as utils from '../utils/utils';
import * as coder from 'ethereumjs-abi';
import * as convert from '../utils/convert';
import * as grpc from '@grpc/grpc-js';
import sha3 from '../utils/sha3';
import { TxInput, CallTx, ContractMeta } from '../../../proto/payload_pb';
import { TxExecution, Result } from '../../../proto/exec_pb';
import { Burrow, Error } from '../burrow';
import { Envelope } from '../../../proto/txs_pb';
import { Function, FunctionInput, FunctionOutput } from 'solc';
import { ABI, Contract } from "./contract";

type FunctionIO = FunctionInput & FunctionOutput;

export const DEFAULT_GAS = 1111111111;

type TransactionResult = {
  contractAddress: string
  height: number
  index: number
  hash: string
  type: number
  result: Result.AsObject
  tx: Envelope.AsObject
  caller: string | string[]
} & DecodedResult

type DecodedResult = {
  raw?: any[]
  values?: Record<string, any>
}

export type Handler = (result: TransactionResult) => any

function fnSignature(abi: Function) {
  const name = utils.transformToFullName(abi)
  return sha3(name).slice(0, 8)
}

const types = (args: Array<FunctionIO>) => args.map(arg => arg.type);

function txPayload(data: string, account: string, address: string, contract?: Contract): CallTx {
  const input = new TxInput();
  input.setAddress(Buffer.from(account, 'hex'));
  input.setAmount(0);

  const payload = new CallTx();
  payload.setInput(input);
  if (address) payload.setAddress(Buffer.from(address, 'hex'));
  payload.setGaslimit(DEFAULT_GAS);
  payload.setFee(0);
  const code = Buffer.from(data, 'hex');
  // if we are deploying and it looks like wasm, it must be wasm. Note that
  // evm opcode 0 is stop, so this would not make any sense.
  const wasm_magic = Buffer.from('\0asm');
  if (!address && !Buffer.compare(code.slice(0, 4), wasm_magic)) {
    payload.setWasm(code);
  } else {
    payload.setData(code);
  }
  // If address is null then we are creating a new contract, if we have the deployedBytecode then send it with the ABI
  if (!address && contract.code.deployedBytecode) {
    const meta = new ContractMeta()
    // TODO: document/formalise the expected structure of the contract metadata
    meta.setMeta(JSON.stringify({ Abi: contract.abi }))
    const codeHash = (new Keccak(256)).update(contract.code.deployedBytecode, "hex").digest()
    meta.setCodehash(codeHash)
    payload.setContractmetaList([meta])
  }

  return payload
}

const encodeF = function (abi: Function, args: Array<any>, bytecode: string): string {
  let abiInputs: string[];
  if (abi) {
    abiInputs = types(abi.inputs)
    args = convert.burrowToAbi(abiInputs, args) // If abi is passed, convert values accordingly
  }

  // If bytecode provided then this is a creation call, bytecode goes first
  if (bytecode) {
    let data = bytecode
    if (abi) data += convert.bytesTB(coder.rawEncode(abiInputs, args))
    return data
  } else {
    return fnSignature(abi) + convert.bytesTB(coder.rawEncode(abiInputs, args))
  }
}

const decodeF = function (abi: Function, output: Uint8Array): DecodedResult {
  if (!output) return

  let outputs = abi.outputs;
  let outputTypes = types(outputs);

  // Decode raw bytes to arguments
  let raw = convert.abiToBurrow(outputTypes, coder.rawDecode(outputTypes, Buffer.from(output)));
  let result: DecodedResult = { raw: raw.slice() }

  result.values = outputs.reduce(function (acc, current) {
    let value = raw.shift();
    if (current.name) {
      acc[current.name] = value;
    }
    return acc;
  }, {});

  return result;
}

export const SolidityFunction = function (abi: Function, burrow: Burrow) {
  let isConstructor = (abi == null || abi.type === 'constructor');
  let name: string;
  let displayName: string;
  let typeName: string;
  if (!isConstructor) {
    name = utils.transformToFullName(abi);
    displayName = utils.extractDisplayName(name);
    typeName = utils.extractTypeName(name);
  }

  // It might seem weird to include copies of the functions in here and above
  // My reason is the code above can be used "functionally" whereas this version
  // Uses implicit attributes of this object.
  // I want to keep them separate in the case that we want to move all the functional
  // components together and maybe even... write tests for them (gasp!)
  const encode = function () {
    // Call should always be bound to a Contract
    const contract: Contract = this;
    let args = Array.prototype.slice.call(arguments)
    return encodeF(abi, args, isConstructor ? contract.code.bytecode : null)
  }

  const decode = function (data) {
    // Call should always be bound to a Contract
    const contract: Contract = this;
    return decodeF(abi, data)
  }

  const call = async function (isSim: boolean, handler: Handler, address: string, ...args: any[]) {
    // Call should always be bound to a Contract
    const contract: Contract = this;

    handler = handler || function (result) {
      return result
    };
    address = address || contract.address;
    if (isConstructor) {
      address = null
    }


    let P = new Promise<TransactionResult>(function (resolve, reject) {
      if (address == null && !isConstructor) reject(new Error('Address not provided to call'))
      if (abi != null && abi.inputs.length !== args.length) reject(new Error('Insufficient arguments passed to function: ' + (isConstructor ? 'constructor' : name)))
      // Post process the return
      let post = function (error: Error, result: TxExecution) {
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
        let returnObj: TransactionResult = {
          contractAddress: Buffer.from(result.getReceipt().getContractaddress_asU8()).toString('hex').toUpperCase(),
          height: result.getHeader().getHeight(),
          index: result.getHeader().getIndex(),
          hash: Buffer.from(result.getHeader().getTxhash_asU8()).toString('hex').toUpperCase(),
          type: result.getHeader().getTxtype(),
          result: result.getResult().toObject(),
          tx: result.getEnvelope().toObject(),
          caller: convert.recApply(convert.bytesTB, result.getEnvelope().getSignatoriesList().map(sig => sig.getAddress_asU8())),
        }

        // Unpack return arguments
        if (!isConstructor) {
          try {
            let { raw, values } = decodeF(abi, result.getResult().getReturn_asU8());
            returnObj.raw = raw;
            returnObj.values = values;
          } catch (e) {
            return reject(e)
          }
        }

        return resolve(returnObj);
      }

      // Decide if to make a "call" or a "transaction"
      // For the moment we need to use the burrowtoweb3 function to prefix bytes with 0x
      // otherwise the coder will give an error with bignumber not a number
      // TODO investigate if other libs or an updated lib will fix this
      // let data = encodeF(abi, utils.burrowToWeb3(args), isCon ? self.code : null)
      let data = encodeF(abi, args, isConstructor ? contract.code.bytecode : null)
      let payload = txPayload(data, burrow.account, address, contract)

      if (isSim) {
        burrow.pipe.call(payload, post)
      } else {
        burrow.pipe.transact(payload, post)
      }
    })

    const result = await P;
    return handler(result);
  }

  return { displayName, typeName, call, encode, decode }
}
