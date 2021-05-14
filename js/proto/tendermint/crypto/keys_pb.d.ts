// package: tendermint.crypto
// file: tendermint/crypto/keys.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";

export class PublicKey extends jspb.Message { 

    hasEd25519(): boolean;
    clearEd25519(): void;
    getEd25519(): Uint8Array | string;
    getEd25519_asU8(): Uint8Array;
    getEd25519_asB64(): string;
    setEd25519(value: Uint8Array | string): PublicKey;

    hasSecp256k1(): boolean;
    clearSecp256k1(): void;
    getSecp256k1(): Uint8Array | string;
    getSecp256k1_asU8(): Uint8Array;
    getSecp256k1_asB64(): string;
    setSecp256k1(value: Uint8Array | string): PublicKey;

    getSumCase(): PublicKey.SumCase;

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
        ed25519: Uint8Array | string,
        secp256k1: Uint8Array | string,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        ED25519 = 1,
        SECP256K1 = 2,
    }

}
