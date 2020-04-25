// package: keys
// file: keys.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as crypto_pb from "./crypto_pb";

export class ListRequest extends jspb.Message { 
    getKeyname(): string;
    setKeyname(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListRequest): ListRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListRequest;
    static deserializeBinaryFromReader(message: ListRequest, reader: jspb.BinaryReader): ListRequest;
}

export namespace ListRequest {
    export type AsObject = {
        keyname: string,
    }
}

export class VerifyResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): VerifyResponse.AsObject;
    static toObject(includeInstance: boolean, msg: VerifyResponse): VerifyResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: VerifyResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): VerifyResponse;
    static deserializeBinaryFromReader(message: VerifyResponse, reader: jspb.BinaryReader): VerifyResponse;
}

export namespace VerifyResponse {
    export type AsObject = {
    }
}

export class RemoveNameResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RemoveNameResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RemoveNameResponse): RemoveNameResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RemoveNameResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RemoveNameResponse;
    static deserializeBinaryFromReader(message: RemoveNameResponse, reader: jspb.BinaryReader): RemoveNameResponse;
}

export namespace RemoveNameResponse {
    export type AsObject = {
    }
}

export class AddNameResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddNameResponse.AsObject;
    static toObject(includeInstance: boolean, msg: AddNameResponse): AddNameResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddNameResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddNameResponse;
    static deserializeBinaryFromReader(message: AddNameResponse, reader: jspb.BinaryReader): AddNameResponse;
}

export namespace AddNameResponse {
    export type AsObject = {
    }
}

export class RemoveNameRequest extends jspb.Message { 
    getKeyname(): string;
    setKeyname(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RemoveNameRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RemoveNameRequest): RemoveNameRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RemoveNameRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RemoveNameRequest;
    static deserializeBinaryFromReader(message: RemoveNameRequest, reader: jspb.BinaryReader): RemoveNameRequest;
}

export namespace RemoveNameRequest {
    export type AsObject = {
        keyname: string,
    }
}

export class GenRequest extends jspb.Message { 
    getPassphrase(): string;
    setPassphrase(value: string): void;

    getCurvetype(): string;
    setCurvetype(value: string): void;

    getKeyname(): string;
    setKeyname(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GenRequest): GenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GenRequest;
    static deserializeBinaryFromReader(message: GenRequest, reader: jspb.BinaryReader): GenRequest;
}

export namespace GenRequest {
    export type AsObject = {
        passphrase: string,
        curvetype: string,
        keyname: string,
    }
}

export class GenResponse extends jspb.Message { 
    getAddress(): string;
    setAddress(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GenResponse): GenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GenResponse;
    static deserializeBinaryFromReader(message: GenResponse, reader: jspb.BinaryReader): GenResponse;
}

export namespace GenResponse {
    export type AsObject = {
        address: string,
    }
}

export class PubRequest extends jspb.Message { 
    getAddress(): string;
    setAddress(value: string): void;

    getName(): string;
    setName(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PubRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PubRequest): PubRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PubRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PubRequest;
    static deserializeBinaryFromReader(message: PubRequest, reader: jspb.BinaryReader): PubRequest;
}

export namespace PubRequest {
    export type AsObject = {
        address: string,
        name: string,
    }
}

export class PubResponse extends jspb.Message { 
    getPublickey(): Uint8Array | string;
    getPublickey_asU8(): Uint8Array;
    getPublickey_asB64(): string;
    setPublickey(value: Uint8Array | string): void;

    getCurvetype(): string;
    setCurvetype(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PubResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PubResponse): PubResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PubResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PubResponse;
    static deserializeBinaryFromReader(message: PubResponse, reader: jspb.BinaryReader): PubResponse;
}

export namespace PubResponse {
    export type AsObject = {
        publickey: Uint8Array | string,
        curvetype: string,
    }
}

export class ImportJSONRequest extends jspb.Message { 
    getPassphrase(): string;
    setPassphrase(value: string): void;

