// GENERATED CODE -- DO NOT EDIT!

// package: rpcevents
// file: rpcevents.proto

import * as rpcevents_pb from "./rpcevents_pb";
import * as exec_pb from "./exec_pb";
import * as grpc from "grpc";

interface IExecutionEventsService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  stream: grpc.MethodDefinition<rpcevents_pb.BlocksRequest, exec_pb.StreamEvent>;
  tx: grpc.MethodDefinition<rpcevents_pb.TxRequest, exec_pb.TxExecution>;
  events: grpc.MethodDefinition<rpcevents_pb.BlocksRequest, rpcevents_pb.EventsResponse>;
}

export const ExecutionEventsService: IExecutionEventsService;

export class ExecutionEventsClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  stream(argument: rpcevents_pb.BlocksRequest, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<exec_pb.StreamEvent>;
  stream(argument: rpcevents_pb.BlocksRequest, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<exec_pb.StreamEvent>;
  tx(argument: rpcevents_pb.TxRequest, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  tx(argument: rpcevents_pb.TxRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  tx(argument: rpcevents_pb.TxRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<exec_pb.TxExecution>): grpc.ClientUnaryCall;
  events(argument: rpcevents_pb.BlocksRequest, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<rpcevents_pb.EventsResponse>;
  events(argument: rpcevents_pb.BlocksRequest, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<rpcevents_pb.EventsResponse>;
}
