// GENERATED CODE -- DO NOT EDIT!

// package: rpcquery
// file: rpcquery.proto

import * as rpcquery_pb from "./rpcquery_pb";
import * as github_com_tendermint_tendermint_abci_types_types_pb from "./github.com/tendermint/tendermint/abci/types/types_pb";
import * as names_pb from "./names_pb";
import * as acm_pb from "./acm_pb";
import * as rpc_pb from "./rpc_pb";
import * as payload_pb from "./payload_pb";
import * as grpc from "grpc";

interface IQueryService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  status: grpc.MethodDefinition<rpcquery_pb.StatusParam, rpc_pb.ResultStatus>;
  getAccount: grpc.MethodDefinition<rpcquery_pb.GetAccountParam, acm_pb.Account>;
  getMetadata: grpc.MethodDefinition<rpcquery_pb.GetMetadataParam, rpcquery_pb.MetadataResult>;
  getStorage: grpc.MethodDefinition<rpcquery_pb.GetStorageParam, rpcquery_pb.StorageValue>;
  listAccounts: grpc.MethodDefinition<rpcquery_pb.ListAccountsParam, acm_pb.Account>;
  getName: grpc.MethodDefinition<rpcquery_pb.GetNameParam, names_pb.Entry>;
  listNames: grpc.MethodDefinition<rpcquery_pb.ListNamesParam, names_pb.Entry>;
  getNetworkRegistry: grpc.MethodDefinition<rpcquery_pb.GetNetworkRegistryParam, rpcquery_pb.NetworkRegistry>;
  getValidatorSet: grpc.MethodDefinition<rpcquery_pb.GetValidatorSetParam, rpcquery_pb.ValidatorSet>;
  getValidatorSetHistory: grpc.MethodDefinition<rpcquery_pb.GetValidatorSetHistoryParam, rpcquery_pb.ValidatorSetHistory>;
  getProposal: grpc.MethodDefinition<rpcquery_pb.GetProposalParam, payload_pb.Ballot>;
  listProposals: grpc.MethodDefinition<rpcquery_pb.ListProposalsParam, rpcquery_pb.ProposalResult>;
  getStats: grpc.MethodDefinition<rpcquery_pb.GetStatsParam, rpcquery_pb.Stats>;
  getBlockHeader: grpc.MethodDefinition<rpcquery_pb.GetBlockParam, github_com_tendermint_tendermint_abci_types_types_pb.Header>;
}

export const QueryService: IQueryService;

