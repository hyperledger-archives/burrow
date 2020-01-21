// GENERATED CODE -- DO NOT EDIT!

// package: keys
// file: keys.proto

import * as keys_pb from "./keys_pb";
import * as grpc from "grpc";

interface IKeysService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  generateKey: grpc.MethodDefinition<keys_pb.GenRequest, keys_pb.GenResponse>;
  publicKey: grpc.MethodDefinition<keys_pb.PubRequest, keys_pb.PubResponse>;
  sign: grpc.MethodDefinition<keys_pb.SignRequest, keys_pb.SignResponse>;
  verify: grpc.MethodDefinition<keys_pb.VerifyRequest, keys_pb.VerifyResponse>;
  import: grpc.MethodDefinition<keys_pb.ImportRequest, keys_pb.ImportResponse>;
  importJSON: grpc.MethodDefinition<keys_pb.ImportJSONRequest, keys_pb.ImportResponse>;
  export: grpc.MethodDefinition<keys_pb.ExportRequest, keys_pb.ExportResponse>;
  hash: grpc.MethodDefinition<keys_pb.HashRequest, keys_pb.HashResponse>;
  removeName: grpc.MethodDefinition<keys_pb.RemoveNameRequest, keys_pb.RemoveNameResponse>;
  list: grpc.MethodDefinition<keys_pb.ListRequest, keys_pb.ListResponse>;
  addName: grpc.MethodDefinition<keys_pb.AddNameRequest, keys_pb.AddNameResponse>;
}

export const KeysService: IKeysService;

export class KeysClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  generateKey(argument: keys_pb.GenRequest, callback: grpc.requestCallback<keys_pb.GenResponse>): grpc.ClientUnaryCall;
  generateKey(argument: keys_pb.GenRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.GenResponse>): grpc.ClientUnaryCall;
  generateKey(argument: keys_pb.GenRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.GenResponse>): grpc.ClientUnaryCall;
  publicKey(argument: keys_pb.PubRequest, callback: grpc.requestCallback<keys_pb.PubResponse>): grpc.ClientUnaryCall;
  publicKey(argument: keys_pb.PubRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.PubResponse>): grpc.ClientUnaryCall;
  publicKey(argument: keys_pb.PubRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.PubResponse>): grpc.ClientUnaryCall;
  sign(argument: keys_pb.SignRequest, callback: grpc.requestCallback<keys_pb.SignResponse>): grpc.ClientUnaryCall;
  sign(argument: keys_pb.SignRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.SignResponse>): grpc.ClientUnaryCall;
  sign(argument: keys_pb.SignRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.SignResponse>): grpc.ClientUnaryCall;
  verify(argument: keys_pb.VerifyRequest, callback: grpc.requestCallback<keys_pb.VerifyResponse>): grpc.ClientUnaryCall;
  verify(argument: keys_pb.VerifyRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.VerifyResponse>): grpc.ClientUnaryCall;
  verify(argument: keys_pb.VerifyRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.VerifyResponse>): grpc.ClientUnaryCall;
  import(argument: keys_pb.ImportRequest, callback: grpc.requestCallback<keys_pb.ImportResponse>): grpc.ClientUnaryCall;
  import(argument: keys_pb.ImportRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ImportResponse>): grpc.ClientUnaryCall;
  import(argument: keys_pb.ImportRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ImportResponse>): grpc.ClientUnaryCall;
  importJSON(argument: keys_pb.ImportJSONRequest, callback: grpc.requestCallback<keys_pb.ImportResponse>): grpc.ClientUnaryCall;
  importJSON(argument: keys_pb.ImportJSONRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ImportResponse>): grpc.ClientUnaryCall;
  importJSON(argument: keys_pb.ImportJSONRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ImportResponse>): grpc.ClientUnaryCall;
  export(argument: keys_pb.ExportRequest, callback: grpc.requestCallback<keys_pb.ExportResponse>): grpc.ClientUnaryCall;
  export(argument: keys_pb.ExportRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ExportResponse>): grpc.ClientUnaryCall;
  export(argument: keys_pb.ExportRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ExportResponse>): grpc.ClientUnaryCall;
  hash(argument: keys_pb.HashRequest, callback: grpc.requestCallback<keys_pb.HashResponse>): grpc.ClientUnaryCall;
  hash(argument: keys_pb.HashRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.HashResponse>): grpc.ClientUnaryCall;
  hash(argument: keys_pb.HashRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.HashResponse>): grpc.ClientUnaryCall;
  removeName(argument: keys_pb.RemoveNameRequest, callback: grpc.requestCallback<keys_pb.RemoveNameResponse>): grpc.ClientUnaryCall;
  removeName(argument: keys_pb.RemoveNameRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.RemoveNameResponse>): grpc.ClientUnaryCall;
  removeName(argument: keys_pb.RemoveNameRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.RemoveNameResponse>): grpc.ClientUnaryCall;
  list(argument: keys_pb.ListRequest, callback: grpc.requestCallback<keys_pb.ListResponse>): grpc.ClientUnaryCall;
  list(argument: keys_pb.ListRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ListResponse>): grpc.ClientUnaryCall;
  list(argument: keys_pb.ListRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.ListResponse>): grpc.ClientUnaryCall;
  addName(argument: keys_pb.AddNameRequest, callback: grpc.requestCallback<keys_pb.AddNameResponse>): grpc.ClientUnaryCall;
  addName(argument: keys_pb.AddNameRequest, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.AddNameResponse>): grpc.ClientUnaryCall;
  addName(argument: keys_pb.AddNameRequest, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<keys_pb.AddNameResponse>): grpc.ClientUnaryCall;
}
