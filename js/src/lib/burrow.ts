import * as grpc from "@grpc/grpc-js";
import {makeClientConstructor} from "@grpc/grpc-js";
import {IExecutionEventsClient} from "../../proto/rpcevents_grpc_pb";
import {IQueryClient} from '../../proto/rpcquery_grpc_pb';
import {ITransactClient} from '../../proto/rpctransact_grpc_pb';
import {ContractManager} from './contracts/manager';
import {Events} from './events';
import {Namereg} from './namereg';
import {Pipe} from './pipe';

export type Error = grpc.ServiceError;

export class Burrow {
  URL: string;
  account: string;

  events: Events;
  pipe: Pipe;
  contracts: ContractManager;
  namereg: Namereg;

  ec: IExecutionEventsClient;
  tc: ITransactClient;
  qc: IQueryClient;

  constructor(URL: string, account: string) {
    this.URL = URL;
    this.account = account;


    this.ec = newClient('rpcevents', 'ExecutionEvents', URL);
    this.tc = newClient('rpctransact', 'Transact', URL);
    this.qc = newClient('rpcquery', 'Query', URL);

    // This is the execution events streaming service running on top of the raw streaming function.
    this.events = new Events(this.ec);

    // Contracts stuff running on top of grpc
    this.pipe = new Pipe(this.tc, this.events);
    this.contracts = new ContractManager(this);

    this.namereg = new Namereg(this.tc, this.qc, this.account);
  }
}

function newClient<T>(pkg: string, service: string, URL: string): T {
  const imp = require(`../../proto/${pkg}_grpc_pb`);
  // As per https://github.com/agreatfool/grpc_tools_node_protoc_ts/blob/master/examples/src/grpcjs/client.ts
  const cons = makeClientConstructor(imp[`${pkg}.${service}`], service);
  // ugh
  return new cons(URL, grpc.credentials.createInsecure()) as unknown as T
}
