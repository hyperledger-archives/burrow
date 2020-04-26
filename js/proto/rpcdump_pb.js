/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

var github_com_gogo_protobuf_gogoproto_gogo_pb = require('./github.com/gogo/protobuf/gogoproto/gogo_pb.js');
goog.object.extend(proto, github_com_gogo_protobuf_gogoproto_gogo_pb);
var dump_pb = require('./dump_pb.js');
goog.object.extend(proto, dump_pb);
goog.exportSymbol('proto.rpcdump.GetDumpParam', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.rpcdump.GetDumpParam = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.rpcdump.GetDumpParam, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.rpcdump.GetDumpParam.displayName = 'proto.rpcdump.GetDumpParam';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto suitable for use in Soy templates.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     com.google.apps.jspb.JsClassTemplate.JS_RESERVED_WORDS.
 * @param {boolean=} opt_includeInstance Whether to include the JSPB instance
 *     for transitional soy proto support: http://goto/soy-param-migration
 * @return {!Object}
 */
proto.rpcdump.GetDumpParam.prototype.toObject = function(opt_includeInstance) {
  return proto.rpcdump.GetDumpParam.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Whether to include the JSPB
 *     instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.rpcdump.GetDumpParam} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.rpcdump.GetDumpParam.toObject = function(includeInstance, msg) {
  var f, obj = {
    height: jspb.Message.getFieldWithDefault(msg, 1, 0)
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.rpcdump.GetDumpParam}
 */
proto.rpcdump.GetDumpParam.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.rpcdump.GetDumpParam;
  return proto.rpcdump.GetDumpParam.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.rpcdump.GetDumpParam} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.rpcdump.GetDumpParam}
 */
proto.rpcdump.GetDumpParam.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readUint64());
      msg.setHeight(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.rpcdump.GetDumpParam.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.rpcdump.GetDumpParam.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.rpcdump.GetDumpParam} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.rpcdump.GetDumpParam.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getHeight();
  if (f !== 0) {
    writer.writeUint64(
      1,
      f
    );
  }
};


/**
 * optional uint64 height = 1;
 * @return {number}
 */
proto.rpcdump.GetDumpParam.prototype.getHeight = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/** @param {number} value */
proto.rpcdump.GetDumpParam.prototype.setHeight = function(value) {
  jspb.Message.setProto3IntField(this, 1, value);
};


goog.object.extend(exports, proto.rpcdump);
