// package: tendermint.p2p
// file: tendermint/p2p/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";

export class NetAddress extends jspb.Message { 
    getId(): string;
    setId(value: string): NetAddress;
    getIp(): string;
    setIp(value: string): NetAddress;
    getPort(): number;
    setPort(value: number): NetAddress;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NetAddress.AsObject;
    static toObject(includeInstance: boolean, msg: NetAddress): NetAddress.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NetAddress, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NetAddress;
    static deserializeBinaryFromReader(message: NetAddress, reader: jspb.BinaryReader): NetAddress;
}

export namespace NetAddress {
    export type AsObject = {
        id: string,
        ip: string,
        port: number,
    }
}

export class ProtocolVersion extends jspb.Message { 
    getP2p(): number;
    setP2p(value: number): ProtocolVersion;
    getBlock(): number;
    setBlock(value: number): ProtocolVersion;
    getApp(): number;
    setApp(value: number): ProtocolVersion;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProtocolVersion.AsObject;
    static toObject(includeInstance: boolean, msg: ProtocolVersion): ProtocolVersion.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProtocolVersion, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProtocolVersion;
    static deserializeBinaryFromReader(message: ProtocolVersion, reader: jspb.BinaryReader): ProtocolVersion;
}

export namespace ProtocolVersion {
    export type AsObject = {
        p2p: number,
        block: number,
        app: number,
    }
}

export class DefaultNodeInfo extends jspb.Message { 

    hasProtocolVersion(): boolean;
    clearProtocolVersion(): void;
    getProtocolVersion(): ProtocolVersion | undefined;
    setProtocolVersion(value?: ProtocolVersion): DefaultNodeInfo;
    getDefaultNodeId(): string;
    setDefaultNodeId(value: string): DefaultNodeInfo;
    getListenAddr(): string;
    setListenAddr(value: string): DefaultNodeInfo;
    getNetwork(): string;
    setNetwork(value: string): DefaultNodeInfo;
    getVersion(): string;
    setVersion(value: string): DefaultNodeInfo;
    getChannels(): Uint8Array | string;
    getChannels_asU8(): Uint8Array;
    getChannels_asB64(): string;
    setChannels(value: Uint8Array | string): DefaultNodeInfo;
    getMoniker(): string;
    setMoniker(value: string): DefaultNodeInfo;

    hasOther(): boolean;
    clearOther(): void;
    getOther(): DefaultNodeInfoOther | undefined;
    setOther(value?: DefaultNodeInfoOther): DefaultNodeInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DefaultNodeInfo.AsObject;
    static toObject(includeInstance: boolean, msg: DefaultNodeInfo): DefaultNodeInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DefaultNodeInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DefaultNodeInfo;
    static deserializeBinaryFromReader(message: DefaultNodeInfo, reader: jspb.BinaryReader): DefaultNodeInfo;
}

export namespace DefaultNodeInfo {
    export type AsObject = {
        protocolVersion?: ProtocolVersion.AsObject,
        defaultNodeId: string,
        listenAddr: string,
        network: string,
        version: string,
        channels: Uint8Array | string,
        moniker: string,
        other?: DefaultNodeInfoOther.AsObject,
    }
}

export class DefaultNodeInfoOther extends jspb.Message { 
    getTxIndex(): string;
    setTxIndex(value: string): DefaultNodeInfoOther;
    getRpcAddress(): string;
    setRpcAddress(value: string): DefaultNodeInfoOther;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DefaultNodeInfoOther.AsObject;
    static toObject(includeInstance: boolean, msg: DefaultNodeInfoOther): DefaultNodeInfoOther.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DefaultNodeInfoOther, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DefaultNodeInfoOther;
    static deserializeBinaryFromReader(message: DefaultNodeInfoOther, reader: jspb.BinaryReader): DefaultNodeInfoOther;
}

export namespace DefaultNodeInfoOther {
    export type AsObject = {
        txIndex: string,
        rpcAddress: string,
    }
}
