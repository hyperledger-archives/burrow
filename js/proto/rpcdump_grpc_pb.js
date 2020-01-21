// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('grpc');
var rpcdump_pb = require('./rpcdump_pb.js');
var github_com_gogo_protobuf_gogoproto_gogo_pb = require('./github.com/gogo/protobuf/gogoproto/gogo_pb.js');
var dump_pb = require('./dump_pb.js');

function serialize_dump_Dump(arg) {
  if (!(arg instanceof dump_pb.Dump)) {
    throw new Error('Expected argument of type dump.Dump');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_dump_Dump(buffer_arg) {
  return dump_pb.Dump.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_rpcdump_GetDumpParam(arg) {
  if (!(arg instanceof rpcdump_pb.GetDumpParam)) {
    throw new Error('Expected argument of type rpcdump.GetDumpParam');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_rpcdump_GetDumpParam(buffer_arg) {
  return rpcdump_pb.GetDumpParam.deserializeBinary(new Uint8Array(buffer_arg));
}


var DumpService = exports.DumpService = {
  getDump: {
    path: '/rpcdump.Dump/GetDump',
    requestStream: false,
    responseStream: true,
    requestType: rpcdump_pb.GetDumpParam,
    responseType: dump_pb.Dump,
    requestSerialize: serialize_rpcdump_GetDumpParam,
    requestDeserialize: deserialize_rpcdump_GetDumpParam,
    responseSerialize: serialize_dump_Dump,
    responseDeserialize: deserialize_dump_Dump,
  },
};

exports.DumpClient = grpc.makeGenericClientConstructor(DumpService);
