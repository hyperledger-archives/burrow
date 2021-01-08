// package: tendermint.abci
// file: tendermint/abci/types.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as tendermint_abci_types_pb from "../../tendermint/abci/types_pb";
import * as tendermint_crypto_proof_pb from "../../tendermint/crypto/proof_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as tendermint_crypto_keys_pb from "../../tendermint/crypto/keys_pb";
import * as tendermint_types_params_pb from "../../tendermint/types/params_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";

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
    listSnapshots: IABCIApplicationService_IListSnapshots;
    offerSnapshot: IABCIApplicationService_IOfferSnapshot;
    loadSnapshotChunk: IABCIApplicationService_ILoadSnapshotChunk;
    applySnapshotChunk: IABCIApplicationService_IApplySnapshotChunk;
}

interface IABCIApplicationService_IEcho extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestEcho, tendermint_abci_types_pb.ResponseEcho> {
    path: "/tendermint.abci.ABCIApplication/Echo";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestEcho>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestEcho>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseEcho>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseEcho>;
}
interface IABCIApplicationService_IFlush extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestFlush, tendermint_abci_types_pb.ResponseFlush> {
    path: "/tendermint.abci.ABCIApplication/Flush";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestFlush>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestFlush>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseFlush>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseFlush>;
}
interface IABCIApplicationService_IInfo extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestInfo, tendermint_abci_types_pb.ResponseInfo> {
    path: "/tendermint.abci.ABCIApplication/Info";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestInfo>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestInfo>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseInfo>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseInfo>;
}
interface IABCIApplicationService_ISetOption extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestSetOption, tendermint_abci_types_pb.ResponseSetOption> {
    path: "/tendermint.abci.ABCIApplication/SetOption";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestSetOption>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestSetOption>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseSetOption>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseSetOption>;
}
interface IABCIApplicationService_IDeliverTx extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestDeliverTx, tendermint_abci_types_pb.ResponseDeliverTx> {
    path: "/tendermint.abci.ABCIApplication/DeliverTx";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestDeliverTx>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestDeliverTx>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseDeliverTx>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseDeliverTx>;
}
interface IABCIApplicationService_ICheckTx extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestCheckTx, tendermint_abci_types_pb.ResponseCheckTx> {
    path: "/tendermint.abci.ABCIApplication/CheckTx";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestCheckTx>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestCheckTx>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseCheckTx>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseCheckTx>;
}
interface IABCIApplicationService_IQuery extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestQuery, tendermint_abci_types_pb.ResponseQuery> {
    path: "/tendermint.abci.ABCIApplication/Query";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestQuery>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestQuery>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseQuery>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseQuery>;
}
interface IABCIApplicationService_ICommit extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestCommit, tendermint_abci_types_pb.ResponseCommit> {
    path: "/tendermint.abci.ABCIApplication/Commit";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestCommit>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestCommit>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseCommit>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseCommit>;
}
interface IABCIApplicationService_IInitChain extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestInitChain, tendermint_abci_types_pb.ResponseInitChain> {
    path: "/tendermint.abci.ABCIApplication/InitChain";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestInitChain>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestInitChain>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseInitChain>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseInitChain>;
}
interface IABCIApplicationService_IBeginBlock extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestBeginBlock, tendermint_abci_types_pb.ResponseBeginBlock> {
    path: "/tendermint.abci.ABCIApplication/BeginBlock";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestBeginBlock>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestBeginBlock>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseBeginBlock>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseBeginBlock>;
}
interface IABCIApplicationService_IEndBlock extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestEndBlock, tendermint_abci_types_pb.ResponseEndBlock> {
    path: "/tendermint.abci.ABCIApplication/EndBlock";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestEndBlock>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestEndBlock>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseEndBlock>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseEndBlock>;
}
interface IABCIApplicationService_IListSnapshots extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestListSnapshots, tendermint_abci_types_pb.ResponseListSnapshots> {
    path: "/tendermint.abci.ABCIApplication/ListSnapshots";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestListSnapshots>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestListSnapshots>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseListSnapshots>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseListSnapshots>;
}
interface IABCIApplicationService_IOfferSnapshot extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestOfferSnapshot, tendermint_abci_types_pb.ResponseOfferSnapshot> {
    path: "/tendermint.abci.ABCIApplication/OfferSnapshot";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestOfferSnapshot>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestOfferSnapshot>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseOfferSnapshot>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseOfferSnapshot>;
}
interface IABCIApplicationService_ILoadSnapshotChunk extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestLoadSnapshotChunk, tendermint_abci_types_pb.ResponseLoadSnapshotChunk> {
    path: "/tendermint.abci.ABCIApplication/LoadSnapshotChunk";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestLoadSnapshotChunk>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestLoadSnapshotChunk>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseLoadSnapshotChunk>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseLoadSnapshotChunk>;
}
interface IABCIApplicationService_IApplySnapshotChunk extends grpc.MethodDefinition<tendermint_abci_types_pb.RequestApplySnapshotChunk, tendermint_abci_types_pb.ResponseApplySnapshotChunk> {
    path: "/tendermint.abci.ABCIApplication/ApplySnapshotChunk";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<tendermint_abci_types_pb.RequestApplySnapshotChunk>;
    requestDeserialize: grpc.deserialize<tendermint_abci_types_pb.RequestApplySnapshotChunk>;
    responseSerialize: grpc.serialize<tendermint_abci_types_pb.ResponseApplySnapshotChunk>;
    responseDeserialize: grpc.deserialize<tendermint_abci_types_pb.ResponseApplySnapshotChunk>;
}

