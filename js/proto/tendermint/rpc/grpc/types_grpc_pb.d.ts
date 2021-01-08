// package: tendermint.rpc.grpc
// file: tendermint/rpc/grpc/types.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as tendermint_rpc_grpc_types_pb from "../../../tendermint/rpc/grpc/types_pb";
import * as tendermint_abci_types_pb from "../../../tendermint/abci/types_pb";

interface IBroadcastAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    ping: IBroadcastAPIService_IPing;
    broadcastTx: IBroadcastAPIService_IBroadcastTx;
}

interface IBroadcastAPIService_IPing extends grpc.MethodDefinition<tendermint_rpc_grpc_types_pb.RequestPing, tendermint_rpc_grpc_types_pb.ResponsePing> {
    path: "/tendermint.rpc.grpc.BroadcastAPI/Ping";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_rpc_grpc_types_pb.RequestPing>;
    requestDeserialize: grpc.deserialize<tendermint_rpc_grpc_types_pb.RequestPing>;
    responseSerialize: grpc.serialize<tendermint_rpc_grpc_types_pb.ResponsePing>;
    responseDeserialize: grpc.deserialize<tendermint_rpc_grpc_types_pb.ResponsePing>;
}
interface IBroadcastAPIService_IBroadcastTx extends grpc.MethodDefinition<tendermint_rpc_grpc_types_pb.RequestBroadcastTx, tendermint_rpc_grpc_types_pb.ResponseBroadcastTx> {
    path: "/tendermint.rpc.grpc.BroadcastAPI/BroadcastTx";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_rpc_grpc_types_pb.RequestBroadcastTx>;
    requestDeserialize: grpc.deserialize<tendermint_rpc_grpc_types_pb.RequestBroadcastTx>;
    responseSerialize: grpc.serialize<tendermint_rpc_grpc_types_pb.ResponseBroadcastTx>;
    responseDeserialize: grpc.deserialize<tendermint_rpc_grpc_types_pb.ResponseBroadcastTx>;
}

export const BroadcastAPIService: IBroadcastAPIService;

export interface IBroadcastAPIServer extends grpc.UntypedServiceImplementation {
    ping: grpc.handleUnaryCall<tendermint_rpc_grpc_types_pb.RequestPing, tendermint_rpc_grpc_types_pb.ResponsePing>;
    broadcastTx: grpc.handleUnaryCall<tendermint_rpc_grpc_types_pb.RequestBroadcastTx, tendermint_rpc_grpc_types_pb.ResponseBroadcastTx>;
}

export interface IBroadcastAPIClient {
    ping(request: tendermint_rpc_grpc_types_pb.RequestPing, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponsePing) => void): grpc.ClientUnaryCall;
    ping(request: tendermint_rpc_grpc_types_pb.RequestPing, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponsePing) => void): grpc.ClientUnaryCall;
    ping(request: tendermint_rpc_grpc_types_pb.RequestPing, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponsePing) => void): grpc.ClientUnaryCall;
    broadcastTx(request: tendermint_rpc_grpc_types_pb.RequestBroadcastTx, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx) => void): grpc.ClientUnaryCall;
    broadcastTx(request: tendermint_rpc_grpc_types_pb.RequestBroadcastTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx) => void): grpc.ClientUnaryCall;
    broadcastTx(request: tendermint_rpc_grpc_types_pb.RequestBroadcastTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx) => void): grpc.ClientUnaryCall;
}

export class BroadcastAPIClient extends grpc.Client implements IBroadcastAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public ping(request: tendermint_rpc_grpc_types_pb.RequestPing, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponsePing) => void): grpc.ClientUnaryCall;
    public ping(request: tendermint_rpc_grpc_types_pb.RequestPing, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponsePing) => void): grpc.ClientUnaryCall;
    public ping(request: tendermint_rpc_grpc_types_pb.RequestPing, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponsePing) => void): grpc.ClientUnaryCall;
    public broadcastTx(request: tendermint_rpc_grpc_types_pb.RequestBroadcastTx, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx) => void): grpc.ClientUnaryCall;
    public broadcastTx(request: tendermint_rpc_grpc_types_pb.RequestBroadcastTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx) => void): grpc.ClientUnaryCall;
    public broadcastTx(request: tendermint_rpc_grpc_types_pb.RequestBroadcastTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx) => void): grpc.ClientUnaryCall;
}
