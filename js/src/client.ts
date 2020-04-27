import * as coder from 'ethereumjs-abi';
import { Readable } from 'stream';
import { TxExecution } from '../proto/exec_pb';
import { CallTx, TxInput } from '../proto/payload_pb'
import { Burrow } from './burrow'
import { callTx } from "./contracts/call";
import { Contract } from "./contracts/contract";
import * as convert from './utils/convert';

export type Interceptor = (result: TxExecution) => Promise<TxExecution>;

export class Client extends Burrow {
  interceptor: Interceptor;

  constructor(url: string, account: string) {
    super(url, account);

    this.interceptor = async (data) => data;
  }

  deploy(msg: CallTx, callback: (err: Error, addr: Uint8Array) => void) {
    this.pipe.transact(msg, (err, exec) => {
      if (err) callback(err, null);
      else if (exec.hasException()) callback(new Error(exec.getException().getException()), null);
      else callback(null, exec.getReceipt().getContractaddress_asU8());
    })
  }

  call(msg: CallTx, callback: (err: Error, exec: Uint8Array) => void) {
    this.pipe.transact(msg, (err, exec) => {
      if (err) callback(err, null);
      else if (exec.hasException()) callback(new Error(exec.getException().getException()), null);
      else this.interceptor(exec).then(exec => callback(null, exec.getResult().getReturn_asU8()));
    })
  }

  callSim(msg: CallTx, callback: (err: Error, exec: Uint8Array) => void) {
    this.pipe.call(msg, (err, exec) => {
      if (err) callback(err, null);
      else if (exec.hasException()) callback(new Error(exec.getException().getException()), null);
      else this.interceptor(exec).then(exec => callback(null, exec.getResult().getReturn_asU8()));
    })
  }

  listen(signature: string, address: string, callback: (err: Error, event: any) => void): Readable {
    return this.events.subContractEvents(address, signature, callback)
  }

  payload(data: string, address?: string, contract?: Contract): CallTx {
    return callTx(data, this.account, address, contract)
  }

  encode(name: string, inputs: string[], ...args: any[]): string {
    args = convert.burrowToAbi(inputs, args);
    return name + convert.bytesTB(coder.rawEncode(inputs, args));
  }

  decode(data: Uint8Array, outputs: string[]): any {
    return convert.abiToBurrow(outputs, coder.rawDecode(outputs, Buffer.from(data)));
  }
}
