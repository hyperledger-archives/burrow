// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var tendermint_rpc_grpc_types_pb = require('../../../tendermint/rpc/grpc/types_pb.js');
var tendermint_abci_types_pb = require('../../../tendermint/abci/types_pb.js');

function serialize_tendermint_rpc_grpc_RequestBroadcastTx(arg) {
  if (!(arg instanceof tendermint_rpc_grpc_types_pb.RequestBroadcastTx)) {
    throw new Error('Expected argument of type tendermint.rpc.grpc.RequestBroadcastTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_rpc_grpc_RequestBroadcastTx(buffer_arg) {
  return tendermint_rpc_grpc_types_pb.RequestBroadcastTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_rpc_grpc_RequestPing(arg) {
  if (!(arg instanceof tendermint_rpc_grpc_types_pb.RequestPing)) {
    throw new Error('Expected argument of type tendermint.rpc.grpc.RequestPing');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_rpc_grpc_RequestPing(buffer_arg) {
  return tendermint_rpc_grpc_types_pb.RequestPing.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_rpc_grpc_ResponseBroadcastTx(arg) {
  if (!(arg instanceof tendermint_rpc_grpc_types_pb.ResponseBroadcastTx)) {
    throw new Error('Expected argument of type tendermint.rpc.grpc.ResponseBroadcastTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_rpc_grpc_ResponseBroadcastTx(buffer_arg) {
  return tendermint_rpc_grpc_types_pb.ResponseBroadcastTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_tendermint_rpc_grpc_ResponsePing(arg) {
  if (!(arg instanceof tendermint_rpc_grpc_types_pb.ResponsePing)) {
    throw new Error('Expected argument of type tendermint.rpc.grpc.ResponsePing');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_tendermint_rpc_grpc_ResponsePing(buffer_arg) {
  return tendermint_rpc_grpc_types_pb.ResponsePing.deserializeBinary(new Uint8Array(buffer_arg));
}


// ----------------------------------------
// Service Definition
//
var BroadcastAPIService = exports.BroadcastAPIService = {
  ping: {
    path: '/tendermint.rpc.grpc.BroadcastAPI/Ping',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_rpc_grpc_types_pb.RequestPing,
    responseType: tendermint_rpc_grpc_types_pb.ResponsePing,
    requestSerialize: serialize_tendermint_rpc_grpc_RequestPing,
    requestDeserialize: deserialize_tendermint_rpc_grpc_RequestPing,
    responseSerialize: serialize_tendermint_rpc_grpc_ResponsePing,
    responseDeserialize: deserialize_tendermint_rpc_grpc_ResponsePing,
  },
  broadcastTx: {
    path: '/tendermint.rpc.grpc.BroadcastAPI/BroadcastTx',
    requestStream: false,
    responseStream: false,
    requestType: tendermint_rpc_grpc_types_pb.RequestBroadcastTx,
    responseType: tendermint_rpc_grpc_types_pb.ResponseBroadcastTx,
    requestSerialize: serialize_tendermint_rpc_grpc_RequestBroadcastTx,
    requestDeserialize: deserialize_tendermint_rpc_grpc_RequestBroadcastTx,
    responseSerialize: serialize_tendermint_rpc_grpc_ResponseBroadcastTx,
    responseDeserialize: deserialize_tendermint_rpc_grpc_ResponseBroadcastTx,
  },
};

exports.BroadcastAPIClient = grpc.makeGenericClientConstructor(BroadcastAPIService);
