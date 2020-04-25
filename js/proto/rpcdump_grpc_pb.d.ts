// package: rpcdump
// file: rpcdump.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as rpcdump_pb from "./rpcdump_pb";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as dump_pb from "./dump_pb";

interface IDumpService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    getDump: IDumpService_IGetDump;
}

interface IDumpService_IGetDump extends grpc.MethodDefinition<rpcdump_pb.GetDumpParam, dump_pb.Dump> {
    path: string; // "/rpcdump.Dump/GetDump"
    requestStream: boolean; // false
    responseStream: boolean; // true
    requestSerialize: grpc.serialize<rpcdump_pb.GetDumpParam>;
    requestDeserialize: grpc.deserialize<rpcdump_pb.GetDumpParam>;
    responseSerialize: grpc.serialize<dump_pb.Dump>;
    responseDeserialize: grpc.deserialize<dump_pb.Dump>;
}

export const DumpService: IDumpService;

export interface IDumpServer {
    getDump: grpc.handleServerStreamingCall<rpcdump_pb.GetDumpParam, dump_pb.Dump>;
}

export interface IDumpClient {
    getDump(request: rpcdump_pb.GetDumpParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<dump_pb.Dump>;
    getDump(request: rpcdump_pb.GetDumpParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<dump_pb.Dump>;
}

export class DumpClient extends grpc.Client implements IDumpClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public getDump(request: rpcdump_pb.GetDumpParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<dump_pb.Dump>;
    public getDump(request: rpcdump_pb.GetDumpParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<dump_pb.Dump>;
}