    getJson(): string;
    setJson(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ImportJSONRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ImportJSONRequest): ImportJSONRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ImportJSONRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ImportJSONRequest;
    static deserializeBinaryFromReader(message: ImportJSONRequest, reader: jspb.BinaryReader): ImportJSONRequest;
}

export namespace ImportJSONRequest {
    export type AsObject = {
        passphrase: string,
        json: string,
    }
}

export class ImportResponse extends jspb.Message { 
    getAddress(): string;
    setAddress(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ImportResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ImportResponse): ImportResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ImportResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ImportResponse;
    static deserializeBinaryFromReader(message: ImportResponse, reader: jspb.BinaryReader): ImportResponse;
}

export namespace ImportResponse {
    export type AsObject = {
        address: string,
    }
}

export class ImportRequest extends jspb.Message { 
    getPassphrase(): string;
    setPassphrase(value: string): void;

    getName(): string;
    setName(value: string): void;

    getCurvetype(): string;
    setCurvetype(value: string): void;

    getKeybytes(): Uint8Array | string;
    getKeybytes_asU8(): Uint8Array;
    getKeybytes_asB64(): string;
    setKeybytes(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ImportRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ImportRequest): ImportRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ImportRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ImportRequest;
    static deserializeBinaryFromReader(message: ImportRequest, reader: jspb.BinaryReader): ImportRequest;
}

export namespace ImportRequest {
    export type AsObject = {
        passphrase: string,
        name: string,
        curvetype: string,
        keybytes: Uint8Array | string,
    }
}

export class ExportRequest extends jspb.Message { 
    getPassphrase(): string;
    setPassphrase(value: string): void;

    getName(): string;
    setName(value: string): void;

    getAddress(): string;
    setAddress(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExportRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ExportRequest): ExportRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExportRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExportRequest;
    static deserializeBinaryFromReader(message: ExportRequest, reader: jspb.BinaryReader): ExportRequest;
}

export namespace ExportRequest {
    export type AsObject = {
        passphrase: string,
        name: string,
        address: string,
    }
}

export class ExportResponse extends jspb.Message { 
    getPublickey(): Uint8Array | string;
    getPublickey_asU8(): Uint8Array;
    getPublickey_asB64(): string;
    setPublickey(value: Uint8Array | string): void;

    getPrivatekey(): Uint8Array | string;
    getPrivatekey_asU8(): Uint8Array;
    getPrivatekey_asB64(): string;
    setPrivatekey(value: Uint8Array | string): void;

    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;

    getCurvetype(): string;
    setCurvetype(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExportResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ExportResponse): ExportResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExportResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExportResponse;
    static deserializeBinaryFromReader(message: ExportResponse, reader: jspb.BinaryReader): ExportResponse;
}

export namespace ExportResponse {
    export type AsObject = {
        publickey: Uint8Array | string,
        privatekey: Uint8Array | string,
        address: Uint8Array | string,
        curvetype: string,
    }
}

export class SignRequest extends jspb.Message { 
    getPassphrase(): string;
    setPassphrase(value: string): void;

    getAddress(): string;
    setAddress(value: string): void;

    getName(): string;
    setName(value: string): void;

    getMessage(): Uint8Array | string;
    getMessage_asU8(): Uint8Array;
    getMessage_asB64(): string;
    setMessage(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SignRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SignRequest): SignRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SignRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SignRequest;
    static deserializeBinaryFromReader(message: SignRequest, reader: jspb.BinaryReader): SignRequest;
}

export namespace SignRequest {
    export type AsObject = {
        passphrase: string,
        address: string,
        name: string,
        message: Uint8Array | string,
    }
}

export class SignResponse extends jspb.Message { 

    hasSignature(): boolean;
    clearSignature(): void;
    getSignature(): crypto_pb.Signature | undefined;
    setSignature(value?: crypto_pb.Signature): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SignResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SignResponse): SignResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SignResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SignResponse;
    static deserializeBinaryFromReader(message: SignResponse, reader: jspb.BinaryReader): SignResponse;
}

