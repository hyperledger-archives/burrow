// package: keys
// file: keys.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as keys_pb from "./keys_pb";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as crypto_pb from "./crypto_pb";

interface IKeysService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    generateKey: IKeysService_IGenerateKey;
    publicKey: IKeysService_IPublicKey;
    sign: IKeysService_ISign;
    verify: IKeysService_IVerify;
    import: IKeysService_IImport;
    importJSON: IKeysService_IImportJSON;
    export: IKeysService_IExport;
    hash: IKeysService_IHash;
    removeName: IKeysService_IRemoveName;
    list: IKeysService_IList;
    addName: IKeysService_IAddName;
}

interface IKeysService_IGenerateKey extends grpc.MethodDefinition<keys_pb.GenRequest, keys_pb.GenResponse> {
    path: "/keys.Keys/GenerateKey";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.GenRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.GenRequest>;
    responseSerialize: grpc.serialize<keys_pb.GenResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.GenResponse>;
}
interface IKeysService_IPublicKey extends grpc.MethodDefinition<keys_pb.PubRequest, keys_pb.PubResponse> {
    path: "/keys.Keys/PublicKey";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.PubRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.PubRequest>;
    responseSerialize: grpc.serialize<keys_pb.PubResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.PubResponse>;
}
interface IKeysService_ISign extends grpc.MethodDefinition<keys_pb.SignRequest, keys_pb.SignResponse> {
    path: "/keys.Keys/Sign";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.SignRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.SignRequest>;
    responseSerialize: grpc.serialize<keys_pb.SignResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.SignResponse>;
}
interface IKeysService_IVerify extends grpc.MethodDefinition<keys_pb.VerifyRequest, keys_pb.VerifyResponse> {
    path: "/keys.Keys/Verify";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.VerifyRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.VerifyRequest>;
    responseSerialize: grpc.serialize<keys_pb.VerifyResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.VerifyResponse>;
}
interface IKeysService_IImport extends grpc.MethodDefinition<keys_pb.ImportRequest, keys_pb.ImportResponse> {
    path: "/keys.Keys/Import";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.ImportRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.ImportRequest>;
    responseSerialize: grpc.serialize<keys_pb.ImportResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.ImportResponse>;
}
interface IKeysService_IImportJSON extends grpc.MethodDefinition<keys_pb.ImportJSONRequest, keys_pb.ImportResponse> {
    path: "/keys.Keys/ImportJSON";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.ImportJSONRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.ImportJSONRequest>;
    responseSerialize: grpc.serialize<keys_pb.ImportResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.ImportResponse>;
}
interface IKeysService_IExport extends grpc.MethodDefinition<keys_pb.ExportRequest, keys_pb.ExportResponse> {
    path: "/keys.Keys/Export";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.ExportRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.ExportRequest>;
    responseSerialize: grpc.serialize<keys_pb.ExportResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.ExportResponse>;
}
interface IKeysService_IHash extends grpc.MethodDefinition<keys_pb.HashRequest, keys_pb.HashResponse> {
    path: "/keys.Keys/Hash";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.HashRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.HashRequest>;
    responseSerialize: grpc.serialize<keys_pb.HashResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.HashResponse>;
}
interface IKeysService_IRemoveName extends grpc.MethodDefinition<keys_pb.RemoveNameRequest, keys_pb.RemoveNameResponse> {
    path: "/keys.Keys/RemoveName";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.RemoveNameRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.RemoveNameRequest>;
    responseSerialize: grpc.serialize<keys_pb.RemoveNameResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.RemoveNameResponse>;
}
interface IKeysService_IList extends grpc.MethodDefinition<keys_pb.ListRequest, keys_pb.ListResponse> {
    path: "/keys.Keys/List";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.ListRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.ListRequest>;
    responseSerialize: grpc.serialize<keys_pb.ListResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.ListResponse>;
}
interface IKeysService_IAddName extends grpc.MethodDefinition<keys_pb.AddNameRequest, keys_pb.AddNameResponse> {
    path: "/keys.Keys/AddName";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<keys_pb.AddNameRequest>;
    requestDeserialize: grpc.deserialize<keys_pb.AddNameRequest>;
    responseSerialize: grpc.serialize<keys_pb.AddNameResponse>;
    responseDeserialize: grpc.deserialize<keys_pb.AddNameResponse>;
}

