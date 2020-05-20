// package: rpcquery
// file: rpcquery.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as rpcquery_pb from "./rpcquery_pb";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as github_com_tendermint_tendermint_abci_types_types_pb from "./github.com/tendermint/tendermint/abci/types/types_pb";
import * as names_pb from "./names_pb";
import * as acm_pb from "./acm_pb";
import * as validator_pb from "./validator_pb";
import * as registry_pb from "./registry_pb";
import * as rpc_pb from "./rpc_pb";
import * as payload_pb from "./payload_pb";

interface IQueryService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    status: IQueryService_IStatus;
    getAccount: IQueryService_IGetAccount;
    getMetadata: IQueryService_IGetMetadata;
    getStorage: IQueryService_IGetStorage;
    listAccounts: IQueryService_IListAccounts;
    getName: IQueryService_IGetName;
    listNames: IQueryService_IListNames;
    getNetworkRegistry: IQueryService_IGetNetworkRegistry;
    getValidatorSet: IQueryService_IGetValidatorSet;
    getValidatorSetHistory: IQueryService_IGetValidatorSetHistory;
    getProposal: IQueryService_IGetProposal;
    listProposals: IQueryService_IListProposals;
    getStats: IQueryService_IGetStats;
    getBlockHeader: IQueryService_IGetBlockHeader;
}

