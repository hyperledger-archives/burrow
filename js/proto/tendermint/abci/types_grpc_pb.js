// GENERATED CODE -- DO NOT EDIT!

'use strict';
var tendermint_abci_types_pb = require('../../tendermint/abci/types_pb.js');
var tendermint_crypto_proof_pb = require('../../tendermint/crypto/proof_pb.js');
var tendermint_types_types_pb = require('../../tendermint/types/types_pb.js');
var tendermint_crypto_keys_pb = require('../../tendermint/crypto/keys_pb.js');
var tendermint_types_params_pb = require('../../tendermint/types/params_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var gogoproto_gogo_pb = require('../../gogoproto/gogo_pb.js');

function serialize_tendermint_abci_RequestApplySnapshotChunk(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestApplySnapshotChunk)) {
    throw new Error('Expected argument of type tendermint.abci.RequestApplySnapshotChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestApplySnapshotChunk(buffer_arg) {
  return tendermint_abci_types_pb.RequestApplySnapshotChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestBeginBlock(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestBeginBlock)) {
    throw new Error('Expected argument of type tendermint.abci.RequestBeginBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestBeginBlock(buffer_arg) {
  return tendermint_abci_types_pb.RequestBeginBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestCheckTx(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestCheckTx)) {
    throw new Error('Expected argument of type tendermint.abci.RequestCheckTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestCheckTx(buffer_arg) {
  return tendermint_abci_types_pb.RequestCheckTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestCommit(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestCommit)) {
    throw new Error('Expected argument of type tendermint.abci.RequestCommit');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestCommit(buffer_arg) {
  return tendermint_abci_types_pb.RequestCommit.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestDeliverTx(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestDeliverTx)) {
    throw new Error('Expected argument of type tendermint.abci.RequestDeliverTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestDeliverTx(buffer_arg) {
  return tendermint_abci_types_pb.RequestDeliverTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestEcho(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestEcho)) {
    throw new Error('Expected argument of type tendermint.abci.RequestEcho');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestEcho(buffer_arg) {
  return tendermint_abci_types_pb.RequestEcho.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestEndBlock(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestEndBlock)) {
    throw new Error('Expected argument of type tendermint.abci.RequestEndBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestEndBlock(buffer_arg) {
  return tendermint_abci_types_pb.RequestEndBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestFlush(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestFlush)) {
    throw new Error('Expected argument of type tendermint.abci.RequestFlush');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestFlush(buffer_arg) {
  return tendermint_abci_types_pb.RequestFlush.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestInfo(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestInfo)) {
    throw new Error('Expected argument of type tendermint.abci.RequestInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestInfo(buffer_arg) {
  return tendermint_abci_types_pb.RequestInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestInitChain(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestInitChain)) {
    throw new Error('Expected argument of type tendermint.abci.RequestInitChain');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestInitChain(buffer_arg) {
  return tendermint_abci_types_pb.RequestInitChain.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestListSnapshots(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestListSnapshots)) {
    throw new Error('Expected argument of type tendermint.abci.RequestListSnapshots');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestListSnapshots(buffer_arg) {
  return tendermint_abci_types_pb.RequestListSnapshots.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestLoadSnapshotChunk(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestLoadSnapshotChunk)) {
    throw new Error('Expected argument of type tendermint.abci.RequestLoadSnapshotChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestLoadSnapshotChunk(buffer_arg) {
  return tendermint_abci_types_pb.RequestLoadSnapshotChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestOfferSnapshot(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestOfferSnapshot)) {
    throw new Error('Expected argument of type tendermint.abci.RequestOfferSnapshot');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestOfferSnapshot(buffer_arg) {
  return tendermint_abci_types_pb.RequestOfferSnapshot.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestQuery(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestQuery)) {
    throw new Error('Expected argument of type tendermint.abci.RequestQuery');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestQuery(buffer_arg) {
  return tendermint_abci_types_pb.RequestQuery.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_RequestSetOption(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.RequestSetOption)) {
    throw new Error('Expected argument of type tendermint.abci.RequestSetOption');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_RequestSetOption(buffer_arg) {
  return tendermint_abci_types_pb.RequestSetOption.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseApplySnapshotChunk(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseApplySnapshotChunk)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseApplySnapshotChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseApplySnapshotChunk(buffer_arg) {
  return tendermint_abci_types_pb.ResponseApplySnapshotChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseBeginBlock(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseBeginBlock)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseBeginBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseBeginBlock(buffer_arg) {
  return tendermint_abci_types_pb.ResponseBeginBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseCheckTx(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseCheckTx)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseCheckTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseCheckTx(buffer_arg) {
  return tendermint_abci_types_pb.ResponseCheckTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseCommit(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseCommit)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseCommit');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseCommit(buffer_arg) {
  return tendermint_abci_types_pb.ResponseCommit.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseDeliverTx(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseDeliverTx)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseDeliverTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseDeliverTx(buffer_arg) {
  return tendermint_abci_types_pb.ResponseDeliverTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseEcho(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseEcho)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseEcho');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseEcho(buffer_arg) {
  return tendermint_abci_types_pb.ResponseEcho.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseEndBlock(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseEndBlock)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseEndBlock');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseEndBlock(buffer_arg) {
  return tendermint_abci_types_pb.ResponseEndBlock.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseFlush(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseFlush)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseFlush');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseFlush(buffer_arg) {
  return tendermint_abci_types_pb.ResponseFlush.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseInfo(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseInfo)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseInfo(buffer_arg) {
  return tendermint_abci_types_pb.ResponseInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseInitChain(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseInitChain)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseInitChain');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseInitChain(buffer_arg) {
  return tendermint_abci_types_pb.ResponseInitChain.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseListSnapshots(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseListSnapshots)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseListSnapshots');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseListSnapshots(buffer_arg) {
  return tendermint_abci_types_pb.ResponseListSnapshots.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseLoadSnapshotChunk(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseLoadSnapshotChunk)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseLoadSnapshotChunk');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseLoadSnapshotChunk(buffer_arg) {
  return tendermint_abci_types_pb.ResponseLoadSnapshotChunk.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseOfferSnapshot(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseOfferSnapshot)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseOfferSnapshot');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseOfferSnapshot(buffer_arg) {
  return tendermint_abci_types_pb.ResponseOfferSnapshot.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseQuery(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseQuery)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseQuery');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseQuery(buffer_arg) {
  return tendermint_abci_types_pb.ResponseQuery.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_abci_ResponseSetOption(arg) {
  if (!(arg instanceof tendermint_abci_types_pb.ResponseSetOption)) {
    throw new Error('Expected argument of type tendermint.abci.ResponseSetOption');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_abci_ResponseSetOption(buffer_arg) {
  return tendermint_abci_types_pb.ResponseSetOption.deserializeBinary(new Uint8Array(buffer_arg));
}


// ----------------------------------------
// Service Definition
//
var ABCIApplicationService = exports['tendermint.abci.ABCIApplication'] = {
  echo: {
    path: '/tendermint.abci.ABCIApplication/Echo',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestEcho,
    responseType: tendermint_abci_types_pb.ResponseEcho,
    requestSerialize: serialize_tendermint_abci_RequestEcho,
    requestDeserialize: deserialize_tendermint_abci_RequestEcho,
    responseSerialize: serialize_tendermint_abci_ResponseEcho,
    responseDeserialize: deserialize_tendermint_abci_ResponseEcho,
  },
  flush: {
    path: '/tendermint.abci.ABCIApplication/Flush',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestFlush,
    responseType: tendermint_abci_types_pb.ResponseFlush,
    requestSerialize: serialize_tendermint_abci_RequestFlush,
    requestDeserialize: deserialize_tendermint_abci_RequestFlush,
    responseSerialize: serialize_tendermint_abci_ResponseFlush,
    responseDeserialize: deserialize_tendermint_abci_ResponseFlush,
  },
  info: {
    path: '/tendermint.abci.ABCIApplication/Info',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestInfo,
    responseType: tendermint_abci_types_pb.ResponseInfo,
    requestSerialize: serialize_tendermint_abci_RequestInfo,
    requestDeserialize: deserialize_tendermint_abci_RequestInfo,
    responseSerialize: serialize_tendermint_abci_ResponseInfo,
    responseDeserialize: deserialize_tendermint_abci_ResponseInfo,
  },
  setOption: {
    path: '/tendermint.abci.ABCIApplication/SetOption',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestSetOption,
    responseType: tendermint_abci_types_pb.ResponseSetOption,
    requestSerialize: serialize_tendermint_abci_RequestSetOption,
    requestDeserialize: deserialize_tendermint_abci_RequestSetOption,
    responseSerialize: serialize_tendermint_abci_ResponseSetOption,
    responseDeserialize: deserialize_tendermint_abci_ResponseSetOption,
  },
  deliverTx: {
    path: '/tendermint.abci.ABCIApplication/DeliverTx',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestDeliverTx,
    responseType: tendermint_abci_types_pb.ResponseDeliverTx,
    requestSerialize: serialize_tendermint_abci_RequestDeliverTx,
    requestDeserialize: deserialize_tendermint_abci_RequestDeliverTx,
    responseSerialize: serialize_tendermint_abci_ResponseDeliverTx,
    responseDeserialize: deserialize_tendermint_abci_ResponseDeliverTx,
  },
  checkTx: {
    path: '/tendermint.abci.ABCIApplication/CheckTx',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestCheckTx,
    responseType: tendermint_abci_types_pb.ResponseCheckTx,
    requestSerialize: serialize_tendermint_abci_RequestCheckTx,
    requestDeserialize: deserialize_tendermint_abci_RequestCheckTx,
    responseSerialize: serialize_tendermint_abci_ResponseCheckTx,
    responseDeserialize: deserialize_tendermint_abci_ResponseCheckTx,
  },
  query: {
    path: '/tendermint.abci.ABCIApplication/Query',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestQuery,
    responseType: tendermint_abci_types_pb.ResponseQuery,
    requestSerialize: serialize_tendermint_abci_RequestQuery,
    requestDeserialize: deserialize_tendermint_abci_RequestQuery,
    responseSerialize: serialize_tendermint_abci_ResponseQuery,
    responseDeserialize: deserialize_tendermint_abci_ResponseQuery,
  },
  commit: {
    path: '/tendermint.abci.ABCIApplication/Commit',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestCommit,
    responseType: tendermint_abci_types_pb.ResponseCommit,
    requestSerialize: serialize_tendermint_abci_RequestCommit,
    requestDeserialize: deserialize_tendermint_abci_RequestCommit,
    responseSerialize: serialize_tendermint_abci_ResponseCommit,
    responseDeserialize: deserialize_tendermint_abci_ResponseCommit,
  },
  initChain: {
    path: '/tendermint.abci.ABCIApplication/InitChain',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestInitChain,
    responseType: tendermint_abci_types_pb.ResponseInitChain,
    requestSerialize: serialize_tendermint_abci_RequestInitChain,
    requestDeserialize: deserialize_tendermint_abci_RequestInitChain,
    responseSerialize: serialize_tendermint_abci_ResponseInitChain,
    responseDeserialize: deserialize_tendermint_abci_ResponseInitChain,
  },
  beginBlock: {
    path: '/tendermint.abci.ABCIApplication/BeginBlock',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestBeginBlock,
    responseType: tendermint_abci_types_pb.ResponseBeginBlock,
    requestSerialize: serialize_tendermint_abci_RequestBeginBlock,
    requestDeserialize: deserialize_tendermint_abci_RequestBeginBlock,
    responseSerialize: serialize_tendermint_abci_ResponseBeginBlock,
    responseDeserialize: deserialize_tendermint_abci_ResponseBeginBlock,
  },
  endBlock: {
    path: '/tendermint.abci.ABCIApplication/EndBlock',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestEndBlock,
    responseType: tendermint_abci_types_pb.ResponseEndBlock,
    requestSerialize: serialize_tendermint_abci_RequestEndBlock,
    requestDeserialize: deserialize_tendermint_abci_RequestEndBlock,
    responseSerialize: serialize_tendermint_abci_ResponseEndBlock,
    responseDeserialize: deserialize_tendermint_abci_ResponseEndBlock,
  },
  listSnapshots: {
    path: '/tendermint.abci.ABCIApplication/ListSnapshots',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestListSnapshots,
    responseType: tendermint_abci_types_pb.ResponseListSnapshots,
    requestSerialize: serialize_tendermint_abci_RequestListSnapshots,
    requestDeserialize: deserialize_tendermint_abci_RequestListSnapshots,
    responseSerialize: serialize_tendermint_abci_ResponseListSnapshots,
    responseDeserialize: deserialize_tendermint_abci_ResponseListSnapshots,
  },
  offerSnapshot: {
    path: '/tendermint.abci.ABCIApplication/OfferSnapshot',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestOfferSnapshot,
    responseType: tendermint_abci_types_pb.ResponseOfferSnapshot,
    requestSerialize: serialize_tendermint_abci_RequestOfferSnapshot,
    requestDeserialize: deserialize_tendermint_abci_RequestOfferSnapshot,
    responseSerialize: serialize_tendermint_abci_ResponseOfferSnapshot,
    responseDeserialize: deserialize_tendermint_abci_ResponseOfferSnapshot,
  },
  loadSnapshotChunk: {
    path: '/tendermint.abci.ABCIApplication/LoadSnapshotChunk',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestLoadSnapshotChunk,
    responseType: tendermint_abci_types_pb.ResponseLoadSnapshotChunk,
    requestSerialize: serialize_tendermint_abci_RequestLoadSnapshotChunk,
    requestDeserialize: deserialize_tendermint_abci_RequestLoadSnapshotChunk,
    responseSerialize: serialize_tendermint_abci_ResponseLoadSnapshotChunk,
    responseDeserialize: deserialize_tendermint_abci_ResponseLoadSnapshotChunk,
  },
  applySnapshotChunk: {
    path: '/tendermint.abci.ABCIApplication/ApplySnapshotChunk',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_abci_types_pb.RequestApplySnapshotChunk,
    responseType: tendermint_abci_types_pb.ResponseApplySnapshotChunk,
    requestSerialize: serialize_tendermint_abci_RequestApplySnapshotChunk,
    requestDeserialize: deserialize_tendermint_abci_RequestApplySnapshotChunk,
    responseSerialize: serialize_tendermint_abci_ResponseApplySnapshotChunk,
    responseDeserialize: deserialize_tendermint_abci_ResponseApplySnapshotChunk,
  },
};

