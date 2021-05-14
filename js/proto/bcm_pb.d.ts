// package: bcm
// file: bcm.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";

export class SyncInfo extends jspb.Message { 
    getLatestblockheight(): number;
    setLatestblockheight(value: number): SyncInfo;
    getLatestblockhash(): Uint8Array | string;
    getLatestblockhash_asU8(): Uint8Array;
    getLatestblockhash_asB64(): string;
    setLatestblockhash(value: Uint8Array | string): SyncInfo;
    getLatestapphash(): Uint8Array | string;
    getLatestapphash_asU8(): Uint8Array;
    getLatestapphash_asB64(): string;
    setLatestapphash(value: Uint8Array | string): SyncInfo;

    hasLatestblocktime(): boolean;
    clearLatestblocktime(): void;
    getLatestblocktime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setLatestblocktime(value?: google_protobuf_timestamp_pb.Timestamp): SyncInfo;

    hasLatestblockseentime(): boolean;
    clearLatestblockseentime(): void;
    getLatestblockseentime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setLatestblockseentime(value?: google_protobuf_timestamp_pb.Timestamp): SyncInfo;

    hasLatestblockduration(): boolean;
    clearLatestblockduration(): void;
    getLatestblockduration(): google_protobuf_duration_pb.Duration | undefined;
    setLatestblockduration(value?: google_protobuf_duration_pb.Duration): SyncInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SyncInfo.AsObject;
    static toObject(includeInstance: boolean, msg: SyncInfo): SyncInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SyncInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SyncInfo;
    static deserializeBinaryFromReader(message: SyncInfo, reader: jspb.BinaryReader): SyncInfo;
}

export namespace SyncInfo {
    export type AsObject = {
        latestblockheight: number,
        latestblockhash: Uint8Array | string,
        latestapphash: Uint8Array | string,
        latestblocktime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        latestblockseentime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        latestblockduration?: google_protobuf_duration_pb.Duration.AsObject,
    }
}

export class PersistedState extends jspb.Message { 
    getApphashafterlastblock(): Uint8Array | string;
    getApphashafterlastblock_asU8(): Uint8Array;
    getApphashafterlastblock_asB64(): string;
    setApphashafterlastblock(value: Uint8Array | string): PersistedState;

    hasLastblocktime(): boolean;
    clearLastblocktime(): void;
    getLastblocktime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setLastblocktime(value?: google_protobuf_timestamp_pb.Timestamp): PersistedState;
    getLastblockheight(): number;
    setLastblockheight(value: number): PersistedState;
    getGenesishash(): Uint8Array | string;
    getGenesishash_asU8(): Uint8Array;
    getGenesishash_asB64(): string;
    setGenesishash(value: Uint8Array | string): PersistedState;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PersistedState.AsObject;
    static toObject(includeInstance: boolean, msg: PersistedState): PersistedState.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PersistedState, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PersistedState;
    static deserializeBinaryFromReader(message: PersistedState, reader: jspb.BinaryReader): PersistedState;
}

export namespace PersistedState {
    export type AsObject = {
        apphashafterlastblock: Uint8Array | string,
        lastblocktime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        lastblockheight: number,
        genesishash: Uint8Array | string,
    }
}