interface IQueryService_IStatus extends grpc.MethodDefinition<rpcquery_pb.StatusParam, rpc_pb.ResultStatus> {
    path: string; // "/rpcquery.Query/Status"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.StatusParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.StatusParam>;
    responseSerialize: grpc.serialize<rpc_pb.ResultStatus>;
    responseDeserialize: grpc.deserialize<rpc_pb.ResultStatus>;
}
interface IQueryService_IGetAccount extends grpc.MethodDefinition<rpcquery_pb.GetAccountParam, acm_pb.Account> {
    path: string; // "/rpcquery.Query/GetAccount"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetAccountParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetAccountParam>;
    responseSerialize: grpc.serialize<acm_pb.Account>;
    responseDeserialize: grpc.deserialize<acm_pb.Account>;
}
interface IQueryService_IGetMetadata extends grpc.MethodDefinition<rpcquery_pb.GetMetadataParam, rpcquery_pb.MetadataResult> {
    path: string; // "/rpcquery.Query/GetMetadata"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetMetadataParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetMetadataParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.MetadataResult>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.MetadataResult>;
}
interface IQueryService_IGetStorage extends grpc.MethodDefinition<rpcquery_pb.GetStorageParam, rpcquery_pb.StorageValue> {
    path: string; // "/rpcquery.Query/GetStorage"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetStorageParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetStorageParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.StorageValue>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.StorageValue>;
}
interface IQueryService_IListAccounts extends grpc.MethodDefinition<rpcquery_pb.ListAccountsParam, acm_pb.Account> {
    path: string; // "/rpcquery.Query/ListAccounts"
    requestStream: boolean; // false
    responseStream: boolean; // true
    requestSerialize: grpc.serialize<rpcquery_pb.ListAccountsParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.ListAccountsParam>;
    responseSerialize: grpc.serialize<acm_pb.Account>;
    responseDeserialize: grpc.deserialize<acm_pb.Account>;
}
interface IQueryService_IGetName extends grpc.MethodDefinition<rpcquery_pb.GetNameParam, names_pb.Entry> {
    path: string; // "/rpcquery.Query/GetName"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetNameParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetNameParam>;
    responseSerialize: grpc.serialize<names_pb.Entry>;
    responseDeserialize: grpc.deserialize<names_pb.Entry>;
}
interface IQueryService_IListNames extends grpc.MethodDefinition<rpcquery_pb.ListNamesParam, names_pb.Entry> {
    path: string; // "/rpcquery.Query/ListNames"
    requestStream: boolean; // false
    responseStream: boolean; // true
    requestSerialize: grpc.serialize<rpcquery_pb.ListNamesParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.ListNamesParam>;
    responseSerialize: grpc.serialize<names_pb.Entry>;
    responseDeserialize: grpc.deserialize<names_pb.Entry>;
}
interface IQueryService_IGetNetworkRegistry extends grpc.MethodDefinition<rpcquery_pb.GetNetworkRegistryParam, rpcquery_pb.NetworkRegistry> {
    path: string; // "/rpcquery.Query/GetNetworkRegistry"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetNetworkRegistryParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetNetworkRegistryParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.NetworkRegistry>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.NetworkRegistry>;
}
interface IQueryService_IGetValidatorSet extends grpc.MethodDefinition<rpcquery_pb.GetValidatorSetParam, rpcquery_pb.ValidatorSet> {
    path: string; // "/rpcquery.Query/GetValidatorSet"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetValidatorSetParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetValidatorSetParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.ValidatorSet>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.ValidatorSet>;
}
interface IQueryService_IGetValidatorSetHistory extends grpc.MethodDefinition<rpcquery_pb.GetValidatorSetHistoryParam, rpcquery_pb.ValidatorSetHistory> {
    path: string; // "/rpcquery.Query/GetValidatorSetHistory"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetValidatorSetHistoryParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetValidatorSetHistoryParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.ValidatorSetHistory>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.ValidatorSetHistory>;
}
interface IQueryService_IGetProposal extends grpc.MethodDefinition<rpcquery_pb.GetProposalParam, payload_pb.Ballot> {
    path: string; // "/rpcquery.Query/GetProposal"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetProposalParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetProposalParam>;
    responseSerialize: grpc.serialize<payload_pb.Ballot>;
    responseDeserialize: grpc.deserialize<payload_pb.Ballot>;
}
interface IQueryService_IListProposals extends grpc.MethodDefinition<rpcquery_pb.ListProposalsParam, rpcquery_pb.ProposalResult> {
    path: string; // "/rpcquery.Query/ListProposals"
    requestStream: boolean; // false
    responseStream: boolean; // true
    requestSerialize: grpc.serialize<rpcquery_pb.ListProposalsParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.ListProposalsParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.ProposalResult>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.ProposalResult>;
}
interface IQueryService_IGetStats extends grpc.MethodDefinition<rpcquery_pb.GetStatsParam, rpcquery_pb.Stats> {
    path: string; // "/rpcquery.Query/GetStats"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetStatsParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetStatsParam>;
    responseSerialize: grpc.serialize<rpcquery_pb.Stats>;
    responseDeserialize: grpc.deserialize<rpcquery_pb.Stats>;
}
interface IQueryService_IGetBlockHeader extends grpc.MethodDefinition<rpcquery_pb.GetBlockParam, github_com_tendermint_tendermint_abci_types_types_pb.Header> {
    path: string; // "/rpcquery.Query/GetBlockHeader"
    requestStream: boolean; // false
    responseStream: boolean; // false
    requestSerialize: grpc.serialize<rpcquery_pb.GetBlockParam>;
    requestDeserialize: grpc.deserialize<rpcquery_pb.GetBlockParam>;
    responseSerialize: grpc.serialize<github_com_tendermint_tendermint_abci_types_types_pb.Header>;
    responseDeserialize: grpc.deserialize<github_com_tendermint_tendermint_abci_types_types_pb.Header>;
}

export const QueryService: IQueryService;

