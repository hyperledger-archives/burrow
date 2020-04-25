// package: common
// file: github.com/tendermint/tendermint/libs/common/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "../../../../../github.com/gogo/protobuf/gogoproto/gogo_pb";

export class KVPair extends jspb.Message { 
    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): void;

    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): KVPair.AsObject;
    static toObject(includeInstance: boolean, msg: KVPair): KVPair.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: KVPair, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): KVPair;
    static deserializeBinaryFromReader(message: KVPair, reader: jspb.BinaryReader): KVPair;
}

export namespace KVPair {
    export type AsObject = {
        key: Uint8Array | string,
        value: Uint8Array | string,
    }
}

export class KI64Pair extends jspb.Message { 
    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): void;

    getValue(): number;
    setValue(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): KI64Pair.AsObject;
    static toObject(includeInstance: boolean, msg: KI64Pair): KI64Pair.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: KI64Pair, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): KI64Pair;
    static deserializeBinaryFromReader(message: KI64Pair, reader: jspb.BinaryReader): KI64Pair;
}

export namespace KI64Pair {
    export type AsObject = {
        key: Uint8Array | string,
        value: number,
    }
}
