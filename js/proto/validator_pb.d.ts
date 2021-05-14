// package: validator
// file: validator.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as crypto_pb from "./crypto_pb";

export class Validator extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): Validator;

    hasPublickey(): boolean;
    clearPublickey(): void;
    getPublickey(): crypto_pb.PublicKey | undefined;
    setPublickey(value?: crypto_pb.PublicKey): Validator;
    getPower(): number;
    setPower(value: number): Validator;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Validator.AsObject;
    static toObject(includeInstance: boolean, msg: Validator): Validator.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Validator, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Validator;
    static deserializeBinaryFromReader(message: Validator, reader: jspb.BinaryReader): Validator;
}

export namespace Validator {
    export type AsObject = {
        address: Uint8Array | string,
        publickey?: crypto_pb.PublicKey.AsObject,
        power: number,
    }
}