export interface IQueryServer {
    status: grpc.handleUnaryCall<rpcquery_pb.StatusParam, rpc_pb.ResultStatus>;
    getAccount: grpc.handleUnaryCall<rpcquery_pb.GetAccountParam, acm_pb.Account>;
    getMetadata: grpc.handleUnaryCall<rpcquery_pb.GetMetadataParam, rpcquery_pb.MetadataResult>;
    getStorage: grpc.handleUnaryCall<rpcquery_pb.GetStorageParam, rpcquery_pb.StorageValue>;
    listAccounts: grpc.handleServerStreamingCall<rpcquery_pb.ListAccountsParam, acm_pb.Account>;
    getName: grpc.handleUnaryCall<rpcquery_pb.GetNameParam, names_pb.Entry>;
    listNames: grpc.handleServerStreamingCall<rpcquery_pb.ListNamesParam, names_pb.Entry>;
    getNetworkRegistry: grpc.handleUnaryCall<rpcquery_pb.GetNetworkRegistryParam, rpcquery_pb.NetworkRegistry>;
    getValidatorSet: grpc.handleUnaryCall<rpcquery_pb.GetValidatorSetParam, rpcquery_pb.ValidatorSet>;
    getValidatorSetHistory: grpc.handleUnaryCall<rpcquery_pb.GetValidatorSetHistoryParam, rpcquery_pb.ValidatorSetHistory>;
    getProposal: grpc.handleUnaryCall<rpcquery_pb.GetProposalParam, payload_pb.Ballot>;
    listProposals: grpc.handleServerStreamingCall<rpcquery_pb.ListProposalsParam, rpcquery_pb.ProposalResult>;
    getStats: grpc.handleUnaryCall<rpcquery_pb.GetStatsParam, rpcquery_pb.Stats>;
    getBlockHeader: grpc.handleUnaryCall<rpcquery_pb.GetBlockParam, github_com_tendermint_tendermint_abci_types_types_pb.Header>;
}

export interface IQueryClient {
    status(request: rpcquery_pb.StatusParam, callback: (error: grpc.ServiceError | null, response: rpc_pb.ResultStatus) => void): grpc.ClientUnaryCall;
    status(request: rpcquery_pb.StatusParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpc_pb.ResultStatus) => void): grpc.ClientUnaryCall;
    status(request: rpcquery_pb.StatusParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpc_pb.ResultStatus) => void): grpc.ClientUnaryCall;
    getAccount(request: rpcquery_pb.GetAccountParam, callback: (error: grpc.ServiceError | null, response: acm_pb.Account) => void): grpc.ClientUnaryCall;
    getAccount(request: rpcquery_pb.GetAccountParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: acm_pb.Account) => void): grpc.ClientUnaryCall;
    getAccount(request: rpcquery_pb.GetAccountParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: acm_pb.Account) => void): grpc.ClientUnaryCall;
    getMetadata(request: rpcquery_pb.GetMetadataParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.MetadataResult) => void): grpc.ClientUnaryCall;
    getMetadata(request: rpcquery_pb.GetMetadataParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.MetadataResult) => void): grpc.ClientUnaryCall;
    getMetadata(request: rpcquery_pb.GetMetadataParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.MetadataResult) => void): grpc.ClientUnaryCall;
    getStorage(request: rpcquery_pb.GetStorageParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.StorageValue) => void): grpc.ClientUnaryCall;
    getStorage(request: rpcquery_pb.GetStorageParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.StorageValue) => void): grpc.ClientUnaryCall;
    getStorage(request: rpcquery_pb.GetStorageParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.StorageValue) => void): grpc.ClientUnaryCall;
    listAccounts(request: rpcquery_pb.ListAccountsParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<acm_pb.Account>;
    listAccounts(request: rpcquery_pb.ListAccountsParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<acm_pb.Account>;
    getName(request: rpcquery_pb.GetNameParam, callback: (error: grpc.ServiceError | null, response: names_pb.Entry) => void): grpc.ClientUnaryCall;
    getName(request: rpcquery_pb.GetNameParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: names_pb.Entry) => void): grpc.ClientUnaryCall;
    getName(request: rpcquery_pb.GetNameParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: names_pb.Entry) => void): grpc.ClientUnaryCall;
    listNames(request: rpcquery_pb.ListNamesParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<names_pb.Entry>;
    listNames(request: rpcquery_pb.ListNamesParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<names_pb.Entry>;
    getNetworkRegistry(request: rpcquery_pb.GetNetworkRegistryParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.NetworkRegistry) => void): grpc.ClientUnaryCall;
    getNetworkRegistry(request: rpcquery_pb.GetNetworkRegistryParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.NetworkRegistry) => void): grpc.ClientUnaryCall;
    getNetworkRegistry(request: rpcquery_pb.GetNetworkRegistryParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.NetworkRegistry) => void): grpc.ClientUnaryCall;
    getValidatorSet(request: rpcquery_pb.GetValidatorSetParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSet) => void): grpc.ClientUnaryCall;
    getValidatorSet(request: rpcquery_pb.GetValidatorSetParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSet) => void): grpc.ClientUnaryCall;
    getValidatorSet(request: rpcquery_pb.GetValidatorSetParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSet) => void): grpc.ClientUnaryCall;
    getValidatorSetHistory(request: rpcquery_pb.GetValidatorSetHistoryParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSetHistory) => void): grpc.ClientUnaryCall;
    getValidatorSetHistory(request: rpcquery_pb.GetValidatorSetHistoryParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSetHistory) => void): grpc.ClientUnaryCall;
    getValidatorSetHistory(request: rpcquery_pb.GetValidatorSetHistoryParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSetHistory) => void): grpc.ClientUnaryCall;
    getProposal(request: rpcquery_pb.GetProposalParam, callback: (error: grpc.ServiceError | null, response: payload_pb.Ballot) => void): grpc.ClientUnaryCall;
    getProposal(request: rpcquery_pb.GetProposalParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: payload_pb.Ballot) => void): grpc.ClientUnaryCall;
    getProposal(request: rpcquery_pb.GetProposalParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: payload_pb.Ballot) => void): grpc.ClientUnaryCall;
    listProposals(request: rpcquery_pb.ListProposalsParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcquery_pb.ProposalResult>;
    listProposals(request: rpcquery_pb.ListProposalsParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcquery_pb.ProposalResult>;
    getStats(request: rpcquery_pb.GetStatsParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.Stats) => void): grpc.ClientUnaryCall;
    getStats(request: rpcquery_pb.GetStatsParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.Stats) => void): grpc.ClientUnaryCall;
    getStats(request: rpcquery_pb.GetStatsParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.Stats) => void): grpc.ClientUnaryCall;
    getBlockHeader(request: rpcquery_pb.GetBlockParam, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.Header) => void): grpc.ClientUnaryCall;
    getBlockHeader(request: rpcquery_pb.GetBlockParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.Header) => void): grpc.ClientUnaryCall;
    getBlockHeader(request: rpcquery_pb.GetBlockParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.Header) => void): grpc.ClientUnaryCall;
}