export class QueryClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  status(argument: rpcquery_pb.StatusParam, callback: grpc.requestCallback<rpc_pb.ResultStatus>): grpc.ClientUnaryCall;
  status(argument: rpcquery_pb.StatusParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpc_pb.ResultStatus>): grpc.ClientUnaryCall;
  status(argument: rpcquery_pb.StatusParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpc_pb.ResultStatus>): grpc.ClientUnaryCall;
  getAccount(argument: rpcquery_pb.GetAccountParam, callback: grpc.requestCallback<acm_pb.Account>): grpc.ClientUnaryCall;
  getAccount(argument: rpcquery_pb.GetAccountParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<acm_pb.Account>): grpc.ClientUnaryCall;
  getAccount(argument: rpcquery_pb.GetAccountParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<acm_pb.Account>): grpc.ClientUnaryCall;
  getMetadata(argument: rpcquery_pb.GetMetadataParam, callback: grpc.requestCallback<rpcquery_pb.MetadataResult>): grpc.ClientUnaryCall;
  getMetadata(argument: rpcquery_pb.GetMetadataParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.MetadataResult>): grpc.ClientUnaryCall;
  getMetadata(argument: rpcquery_pb.GetMetadataParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.MetadataResult>): grpc.ClientUnaryCall;
  getStorage(argument: rpcquery_pb.GetStorageParam, callback: grpc.requestCallback<rpcquery_pb.StorageValue>): grpc.ClientUnaryCall;
  getStorage(argument: rpcquery_pb.GetStorageParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.StorageValue>): grpc.ClientUnaryCall;
  getStorage(argument: rpcquery_pb.GetStorageParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.StorageValue>): grpc.ClientUnaryCall;
  listAccounts(argument: rpcquery_pb.ListAccountsParam, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<acm_pb.Account>;
  listAccounts(argument: rpcquery_pb.ListAccountsParam, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<acm_pb.Account>;
  getName(argument: rpcquery_pb.GetNameParam, callback: grpc.requestCallback<names_pb.Entry>): grpc.ClientUnaryCall;
  getName(argument: rpcquery_pb.GetNameParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<names_pb.Entry>): grpc.ClientUnaryCall;
  getName(argument: rpcquery_pb.GetNameParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<names_pb.Entry>): grpc.ClientUnaryCall;
  listNames(argument: rpcquery_pb.ListNamesParam, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<names_pb.Entry>;
  listNames(argument: rpcquery_pb.ListNamesParam, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<names_pb.Entry>;
  getNetworkRegistry(argument: rpcquery_pb.GetNetworkRegistryParam, callback: grpc.requestCallback<rpcquery_pb.NetworkRegistry>): grpc.ClientUnaryCall;
  getNetworkRegistry(argument: rpcquery_pb.GetNetworkRegistryParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.NetworkRegistry>): grpc.ClientUnaryCall;
  getNetworkRegistry(argument: rpcquery_pb.GetNetworkRegistryParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.NetworkRegistry>): grpc.ClientUnaryCall;
  getValidatorSet(argument: rpcquery_pb.GetValidatorSetParam, callback: grpc.requestCallback<rpcquery_pb.ValidatorSet>): grpc.ClientUnaryCall;
  getValidatorSet(argument: rpcquery_pb.GetValidatorSetParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.ValidatorSet>): grpc.ClientUnaryCall;
  getValidatorSet(argument: rpcquery_pb.GetValidatorSetParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.ValidatorSet>): grpc.ClientUnaryCall;
  getValidatorSetHistory(argument: rpcquery_pb.GetValidatorSetHistoryParam, callback: grpc.requestCallback<rpcquery_pb.ValidatorSetHistory>): grpc.ClientUnaryCall;
  getValidatorSetHistory(argument: rpcquery_pb.GetValidatorSetHistoryParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.ValidatorSetHistory>): grpc.ClientUnaryCall;
  getValidatorSetHistory(argument: rpcquery_pb.GetValidatorSetHistoryParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.ValidatorSetHistory>): grpc.ClientUnaryCall;
  getProposal(argument: rpcquery_pb.GetProposalParam, callback: grpc.requestCallback<payload_pb.Ballot>): grpc.ClientUnaryCall;
  getProposal(argument: rpcquery_pb.GetProposalParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<payload_pb.Ballot>): grpc.ClientUnaryCall;
  getProposal(argument: rpcquery_pb.GetProposalParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<payload_pb.Ballot>): grpc.ClientUnaryCall;
  listProposals(argument: rpcquery_pb.ListProposalsParam, metadataOrOptions?: grpc.Metadata | grpc.CallOptions | null): grpc.ClientReadableStream<rpcquery_pb.ProposalResult>;
  listProposals(argument: rpcquery_pb.ListProposalsParam, metadata?: grpc.Metadata | null, options?: grpc.CallOptions | null): grpc.ClientReadableStream<rpcquery_pb.ProposalResult>;
  getStats(argument: rpcquery_pb.GetStatsParam, callback: grpc.requestCallback<rpcquery_pb.Stats>): grpc.ClientUnaryCall;
  getStats(argument: rpcquery_pb.GetStatsParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.Stats>): grpc.ClientUnaryCall;
  getStats(argument: rpcquery_pb.GetStatsParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<rpcquery_pb.Stats>): grpc.ClientUnaryCall;
  getBlockHeader(argument: rpcquery_pb.GetBlockParam, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.Header>): grpc.ClientUnaryCall;
  getBlockHeader(argument: rpcquery_pb.GetBlockParam, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.Header>): grpc.ClientUnaryCall;
  getBlockHeader(argument: rpcquery_pb.GetBlockParam, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.Header>): grpc.ClientUnaryCall;
}
