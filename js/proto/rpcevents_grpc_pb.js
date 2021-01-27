// GENERATED CODE -- DO NOT EDIT!

'use strict';
var rpcevents_pb = require('./rpcevents_pb.js');
var gogoproto_gogo_pb = require('./gogoproto/gogo_pb.js');
var exec_pb = require('./exec_pb.js');

function serialize_exec_StreamEvent(arg) {
  if (!(arg instanceof exec_pb.StreamEvent)) {
    throw new Error('Expected argument of type exec.StreamEvent');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_exec_StreamEvent(buffer_arg) {
  return exec_pb.StreamEvent.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_exec_TxExecution(arg) {
  if (!(arg instanceof exec_pb.TxExecution)) {
    throw new Error('Expected argument of type exec.TxExecution');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_exec_TxExecution(buffer_arg) {
  return exec_pb.TxExecution.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcevents_BlocksRequest(arg) {
  if (!(arg instanceof rpcevents_pb.BlocksRequest)) {
    throw new Error('Expected argument of type rpcevents.BlocksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcevents_BlocksRequest(buffer_arg) {
  return rpcevents_pb.BlocksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcevents_EventsResponse(arg) {
  if (!(arg instanceof rpcevents_pb.EventsResponse)) {
    throw new Error('Expected argument of type rpcevents.EventsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcevents_EventsResponse(buffer_arg) {
  return rpcevents_pb.EventsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcevents_TxRequest(arg) {
  if (!(arg instanceof rpcevents_pb.TxRequest)) {
    throw new Error('Expected argument of type rpcevents.TxRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcevents_TxRequest(buffer_arg) {
  return rpcevents_pb.TxRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// --------------------------------------------------
// Execution events
var ExecutionEventsService = exports['rpcevents.ExecutionEvents'] = {
  // Get StreamEvents (including transactions) for a range of block heights
stream: {
    path: '/rpcevents.ExecutionEvents/Stream',
    requestStream: false,
    responseStream: true,
    requestType: rpcevents_pb.BlocksRequest,
    responseType: exec_pb.StreamEvent,
    requestSerialize: serialize_rpcevents_BlocksRequest,
    requestDeserialize: deserialize_rpcevents_BlocksRequest,
    responseSerialize: serialize_exec_StreamEvent,
    responseDeserialize: deserialize_exec_StreamEvent,
  },
  // Get a particular TxExecution by hash
tx: {
    path: '/rpcevents.ExecutionEvents/Tx',
    requestStream: false,
    responseStream: false,
    requestType: rpcevents_pb.TxRequest,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_rpcevents_TxRequest,
    requestDeserialize: deserialize_rpcevents_TxRequest,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // GetEvents provides events streaming one block at a time - that is all events emitted in a particular block
// are guaranteed to be delivered in each GetEventsResponse
events: {
    path: '/rpcevents.ExecutionEvents/Events',
    requestStream: false,
    responseStream: true,
    requestType: rpcevents_pb.BlocksRequest,
    responseType: rpcevents_pb.EventsResponse,
    requestSerialize: serialize_rpcevents_BlocksRequest,
    requestDeserialize: deserialize_rpcevents_BlocksRequest,
    responseSerialize: serialize_rpcevents_EventsResponse,
    responseDeserialize: deserialize_rpcevents_EventsResponse,
  },
};