export class QueryClient extends grpc.Client implements IQueryClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public status(request: rpcquery_pb.StatusParam, callback: (error: grpc.ServiceError | null, response: rpc_pb.ResultStatus) => void): grpc.ClientUnaryCall;
    public status(request: rpcquery_pb.StatusParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpc_pb.ResultStatus) => void): grpc.ClientUnaryCall;
    public status(request: rpcquery_pb.StatusParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpc_pb.ResultStatus) => void): grpc.ClientUnaryCall;
    public getAccount(request: rpcquery_pb.GetAccountParam, callback: (error: grpc.ServiceError | null, response: acm_pb.Account) => void): grpc.ClientUnaryCall;
    public getAccount(request: rpcquery_pb.GetAccountParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: acm_pb.Account) => void): grpc.ClientUnaryCall;
    public getAccount(request: rpcquery_pb.GetAccountParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: acm_pb.Account) => void): grpc.ClientUnaryCall;
    public getMetadata(request: rpcquery_pb.GetMetadataParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.MetadataResult) => void): grpc.ClientUnaryCall;
    public getMetadata(request: rpcquery_pb.GetMetadataParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.MetadataResult) => void): grpc.ClientUnaryCall;
    public getMetadata(request: rpcquery_pb.GetMetadataParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.MetadataResult) => void): grpc.ClientUnaryCall;
    public getStorage(request: rpcquery_pb.GetStorageParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.StorageValue) => void): grpc.ClientUnaryCall;
    public getStorage(request: rpcquery_pb.GetStorageParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.StorageValue) => void): grpc.ClientUnaryCall;
    public getStorage(request: rpcquery_pb.GetStorageParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.StorageValue) => void): grpc.ClientUnaryCall;
    public listAccounts(request: rpcquery_pb.ListAccountsParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<acm_pb.Account>;
    public listAccounts(request: rpcquery_pb.ListAccountsParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<acm_pb.Account>;
    public getName(request: rpcquery_pb.GetNameParam, callback: (error: grpc.ServiceError | null, response: names_pb.Entry) => void): grpc.ClientUnaryCall;
    public getName(request: rpcquery_pb.GetNameParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: names_pb.Entry) => void): grpc.ClientUnaryCall;
    public getName(request: rpcquery_pb.GetNameParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: names_pb.Entry) => void): grpc.ClientUnaryCall;
    public listNames(request: rpcquery_pb.ListNamesParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<names_pb.Entry>;
    public listNames(request: rpcquery_pb.ListNamesParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<names_pb.Entry>;
    public getNetworkRegistry(request: rpcquery_pb.GetNetworkRegistryParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.NetworkRegistry) => void): grpc.ClientUnaryCall;
    public getNetworkRegistry(request: rpcquery_pb.GetNetworkRegistryParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.NetworkRegistry) => void): grpc.ClientUnaryCall;
    public getNetworkRegistry(request: rpcquery_pb.GetNetworkRegistryParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.NetworkRegistry) => void): grpc.ClientUnaryCall;
    public getValidatorSet(request: rpcquery_pb.GetValidatorSetParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSet) => void): grpc.ClientUnaryCall;
    public getValidatorSet(request: rpcquery_pb.GetValidatorSetParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSet) => void): grpc.ClientUnaryCall;
    public getValidatorSet(request: rpcquery_pb.GetValidatorSetParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSet) => void): grpc.ClientUnaryCall;
    public getValidatorSetHistory(request: rpcquery_pb.GetValidatorSetHistoryParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSetHistory) => void): grpc.ClientUnaryCall;
    public getValidatorSetHistory(request: rpcquery_pb.GetValidatorSetHistoryParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSetHistory) => void): grpc.ClientUnaryCall;
    public getValidatorSetHistory(request: rpcquery_pb.GetValidatorSetHistoryParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.ValidatorSetHistory) => void): grpc.ClientUnaryCall;
    public getProposal(request: rpcquery_pb.GetProposalParam, callback: (error: grpc.ServiceError | null, response: payload_pb.Ballot) => void): grpc.ClientUnaryCall;
    public getProposal(request: rpcquery_pb.GetProposalParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: payload_pb.Ballot) => void): grpc.ClientUnaryCall;
    public getProposal(request: rpcquery_pb.GetProposalParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: payload_pb.Ballot) => void): grpc.ClientUnaryCall;
    public listProposals(request: rpcquery_pb.ListProposalsParam, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcquery_pb.ProposalResult>;
    public listProposals(request: rpcquery_pb.ListProposalsParam, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<rpcquery_pb.ProposalResult>;
    public getStats(request: rpcquery_pb.GetStatsParam, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.Stats) => void): grpc.ClientUnaryCall;
    public getStats(request: rpcquery_pb.GetStatsParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.Stats) => void): grpc.ClientUnaryCall;
    public getStats(request: rpcquery_pb.GetStatsParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: rpcquery_pb.Stats) => void): grpc.ClientUnaryCall;
    public getBlockHeader(request: rpcquery_pb.GetBlockParam, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.Header) => void): grpc.ClientUnaryCall;
    public getBlockHeader(request: rpcquery_pb.GetBlockParam, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.Header) => void): grpc.ClientUnaryCall;
    public getBlockHeader(request: rpcquery_pb.GetBlockParam, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: github_com_tendermint_tendermint_abci_types_types_pb.Header) => void): grpc.ClientUnaryCall;
}