export namespace SignResponse {
    export type AsObject = {
        signature?: crypto_pb.Signature.AsObject,
    }
}

export class VerifyRequest extends jspb.Message { 
    getPublickey(): Uint8Array | string;
    getPublickey_asU8(): Uint8Array;
    getPublickey_asB64(): string;
    setPublickey(value: Uint8Array | string): void;

    getMessage(): Uint8Array | string;
    getMessage_asU8(): Uint8Array;
    getMessage_asB64(): string;
    setMessage(value: Uint8Array | string): void;


    hasSignature(): boolean;
    clearSignature(): void;
    getSignature(): crypto_pb.Signature | undefined;
    setSignature(value?: crypto_pb.Signature): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): VerifyRequest.AsObject;
    static toObject(includeInstance: boolean, msg: VerifyRequest): VerifyRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: VerifyRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): VerifyRequest;
    static deserializeBinaryFromReader(message: VerifyRequest, reader: jspb.BinaryReader): VerifyRequest;
}

export namespace VerifyRequest {
    export type AsObject = {
        publickey: Uint8Array | string,
        message: Uint8Array | string,
        signature?: crypto_pb.Signature.AsObject,
    }
}

export class HashRequest extends jspb.Message { 
    getHashtype(): string;
    setHashtype(value: string): void;

    getMessage(): Uint8Array | string;
    getMessage_asU8(): Uint8Array;
    getMessage_asB64(): string;
    setMessage(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HashRequest.AsObject;
    static toObject(includeInstance: boolean, msg: HashRequest): HashRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HashRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HashRequest;
    static deserializeBinaryFromReader(message: HashRequest, reader: jspb.BinaryReader): HashRequest;
}

export namespace HashRequest {
    export type AsObject = {
        hashtype: string,
        message: Uint8Array | string,
    }
}

export class HashResponse extends jspb.Message { 
    getHash(): string;
    setHash(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HashResponse.AsObject;
    static toObject(includeInstance: boolean, msg: HashResponse): HashResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HashResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HashResponse;
    static deserializeBinaryFromReader(message: HashResponse, reader: jspb.BinaryReader): HashResponse;
}

export namespace HashResponse {
    export type AsObject = {
        hash: string,
    }
}

export class KeyID extends jspb.Message { 
    getAddress(): string;
    setAddress(value: string): void;

    clearKeynameList(): void;
    getKeynameList(): Array<string>;
    setKeynameList(value: Array<string>): void;
    addKeyname(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): KeyID.AsObject;
    static toObject(includeInstance: boolean, msg: KeyID): KeyID.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: KeyID, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): KeyID;
    static deserializeBinaryFromReader(message: KeyID, reader: jspb.BinaryReader): KeyID;
}

export namespace KeyID {
    export type AsObject = {
        address: string,
        keynameList: Array<string>,
    }
}

export class ListResponse extends jspb.Message { 
    clearKeyList(): void;
    getKeyList(): Array<KeyID>;
    setKeyList(value: Array<KeyID>): void;
    addKey(value?: KeyID, index?: number): KeyID;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListResponse): ListResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListResponse;
    static deserializeBinaryFromReader(message: ListResponse, reader: jspb.BinaryReader): ListResponse;
}

export namespace ListResponse {
    export type AsObject = {
        keyList: Array<KeyID.AsObject>,
    }
}

export class AddNameRequest extends jspb.Message { 
    getKeyname(): string;
    setKeyname(value: string): void;

    getAddress(): string;
    setAddress(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddNameRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AddNameRequest): AddNameRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddNameRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddNameRequest;
    static deserializeBinaryFromReader(message: AddNameRequest, reader: jspb.BinaryReader): AddNameRequest;
}

export namespace AddNameRequest {
    export type AsObject = {
        keyname: string,
        address: string,
    }
}
