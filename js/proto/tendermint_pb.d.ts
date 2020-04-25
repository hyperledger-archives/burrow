// package: tendermint
// file: tendermint.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";

export class NodeInfo extends jspb.Message { 
    getId(): Uint8Array | string;
    getId_asU8(): Uint8Array;
    getId_asB64(): string;
    setId(value: Uint8Array | string): void;

    getListenaddress(): string;
    setListenaddress(value: string): void;

    getNetwork(): string;
    setNetwork(value: string): void;

    getVersion(): string;
    setVersion(value: string): void;

    getChannels(): Uint8Array | string;
    getChannels_asU8(): Uint8Array;
    getChannels_asB64(): string;
    setChannels(value: Uint8Array | string): void;

    getMoniker(): string;
    setMoniker(value: string): void;

    getRpcaddress(): string;
    setRpcaddress(value: string): void;

    getTxindex(): string;
    setTxindex(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NodeInfo.AsObject;
    static toObject(includeInstance: boolean, msg: NodeInfo): NodeInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NodeInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NodeInfo;
    static deserializeBinaryFromReader(message: NodeInfo, reader: jspb.BinaryReader): NodeInfo;
}

export namespace NodeInfo {
    export type AsObject = {
        id: Uint8Array | string,
        listenaddress: string,
        network: string,
        version: string,
        channels: Uint8Array | string,
        moniker: string,
        rpcaddress: string,
        txindex: string,
    }
}
