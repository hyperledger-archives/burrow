import { TxExecution } from '../proto/exec_pb';
import { Entry } from '../proto/names_pb';
import { NameTx, TxInput } from '../proto/payload_pb';
import { IQueryClient } from '../proto/rpcquery_grpc_pb';
import { GetNameParam } from '../proto/rpcquery_pb';
import { ITransactClient } from '../proto/rpctransact_grpc_pb';

export class Namereg {
  constructor(private transact: ITransactClient, private query: IQueryClient, private account: string) {}

  set(name: string, data: string, lease = 50000, fee = 5000): Promise<TxExecution> {
    const input = new TxInput();
    input.setAddress(Buffer.from(this.account, 'hex'));
    input.setAmount(lease);

    const payload = new NameTx();
    payload.setInput(input);
    payload.setName(name);
    payload.setData(data);
    payload.setFee(fee);

    return new Promise((resolve, reject) => {
      this.transact.nameTxSync(payload, (err, txe) => {
        if (err) {
          reject(err);
        }
        resolve(txe);
      });
    });
  }

  get(name: string): Promise<Entry> {
    const payload = new GetNameParam();
    payload.setName(name);
    return new Promise((resolve, reject) => {
      this.query.getName(payload, (err, entry) => {
        if (err) {
          reject(err);
        }
        resolve(entry);
      });
    });
  }
}