export const ABCIApplicationService: IABCIApplicationService;

export interface IABCIApplicationServer extends grpc.UntypedServiceImplementation {
    echo: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestEcho, tendermint_abci_types_pb.ResponseEcho>;
    flush: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestFlush, tendermint_abci_types_pb.ResponseFlush>;
    info: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestInfo, tendermint_abci_types_pb.ResponseInfo>;
    setOption: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestSetOption, tendermint_abci_types_pb.ResponseSetOption>;
    deliverTx: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestDeliverTx, tendermint_abci_types_pb.ResponseDeliverTx>;
    checkTx: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestCheckTx, tendermint_abci_types_pb.ResponseCheckTx>;
    query: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestQuery, tendermint_abci_types_pb.ResponseQuery>;
    commit: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestCommit, tendermint_abci_types_pb.ResponseCommit>;
    initChain: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestInitChain, tendermint_abci_types_pb.ResponseInitChain>;
    beginBlock: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestBeginBlock, tendermint_abci_types_pb.ResponseBeginBlock>;
    endBlock: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestEndBlock, tendermint_abci_types_pb.ResponseEndBlock>;
    listSnapshots: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestListSnapshots, tendermint_abci_types_pb.ResponseListSnapshots>;
    offerSnapshot: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestOfferSnapshot, tendermint_abci_types_pb.ResponseOfferSnapshot>;
    loadSnapshotChunk: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestLoadSnapshotChunk, tendermint_abci_types_pb.ResponseLoadSnapshotChunk>;
    applySnapshotChunk: grpc.handleUnaryCall<tendermint_abci_types_pb.RequestApplySnapshotChunk, tendermint_abci_types_pb.ResponseApplySnapshotChunk>;
}

