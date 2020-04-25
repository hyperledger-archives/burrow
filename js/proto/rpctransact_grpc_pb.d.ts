// package: rpctransact
// file: rpctransact.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as rpctransact_pb from "./rpctransact_pb";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as exec_pb from "./exec_pb";
import * as payload_pb from "./payload_pb";
import * as txs_pb from "./txs_pb";

interface ITransactService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    broadcastTxSync: ITransactService_IBroadcastTxSync;
    broadcastTxAsync: ITransactService_IBroadcastTxAsync;
    signTx: ITransactService_ISignTx;
    formulateTx: ITransactService_IFormulateTx;
    callTxSync: ITransactService_ICallTxSync;
    callTxAsync: ITransactService_ICallTxAsync;
    callTxSim: ITransactService_ICallTxSim;
    callCodeSim: ITransactService_ICallCodeSim;
    sendTxSync: ITransactService_ISendTxSync;
    sendTxAsync: ITransactService_ISendTxAsync;
    nameTxSync: ITransactService_INameTxSync;
    nameTxAsync: ITransactService_INameTxAsync;
}

interface ITransactService_IBroadcastTxSync extends grpc.MethodDefinition<rpctransact_pb.TxEnvelopeParam, exec_pb.TxExecution> {
    path: string; // "/rpctransact.Transact/BroadcastTxSync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpctransact_pb.TxEnvelopeParam>;
    requestDeserialize: grpc.deserialize<rpctransact_pb.TxEnvelopeParam>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface ITransactService_IBroadcastTxAsync extends grpc.MethodDefinition<rpctransact_pb.TxEnvelopeParam, txs_pb.Receipt> {
    path: string; // "/rpctransact.Transact/BroadcastTxAsync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpctransact_pb.TxEnvelopeParam>;
    requestDeserialize: grpc.deserialize<rpctransact_pb.TxEnvelopeParam>;
    responseSerialize: grpc.serialize<txs_pb.Receipt>;
    responseDeserialize: grpc.deserialize<txs_pb.Receipt>;
}
interface ITransactService_ISignTx extends grpc.MethodDefinition<rpctransact_pb.TxEnvelopeParam, rpctransact_pb.TxEnvelope> {
    path: string; // "/rpctransact.Transact/SignTx"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpctransact_pb.TxEnvelopeParam>;
    requestDeserialize: grpc.deserialize<rpctransact_pb.TxEnvelopeParam>;
    responseSerialize: grpc.serialize<rpctransact_pb.TxEnvelope>;
    responseDeserialize: grpc.deserialize<rpctransact_pb.TxEnvelope>;
}
interface ITransactService_IFormulateTx extends grpc.MethodDefinition<payload_pb.Any, rpctransact_pb.TxEnvelope> {
    path: string; // "/rpctransact.Transact/FormulateTx"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.Any>;
    requestDeserialize: grpc.deserialize<payload_pb.Any>;
    responseSerialize: grpc.serialize<rpctransact_pb.TxEnvelope>;
    responseDeserialize: grpc.deserialize<rpctransact_pb.TxEnvelope>;
}
interface ITransactService_ICallTxSync extends grpc.MethodDefinition<payload_pb.CallTx, exec_pb.TxExecution> {
    path: string; // "/rpctransact.Transact/CallTxSync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.CallTx>;
    requestDeserialize: grpc.deserialize<payload_pb.CallTx>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface ITransactService_ICallTxAsync extends grpc.MethodDefinition<payload_pb.CallTx, txs_pb.Receipt> {
    path: string; // "/rpctransact.Transact/CallTxAsync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.CallTx>;
    requestDeserialize: grpc.deserialize<payload_pb.CallTx>;
    responseSerialize: grpc.serialize<txs_pb.Receipt>;
    responseDeserialize: grpc.deserialize<txs_pb.Receipt>;
}
interface ITransactService_ICallTxSim extends grpc.MethodDefinition<payload_pb.CallTx, exec_pb.TxExecution> {
    path: string; // "/rpctransact.Transact/CallTxSim"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.CallTx>;
    requestDeserialize: grpc.deserialize<payload_pb.CallTx>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface ITransactService_ICallCodeSim extends grpc.MethodDefinition<rpctransact_pb.CallCodeParam, exec_pb.TxExecution> {
    path: string; // "/rpctransact.Transact/CallCodeSim"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpctransact_pb.CallCodeParam>;
    requestDeserialize: grpc.deserialize<rpctransact_pb.CallCodeParam>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface ITransactService_ISendTxSync extends grpc.MethodDefinition<payload_pb.SendTx, exec_pb.TxExecution> {
    path: string; // "/rpctransact.Transact/SendTxSync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.SendTx>;
    requestDeserialize: grpc.deserialize<payload_pb.SendTx>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface ITransactService_ISendTxAsync extends grpc.MethodDefinition<payload_pb.SendTx, txs_pb.Receipt> {
    path: string; // "/rpctransact.Transact/SendTxAsync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.SendTx>;
    requestDeserialize: grpc.deserialize<payload_pb.SendTx>;
    responseSerialize: grpc.serialize<txs_pb.Receipt>;
    responseDeserialize: grpc.deserialize<txs_pb.Receipt>;
}
interface ITransactService_INameTxSync extends grpc.MethodDefinition<payload_pb.NameTx, exec_pb.TxExecution> {
    path: string; // "/rpctransact.Transact/NameTxSync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.NameTx>;
    requestDeserialize: grpc.deserialize<payload_pb.NameTx>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface ITransactService_INameTxAsync extends grpc.MethodDefinition<payload_pb.NameTx, txs_pb.Receipt> {
    path: string; // "/rpctransact.Transact/NameTxAsync"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<payload_pb.NameTx>;
    requestDeserialize: grpc.deserialize<payload_pb.NameTx>;
    responseSerialize: grpc.serialize<txs_pb.Receipt>;
    responseDeserialize: grpc.deserialize<txs_pb.Receipt>;
}

export const TransactService: ITransactService;

export interface ITransactServer {
    broadcastTxSync: grpc.handleUnaryCall<rpctransact_pb.TxEnvelopeParam, exec_pb.TxExecution>;
    broadcastTxAsync: grpc.handleUnaryCall<rpctransact_pb.TxEnvelopeParam, txs_pb.Receipt>;
    signTx: grpc.handleUnaryCall<rpctransact_pb.TxEnvelopeParam, rpctransact_pb.TxEnvelope>;
    formulateTx: grpc.handleUnaryCall<payload_pb.Any, rpctransact_pb.TxEnvelope>;
    callTxSync: grpc.handleUnaryCall<payload_pb.CallTx, exec_pb.TxExecution>;
    callTxAsync: grpc.handleUnaryCall<payload_pb.CallTx, txs_pb.Receipt>;
    callTxSim: grpc.handleUnaryCall<payload_pb.CallTx, exec_pb.TxExecution>;
    callCodeSim: grpc.handleUnaryCall<rpctransact_pb.CallCodeParam, exec_pb.TxExecution>;
    sendTxSync: grpc.handleUnaryCall<payload_pb.SendTx, exec_pb.TxExecution>;
    sendTxAsync: grpc.handleUnaryCall<payload_pb.SendTx, txs_pb.Receipt>;
    nameTxSync: grpc.handleUnaryCall<payload_pb.NameTx, exec_pb.TxExecution>;
    nameTxAsync: grpc.handleUnaryCall<payload_pb.NameTx, txs_pb.Receipt>;
}

export interface ITransactClient {
    broadcastTxSync(request: rpctransact_pb.TxEnvelopeParam, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    broadcastTxSync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    broadcastTxSync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    broadcastTxAsync(request: rpctransact_pb.TxEnvelopeParam, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    broadcastTxAsync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    broadcastTxAsync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    signTx(request: rpctransact_pb.TxEnvelopeParam, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    signTx(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    signTx(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    formulateTx(request: payload_pb.Any, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    formulateTx(request: payload_pb.Any, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    formulateTx(request: payload_pb.Any, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    callTxSync(request: payload_pb.CallTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callTxSync(request: payload_pb.CallTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callTxSync(request: payload_pb.CallTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callTxAsync(request: payload_pb.CallTx, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    callTxAsync(request: payload_pb.CallTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    callTxAsync(request: payload_pb.CallTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    callTxSim(request: payload_pb.CallTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callTxSim(request: payload_pb.CallTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callTxSim(request: payload_pb.CallTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callCodeSim(request: rpctransact_pb.CallCodeParam, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callCodeSim(request: rpctransact_pb.CallCodeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    callCodeSim(request: rpctransact_pb.CallCodeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    sendTxSync(request: payload_pb.SendTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    sendTxSync(request: payload_pb.SendTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    sendTxSync(request: payload_pb.SendTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    sendTxAsync(request: payload_pb.SendTx, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    sendTxAsync(request: payload_pb.SendTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    sendTxAsync(request: payload_pb.SendTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    nameTxSync(request: payload_pb.NameTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    nameTxSync(request: payload_pb.NameTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    nameTxSync(request: payload_pb.NameTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    nameTxAsync(request: payload_pb.NameTx, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    nameTxAsync(request: payload_pb.NameTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    nameTxAsync(request: payload_pb.NameTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
}

export class TransactClient extends grpc.Client implements ITransactClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public broadcastTxSync(request: rpctransact_pb.TxEnvelopeParam, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public broadcastTxSync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public broadcastTxSync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public broadcastTxAsync(request: rpctransact_pb.TxEnvelopeParam, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public broadcastTxAsync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public broadcastTxAsync(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public signTx(request: rpctransact_pb.TxEnvelopeParam, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    public signTx(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    public signTx(request: rpctransact_pb.TxEnvelopeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    public formulateTx(request: payload_pb.Any, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    public formulateTx(request: payload_pb.Any, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    public formulateTx(request: payload_pb.Any, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpctransact_pb.TxEnvelope) => void): grpc.ClientUnaryCall;
    public callTxSync(request: payload_pb.CallTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callTxSync(request: payload_pb.CallTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callTxSync(request: payload_pb.CallTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callTxAsync(request: payload_pb.CallTx, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public callTxAsync(request: payload_pb.CallTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public callTxAsync(request: payload_pb.CallTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public callTxSim(request: payload_pb.CallTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callTxSim(request: payload_pb.CallTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callTxSim(request: payload_pb.CallTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callCodeSim(request: rpctransact_pb.CallCodeParam, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callCodeSim(request: rpctransact_pb.CallCodeParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public callCodeSim(request: rpctransact_pb.CallCodeParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public sendTxSync(request: payload_pb.SendTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public sendTxSync(request: payload_pb.SendTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public sendTxSync(request: payload_pb.SendTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public sendTxAsync(request: payload_pb.SendTx, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public sendTxAsync(request: payload_pb.SendTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public sendTxAsync(request: payload_pb.SendTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public nameTxSync(request: payload_pb.NameTx, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public nameTxSync(request: payload_pb.NameTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public nameTxSync(request: payload_pb.NameTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public nameTxAsync(request: payload_pb.NameTx, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public nameTxAsync(request: payload_pb.NameTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
    public nameTxAsync(request: payload_pb.NameTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: txs_pb.Receipt) => void): grpc.ClientUnaryCall;
}
