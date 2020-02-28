// GENERATED CODE -- DO NOT EDIT!

// package: rpcdump
// file: rpcdump.proto

import * as rpcdump_pb from "./rpcdump_pb";
import * as dump_pb from "./dump_pb";
import * as grpc from "grpc";

interface IDumpService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  getDump: grpc.MethodDefinition<rpcdump_pb.GetDumpParam, dump_pb.Dump>;
}

export const DumpService: IDumpService;

export class DumpClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  getDump(argument: rpcdump_pb.GetDumpParam, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<dump_pb.Dump>;
  getDump(argument: rpcdump_pb.GetDumpParam, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<dump_pb.Dump>;
}
