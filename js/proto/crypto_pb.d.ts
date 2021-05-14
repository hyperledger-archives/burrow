// package: crypto
// file: crypto.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";

export class PublicKey extends jspb.Message { 
    getCurvetype(): number;
    setCurvetype(value: number): PublicKey;
    getPublickey(): Uint8Array | string;
    getPublickey_asU8(): Uint8Array;
    getPublickey_asB64(): string;
    setPublickey(value: Uint8Array | string): PublicKey;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PublicKey.AsObject;
    static toObject(includeInstance: boolean, msg: PublicKey): PublicKey.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PublicKey, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PublicKey;
    static deserializeBinaryFromReader(message: PublicKey, reader: jspb.BinaryReader): PublicKey;
}

export namespace PublicKey {
    export type AsObject = {
        curvetype: number,
        publickey: Uint8Array | string,
    }
}

export class PrivateKey extends jspb.Message { 
    getCurvetype(): number;
    setCurvetype(value: number): PrivateKey;
    getPublickey(): Uint8Array | string;
    getPublickey_asU8(): Uint8Array;
    getPublickey_asB64(): string;
    setPublickey(value: Uint8Array | string): PrivateKey;
    getPrivatekey(): Uint8Array | string;
    getPrivatekey_asU8(): Uint8Array;
    getPrivatekey_asB64(): string;
    setPrivatekey(value: Uint8Array | string): PrivateKey;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PrivateKey.AsObject;
    static toObject(includeInstance: boolean, msg: PrivateKey): PrivateKey.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PrivateKey, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PrivateKey;
    static deserializeBinaryFromReader(message: PrivateKey, reader: jspb.BinaryReader): PrivateKey;
}

export namespace PrivateKey {
    export type AsObject = {
        curvetype: number,
        publickey: Uint8Array | string,
        privatekey: Uint8Array | string,
    }
}

export class Signature extends jspb.Message { 
    getCurvetype(): number;
    setCurvetype(value: number): Signature;
    getSignature(): Uint8Array | string;
    getSignature_asU8(): Uint8Array;
    getSignature_asB64(): string;
    setSignature(value: Uint8Array | string): Signature;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Signature.AsObject;
    static toObject(includeInstance: boolean, msg: Signature): Signature.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Signature, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Signature;
    static deserializeBinaryFromReader(message: Signature, reader: jspb.BinaryReader): Signature;
}

export namespace Signature {
    export type AsObject = {
        curvetype: number,
        signature: Uint8Array | string,
    }
}
