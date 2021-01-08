// package: tendermint.p2p
// file: tendermint/p2p/pex.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as tendermint_p2p_types_pb from "../../tendermint/p2p/types_pb";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";

export class PexRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PexRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PexRequest): PexRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PexRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PexRequest;
    static deserializeBinaryFromReader(message: PexRequest, reader: jspb.BinaryReader): PexRequest;
}

export namespace PexRequest {
    export type AsObject = {
    }
}

export class PexAddrs extends jspb.Message { 
    clearAddrsList(): void;
    getAddrsList(): Array<tendermint_p2p_types_pb.NetAddress>;
    setAddrsList(value: Array<tendermint_p2p_types_pb.NetAddress>): PexAddrs;
    addAddrs(value?: tendermint_p2p_types_pb.NetAddress, index?: number): tendermint_p2p_types_pb.NetAddress;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PexAddrs.AsObject;
    static toObject(includeInstance: boolean, msg: PexAddrs): PexAddrs.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PexAddrs, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PexAddrs;
    static deserializeBinaryFromReader(message: PexAddrs, reader: jspb.BinaryReader): PexAddrs;
}

export namespace PexAddrs {
    export type AsObject = {
        addrsList: Array<tendermint_p2p_types_pb.NetAddress.AsObject>,
    }
}

export class Message extends jspb.Message { 

    hasPexRequest(): boolean;
    clearPexRequest(): void;
    getPexRequest(): PexRequest | undefined;
    setPexRequest(value?: PexRequest): Message;


    hasPexAddrs(): boolean;
    clearPexAddrs(): void;
    getPexAddrs(): PexAddrs | undefined;
    setPexAddrs(value?: PexAddrs): Message;


    getSumCase(): Message.SumCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Message.AsObject;
    static toObject(includeInstance: boolean, msg: Message): Message.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Message, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Message;
    static deserializeBinaryFromReader(message: Message, reader: jspb.BinaryReader): Message;
}

export namespace Message {
    export type AsObject = {
        pexRequest?: PexRequest.AsObject,
        pexAddrs?: PexAddrs.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
    
    PEX_REQUEST = 1,

    PEX_ADDRS = 2,

    }

}
