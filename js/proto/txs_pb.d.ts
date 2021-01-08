// package: txs
// file: txs.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as crypto_pb from "./crypto_pb";

export class Envelope extends jspb.Message { 
    clearSignatoriesList(): void;
    getSignatoriesList(): Array<Signatory>;
    setSignatoriesList(value: Array<Signatory>): Envelope;
    addSignatories(value?: Signatory, index?: number): Signatory;

    getTx(): Uint8Array | string;
    getTx_asU8(): Uint8Array;
    getTx_asB64(): string;
    setTx(value: Uint8Array | string): Envelope;

    getEncoding(): Envelope.EncodingType;
    setEncoding(value: Envelope.EncodingType): Envelope;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Envelope.AsObject;
    static toObject(includeInstance: boolean, msg: Envelope): Envelope.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Envelope, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Envelope;
    static deserializeBinaryFromReader(message: Envelope, reader: jspb.BinaryReader): Envelope;
}

export namespace Envelope {
    export type AsObject = {
        signatoriesList: Array<Signatory.AsObject>,
        tx: Uint8Array | string,
        encoding: Envelope.EncodingType,
    }

    export enum EncodingType {
    JSON = 0,
    RLP = 1,
    }

}

export class Signatory extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): Signatory;


    hasPublickey(): boolean;
    clearPublickey(): void;
    getPublickey(): crypto_pb.PublicKey | undefined;
    setPublickey(value?: crypto_pb.PublicKey): Signatory;


    hasSignature(): boolean;
    clearSignature(): void;
    getSignature(): crypto_pb.Signature | undefined;
    setSignature(value?: crypto_pb.Signature): Signatory;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Signatory.AsObject;
    static toObject(includeInstance: boolean, msg: Signatory): Signatory.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Signatory, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Signatory;
    static deserializeBinaryFromReader(message: Signatory, reader: jspb.BinaryReader): Signatory;
}

export namespace Signatory {
    export type AsObject = {
        address: Uint8Array | string,
        publickey?: crypto_pb.PublicKey.AsObject,
        signature?: crypto_pb.Signature.AsObject,
    }
}

export class Receipt extends jspb.Message { 
    getTxtype(): number;
    setTxtype(value: number): Receipt;

    getTxhash(): Uint8Array | string;
    getTxhash_asU8(): Uint8Array;
    getTxhash_asB64(): string;
    setTxhash(value: Uint8Array | string): Receipt;

    getCreatescontract(): boolean;
    setCreatescontract(value: boolean): Receipt;

    getContractaddress(): Uint8Array | string;
    getContractaddress_asU8(): Uint8Array;
    getContractaddress_asB64(): string;
    setContractaddress(value: Uint8Array | string): Receipt;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Receipt.AsObject;
    static toObject(includeInstance: boolean, msg: Receipt): Receipt.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Receipt, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Receipt;
    static deserializeBinaryFromReader(message: Receipt, reader: jspb.BinaryReader): Receipt;
}

export namespace Receipt {
    export type AsObject = {
        txtype: number,
        txhash: Uint8Array | string,
        createscontract: boolean,
        contractaddress: Uint8Array | string,
    }
}