export const KeysService: IKeysService;

export interface IKeysServer extends grpc.UntypedServiceImplementation {
    generateKey: grpc.handleUnaryCall<keys_pb.GenRequest, keys_pb.GenResponse>;
    publicKey: grpc.handleUnaryCall<keys_pb.PubRequest, keys_pb.PubResponse>;
    sign: grpc.handleUnaryCall<keys_pb.SignRequest, keys_pb.SignResponse>;
    verify: grpc.handleUnaryCall<keys_pb.VerifyRequest, keys_pb.VerifyResponse>;
    import: grpc.handleUnaryCall<keys_pb.ImportRequest, keys_pb.ImportResponse>;
    importJSON: grpc.handleUnaryCall<keys_pb.ImportJSONRequest, keys_pb.ImportResponse>;
    export: grpc.handleUnaryCall<keys_pb.ExportRequest, keys_pb.ExportResponse>;
    hash: grpc.handleUnaryCall<keys_pb.HashRequest, keys_pb.HashResponse>;
    removeName: grpc.handleUnaryCall<keys_pb.RemoveNameRequest, keys_pb.RemoveNameResponse>;
    list: grpc.handleUnaryCall<keys_pb.ListRequest, keys_pb.ListResponse>;
    addName: grpc.handleUnaryCall<keys_pb.AddNameRequest, keys_pb.AddNameResponse>;
}

export interface IKeysClient {
    generateKey(request: keys_pb.GenRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.GenResponse) => void): grpc.ClientUnaryCall;
    generateKey(request: keys_pb.GenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.GenResponse) => void): grpc.ClientUnaryCall;
    generateKey(request: keys_pb.GenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.GenResponse) => void): grpc.ClientUnaryCall;
    publicKey(request: keys_pb.PubRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.PubResponse) => void): grpc.ClientUnaryCall;
    publicKey(request: keys_pb.PubRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.PubResponse) => void): grpc.ClientUnaryCall;
    publicKey(request: keys_pb.PubRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.PubResponse) => void): grpc.ClientUnaryCall;
    sign(request: keys_pb.SignRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.SignResponse) => void): grpc.ClientUnaryCall;
    sign(request: keys_pb.SignRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.SignResponse) => void): grpc.ClientUnaryCall;
    sign(request: keys_pb.SignRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.SignResponse) => void): grpc.ClientUnaryCall;
    verify(request: keys_pb.VerifyRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.VerifyResponse) => void): grpc.ClientUnaryCall;
    verify(request: keys_pb.VerifyRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.VerifyResponse) => void): grpc.ClientUnaryCall;
    verify(request: keys_pb.VerifyRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.VerifyResponse) => void): grpc.ClientUnaryCall;
    import(request: keys_pb.ImportRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    import(request: keys_pb.ImportRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    import(request: keys_pb.ImportRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    importJSON(request: keys_pb.ImportJSONRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    importJSON(request: keys_pb.ImportJSONRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    importJSON(request: keys_pb.ImportJSONRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    export(request: keys_pb.ExportRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ExportResponse) => void): grpc.ClientUnaryCall;
    export(request: keys_pb.ExportRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ExportResponse) => void): grpc.ClientUnaryCall;
    export(request: keys_pb.ExportRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ExportResponse) => void): grpc.ClientUnaryCall;
    hash(request: keys_pb.HashRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.HashResponse) => void): grpc.ClientUnaryCall;
    hash(request: keys_pb.HashRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.HashResponse) => void): grpc.ClientUnaryCall;
    hash(request: keys_pb.HashRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.HashResponse) => void): grpc.ClientUnaryCall;
    removeName(request: keys_pb.RemoveNameRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.RemoveNameResponse) => void): grpc.ClientUnaryCall;
    removeName(request: keys_pb.RemoveNameRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.RemoveNameResponse) => void): grpc.ClientUnaryCall;
    removeName(request: keys_pb.RemoveNameRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.RemoveNameResponse) => void): grpc.ClientUnaryCall;
    list(request: keys_pb.ListRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ListResponse) => void): grpc.ClientUnaryCall;
    list(request: keys_pb.ListRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ListResponse) => void): grpc.ClientUnaryCall;
    list(request: keys_pb.ListRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ListResponse) => void): grpc.ClientUnaryCall;
    addName(request: keys_pb.AddNameRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.AddNameResponse) => void): grpc.ClientUnaryCall;
    addName(request: keys_pb.AddNameRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.AddNameResponse) => void): grpc.ClientUnaryCall;
    addName(request: keys_pb.AddNameRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.AddNameResponse) => void): grpc.ClientUnaryCall;
}

