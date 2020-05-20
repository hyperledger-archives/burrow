import {ServiceError} from "@grpc/grpc-js";
import { ITransactClient } from '../../proto/rpctransact_grpc_pb';
import { CallTx } from '../../proto/payload_pb';
import { TxExecution } from '../../proto/exec_pb';
import { Events } from './events';
import { LogEvent } from '../../proto/exec_pb';
import * as grpc from '@grpc/grpc-js';

export type TxCallback = grpc.requestCallback<TxExecution>;

export class Pipe {
  burrow: ITransactClient;
  events: Events;

  constructor(burrow: ITransactClient, events: Events) {
    this.burrow = burrow;
    this.events = events;
  }

  transact(payload: CallTx, callback: TxCallback) {
    return this.burrow.callTxSync(payload, callback)
  }

  call(payload: CallTx, callback: TxCallback) {
    this.burrow.callTxSim(payload, callback)
  }

  eventSub(accountAddress: string, signature: string, callback: (err: ServiceError, log: LogEvent) => void) {
    return this.events.subContractEvents(accountAddress, signature, callback)
  }
}

