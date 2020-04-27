import { Keccak } from "sha3";
import { Event } from 'solc';
import { ContractMeta } from "../../proto/payload_pb";
import { Burrow } from '../burrow';
import { ABI, Address, isFunction } from "./abi";
import { CompiledCode } from "./compile";
import { SolidityEvent } from './event';
import { Handler } from './function';

export type Meta = { Abi: ABI }

function Meta(abi: ABI): Meta {
  return {Abi: abi}
}

export interface Handlers {
  call?: Handler;
  deploy?: Handler;
}

const defaultHandlers: Handlers = {
  call: function (result) {
    return result.raw
  },
  deploy: function (result) {
    return result.contractAddress
  }
}


// TODO: we could do magical things here `const abi: ABI = <a real ABI>` ... `typeof
export class Contract<T extends { [key: string]: Function }> {
  abi: ABI;
  code: CompiledCode;

  _constructor: any;

  constructor(abi: ABI, code: string | CompiledCode) {

    this.abi = abi;
    this.code = typeof (code) === 'string' ? {bytecode: code} : code;
    this.addFunctionsToContract();
    this.addEventsToContract();
  }

  at(address: Address, burrow: Burrow, handlers?: Handlers): Contract {
    handlers = Object.assign({}, defaultHandlers, handlers);
    const instance = this.members(address, burrow, handlers)
    Object.setPrototypeOf(instance, this)
    return instance
  }

  meta(): ContractMeta[] {
    // We can only calculate the codeHash if we have the deployedCode
    if (this.abi.length == 0 || !this.code.deployedBytecode) {
      return []
    }
    const meta = new ContractMeta()
    const hasher = new Keccak(256);
    const codeHash = hasher.update(this.code.deployedBytecode, "hex").digest()
    meta.setMeta(JSON.stringify(Meta(this.abi)))
    meta.setCodehash(codeHash)
    return [meta]
  }

  async create(...args: unknown[]): Promise<Contract> {
    this.address = await this._constructor(...args);
    return this;
  }

  static attach(abi: ABI, address: Address, burrow: Burrow, handlers?: Handlers): Promise<Contract> {
    const contract = new Contract(abi, null, burrow, handlers)
    return this;
  }

  private members(address: Address, burrow: Burrow, handlers: Handlers): { [p: string]: Function } {
    return {
      ...this.functions(address, burrow, handlers),
      ...this.events(address, burrow, handlers),
    }
  }

  private functions(address: Address, burrow: Burrow, handlers: Handlers): { [p: string]: Function } {

  }

  private events(address: Address, burrow: Burrow, handlers: Handlers): { [p: string]: Function } {

  }

  private addFunctionsToContract(): void {
    this.abi.filter(isFunction).forEach(abi => {
      const solidityFunction = new SolidityFunction(this, abi, this.burrow);
      const {displayName, typeName, call, encode, decode} = solidityFunction;

      if (abi.type === 'constructor') {
        this._constructor = call.bind(contract, false, contract.handlers.deploy, '');
      } else {
        // bind the function call to the contract, specify if call or transact is desired
        const execute = call.bind(contract, abi.constant, contract.handlers.call, null);
        execute.sim = call.bind(contract, true, contract.handlers.call, null);
        // These allow the interface to be used for a generic contract of this type
        execute.at = call.bind(contract, abi.constant, contract.handlers.call);
        execute.atSim = call.bind(contract, true, contract.handlers.call);

        execute.encode = encode.bind(contract);
        execute.decode = decode.bind(contract);

        // Attach to the contract object
        if (!contract[displayName]) {
          contract[displayName] = execute;
        }
        contract[displayName][typeName] = execute;
      }
    })

    // Not every abi has a constructor specification.
    // If it doesn't we force a _constructor with null abi
    if (!contract._constructor) {
      const {call} = SolidityFunction(null, contract.burrow);
      contract._constructor = call.bind(contract, false, contract.handlers.deploy, '');
    }
  }

  private addEventsToContract(): void {
    (contract.abi.filter(json => {
      return json.type === 'event'
    }) as Event[]).forEach(json => {
      const {displayName, typeName, call} = SolidityEvent(json, contract.burrow);

      const execute = call.bind(contract, null);
      execute.once = call.bind(contract, null);
      execute.at = call.bind(contract);
      if (!contract[displayName]) {
        contract[displayName] = execute;
      }
      contract[displayName][typeName] = call.bind(contract);
    })
  }
}


