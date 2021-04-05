// GENERATED CODE -- DO NOT EDIT!

'use strict';
var rpctransact_pb = require('./rpctransact_pb.js');
var gogoproto_gogo_pb = require('./gogoproto/gogo_pb.js');
var google_protobuf_duration_pb = require('google-protobuf/google/protobuf/duration_pb.js');
var exec_pb = require('./exec_pb.js');
var payload_pb = require('./payload_pb.js');
var txs_pb = require('./txs_pb.js');

function serialize_exec_TxExecution(arg) {
  if (!(arg instanceof exec_pb.TxExecution)) {
    throw new Error('Expected argument of type exec.TxExecution');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_exec_TxExecution(buffer_arg) {
  return exec_pb.TxExecution.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_payload_Any(arg) {
  if (!(arg instanceof payload_pb.Any)) {
    throw new Error('Expected argument of type payload.Any');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_payload_Any(buffer_arg) {
  return payload_pb.Any.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_payload_CallTx(arg) {
  if (!(arg instanceof payload_pb.CallTx)) {
    throw new Error('Expected argument of type payload.CallTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_payload_CallTx(buffer_arg) {
  return payload_pb.CallTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_payload_NameTx(arg) {
  if (!(arg instanceof payload_pb.NameTx)) {
    throw new Error('Expected argument of type payload.NameTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_payload_NameTx(buffer_arg) {
  return payload_pb.NameTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_payload_SendTx(arg) {
  if (!(arg instanceof payload_pb.SendTx)) {
    throw new Error('Expected argument of type payload.SendTx');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_payload_SendTx(buffer_arg) {
  return payload_pb.SendTx.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpctransact_CallCodeParam(arg) {
  if (!(arg instanceof rpctransact_pb.CallCodeParam)) {
    throw new Error('Expected argument of type rpctransact.CallCodeParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpctransact_CallCodeParam(buffer_arg) {
  return rpctransact_pb.CallCodeParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpctransact_TxEnvelope(arg) {
  if (!(arg instanceof rpctransact_pb.TxEnvelope)) {
    throw new Error('Expected argument of type rpctransact.TxEnvelope');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpctransact_TxEnvelope(buffer_arg) {
  return rpctransact_pb.TxEnvelope.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpctransact_TxEnvelopeParam(arg) {
  if (!(arg instanceof rpctransact_pb.TxEnvelopeParam)) {
    throw new Error('Expected argument of type rpctransact.TxEnvelopeParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpctransact_TxEnvelopeParam(buffer_arg) {
  return rpctransact_pb.TxEnvelopeParam.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_txs_Receipt(arg) {
  if (!(arg instanceof txs_pb.Receipt)) {
    throw new Error('Expected argument of type txs.Receipt');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_txs_Receipt(buffer_arg) {
  return txs_pb.Receipt.deserializeBinary(new Uint8Array(buffer_arg));
}


// Transaction Service Definition
var TransactService = exports['rpctransact.Transact'] = {
  // Broadcast a transaction to the mempool - if the transaction is not signed signing will be attempted server-side
// and wait for it to be included in block
broadcastTxSync: {
    path: '/rpctransact.Transact/BroadcastTxSync',
    requestStream: false,
    responseStream: false,
    requestType: rpctransact_pb.TxEnvelopeParam,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_rpctransact_TxEnvelopeParam,
    requestDeserialize: deserialize_rpctransact_TxEnvelopeParam,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // Broadcast a transaction to the mempool - if the transaction is not signed signing will be attempted server-side
broadcastTxAsync: {
    path: '/rpctransact.Transact/BroadcastTxAsync',
    requestStream: false,
    responseStream: false,
    requestType: rpctransact_pb.TxEnvelopeParam,
    responseType: txs_pb.Receipt,
    requestSerialize: serialize_rpctransact_TxEnvelopeParam,
    requestDeserialize: deserialize_rpctransact_TxEnvelopeParam,
    responseSerialize: serialize_txs_Receipt,
    responseDeserialize: deserialize_txs_Receipt,
  },
  // Sign transaction server-side
signTx: {
    path: '/rpctransact.Transact/SignTx',
    requestStream: false,
    responseStream: false,
    requestType: rpctransact_pb.TxEnvelopeParam,
    responseType: rpctransact_pb.TxEnvelope,
    requestSerialize: serialize_rpctransact_TxEnvelopeParam,
    requestDeserialize: deserialize_rpctransact_TxEnvelopeParam,
    responseSerialize: serialize_rpctransact_TxEnvelope,
    responseDeserialize: deserialize_rpctransact_TxEnvelope,
  },
  // Formulate a transaction from a Payload and retrun the envelop with the Tx bytes ready to sign
formulateTx: {
    path: '/rpctransact.Transact/FormulateTx',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.Any,
    responseType: rpctransact_pb.TxEnvelope,
    requestSerialize: serialize_payload_Any,
    requestDeserialize: deserialize_payload_Any,
    responseSerialize: serialize_rpctransact_TxEnvelope,
    responseDeserialize: deserialize_rpctransact_TxEnvelope,
  },
  // Formulate and sign a CallTx transaction signed server-side and wait for it to be included in a block, retrieving response
callTxSync: {
    path: '/rpctransact.Transact/CallTxSync',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.CallTx,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_payload_CallTx,
    requestDeserialize: deserialize_payload_CallTx,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // Formulate and sign a CallTx transaction signed server-side
callTxAsync: {
    path: '/rpctransact.Transact/CallTxAsync',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.CallTx,
    responseType: txs_pb.Receipt,
    requestSerialize: serialize_payload_CallTx,
    requestDeserialize: deserialize_payload_CallTx,
    responseSerialize: serialize_txs_Receipt,
    responseDeserialize: deserialize_txs_Receipt,
  },
  // Perform a 'simulated' call of a contract against the current committed EVM state without any changes been saved
// and wait for the transaction to be included in a block
callTxSim: {
    path: '/rpctransact.Transact/CallTxSim',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.CallTx,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_payload_CallTx,
    requestDeserialize: deserialize_payload_CallTx,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // Perform a 'simulated' execution of provided code against the current committed EVM state without any changes been saved
callCodeSim: {
    path: '/rpctransact.Transact/CallCodeSim',
    requestStream: false,
    responseStream: false,
    requestType: rpctransact_pb.CallCodeParam,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_rpctransact_CallCodeParam,
    requestDeserialize: deserialize_rpctransact_CallCodeParam,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // Formulate a SendTx transaction signed server-side and wait for it to be included in a block, retrieving response
sendTxSync: {
    path: '/rpctransact.Transact/SendTxSync',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.SendTx,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_payload_SendTx,
    requestDeserialize: deserialize_payload_SendTx,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // Formulate and  SendTx transaction signed server-side
sendTxAsync: {
    path: '/rpctransact.Transact/SendTxAsync',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.SendTx,
    responseType: txs_pb.Receipt,
    requestSerialize: serialize_payload_SendTx,
    requestDeserialize: deserialize_payload_SendTx,
    responseSerialize: serialize_txs_Receipt,
    responseDeserialize: deserialize_txs_Receipt,
  },
  // Formulate a NameTx signed server-side and wait for it to be included in a block returning the registered name
nameTxSync: {
    path: '/rpctransact.Transact/NameTxSync',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.NameTx,
    responseType: exec_pb.TxExecution,
    requestSerialize: serialize_payload_NameTx,
    requestDeserialize: deserialize_payload_NameTx,
    responseSerialize: serialize_exec_TxExecution,
    responseDeserialize: deserialize_exec_TxExecution,
  },
  // Formulate a NameTx signed server-side
nameTxAsync: {
    path: '/rpctransact.Transact/NameTxAsync',
    requestStream: false,
    responseStream: false,
    requestType: payload_pb.NameTx,
    responseType: txs_pb.Receipt,
    requestSerialize: serialize_payload_NameTx,
    requestDeserialize: deserialize_payload_NameTx,
    responseSerialize: serialize_txs_Receipt,
    responseDeserialize: deserialize_txs_Receipt,
  },
};

