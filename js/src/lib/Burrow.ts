import { Events } from './events';
import { Pipe } from './pipe';
import { ContractManager } from './contracts/manager';
import { Namereg } from './namereg';
import { TransactClient } from '../../proto/rpctransact_grpc_pb';
import { QueryClient } from '../../proto/rpcquery_grpc_pb';
import { ExecutionEventsClient } from '../../proto/rpcevents_grpc_pb';
import * as grpc from 'grpc';

export type Error = grpc.ServiceError;

export class Burrow {
  URL: string;
  account: string;

  events: Events;
  pipe: Pipe;
  contracts: ContractManager;
  namereg: Namereg;

  ec: ExecutionEventsClient;
  tc: TransactClient;
  qc: QueryClient;

  constructor(URL: string, account: string) {
    this.URL = URL;
    this.account = account;

    this.ec = new ExecutionEventsClient(URL, grpc.credentials.createInsecure());
    this.tc = new TransactClient(URL, grpc.credentials.createInsecure());
    this.qc = new QueryClient(URL, grpc.credentials.createInsecure());
  
    // This is the execution events streaming service running on top of the raw streaming function.
    this.events = new Events(this.ec);
  
    // Contracts stuff running on top of grpc
    this.pipe = new Pipe(this.tc, this.events);
    this.contracts = new ContractManager(this);

    this.namereg = new Namereg(merge(this.tc, this.qc), this.account);
  }
}

function merge(...args: any[]): any {
  const next = {};
  for (const obj of args) {
      for (const key in obj) {
        next[key] = obj[key];
      }
  }
  return next;
};
