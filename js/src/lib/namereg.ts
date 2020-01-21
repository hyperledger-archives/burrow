import { QueryClient } from '../../proto/rpcquery_grpc_pb';
import { GetNameParam } from '../../proto/rpcquery_pb';
import { Entry } from '../../proto/names_pb';
import { TransactClient } from '../../proto/rpctransact_grpc_pb';
import { TxInput, NameTx } from '../../proto/payload_pb';
import { TxExecution } from '../../proto/exec_pb';
import * as grpc from 'grpc';

export class Namereg {
  burrow: TransactClient & QueryClient;
  account: string;

  constructor(burrow: TransactClient & QueryClient, account: string) {
    this.burrow = burrow;
    this.account = account;
  }

  set(name: string, data: string, lease = 50000, fee = 5000, callback: grpc.requestCallback<TxExecution>) {
    const input = new TxInput();
    input.setAddress(Buffer.from(this.account, 'hex'));
    input.setAmount(lease);

    const payload = new NameTx();
    payload.setInput(input);
    payload.setName(name);
    payload.setData(data);
    payload.setFee(fee);

    return this.burrow.nameTxSync(payload, callback);
  }

  get(name: string, callback: grpc.requestCallback<Entry>) {
    const payload = new GetNameParam();
    payload.setName(name);
    return this.burrow.getName(payload, callback)
  }
}