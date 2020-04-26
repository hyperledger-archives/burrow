// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var rpcquery_pb = require('./rpcquery_pb.js');
var github_com_gogo_protobuf_gogoproto_gogo_pb = require('./github.com/gogo/protobuf/gogoproto/gogo_pb.js');
var github_com_tendermint_tendermint_abci_types_types_pb = require('./github.com/tendermint/tendermint/abci/types/types_pb.js');
var names_pb = require('./names_pb.js');
var acm_pb = require('./acm_pb.js');
var validator_pb = require('./validator_pb.js');
var registry_pb = require('./registry_pb.js');
var rpc_pb = require('./rpc_pb.js');
var payload_pb = require('./payload_pb.js');

function serialize_acm_Account(arg) {
  if (!(arg instanceof acm_pb.Account)) {
    throw new Error('Expected argument of type acm.Account');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_acm_Account(buffer_arg) {
  return acm_pb.Account.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_names_Entry(arg) {
  if (!(arg instanceof names_pb.Entry)) {
    throw new Error('Expected argument of type names.Entry');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_names_Entry(buffer_arg) {
  return names_pb.Entry.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_payload_Ballot(arg) {
  if (!(arg instanceof payload_pb.Ballot)) {
    throw new Error('Expected argument of type payload.Ballot');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_payload_Ballot(buffer_arg) {
  return payload_pb.Ballot.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpc_ResultStatus(arg) {
  if (!(arg instanceof rpc_pb.ResultStatus)) {
    throw new Error('Expected argument of type rpc.ResultStatus');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpc_ResultStatus(buffer_arg) {
  return rpc_pb.ResultStatus.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetAccountParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetAccountParam)) {
    throw new Error('Expected argument of type rpcquery.GetAccountParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetAccountParam(buffer_arg) {
  return rpcquery_pb.GetAccountParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetMetadataParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetMetadataParam)) {
    throw new Error('Expected argument of type rpcquery.GetMetadataParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetMetadataParam(buffer_arg) {
  return rpcquery_pb.GetMetadataParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetNameParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetNameParam)) {
    throw new Error('Expected argument of type rpcquery.GetNameParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetNameParam(buffer_arg) {
  return rpcquery_pb.GetNameParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetNetworkRegistryParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetNetworkRegistryParam)) {
    throw new Error('Expected argument of type rpcquery.GetNetworkRegistryParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetNetworkRegistryParam(buffer_arg) {
  return rpcquery_pb.GetNetworkRegistryParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetProposalParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetProposalParam)) {
    throw new Error('Expected argument of type rpcquery.GetProposalParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetProposalParam(buffer_arg) {
  return rpcquery_pb.GetProposalParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetStatsParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetStatsParam)) {
    throw new Error('Expected argument of type rpcquery.GetStatsParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetStatsParam(buffer_arg) {
  return rpcquery_pb.GetStatsParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetStorageParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetStorageParam)) {
    throw new Error('Expected argument of type rpcquery.GetStorageParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetStorageParam(buffer_arg) {
  return rpcquery_pb.GetStorageParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetTendermintBlockHeaderParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetTendermintBlockHeaderParam)) {
    throw new Error('Expected argument of type rpcquery.GetTendermintBlockHeaderParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetTendermintBlockHeaderParam(buffer_arg) {
  return rpcquery_pb.GetTendermintBlockHeaderParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetTendermintValidatorSetParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetTendermintValidatorSetParam)) {
    throw new Error('Expected argument of type rpcquery.GetTendermintValidatorSetParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetTendermintValidatorSetParam(buffer_arg) {
  return rpcquery_pb.GetTendermintValidatorSetParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetValidatorSetHistoryParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetValidatorSetHistoryParam)) {
    throw new Error('Expected argument of type rpcquery.GetValidatorSetHistoryParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetValidatorSetHistoryParam(buffer_arg) {
  return rpcquery_pb.GetValidatorSetHistoryParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_GetValidatorSetParam(arg) {
  if (!(arg instanceof rpcquery_pb.GetValidatorSetParam)) {
    throw new Error('Expected argument of type rpcquery.GetValidatorSetParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_GetValidatorSetParam(buffer_arg) {
  return rpcquery_pb.GetValidatorSetParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_ListAccountsParam(arg) {
  if (!(arg instanceof rpcquery_pb.ListAccountsParam)) {
    throw new Error('Expected argument of type rpcquery.ListAccountsParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_ListAccountsParam(buffer_arg) {
  return rpcquery_pb.ListAccountsParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_ListNamesParam(arg) {
  if (!(arg instanceof rpcquery_pb.ListNamesParam)) {
    throw new Error('Expected argument of type rpcquery.ListNamesParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_ListNamesParam(buffer_arg) {
  return rpcquery_pb.ListNamesParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_ListProposalsParam(arg) {
  if (!(arg instanceof rpcquery_pb.ListProposalsParam)) {
    throw new Error('Expected argument of type rpcquery.ListProposalsParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_ListProposalsParam(buffer_arg) {
  return rpcquery_pb.ListProposalsParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_MetadataResult(arg) {
  if (!(arg instanceof rpcquery_pb.MetadataResult)) {
    throw new Error('Expected argument of type rpcquery.MetadataResult');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_MetadataResult(buffer_arg) {
  return rpcquery_pb.MetadataResult.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_NetworkRegistry(arg) {
  if (!(arg instanceof rpcquery_pb.NetworkRegistry)) {
    throw new Error('Expected argument of type rpcquery.NetworkRegistry');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_NetworkRegistry(buffer_arg) {
  return rpcquery_pb.NetworkRegistry.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_ProposalResult(arg) {
  if (!(arg instanceof rpcquery_pb.ProposalResult)) {
    throw new Error('Expected argument of type rpcquery.ProposalResult');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_ProposalResult(buffer_arg) {
  return rpcquery_pb.ProposalResult.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_Stats(arg) {
  if (!(arg instanceof rpcquery_pb.Stats)) {
    throw new Error('Expected argument of type rpcquery.Stats');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_Stats(buffer_arg) {
  return rpcquery_pb.Stats.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_StatusParam(arg) {
  if (!(arg instanceof rpcquery_pb.StatusParam)) {
    throw new Error('Expected argument of type rpcquery.StatusParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_StatusParam(buffer_arg) {
  return rpcquery_pb.StatusParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_StorageValue(arg) {
  if (!(arg instanceof rpcquery_pb.StorageValue)) {
    throw new Error('Expected argument of type rpcquery.StorageValue');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_StorageValue(buffer_arg) {
  return rpcquery_pb.StorageValue.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_TendermintBlockHeader(arg) {
  if (!(arg instanceof rpcquery_pb.TendermintBlockHeader)) {
    throw new Error('Expected argument of type rpcquery.TendermintBlockHeader');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_TendermintBlockHeader(buffer_arg) {
  return rpcquery_pb.TendermintBlockHeader.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_TendermintValidatorSet(arg) {
  if (!(arg instanceof rpcquery_pb.TendermintValidatorSet)) {
    throw new Error('Expected argument of type rpcquery.TendermintValidatorSet');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_TendermintValidatorSet(buffer_arg) {
  return rpcquery_pb.TendermintValidatorSet.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_ValidatorSet(arg) {
  if (!(arg instanceof rpcquery_pb.ValidatorSet)) {
    throw new Error('Expected argument of type rpcquery.ValidatorSet');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_ValidatorSet(buffer_arg) {
  return rpcquery_pb.ValidatorSet.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcquery_ValidatorSetHistory(arg) {
  if (!(arg instanceof rpcquery_pb.ValidatorSetHistory)) {
    throw new Error('Expected argument of type rpcquery.ValidatorSetHistory');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcquery_ValidatorSetHistory(buffer_arg) {
  return rpcquery_pb.ValidatorSetHistory.deserializeBinary(new Uint8Array(buffer_arg));
}


var QueryService = exports.QueryService = {
  status: {
    path: '/rpcquery.Query/Status',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.StatusParam,
    responseType: rpc_pb.ResultStatus,
    requestSerialize: serialize_rpcquery_StatusParam,
    requestDeserialize: deserialize_rpcquery_StatusParam,
    responseSerialize: serialize_rpc_ResultStatus,
    responseDeserialize: deserialize_rpc_ResultStatus,
  },
  getAccount: {
    path: '/rpcquery.Query/GetAccount',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetAccountParam,
    responseType: acm_pb.Account,
    requestSerialize: serialize_rpcquery_GetAccountParam,
    requestDeserialize: deserialize_rpcquery_GetAccountParam,
    responseSerialize: serialize_acm_Account,
    responseDeserialize: deserialize_acm_Account,
  },
  getMetadata: {
    path: '/rpcquery.Query/GetMetadata',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetMetadataParam,
    responseType: rpcquery_pb.MetadataResult,
    requestSerialize: serialize_rpcquery_GetMetadataParam,
    requestDeserialize: deserialize_rpcquery_GetMetadataParam,
    responseSerialize: serialize_rpcquery_MetadataResult,
    responseDeserialize: deserialize_rpcquery_MetadataResult,
  },
  getStorage: {
    path: '/rpcquery.Query/GetStorage',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetStorageParam,
    responseType: rpcquery_pb.StorageValue,
    requestSerialize: serialize_rpcquery_GetStorageParam,
    requestDeserialize: deserialize_rpcquery_GetStorageParam,
    responseSerialize: serialize_rpcquery_StorageValue,
    responseDeserialize: deserialize_rpcquery_StorageValue,
  },
  listAccounts: {
    path: '/rpcquery.Query/ListAccounts',
    requestStream: false,
    responseStream: true,
    requestType: rpcquery_pb.ListAccountsParam,
    responseType: acm_pb.Account,
    requestSerialize: serialize_rpcquery_ListAccountsParam,
    requestDeserialize: deserialize_rpcquery_ListAccountsParam,
    responseSerialize: serialize_acm_Account,
    responseDeserialize: deserialize_acm_Account,
  },
  getName: {
    path: '/rpcquery.Query/GetName',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetNameParam,
    responseType: names_pb.Entry,
    requestSerialize: serialize_rpcquery_GetNameParam,
    requestDeserialize: deserialize_rpcquery_GetNameParam,
    responseSerialize: serialize_names_Entry,
    responseDeserialize: deserialize_names_Entry,
  },
  listNames: {
    path: '/rpcquery.Query/ListNames',
    requestStream: false,
    responseStream: true,
    requestType: rpcquery_pb.ListNamesParam,
    responseType: names_pb.Entry,
    requestSerialize: serialize_rpcquery_ListNamesParam,
    requestDeserialize: deserialize_rpcquery_ListNamesParam,
    responseSerialize: serialize_names_Entry,
    responseDeserialize: deserialize_names_Entry,
  },
  // GetNetworkRegistry returns for each validator address, the list of their identified node at the current state
getNetworkRegistry: {
    path: '/rpcquery.Query/GetNetworkRegistry',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetNetworkRegistryParam,
    responseType: rpcquery_pb.NetworkRegistry,
    requestSerialize: serialize_rpcquery_GetNetworkRegistryParam,
    requestDeserialize: deserialize_rpcquery_GetNetworkRegistryParam,
    responseSerialize: serialize_rpcquery_NetworkRegistry,
    responseDeserialize: deserialize_rpcquery_NetworkRegistry,
  },
  getValidatorSet: {
    path: '/rpcquery.Query/GetValidatorSet',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetValidatorSetParam,
    responseType: rpcquery_pb.ValidatorSet,
    requestSerialize: serialize_rpcquery_GetValidatorSetParam,
    requestDeserialize: deserialize_rpcquery_GetValidatorSetParam,
    responseSerialize: serialize_rpcquery_ValidatorSet,
    responseDeserialize: deserialize_rpcquery_ValidatorSet,
  },
  getValidatorSetHistory: {
    path: '/rpcquery.Query/GetValidatorSetHistory',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetValidatorSetHistoryParam,
    responseType: rpcquery_pb.ValidatorSetHistory,
    requestSerialize: serialize_rpcquery_GetValidatorSetHistoryParam,
    requestDeserialize: deserialize_rpcquery_GetValidatorSetHistoryParam,
    responseSerialize: serialize_rpcquery_ValidatorSetHistory,
    responseDeserialize: deserialize_rpcquery_ValidatorSetHistory,
  },
  getProposal: {
    path: '/rpcquery.Query/GetProposal',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetProposalParam,
    responseType: payload_pb.Ballot,
    requestSerialize: serialize_rpcquery_GetProposalParam,
    requestDeserialize: deserialize_rpcquery_GetProposalParam,
    responseSerialize: serialize_payload_Ballot,
    responseDeserialize: deserialize_payload_Ballot,
  },
  listProposals: {
    path: '/rpcquery.Query/ListProposals',
    requestStream: false,
    responseStream: true,
    requestType: rpcquery_pb.ListProposalsParam,
    responseType: rpcquery_pb.ProposalResult,
    requestSerialize: serialize_rpcquery_ListProposalsParam,
    requestDeserialize: deserialize_rpcquery_ListProposalsParam,
    responseSerialize: serialize_rpcquery_ProposalResult,
    responseDeserialize: deserialize_rpcquery_ProposalResult,
  },
  getStats: {
    path: '/rpcquery.Query/GetStats',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetStatsParam,
    responseType: rpcquery_pb.Stats,
    requestSerialize: serialize_rpcquery_GetStatsParam,
    requestDeserialize: deserialize_rpcquery_GetStatsParam,
    responseSerialize: serialize_rpcquery_Stats,
    responseDeserialize: deserialize_rpcquery_Stats,
  },
  getTendermintBlockHeader: {
    path: '/rpcquery.Query/GetTendermintBlockHeader',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetTendermintBlockHeaderParam,
    responseType: rpcquery_pb.TendermintBlockHeader,
    requestSerialize: serialize_rpcquery_GetTendermintBlockHeaderParam,
    requestDeserialize: deserialize_rpcquery_GetTendermintBlockHeaderParam,
    responseSerialize: serialize_rpcquery_TendermintBlockHeader,
    responseDeserialize: deserialize_rpcquery_TendermintBlockHeader,
  },
  getTendermintValidatorSet: {
    path: '/rpcquery.Query/GetTendermintValidatorSet',
    requestStream: false,
    responseStream: false,
    requestType: rpcquery_pb.GetTendermintValidatorSetParam,
    responseType: rpcquery_pb.TendermintValidatorSet,
    requestSerialize: serialize_rpcquery_GetTendermintValidatorSetParam,
    requestDeserialize: deserialize_rpcquery_GetTendermintValidatorSetParam,
    responseSerialize: serialize_rpcquery_TendermintValidatorSet,
    responseDeserialize: deserialize_rpcquery_TendermintValidatorSet,
  },
};

exports.QueryClient = grpc.makeGenericClientConstructor(QueryService);
