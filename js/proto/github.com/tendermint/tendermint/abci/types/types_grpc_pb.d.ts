// GENERATED CODE -- DO NOT EDIT!

// package: types
// file: github.com/tendermint/tendermint/abci/types/types.proto

import * as github_com_tendermint_tendermint_abci_types_types_pb from "../../../../../github.com/tendermint/tendermint/abci/types/types_pb";
import * as grpc from "grpc";

interface IABCIApplicationService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
  echo: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>;
  flush: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>;
  info: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>;
  setOption: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>;
  deliverTx: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>;
  checkTx: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>;
  query: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>;
  commit: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>;
  initChain: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>;
  beginBlock: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>;
  endBlock: grpc.MethodDefinition<github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>;
}

export const ABCIApplicationService: IABCIApplicationService;

export class ABCIApplicationClient extends grpc.Client {
  constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
  echo(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>): grpc.ClientUnaryCall;
  echo(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>): grpc.ClientUnaryCall;
  echo(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho>): grpc.ClientUnaryCall;
  flush(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>): grpc.ClientUnaryCall;
  flush(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>): grpc.ClientUnaryCall;
  flush(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush>): grpc.ClientUnaryCall;
  info(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>): grpc.ClientUnaryCall;
  info(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>): grpc.ClientUnaryCall;
  info(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo>): grpc.ClientUnaryCall;
  setOption(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>): grpc.ClientUnaryCall;
  setOption(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>): grpc.ClientUnaryCall;
  setOption(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption>): grpc.ClientUnaryCall;
  deliverTx(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>): grpc.ClientUnaryCall;
  deliverTx(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>): grpc.ClientUnaryCall;
  deliverTx(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx>): grpc.ClientUnaryCall;
  checkTx(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>): grpc.ClientUnaryCall;
  checkTx(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>): grpc.ClientUnaryCall;
  checkTx(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx>): grpc.ClientUnaryCall;
  query(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>): grpc.ClientUnaryCall;
  query(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>): grpc.ClientUnaryCall;
  query(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery>): grpc.ClientUnaryCall;
  commit(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>): grpc.ClientUnaryCall;
  commit(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>): grpc.ClientUnaryCall;
  commit(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit>): grpc.ClientUnaryCall;
  initChain(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>): grpc.ClientUnaryCall;
  initChain(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>): grpc.ClientUnaryCall;
  initChain(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain>): grpc.ClientUnaryCall;
  beginBlock(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>): grpc.ClientUnaryCall;
  beginBlock(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>): grpc.ClientUnaryCall;
  beginBlock(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock>): grpc.ClientUnaryCall;
  endBlock(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>): grpc.ClientUnaryCall;
  endBlock(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, metadataOrOptions: grpc.Metadata | grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>): grpc.ClientUnaryCall;
  endBlock(argument: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock, metadata: grpc.Metadata | null, options: grpc.CallOptions | null, callback: grpc.requestCallback<github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock>): grpc.ClientUnaryCall;
}
