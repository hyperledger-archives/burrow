import { Event, Function } from 'solc';
import { GetMetadataParam } from '../../proto/rpcquery_pb';
import { Burrow } from '../burrow';
import { CompiledCode } from "./compile";
import { Contract, Handlers } from './contract';

type FunctionOrEvent = Function | Event;

export class ContractManager {
  burrow: Burrow;

  constructor(burrow: Burrow) {
    this.burrow = burrow;
  }

  async deploy(abi: Array<FunctionOrEvent>, byteCode: string | CompiledCode,
               handlers?: Handlers, ...args: unknown[]): Promise<Contract> {
    const contract = new Contract(abi, byteCode, this.burrow, handlers)
    contract.address = await contract.create(...args);
    return contract;
  }

  /**
   * Looks up the ABI for a deployed contract from Burrow's contract metadata store.
   * Contract metadata is only stored when provided by the contract deployer so is not guaranteed to exist.
   *
   * @method address
   * @param {string} address
   * @throws an error if no metadata found and contract could not be instantiated
   * @returns {Contract} interface object
   */
  fromAddress(address: string, handlers?: Handlers): Promise<Contract> {
    const msg = new GetMetadataParam();
    msg.setAddress(Buffer.from(address, 'hex'));

    return new Promise((resolve, reject) =>
      this.burrow.qc.getMetadata(msg, (err, res) => {
        if (err) reject(err);
        const metadata = res.getMetadata();
        if (!metadata) {
          throw new Error(`could not find any metadata for account ${address}`)
        }
        const abi = JSON.parse(metadata).Abi;
        // TODO attach rather than new
        resolve(new Contract(abi, null, this.burrow, handlers));
      }))
  }
}
