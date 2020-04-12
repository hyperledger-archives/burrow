import { Contract, Handlers } from './contract';
import { Burrow } from '../burrow';
import { GetMetadataParam } from '../../../proto/rpcquery_pb';
import { Function, Event } from 'solc';

type FunctionOrEvent = Function | Event;

export class ContractManager {
  burrow: Burrow;
  
  constructor(burrow: Burrow) {
    this.burrow = burrow;
  }

  deploy(abi: Array<FunctionOrEvent>, byteCode: string, handlers?: Handlers, ...args: any[]): Promise<Contract> {
    return new Promise((resolve, reject) => {
      let contract = new Contract(abi, byteCode, null, this.burrow, handlers)
      contract._constructor.apply(contract, args).then((address: string) => { 
        contract.address = address;
        resolve(contract);
      });
    });
  }

  /**
   * Creates a contract object interface from an address without ABI.
   * The contract must be deployed using a recent burrow deploy which registers
   * metadata.
   *
   * @method address
   * @param {string} address - default contract address [can be null]
   * @returns {Contract} returns contract interface object
   */
  address(address: string, handlers?: Handlers): Promise<Contract> {
    const msg = new GetMetadataParam();
    msg.setAddress(Buffer.from(address, 'hex'));

    return new Promise((resolve, reject) =>
      this.burrow.qc.getMetadata(msg, (err, res) => {
        if (err) reject(err);
        const abi = JSON.parse(res.getMetadata()).Abi;
        resolve(new Contract(abi, null, address, this.burrow, handlers));
      }))
  }
}