export interface IABCIApplicationClient {
    echo(request: tendermint_abci_types_pb.RequestEcho, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    echo(request: tendermint_abci_types_pb.RequestEcho, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    echo(request: tendermint_abci_types_pb.RequestEcho, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    flush(request: tendermint_abci_types_pb.RequestFlush, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    flush(request: tendermint_abci_types_pb.RequestFlush, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    flush(request: tendermint_abci_types_pb.RequestFlush, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    info(request: tendermint_abci_types_pb.RequestInfo, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    info(request: tendermint_abci_types_pb.RequestInfo, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    info(request: tendermint_abci_types_pb.RequestInfo, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    setOption(request: tendermint_abci_types_pb.RequestSetOption, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    setOption(request: tendermint_abci_types_pb.RequestSetOption, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    setOption(request: tendermint_abci_types_pb.RequestSetOption, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    deliverTx(request: tendermint_abci_types_pb.RequestDeliverTx, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    deliverTx(request: tendermint_abci_types_pb.RequestDeliverTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    deliverTx(request: tendermint_abci_types_pb.RequestDeliverTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    checkTx(request: tendermint_abci_types_pb.RequestCheckTx, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    checkTx(request: tendermint_abci_types_pb.RequestCheckTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    checkTx(request: tendermint_abci_types_pb.RequestCheckTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    query(request: tendermint_abci_types_pb.RequestQuery, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    query(request: tendermint_abci_types_pb.RequestQuery, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    query(request: tendermint_abci_types_pb.RequestQuery, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    commit(request: tendermint_abci_types_pb.RequestCommit, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    commit(request: tendermint_abci_types_pb.RequestCommit, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    commit(request: tendermint_abci_types_pb.RequestCommit, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    initChain(request: tendermint_abci_types_pb.RequestInitChain, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    initChain(request: tendermint_abci_types_pb.RequestInitChain, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    initChain(request: tendermint_abci_types_pb.RequestInitChain, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    beginBlock(request: tendermint_abci_types_pb.RequestBeginBlock, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    beginBlock(request: tendermint_abci_types_pb.RequestBeginBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    beginBlock(request: tendermint_abci_types_pb.RequestBeginBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    endBlock(request: tendermint_abci_types_pb.RequestEndBlock, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    endBlock(request: tendermint_abci_types_pb.RequestEndBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    endBlock(request: tendermint_abci_types_pb.RequestEndBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    listSnapshots(request: tendermint_abci_types_pb.RequestListSnapshots, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseListSnapshots) => void): grpc.ClientUnaryCall;
    listSnapshots(request: tendermint_abci_types_pb.RequestListSnapshots, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseListSnapshots) => void): grpc.ClientUnaryCall;
    listSnapshots(request: tendermint_abci_types_pb.RequestListSnapshots, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseListSnapshots) => void): grpc.ClientUnaryCall;
    offerSnapshot(request: tendermint_abci_types_pb.RequestOfferSnapshot, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseOfferSnapshot) => void): grpc.ClientUnaryCall;
    offerSnapshot(request: tendermint_abci_types_pb.RequestOfferSnapshot, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseOfferSnapshot) => void): grpc.ClientUnaryCall;
    offerSnapshot(request: tendermint_abci_types_pb.RequestOfferSnapshot, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseOfferSnapshot) => void): grpc.ClientUnaryCall;
    loadSnapshotChunk(request: tendermint_abci_types_pb.RequestLoadSnapshotChunk, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseLoadSnapshotChunk) => void): grpc.ClientUnaryCall;
    loadSnapshotChunk(request: tendermint_abci_types_pb.RequestLoadSnapshotChunk, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseLoadSnapshotChunk) => void): grpc.ClientUnaryCall;
    loadSnapshotChunk(request: tendermint_abci_types_pb.RequestLoadSnapshotChunk, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseLoadSnapshotChunk) => void): grpc.ClientUnaryCall;
    applySnapshotChunk(request: tendermint_abci_types_pb.RequestApplySnapshotChunk, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseApplySnapshotChunk) => void): grpc.ClientUnaryCall;
    applySnapshotChunk(request: tendermint_abci_types_pb.RequestApplySnapshotChunk, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseApplySnapshotChunk) => void): grpc.ClientUnaryCall;
    applySnapshotChunk(request: tendermint_abci_types_pb.RequestApplySnapshotChunk, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseApplySnapshotChunk) => void): grpc.ClientUnaryCall;
}

export class ABCIApplicationClient extends grpc.Client implements IABCIApplicationClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public echo(request: tendermint_abci_types_pb.RequestEcho, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    public echo(request: tendermint_abci_types_pb.RequestEcho, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    public echo(request: tendermint_abci_types_pb.RequestEcho, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEcho) => void): grpc.ClientUnaryCall;
    public flush(request: tendermint_abci_types_pb.RequestFlush, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    public flush(request: tendermint_abci_types_pb.RequestFlush, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    public flush(request: tendermint_abci_types_pb.RequestFlush, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseFlush) => void): grpc.ClientUnaryCall;
    public info(request: tendermint_abci_types_pb.RequestInfo, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    public info(request: tendermint_abci_types_pb.RequestInfo, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    public info(request: tendermint_abci_types_pb.RequestInfo, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInfo) => void): grpc.ClientUnaryCall;
    public setOption(request: tendermint_abci_types_pb.RequestSetOption, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    public setOption(request: tendermint_abci_types_pb.RequestSetOption, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    public setOption(request: tendermint_abci_types_pb.RequestSetOption, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseSetOption) => void): grpc.ClientUnaryCall;
    public deliverTx(request: tendermint_abci_types_pb.RequestDeliverTx, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    public deliverTx(request: tendermint_abci_types_pb.RequestDeliverTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    public deliverTx(request: tendermint_abci_types_pb.RequestDeliverTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseDeliverTx) => void): grpc.ClientUnaryCall;
    public checkTx(request: tendermint_abci_types_pb.RequestCheckTx, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    public checkTx(request: tendermint_abci_types_pb.RequestCheckTx, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    public checkTx(request: tendermint_abci_types_pb.RequestCheckTx, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCheckTx) => void): grpc.ClientUnaryCall;
    public query(request: tendermint_abci_types_pb.RequestQuery, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    public query(request: tendermint_abci_types_pb.RequestQuery, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    public query(request: tendermint_abci_types_pb.RequestQuery, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseQuery) => void): grpc.ClientUnaryCall;
    public commit(request: tendermint_abci_types_pb.RequestCommit, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    public commit(request: tendermint_abci_types_pb.RequestCommit, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    public commit(request: tendermint_abci_types_pb.RequestCommit, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseCommit) => void): grpc.ClientUnaryCall;
    public initChain(request: tendermint_abci_types_pb.RequestInitChain, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    public initChain(request: tendermint_abci_types_pb.RequestInitChain, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    public initChain(request: tendermint_abci_types_pb.RequestInitChain, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseInitChain) => void): grpc.ClientUnaryCall;
    public beginBlock(request: tendermint_abci_types_pb.RequestBeginBlock, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    public beginBlock(request: tendermint_abci_types_pb.RequestBeginBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    public beginBlock(request: tendermint_abci_types_pb.RequestBeginBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseBeginBlock) => void): grpc.ClientUnaryCall;
    public endBlock(request: tendermint_abci_types_pb.RequestEndBlock, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    public endBlock(request: tendermint_abci_types_pb.RequestEndBlock, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    public endBlock(request: tendermint_abci_types_pb.RequestEndBlock, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseEndBlock) => void): grpc.ClientUnaryCall;
    public listSnapshots(request: tendermint_abci_types_pb.RequestListSnapshots, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseListSnapshots) => void): grpc.ClientUnaryCall;
    public listSnapshots(request: tendermint_abci_types_pb.RequestListSnapshots, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseListSnapshots) => void): grpc.ClientUnaryCall;
    public listSnapshots(request: tendermint_abci_types_pb.RequestListSnapshots, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseListSnapshots) => void): grpc.ClientUnaryCall;
    public offerSnapshot(request: tendermint_abci_types_pb.RequestOfferSnapshot, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseOfferSnapshot) => void): grpc.ClientUnaryCall;
    public offerSnapshot(request: tendermint_abci_types_pb.RequestOfferSnapshot, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseOfferSnapshot) => void): grpc.ClientUnaryCall;
    public offerSnapshot(request: tendermint_abci_types_pb.RequestOfferSnapshot, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseOfferSnapshot) => void): grpc.ClientUnaryCall;
    public loadSnapshotChunk(request: tendermint_abci_types_pb.RequestLoadSnapshotChunk, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseLoadSnapshotChunk) => void): grpc.ClientUnaryCall;
    public loadSnapshotChunk(request: tendermint_abci_types_pb.RequestLoadSnapshotChunk, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseLoadSnapshotChunk) => void): grpc.ClientUnaryCall;
    public loadSnapshotChunk(request: tendermint_abci_types_pb.RequestLoadSnapshotChunk, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseLoadSnapshotChunk) => void): grpc.ClientUnaryCall;
    public applySnapshotChunk(request: tendermint_abci_types_pb.RequestApplySnapshotChunk, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseApplySnapshotChunk) => void): grpc.ClientUnaryCall;
    public applySnapshotChunk(request: tendermint_abci_types_pb.RequestApplySnapshotChunk, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseApplySnapshotChunk) => void): grpc.ClientUnaryCall;
    public applySnapshotChunk(request: tendermint_abci_types_pb.RequestApplySnapshotChunk, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: tendermint_abci_types_pb.ResponseApplySnapshotChunk) => void): grpc.ClientUnaryCall;
}
