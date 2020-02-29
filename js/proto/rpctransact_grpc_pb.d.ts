// GENERATED CODE -- DO NOT EDIT!

// package: rpctransact
// file: rpctransact.proto

import * as rpctransact_pb from "./rpctransact_pb";
import * as exec_pb from "./exec_pb";
import * as payload_pb from "./payload_pb";
import * as txs_pb from "./txs_pb";
import * as grpc from "grpc";

interface ITransactService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  broadcastTxSync: grpc.MethodDefinition<rpctransact_pb.TxEnvelopeParam, exec_pb.TxExecution>;
  broadcastTxAsync: grpc.MethodDefinition<rpctransact_pb.TxEnvelopeParam, txs_pb.Receipt>;
  signTx: grpc.MethodDefinition<rpctransact_pb.TxEnvelopeParam, rpctransact_pb.TxEnvelope>;
  formulateTx: grpc.MethodDefinition<payload_pb.Any, rpctransact_pb.TxEnvelope>;
  callTxSync: grpc.MethodDefinition<payload_pb.CallTx, exec_pb.TxExecution>;
  callTxAsync: grpc.MethodDefinition<payload_pb.CallTx, txs_pb.Receipt>;
  callTxSim: grpc.MethodDefinition<payload_pb.CallTx, exec_pb.TxExecution>;
  callCodeSim: grpc.MethodDefinition<rpctransact_pb.CallCodeParam, exec_pb.TxExecution>;
  sendTxSync: grpc.MethodDefinition<payload_pb.SendTx, exec_pb.TxExecution>;
  sendTxAsync: grpc.MethodDefinition<payload_pb.SendTx, txs_pb.Receipt>;
  nameTxSync: grpc.MethodDefinition<payload_pb.NameTx, exec_pb.TxExecution>;
  nameTxAsync: grpc.MethodDefinition<payload_pb.NameTx, txs_pb.Receipt>;
}

export const TransactService: ITransactService;

export class TransactClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  broadcastTxSync(argument: rpctransact_pb.TxEnvelopeParam, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  broadcastTxSync(argument: rpctransact_pb.TxEnvelopeParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  broadcastTxSync(argument: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  broadcastTxAsync(argument: rpctransact_pb.TxEnvelopeParam, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  broadcastTxAsync(argument: rpctransact_pb.TxEnvelopeParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  broadcastTxAsync(argument: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  signTx(argument: rpctransact_pb.TxEnvelopeParam, callback: grpc.requestCallback<rpctransact_pb.TxEnvelope>): grpc.ClientUnaryCall;
  signTx(argument: rpctransact_pb.TxEnvelopeParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpctransact_pb.TxEnvelope>): grpc.ClientUnaryCall;
  signTx(argument: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpctransact_pb.TxEnvelope>): grpc.ClientUnaryCall;
  formulateTx(argument: payload_pb.Any, callback: grpc.requestCallback<rpctransact_pb.TxEnvelope>): grpc.ClientUnaryCall;
  formulateTx(argument: payload_pb.Any, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpctransact_pb.TxEnvelope>): grpc.ClientUnaryCall;
  formulateTx(argument: payload_pb.Any, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpctransact_pb.TxEnvelope>): grpc.ClientUnaryCall;
  callTxSync(argument: payload_pb.CallTx, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callTxSync(argument: payload_pb.CallTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callTxSync(argument: payload_pb.CallTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callTxAsync(argument: payload_pb.CallTx, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  callTxAsync(argument: payload_pb.CallTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  callTxAsync(argument: payload_pb.CallTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  callTxSim(argument: payload_pb.CallTx, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callTxSim(argument: payload_pb.CallTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callTxSim(argument: payload_pb.CallTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callCodeSim(argument: rpctransact_pb.CallCodeParam, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callCodeSim(argument: rpctransact_pb.CallCodeParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  callCodeSim(argument: rpctransact_pb.CallCodeParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  sendTxSync(argument: payload_pb.SendTx, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  sendTxSync(argument: payload_pb.SendTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  sendTxSync(argument: payload_pb.SendTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  sendTxAsync(argument: payload_pb.SendTx, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  sendTxAsync(argument: payload_pb.SendTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  sendTxAsync(argument: payload_pb.SendTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  nameTxSync(argument: payload_pb.NameTx, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  nameTxSync(argument: payload_pb.NameTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  nameTxSync(argument: payload_pb.NameTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  nameTxAsync(argument: payload_pb.NameTx, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  nameTxAsync(argument: payload_pb.NameTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
  nameTxAsync(argument: payload_pb.NameTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<txs_pb.Receipt>): grpc.ClientUnaryCall;
}
