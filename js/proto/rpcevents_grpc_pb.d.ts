// package: rpcevents
// file: rpcevents.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as rpcevents_pb from "./rpcevents_pb";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as exec_pb from "./exec_pb";

interface IExecutionEventsService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    stream: IExecutionEventsService_IStream;
    tx: IExecutionEventsService_ITx;
    events: IExecutionEventsService_IEvents;
}

interface IExecutionEventsService_IStream extends grpc.MethodDefinition<rpcevents_pb.BlocksRequest, exec_pb.StreamEvent> {
    path: string; // "/rpcevents.ExecutionEvents/Stream"
    requestStream: boolean; // false
    responseStream: boolean; // true
    requestSerialize: grpc.serialize<rpcevents_pb.BlocksRequest>;
    requestDeserialize: grpc.deserialize<rpcevents_pb.BlocksRequest>;
    responseSerialize: grpc.serialize<exec_pb.StreamEvent>;
    responseDeserialize: grpc.deserialize<exec_pb.StreamEvent>;
}
interface IExecutionEventsService_ITx extends grpc.MethodDefinition<rpcevents_pb.TxRequest, exec_pb.TxExecution> {
    path: string; // "/rpcevents.ExecutionEvents/Tx"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcevents_pb.TxRequest>;
    requestDeserialize: grpc.deserialize<rpcevents_pb.TxRequest>;
    responseSerialize: grpc.serialize<exec_pb.TxExecution>;
    responseDeserialize: grpc.deserialize<exec_pb.TxExecution>;
}
interface IExecutionEventsService_IEvents extends grpc.MethodDefinition<rpcevents_pb.BlocksRequest, rpcevents_pb.EventsResponse> {
    path: string; // "/rpcevents.ExecutionEvents/Events"
    requestStream: boolean; // false
    responseStream: boolean; // true
    requestSerialize: grpc.serialize<rpcevents_pb.BlocksRequest>;
    requestDeserialize: grpc.deserialize<rpcevents_pb.BlocksRequest>;
    responseSerialize: grpc.serialize<rpcevents_pb.EventsResponse>;
    responseDeserialize: grpc.deserialize<rpcevents_pb.EventsResponse>;
}

export const ExecutionEventsService: IExecutionEventsService;

export interface IExecutionEventsServer {
    stream: grpc.handleServerStreamingCall<rpcevents_pb.BlocksRequest, exec_pb.StreamEvent>;
    tx: grpc.handleUnaryCall<rpcevents_pb.TxRequest, exec_pb.TxExecution>;
    events: grpc.handleServerStreamingCall<rpcevents_pb.BlocksRequest, rpcevents_pb.EventsResponse>;
}

export interface IExecutionEventsClient {
    stream(request: rpcevents_pb.BlocksRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<exec_pb.StreamEvent>;
    stream(request: rpcevents_pb.BlocksRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<exec_pb.StreamEvent>;
    tx(request: rpcevents_pb.TxRequest, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    tx(request: rpcevents_pb.TxRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    tx(request: rpcevents_pb.TxRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    events(request: rpcevents_pb.BlocksRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcevents_pb.EventsResponse>;
    events(request: rpcevents_pb.BlocksRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcevents_pb.EventsResponse>;
}

export class ExecutionEventsClient extends grpc.Client implements IExecutionEventsClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public stream(request: rpcevents_pb.BlocksRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<exec_pb.StreamEvent>;
    public stream(request: rpcevents_pb.BlocksRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<exec_pb.StreamEvent>;
    public tx(request: rpcevents_pb.TxRequest, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public tx(request: rpcevents_pb.TxRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public tx(request: rpcevents_pb.TxRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: exec_pb.TxExecution) => void): grpc.ClientUnaryCall;
    public events(request: rpcevents_pb.BlocksRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcevents_pb.EventsResponse>;
    public events(request: rpcevents_pb.BlocksRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcevents_pb.EventsResponse>;
}
