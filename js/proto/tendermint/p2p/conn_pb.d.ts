// package: tendermint.p2p
// file: tendermint/p2p/conn.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as tendermint_crypto_keys_pb from "../../tendermint/crypto/keys_pb";

export class PacketPing extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PacketPing.AsObject;
    static toObject(includeInstance: boolean, msg: PacketPing): PacketPing.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PacketPing, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PacketPing;
    static deserializeBinaryFromReader(message: PacketPing, reader: jspb.BinaryReader): PacketPing;
}

export namespace PacketPing {
    export type AsObject = {
    }
}

export class PacketPong extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PacketPong.AsObject;
    static toObject(includeInstance: boolean, msg: PacketPong): PacketPong.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PacketPong, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PacketPong;
    static deserializeBinaryFromReader(message: PacketPong, reader: jspb.BinaryReader): PacketPong;
}

export namespace PacketPong {
    export type AsObject = {
    }
}

export class PacketMsg extends jspb.Message { 
    getChannelId(): number;
    setChannelId(value: number): PacketMsg;
    getEof(): boolean;
    setEof(value: boolean): PacketMsg;
    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): PacketMsg;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PacketMsg.AsObject;
    static toObject(includeInstance: boolean, msg: PacketMsg): PacketMsg.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PacketMsg, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PacketMsg;
    static deserializeBinaryFromReader(message: PacketMsg, reader: jspb.BinaryReader): PacketMsg;
}

export namespace PacketMsg {
    export type AsObject = {
        channelId: number,
        eof: boolean,
        data: Uint8Array | string,
    }
}

export class Packet extends jspb.Message { 

    hasPacketPing(): boolean;
    clearPacketPing(): void;
    getPacketPing(): PacketPing | undefined;
    setPacketPing(value?: PacketPing): Packet;

    hasPacketPong(): boolean;
    clearPacketPong(): void;
    getPacketPong(): PacketPong | undefined;
    setPacketPong(value?: PacketPong): Packet;

    hasPacketMsg(): boolean;
    clearPacketMsg(): void;
    getPacketMsg(): PacketMsg | undefined;
    setPacketMsg(value?: PacketMsg): Packet;

    getSumCase(): Packet.SumCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Packet.AsObject;
    static toObject(includeInstance: boolean, msg: Packet): Packet.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Packet, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Packet;
    static deserializeBinaryFromReader(message: Packet, reader: jspb.BinaryReader): Packet;
}

export namespace Packet {
    export type AsObject = {
        packetPing?: PacketPing.AsObject,
        packetPong?: PacketPong.AsObject,
        packetMsg?: PacketMsg.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        PACKET_PING = 1,
        PACKET_PONG = 2,
        PACKET_MSG = 3,
    }

}

export class AuthSigMessage extends jspb.Message { 

    hasPubKey(): boolean;
    clearPubKey(): void;
    getPubKey(): tendermint_crypto_keys_pb.PublicKey | undefined;
    setPubKey(value?: tendermint_crypto_keys_pb.PublicKey): AuthSigMessage;
    getSig(): Uint8Array | string;
    getSig_asU8(): Uint8Array;
    getSig_asB64(): string;
    setSig(value: Uint8Array | string): AuthSigMessage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthSigMessage.AsObject;
    static toObject(includeInstance: boolean, msg: AuthSigMessage): AuthSigMessage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthSigMessage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthSigMessage;
    static deserializeBinaryFromReader(message: AuthSigMessage, reader: jspb.BinaryReader): AuthSigMessage;
}

export namespace AuthSigMessage {
    export type AsObject = {
        pubKey?: tendermint_crypto_keys_pb.PublicKey.AsObject,
        sig: Uint8Array | string,
    }
}
