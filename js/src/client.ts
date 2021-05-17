import * as grpc from '@grpc/grpc-js';
import { Interface } from 'ethers/lib/utils';
import { TxExecution } from '../proto/exec_pb';
import { CallTx, ContractMeta } from '../proto/payload_pb';
import { ExecutionEventsClient, IExecutionEventsClient } from '../proto/rpcevents_grpc_pb';
import { BlockRange } from '../proto/rpcevents_pb';
import { IQueryClient, QueryClient } from '../proto/rpcquery_grpc_pb';
import { GetMetadataParam, StatusParam } from '../proto/rpcquery_pb';
import { ITransactClient, TransactClient } from '../proto/rpctransact_grpc_pb';
import { ResultStatus } from '../proto/rpc_pb';
import { ContractCodec, getContractCodec } from './codec';
import { Address } from './contracts/abi';
import { makeCallTx } from './contracts/call';
import { CallOptions, Contract, ContractInstance } from './contracts/contract';
import { toBuffer } from './convert';
import { getException } from './error';
import { EventCallback, Events, EventStream, latestStreamingBlockRange } from './events';
import { Namereg } from './namereg';
import { Provider } from './solts/provider.gd';

type TxCallback = (error: grpc.ServiceError | null, txe: TxExecution) => void;

export type Pipe = (payload: CallTx, callback: TxCallback) => void;

export type Interceptor = (result: TxExecution) => Promise<TxExecution>;

export class Client implements Provider {
  interceptor: Interceptor;
  readonly events: Events;
  readonly namereg: Namereg;

  readonly executionEvents: IExecutionEventsClient;
  readonly transact: ITransactClient;
  readonly query: IQueryClient;

  readonly callPipe: Pipe;
  readonly simPipe: Pipe;

  constructor(public readonly url: string, public readonly account: string) {
    const credentials = grpc.credentials.createInsecure();
    this.executionEvents = new ExecutionEventsClient(url, credentials);
    this.transact = new TransactClient(url, credentials);
    this.query = new QueryClient(url, credentials);
    // This is the execution events streaming service running on top of the raw streaming function.
    this.events = new Events(this.executionEvents);
    // Contracts stuff running on top of grpc
    this.namereg = new Namereg(this.transact, this.query, this.account);
    // NOTE: in general interceptor may be async
    this.interceptor = async (data) => data;

    this.callPipe = this.transact.callTxSync.bind(this.transact);
    this.simPipe = this.transact.callTxSim.bind(this.transact);
  }

  /**
   * Looks up the ABI for a deployed contract from Burrow's contract metadata store.
   * Contract metadata is only stored when provided by the contract deployer so is not guaranteed to exist.
   *
   * @method address
   * @param {string} address
   * @param handler
   * @throws an error if no metadata found and contract could not be instantiated
   * @returns {Contract} interface object
   */
  contractAt(address: string, handler?: CallOptions): Promise<ContractInstance> {
    const msg = new GetMetadataParam();
    msg.setAddress(Buffer.from(address, 'hex'));

    return new Promise((resolve, reject) =>
      this.query.getMetadata(msg, (err, res) => {
        if (err) {
          reject(err);
        }
        const metadata = res.getMetadata();
        if (!metadata) {
          throw new Error(`could not find any metadata for account ${address}`);
        }

        // TODO: parse with io-ts
        const abi = JSON.parse(metadata).Abi;

        const contract = new Contract(abi);
        resolve(contract.at(address, this, handler));
      }),
    );
  }

  callTxSync(callTx: CallTx): Promise<TxExecution> {
    return new Promise((resolve, reject) =>
      this.transact.callTxSync(callTx, (error, txe) => {
        if (error) {
          return reject(error);
        }
        const err = getException(txe);
        if (err) {
          return reject(err);
        }
        return resolve(this.interceptor(txe));
      }),
    );
  }

  callTxSim(callTx: CallTx): Promise<TxExecution> {
    return new Promise((resolve, reject) =>
      this.transact.callTxSim(callTx, (error, txe) => {
        if (error) {
          return reject(error);
        }
        const err = getException(txe);
        if (err) {
          return reject(err);
        }
        return resolve(txe);
      }),
    );
  }

  status(): Promise<ResultStatus> {
    return new Promise((resolve, reject) =>
      this.query.status(new StatusParam(), (err, resp) => (err ? reject(err) : resolve(resp))),
    );
  }

  async latestHeight(): Promise<number> {
    const status = await this.status();
    return status.getSyncinfo()?.getLatestblockheight() ?? 0;
  }

  // Methods below implement the generated codegen provider
  // TODO: should probably generate canonical version of Provider interface somewhere outside of files

  async deploy(msg: CallTx): Promise<Address> {
    const txe = await this.callTxSync(msg);
    const contractAddress = txe.getReceipt()?.getContractaddress_asU8();
    if (!contractAddress) {
      throw new Error(`deploy appears to have succeeded but contract address is missing from result: ${txe}`);
    }
    return Buffer.from(contractAddress).toString('hex').toUpperCase();
  }

  async call(msg: CallTx): Promise<Uint8Array | undefined> {
    const txe = await this.callTxSync(msg);
    return txe.getResult()?.getReturn_asU8();
  }

  async callSim(msg: CallTx): Promise<Uint8Array | undefined> {
    const txe = await this.callTxSim(msg);
    return txe.getResult()?.getReturn_asU8();
  }

  listen(
    signature: string,
    address: string,
    callback: EventCallback,
    range: BlockRange = latestStreamingBlockRange,
  ): EventStream {
    return this.events.listen(range, address, signature, callback);
  }

  payload(data: string | Uint8Array, address?: string, contractMeta: ContractMeta[] = []): CallTx {
    return makeCallTx(typeof data === 'string' ? toBuffer(data) : data, this.account, address, contractMeta);
  }

  contractCodec(contractABI: string): ContractCodec {
    const iface = new Interface(contractABI);
    return getContractCodec(iface);
  }
}
