// package: errors
// file: errors.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";

export class Exception extends jspb.Message { 
    getCode(): number;
    setCode(value: number): Exception;
    getException(): string;
    setException(value: string): Exception;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Exception.AsObject;
    static toObject(includeInstance: boolean, msg: Exception): Exception.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Exception, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Exception;
    static deserializeBinaryFromReader(message: Exception, reader: jspb.BinaryReader): Exception;
}

export namespace Exception {
    export type AsObject = {
        code: number,
        exception: string,
    }
}
