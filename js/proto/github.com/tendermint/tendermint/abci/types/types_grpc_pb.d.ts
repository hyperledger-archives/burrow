// package: types
// file: github.com/tendermint/tendermint/abci/types/types.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as github_com_tendermint_tendermint_abci_types_types_pb from "../../../../../github.com/tendermint/tendermint/abci/types/types_pb";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "../../../../../github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as github_com_tendermint_tendermint_libs_common_types_pb from "../../../../../github.com/tendermint/tendermint/libs/common/types_pb";

interface IABCIApplicationService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    echo: IABCIApplicationService_IEcho;
    flush: IABCIApplicationService_IFlush;
    info: IABCIApplicationService_IInfo;
    setOption: IABCIApplicationService_ISetOption;
    deliverTx: IABCIApplicationService_IDeliverTx;
    checkTx: IABCIApplicationService_ICheckTx;
    query: IABCIApplicationService_IQuery;
    commit: IABCIApplicationService_ICommit;
    initChain: IABCIApplicationService_IInitChain;
    beginBlock: IABCIApplicationService_IBeginBlock;
    endBlock: IABCIApplicationService_IEndBlock;
}

interface IABCIApplicationService_IEcho extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho> {
    path: string; // "/types.ABCIApplication/Echo"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>;
}
interface IABCIApplicationService_IFlush extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush> {
    path: string; // "/types.ABCIApplication/Flush"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>;
}
interface IABCIApplicationService_IInfo extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo> {
    path: string; // "/types.ABCIApplication/Info"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>;
}
interface IABCIApplicationService_ISetOption extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption> {
    path: string; // "/types.ABCIApplication/SetOption"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>;
}
interface IABCIApplicationService_IDeliverTx extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx> {
    path: string; // "/types.ABCIApplication/DeliverTx"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>;
}
interface IABCIApplicationService_ICheckTx extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx> {
    path: string; // "/types.ABCIApplication/CheckTx"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>;
}
interface IABCIApplicationService_IQuery extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery> {
    path: string; // "/types.ABCIApplication/Query"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>;
}
interface IABCIApplicationService_ICommit extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit> {
    path: string; // "/types.ABCIApplication/Commit"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>;
}
interface IABCIApplicationService_IInitChain extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain> {
    path: string; // "/types.ABCIApplication/InitChain"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>;
}
interface IABCIApplicationService_IBeginBlock extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock> {
    path: string; // "/types.ABCIApplication/BeginBlock"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>;
}
interface IABCIApplicationService_IEndBlock extends grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock> {
    path: string; // "/types.ABCIApplication/EndBlock"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock>;
    requestDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>;
}

export const ABCIApplicationService: IABCIApplicationService;

export interface IABCIApplicationServer {
    echo: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>;
    flush: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>;
    info: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>;
    setOption: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>;
    deliverTx: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>;
    checkTx: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>;
    query: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>;
    commit: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>;
    initChain: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>;
    beginBlock: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>;
    endBlock: grpc.handleUnaryCall<github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>;
}

export interface IABCIApplicationClient {
    echo(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    echo(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    echo(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    flush(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    flush(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    flush(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    info(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    info(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    info(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    setOption(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    setOption(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    setOption(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    deliverTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    deliverTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    deliverTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    checkTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    checkTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    checkTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    query(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    query(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    query(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    commit(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    commit(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    commit(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    initChain(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    initChain(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    initChain(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    beginBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    beginBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    beginBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    endBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    endBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    endBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
}

export class ABCIApplicationClient extends grpc.Client implements IABCIApplicationClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public echo(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    public echo(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    public echo(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    public flush(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    public flush(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    public flush(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    public info(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    public info(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    public info(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    public setOption(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    public setOption(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    public setOption(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    public deliverTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    public deliverTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    public deliverTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    public checkTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    public checkTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    public checkTx(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    public query(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    public query(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    public query(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    public commit(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    public commit(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    public commit(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    public initChain(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    public initChain(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    public initChain(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    public beginBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    public beginBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    public beginBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    public endBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    public endBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    public endBlock(request: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
}
