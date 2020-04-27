// package: registry
// file: registry.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";

export class NodeIdentity extends jspb.Message { 
    getMoniker(): string;
    setMoniker(value: string): NodeIdentity;
    getNetworkaddress(): string;
    setNetworkaddress(value: string): NodeIdentity;
    getTendermintnodeid(): Uint8Array | string;
    getTendermintnodeid_asU8(): Uint8Array;
    getTendermintnodeid_asB64(): string;
    setTendermintnodeid(value: Uint8Array | string): NodeIdentity;
    getValidatorpublickey(): Uint8Array | string;
    getValidatorpublickey_asU8(): Uint8Array;
    getValidatorpublickey_asB64(): string;
    setValidatorpublickey(value: Uint8Array | string): NodeIdentity;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NodeIdentity.AsObject;
    static toObject(includeInstance: boolean, msg: NodeIdentity): NodeIdentity.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NodeIdentity, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NodeIdentity;
    static deserializeBinaryFromReader(message: NodeIdentity, reader: jspb.BinaryReader): NodeIdentity;
}

export namespace NodeIdentity {
    export type AsObject = {
        moniker: string,
        networkaddress: string,
        tendermintnodeid: Uint8Array | string,
        validatorpublickey: Uint8Array | string,
    }
}