export class KeysClient extends grpc.Client implements IKeysClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public generateKey(request: keys_pb.GenRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.GenResponse) => void): grpc.ClientUnaryCall;
    public generateKey(request: keys_pb.GenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.GenResponse) => void): grpc.ClientUnaryCall;
    public generateKey(request: keys_pb.GenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.GenResponse) => void): grpc.ClientUnaryCall;
    public publicKey(request: keys_pb.PubRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.PubResponse) => void): grpc.ClientUnaryCall;
    public publicKey(request: keys_pb.PubRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.PubResponse) => void): grpc.ClientUnaryCall;
    public publicKey(request: keys_pb.PubRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.PubResponse) => void): grpc.ClientUnaryCall;
    public sign(request: keys_pb.SignRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.SignResponse) => void): grpc.ClientUnaryCall;
    public sign(request: keys_pb.SignRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.SignResponse) => void): grpc.ClientUnaryCall;
    public sign(request: keys_pb.SignRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.SignResponse) => void): grpc.ClientUnaryCall;
    public verify(request: keys_pb.VerifyRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.VerifyResponse) => void): grpc.ClientUnaryCall;
    public verify(request: keys_pb.VerifyRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.VerifyResponse) => void): grpc.ClientUnaryCall;
    public verify(request: keys_pb.VerifyRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.VerifyResponse) => void): grpc.ClientUnaryCall;
    public import(request: keys_pb.ImportRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    public import(request: keys_pb.ImportRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    public import(request: keys_pb.ImportRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    public importJSON(request: keys_pb.ImportJSONRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    public importJSON(request: keys_pb.ImportJSONRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    public importJSON(request: keys_pb.ImportJSONRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ImportResponse) => void): grpc.ClientUnaryCall;
    public export(request: keys_pb.ExportRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ExportResponse) => void): grpc.ClientUnaryCall;
    public export(request: keys_pb.ExportRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ExportResponse) => void): grpc.ClientUnaryCall;
    public export(request: keys_pb.ExportRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ExportResponse) => void): grpc.ClientUnaryCall;
    public hash(request: keys_pb.HashRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.HashResponse) => void): grpc.ClientUnaryCall;
    public hash(request: keys_pb.HashRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.HashResponse) => void): grpc.ClientUnaryCall;
    public hash(request: keys_pb.HashRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.HashResponse) => void): grpc.ClientUnaryCall;
    public removeName(request: keys_pb.RemoveNameRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.RemoveNameResponse) => void): grpc.ClientUnaryCall;
    public removeName(request: keys_pb.RemoveNameRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.RemoveNameResponse) => void): grpc.ClientUnaryCall;
    public removeName(request: keys_pb.RemoveNameRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.RemoveNameResponse) => void): grpc.ClientUnaryCall;
    public list(request: keys_pb.ListRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.ListResponse) => void): grpc.ClientUnaryCall;
    public list(request: keys_pb.ListRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.ListResponse) => void): grpc.ClientUnaryCall;
    public list(request: keys_pb.ListRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.ListResponse) => void): grpc.ClientUnaryCall;
    public addName(request: keys_pb.AddNameRequest, callback: (error: grpc.ServiceError | null, response: keys_pb.AddNameResponse) => void): grpc.ClientUnaryCall;
    public addName(request: keys_pb.AddNameRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: keys_pb.AddNameResponse) => void): grpc.ClientUnaryCall;
    public addName(request: keys_pb.AddNameRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: keys_pb.AddNameResponse) => void): grpc.ClientUnaryCall;
}
