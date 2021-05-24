import { EventFragment, Fragment, FunctionFragment, Interface, LogDescription } from 'ethers/lib/utils';
import { Keccak } from 'sha3';
import { CallTx, ContractMeta } from '../../proto/payload_pb';
import { Client } from '../client';
import { preEncodeResult, toBuffer } from '../convert';
import { EventStream } from '../events';
import { ABI, Address } from './abi';
import { call, callFunction, CallResult, DecodeResult, makeCallTx } from './call';
import { EventCallback, listen } from './event';

export const meta: unique symbol = Symbol('meta');

export type InstanceMetadata = {
  address: string;
  contract: Contract;
};

export type ContractInstance = {
  [key in string]?: ContractFunction | ContractEvent;
} & {
  // Using a unique symbol as a key here means we cannot clash with any dynamic member of the contract
  [meta]: InstanceMetadata;
};

export type CallOptions = typeof defaultCallOptions;

export const defaultCallOptions = {
  middleware: (callTx: CallTx) => callTx,
  handler: (result: CallResult): unknown => result.result,
} as const;

// Get metadata of a contract instance including its deployed address and the Contract it instantiates
export function getMetadata(instance: ContractInstance): InstanceMetadata {
  return instance[meta];
}

// Since so common
export function getAddress(instance: ContractInstance): Address {
  return instance[meta].address;
}

export type CompiledContract = {
  abi: ABI;
  // Required to deploy a contract
  bytecode?: string;
  // Required to submit an ABI when deploying a contract
  deployedBytecode?: string;
};

export class Contract<T extends ContractInstance | any = any> {
  private readonly iface: Interface;

  constructor(public readonly code: CompiledContract, private readonly childCode: CompiledContract[] = []) {
    this.iface = new Interface(this.code.abi);
  }

  at(address: Address, burrow: Client, options: CallOptions = defaultCallOptions): T {
    const instance: ContractInstance = {
      [meta]: {
        address,
        contract: this,
      },
    };

    for (const frag of Object.values(this.iface.functions)) {
      attachFunction(this.iface, frag, instance, contractFunction(this.iface, frag, burrow, options, address));
    }

    for (const frag of Object.values(this.iface.events)) {
      attachFunction(this.iface, frag, instance, contractEvent(this.iface, frag, burrow, address));
    }

    return instance as T;
  }

  meta(): ContractMeta[] {
    return [this.code, ...this.childCode].map(contractMeta).filter((m): m is ContractMeta => Boolean(m));
  }

  async deploy(burrow: Client, ...args: unknown[]): Promise<T> {
    return this.deployWith(burrow, defaultCallOptions, ...args);
  }

  async deployWith(burrow: Client, options?: Partial<CallOptions>, ...args: unknown[]): Promise<T> {
    const opts = { ...defaultCallOptions, ...options };
    if (!this.code.bytecode) {
      throw new Error(`cannot deploy contract without compiled bytecode`);
    }
    const { middleware } = opts;
    const data = Buffer.concat(
      [this.code.bytecode, this.iface.encodeDeploy(preEncodeResult(args, this.iface.deploy.inputs))].map(toBuffer),
    );
    const tx = middleware(makeCallTx(data, burrow.account, undefined, this.meta()));
    const { contractAddress } = await call(burrow.callPipe, tx);
    return this.at(contractAddress, burrow, opts);
  }
}

// TODO[Silas]: integrate burrow.js with static code/type generation
type GenericFunction<I extends unknown[] = any[], O = any> = (...args: I) => Promise<O>;

// Bare call will execute against contract
type ContractFunction = GenericFunction & {
  sim: GenericFunction;
  at: (address: string) => GenericFunction;
  atSim: (address: string) => GenericFunction;
  encode: (...args: unknown[]) => string;
  decode: (output: Uint8Array) => DecodeResult;
};

export type ContractEvent = ((cb: EventCallback) => EventStream) & {
  at: (address: string, cb: EventCallback) => EventStream;
  once: () => Promise<LogDescription>;
};

function contractFunction(
  iface: Interface,
  frag: FunctionFragment,
  burrow: Client,
  options: CallOptions,
  contractAddress: string,
): ContractFunction {
  const func = (...args: unknown[]) =>
    callFunction(iface, frag, options, burrow.account, burrow.callPipe, contractAddress, args);
  func.sim = (...args: unknown[]) =>
    callFunction(iface, frag, options, burrow.account, burrow.simPipe, contractAddress, args);

  func.at = (address: string) => (...args: unknown[]) =>
    callFunction(iface, frag, options, burrow.account, burrow.callPipe, address, args);
  func.atSim = (address: string) => (...args: unknown[]) =>
    callFunction(iface, frag, options, burrow.account, burrow.simPipe, address, args);

  func.encode = (...args: unknown[]) => iface.encodeFunctionData(frag, args);
  func.decode = (output: Uint8Array) => iface.decodeFunctionResult(frag, output);
  return func;
}

function contractEvent(iface: Interface, frag: EventFragment, burrow: Client, contractAddress: string): ContractEvent {
  const func = (cb: EventCallback) => listen(iface, frag, contractAddress, burrow, cb);
  func.at = (address: string, cb: EventCallback) => listen(iface, frag, address, burrow, cb);
  func.once = () =>
    new Promise<LogDescription>((resolve, reject) =>
      listen(iface, frag, contractAddress, burrow, (err, result) =>
        err ? reject(err) : result ? resolve(result) : reject(new Error(`no EventResult received from callback`)),
      ),
    );
  return func;
}

function contractMeta({ abi, deployedBytecode }: CompiledContract): ContractMeta | void {
  // We can only calculate the codeHash if we have the deployedCode
  if (abi.length == 0 || !deployedBytecode) {
    return undefined;
  }
  const meta = new ContractMeta();
  const hasher = new Keccak(256);
  const codeHash = hasher.update(deployedBytecode, 'hex').digest();
  meta.setMeta(JSON.stringify({ Abi: abi }));
  meta.setCodehash(codeHash);
  return meta;
}

function attachFunction(
  iface: Interface,
  frag: Fragment,
  instance: ContractInstance,
  func: ContractFunction | ContractEvent,
): void {
  if (!instance[frag.name]) {
    instance[frag.name] = func;
  }
  instance[frag.format()] = func;
}
