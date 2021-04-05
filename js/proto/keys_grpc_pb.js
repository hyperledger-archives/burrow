// GENERATED CODE -- DO NOT EDIT!

'use strict';
var keys_pb = require('./keys_pb.js');
var gogoproto_gogo_pb = require('./gogoproto/gogo_pb.js');
var crypto_pb = require('./crypto_pb.js');

function serialize_keys_AddNameRequest(arg) {
  if (!(arg instanceof keys_pb.AddNameRequest)) {
    throw new Error('Expected argument of type keys.AddNameRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_AddNameRequest(buffer_arg) {
  return keys_pb.AddNameRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_AddNameResponse(arg) {
  if (!(arg instanceof keys_pb.AddNameResponse)) {
    throw new Error('Expected argument of type keys.AddNameResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_AddNameResponse(buffer_arg) {
  return keys_pb.AddNameResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ExportRequest(arg) {
  if (!(arg instanceof keys_pb.ExportRequest)) {
    throw new Error('Expected argument of type keys.ExportRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ExportRequest(buffer_arg) {
  return keys_pb.ExportRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ExportResponse(arg) {
  if (!(arg instanceof keys_pb.ExportResponse)) {
    throw new Error('Expected argument of type keys.ExportResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ExportResponse(buffer_arg) {
  return keys_pb.ExportResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_GenRequest(arg) {
  if (!(arg instanceof keys_pb.GenRequest)) {
    throw new Error('Expected argument of type keys.GenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_GenRequest(buffer_arg) {
  return keys_pb.GenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_GenResponse(arg) {
  if (!(arg instanceof keys_pb.GenResponse)) {
    throw new Error('Expected argument of type keys.GenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_GenResponse(buffer_arg) {
  return keys_pb.GenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_HashRequest(arg) {
  if (!(arg instanceof keys_pb.HashRequest)) {
    throw new Error('Expected argument of type keys.HashRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_HashRequest(buffer_arg) {
  return keys_pb.HashRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_HashResponse(arg) {
  if (!(arg instanceof keys_pb.HashResponse)) {
    throw new Error('Expected argument of type keys.HashResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_HashResponse(buffer_arg) {
  return keys_pb.HashResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ImportJSONRequest(arg) {
  if (!(arg instanceof keys_pb.ImportJSONRequest)) {
    throw new Error('Expected argument of type keys.ImportJSONRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ImportJSONRequest(buffer_arg) {
  return keys_pb.ImportJSONRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ImportRequest(arg) {
  if (!(arg instanceof keys_pb.ImportRequest)) {
    throw new Error('Expected argument of type keys.ImportRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ImportRequest(buffer_arg) {
  return keys_pb.ImportRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ImportResponse(arg) {
  if (!(arg instanceof keys_pb.ImportResponse)) {
    throw new Error('Expected argument of type keys.ImportResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ImportResponse(buffer_arg) {
  return keys_pb.ImportResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ListRequest(arg) {
  if (!(arg instanceof keys_pb.ListRequest)) {
    throw new Error('Expected argument of type keys.ListRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ListRequest(buffer_arg) {
  return keys_pb.ListRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_ListResponse(arg) {
  if (!(arg instanceof keys_pb.ListResponse)) {
    throw new Error('Expected argument of type keys.ListResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_ListResponse(buffer_arg) {
  return keys_pb.ListResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_PubRequest(arg) {
  if (!(arg instanceof keys_pb.PubRequest)) {
    throw new Error('Expected argument of type keys.PubRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_PubRequest(buffer_arg) {
  return keys_pb.PubRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_PubResponse(arg) {
  if (!(arg instanceof keys_pb.PubResponse)) {
    throw new Error('Expected argument of type keys.PubResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_PubResponse(buffer_arg) {
  return keys_pb.PubResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_RemoveNameRequest(arg) {
  if (!(arg instanceof keys_pb.RemoveNameRequest)) {
    throw new Error('Expected argument of type keys.RemoveNameRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_RemoveNameRequest(buffer_arg) {
  return keys_pb.RemoveNameRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_RemoveNameResponse(arg) {
  if (!(arg instanceof keys_pb.RemoveNameResponse)) {
    throw new Error('Expected argument of type keys.RemoveNameResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_RemoveNameResponse(buffer_arg) {
  return keys_pb.RemoveNameResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_SignRequest(arg) {
  if (!(arg instanceof keys_pb.SignRequest)) {
    throw new Error('Expected argument of type keys.SignRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_SignRequest(buffer_arg) {
  return keys_pb.SignRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_SignResponse(arg) {
  if (!(arg instanceof keys_pb.SignResponse)) {
    throw new Error('Expected argument of type keys.SignResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_SignResponse(buffer_arg) {
  return keys_pb.SignResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_VerifyRequest(arg) {
  if (!(arg instanceof keys_pb.VerifyRequest)) {
    throw new Error('Expected argument of type keys.VerifyRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_VerifyRequest(buffer_arg) {
  return keys_pb.VerifyRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_keys_VerifyResponse(arg) {
  if (!(arg instanceof keys_pb.VerifyResponse)) {
    throw new Error('Expected argument of type keys.VerifyResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_keys_VerifyResponse(buffer_arg) {
  return keys_pb.VerifyResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var KeysService = exports['keys.Keys'] = {
  generateKey: {
    path: '/keys.Keys/GenerateKey',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.GenRequest,
    responseType: keys_pb.GenResponse,
    requestSerialize: serialize_keys_GenRequest,
    requestDeserialize: deserialize_keys_GenRequest,
    responseSerialize: serialize_keys_GenResponse,
    responseDeserialize: deserialize_keys_GenResponse,
  },
  publicKey: {
    path: '/keys.Keys/PublicKey',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.PubRequest,
    responseType: keys_pb.PubResponse,
    requestSerialize: serialize_keys_PubRequest,
    requestDeserialize: deserialize_keys_PubRequest,
    responseSerialize: serialize_keys_PubResponse,
    responseDeserialize: deserialize_keys_PubResponse,
  },
  sign: {
    path: '/keys.Keys/Sign',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.SignRequest,
    responseType: keys_pb.SignResponse,
    requestSerialize: serialize_keys_SignRequest,
    requestDeserialize: deserialize_keys_SignRequest,
    responseSerialize: serialize_keys_SignResponse,
    responseDeserialize: deserialize_keys_SignResponse,
  },
  verify: {
    path: '/keys.Keys/Verify',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.VerifyRequest,
    responseType: keys_pb.VerifyResponse,
    requestSerialize: serialize_keys_VerifyRequest,
    requestDeserialize: deserialize_keys_VerifyRequest,
    responseSerialize: serialize_keys_VerifyResponse,
    responseDeserialize: deserialize_keys_VerifyResponse,
  },
  import: {
    path: '/keys.Keys/Import',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.ImportRequest,
    responseType: keys_pb.ImportResponse,
    requestSerialize: serialize_keys_ImportRequest,
    requestDeserialize: deserialize_keys_ImportRequest,
    responseSerialize: serialize_keys_ImportResponse,
    responseDeserialize: deserialize_keys_ImportResponse,
  },
  importJSON: {
    path: '/keys.Keys/ImportJSON',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.ImportJSONRequest,
    responseType: keys_pb.ImportResponse,
    requestSerialize: serialize_keys_ImportJSONRequest,
    requestDeserialize: deserialize_keys_ImportJSONRequest,
    responseSerialize: serialize_keys_ImportResponse,
    responseDeserialize: deserialize_keys_ImportResponse,
  },
  export: {
    path: '/keys.Keys/Export',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.ExportRequest,
    responseType: keys_pb.ExportResponse,
    requestSerialize: serialize_keys_ExportRequest,
    requestDeserialize: deserialize_keys_ExportRequest,
    responseSerialize: serialize_keys_ExportResponse,
    responseDeserialize: deserialize_keys_ExportResponse,
  },
  hash: {
    path: '/keys.Keys/Hash',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.HashRequest,
    responseType: keys_pb.HashResponse,
    requestSerialize: serialize_keys_HashRequest,
    requestDeserialize: deserialize_keys_HashRequest,
    responseSerialize: serialize_keys_HashResponse,
    responseDeserialize: deserialize_keys_HashResponse,
  },
  removeName: {
    path: '/keys.Keys/RemoveName',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.RemoveNameRequest,
    responseType: keys_pb.RemoveNameResponse,
    requestSerialize: serialize_keys_RemoveNameRequest,
    requestDeserialize: deserialize_keys_RemoveNameRequest,
    responseSerialize: serialize_keys_RemoveNameResponse,
    responseDeserialize: deserialize_keys_RemoveNameResponse,
  },
  list: {
    path: '/keys.Keys/List',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.ListRequest,
    responseType: keys_pb.ListResponse,
    requestSerialize: serialize_keys_ListRequest,
    requestDeserialize: deserialize_keys_ListRequest,
    responseSerialize: serialize_keys_ListResponse,
    responseDeserialize: deserialize_keys_ListResponse,
  },
  addName: {
    path: '/keys.Keys/AddName',
    requestStream: false,
    responseStream: false,
    requestType: keys_pb.AddNameRequest,
    responseType: keys_pb.AddNameResponse,
    requestSerialize: serialize_keys_AddNameRequest,
    requestDeserialize: deserialize_keys_AddNameRequest,
    responseSerialize: serialize_keys_AddNameResponse,
    responseDeserialize: deserialize_keys_AddNameResponse,
  },
};

