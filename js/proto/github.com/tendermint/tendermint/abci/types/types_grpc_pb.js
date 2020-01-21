// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var github_com_tendermint_tendermint_abci_types_types_pb = require('../../../../../github.com/tendermint/tendermint/abci/types/types_pb.js');
var github_com_gogo_protobuf_gogoproto_gogo_pb = require('../../../../../github.com/gogo/protobuf/gogoproto/gogo_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var github_com_tendermint_tendermint_libs_common_types_pb = require('../../../../../github.com/tendermint/tendermint/libs/common/types_pb.js');

function serialize_types_RequestBeginBlock(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock)) {
    throw new Error('Expected argument of type types.RequestBeginBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestBeginBlock(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestCheckTx(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx)) {
    throw new Error('Expected argument of type types.RequestCheckTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestCheckTx(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestCommit(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit)) {
    throw new Error('Expected argument of type types.RequestCommit');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestCommit(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestDeliverTx(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx)) {
    throw new Error('Expected argument of type types.RequestDeliverTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestDeliverTx(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestEcho(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho)) {
    throw new Error('Expected argument of type types.RequestEcho');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestEcho(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestEndBlock(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock)) {
    throw new Error('Expected argument of type types.RequestEndBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestEndBlock(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestFlush(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush)) {
    throw new Error('Expected argument of type types.RequestFlush');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestFlush(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestInfo(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo)) {
    throw new Error('Expected argument of type types.RequestInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestInfo(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestInitChain(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain)) {
    throw new Error('Expected argument of type types.RequestInitChain');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestInitChain(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestQuery(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery)) {
    throw new Error('Expected argument of type types.RequestQuery');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestQuery(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_RequestSetOption(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption)) {
    throw new Error('Expected argument of type types.RequestSetOption');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_RequestSetOption(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseBeginBlock(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock)) {
    throw new Error('Expected argument of type types.ResponseBeginBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseBeginBlock(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseCheckTx(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx)) {
    throw new Error('Expected argument of type types.ResponseCheckTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseCheckTx(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseCommit(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit)) {
    throw new Error('Expected argument of type types.ResponseCommit');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseCommit(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseDeliverTx(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx)) {
    throw new Error('Expected argument of type types.ResponseDeliverTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseDeliverTx(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseEcho(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho)) {
    throw new Error('Expected argument of type types.ResponseEcho');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseEcho(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseEndBlock(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock)) {
    throw new Error('Expected argument of type types.ResponseEndBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseEndBlock(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseFlush(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush)) {
    throw new Error('Expected argument of type types.ResponseFlush');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseFlush(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseInfo(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo)) {
    throw new Error('Expected argument of type types.ResponseInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseInfo(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseInitChain(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain)) {
    throw new Error('Expected argument of type types.ResponseInitChain');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseInitChain(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseQuery(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery)) {
    throw new Error('Expected argument of type types.ResponseQuery');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseQuery(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_types_ResponseSetOption(arg) {
  if (!(arg instanceof github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption)) {
    throw new Error('Expected argument of type types.ResponseSetOption');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_types_ResponseSetOption(buffer_arg) {
  return github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption.deserializeBinary(new Uint8Array(buffer_arg));
}


// ----------------------------------------
// Service Definition
//
var ABCIApplicationService = exports.ABCIApplicationService = {
  echo: {
    path: '/types.ABCIApplication/Echo',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestEcho,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEcho,
    requestSerialize: serialize_types_RequestEcho,
    requestDeserialize: deserialize_types_RequestEcho,
    responseSerialize: serialize_types_ResponseEcho,
    responseDeserialize: deserialize_types_ResponseEcho,
  },
  flush: {
    path: '/types.ABCIApplication/Flush',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestFlush,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseFlush,
    requestSerialize: serialize_types_RequestFlush,
    requestDeserialize: deserialize_types_RequestFlush,
    responseSerialize: serialize_types_ResponseFlush,
    responseDeserialize: deserialize_types_ResponseFlush,
  },
  info: {
    path: '/types.ABCIApplication/Info',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestInfo,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInfo,
    requestSerialize: serialize_types_RequestInfo,
    requestDeserialize: deserialize_types_RequestInfo,
    responseSerialize: serialize_types_ResponseInfo,
    responseDeserialize: deserialize_types_ResponseInfo,
  },
  setOption: {
    path: '/types.ABCIApplication/SetOption',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestSetOption,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseSetOption,
    requestSerialize: serialize_types_RequestSetOption,
    requestDeserialize: deserialize_types_RequestSetOption,
    responseSerialize: serialize_types_ResponseSetOption,
    responseDeserialize: deserialize_types_ResponseSetOption,
  },
  deliverTx: {
    path: '/types.ABCIApplication/DeliverTx',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestDeliverTx,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseDeliverTx,
    requestSerialize: serialize_types_RequestDeliverTx,
    requestDeserialize: deserialize_types_RequestDeliverTx,
    responseSerialize: serialize_types_ResponseDeliverTx,
    responseDeserialize: deserialize_types_ResponseDeliverTx,
  },
  checkTx: {
    path: '/types.ABCIApplication/CheckTx',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestCheckTx,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCheckTx,
    requestSerialize: serialize_types_RequestCheckTx,
    requestDeserialize: deserialize_types_RequestCheckTx,
    responseSerialize: serialize_types_ResponseCheckTx,
    responseDeserialize: deserialize_types_ResponseCheckTx,
  },
  query: {
    path: '/types.ABCIApplication/Query',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestQuery,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseQuery,
    requestSerialize: serialize_types_RequestQuery,
    requestDeserialize: deserialize_types_RequestQuery,
    responseSerialize: serialize_types_ResponseQuery,
    responseDeserialize: deserialize_types_ResponseQuery,
  },
  commit: {
    path: '/types.ABCIApplication/Commit',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestCommit,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseCommit,
    requestSerialize: serialize_types_RequestCommit,
    requestDeserialize: deserialize_types_RequestCommit,
    responseSerialize: serialize_types_ResponseCommit,
    responseDeserialize: deserialize_types_ResponseCommit,
  },
  initChain: {
    path: '/types.ABCIApplication/InitChain',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestInitChain,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseInitChain,
    requestSerialize: serialize_types_RequestInitChain,
    requestDeserialize: deserialize_types_RequestInitChain,
    responseSerialize: serialize_types_ResponseInitChain,
    responseDeserialize: deserialize_types_ResponseInitChain,
  },
  beginBlock: {
    path: '/types.ABCIApplication/BeginBlock',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestBeginBlock,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseBeginBlock,
    requestSerialize: serialize_types_RequestBeginBlock,
    requestDeserialize: deserialize_types_RequestBeginBlock,
    responseSerialize: serialize_types_ResponseBeginBlock,
    responseDeserialize: deserialize_types_ResponseBeginBlock,
  },
  endBlock: {
    path: '/types.ABCIApplication/EndBlock',
    requestStream: false,
    responseStream: false,
    requestType: github_com_tendermint_tendermint_abci_types_types_pb.RequestEndBlock,
    responseType: github_com_tendermint_tendermint_abci_types_types_pb.ResponseEndBlock,
    requestSerialize: serialize_types_RequestEndBlock,
    requestDeserialize: deserialize_types_RequestEndBlock,
    responseSerialize: serialize_types_ResponseEndBlock,
    responseDeserialize: deserialize_types_ResponseEndBlock,
  },
};

exports.ABCIApplicationClient = grpc.makeGenericClientConstructor(ABCIApplicationService);
