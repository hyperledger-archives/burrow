// package: rpcdump
// file: rpcdump.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as dump_pb from "./dump_pb";

export class GetDumpParam extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): GetDumpParam;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDumpParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetDumpParam): GetDumpParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDumpParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDumpParam;
    static deserializeBinaryFromReader(message: GetDumpParam, reader: jspb.BinaryReader): GetDumpParam;
}

export namespace GetDumpParam {
    export type AsObject = {
        height: number,
    }
}
