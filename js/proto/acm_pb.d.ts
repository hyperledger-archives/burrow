// package: acm
// file: acm.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as permission_pb from "./permission_pb";
import * as crypto_pb from "./crypto_pb";

export class Account extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;


    hasPublickey(): boolean;
    clearPublickey(): void;
    getPublickey(): crypto_pb.PublicKey | undefined;
    setPublickey(value?: crypto_pb.PublicKey): void;

    getSequence(): number;
    setSequence(value: number): void;

    getBalance(): number;
    setBalance(value: number): void;

    getEvmcode(): Uint8Array | string;
    getEvmcode_asU8(): Uint8Array;
    getEvmcode_asB64(): string;
    setEvmcode(value: Uint8Array | string): void;


    hasPermissions(): boolean;
    clearPermissions(): void;
    getPermissions(): permission_pb.AccountPermissions | undefined;
    setPermissions(value?: permission_pb.AccountPermissions): void;

    getWasmcode(): Uint8Array | string;
    getWasmcode_asU8(): Uint8Array;
    getWasmcode_asB64(): string;
    setWasmcode(value: Uint8Array | string): void;

    getNativename(): string;
    setNativename(value: string): void;

    getCodehash(): Uint8Array | string;
    getCodehash_asU8(): Uint8Array;
    getCodehash_asB64(): string;
    setCodehash(value: Uint8Array | string): void;

    clearContractmetaList(): void;
    getContractmetaList(): Array<ContractMeta>;
    setContractmetaList(value: Array<ContractMeta>): void;
    addContractmeta(value?: ContractMeta, index?: number): ContractMeta;

    getForebear(): Uint8Array | string;
    getForebear_asU8(): Uint8Array;
    getForebear_asB64(): string;
    setForebear(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Account.AsObject;
    static toObject(includeInstance: boolean, msg: Account): Account.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Account, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Account;
    static deserializeBinaryFromReader(message: Account, reader: jspb.BinaryReader): Account;
}

export namespace Account {
    export type AsObject = {
        address: Uint8Array | string,
        publickey?: crypto_pb.PublicKey.AsObject,
        sequence: number,
        balance: number,
        evmcode: Uint8Array | string,
        permissions?: permission_pb.AccountPermissions.AsObject,
        wasmcode: Uint8Array | string,
        nativename: string,
        codehash: Uint8Array | string,
        contractmetaList: Array<ContractMeta.AsObject>,
        forebear: Uint8Array | string,
    }
}

export class ContractMeta extends jspb.Message { 
    getCodehash(): Uint8Array | string;
    getCodehash_asU8(): Uint8Array;
    getCodehash_asB64(): string;
    setCodehash(value: Uint8Array | string): void;

    getMetadatahash(): Uint8Array | string;
    getMetadatahash_asU8(): Uint8Array;
    getMetadatahash_asB64(): string;
    setMetadatahash(value: Uint8Array | string): void;

    getMetadata(): string;
    setMetadata(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ContractMeta.AsObject;
    static toObject(includeInstance: boolean, msg: ContractMeta): ContractMeta.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ContractMeta, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ContractMeta;
    static deserializeBinaryFromReader(message: ContractMeta, reader: jspb.BinaryReader): ContractMeta;
}

export namespace ContractMeta {
    export type AsObject = {
        codehash: Uint8Array | string,
        metadatahash: Uint8Array | string,
        metadata: string,
    }
}
