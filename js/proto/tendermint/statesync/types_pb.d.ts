// package: tendermint.statesync
// file: tendermint/statesync/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class Message extends jspb.Message { 

    hasSnapshotsRequest(): boolean;
    clearSnapshotsRequest(): void;
    getSnapshotsRequest(): SnapshotsRequest | undefined;
    setSnapshotsRequest(value?: SnapshotsRequest): Message;

    hasSnapshotsResponse(): boolean;
    clearSnapshotsResponse(): void;
    getSnapshotsResponse(): SnapshotsResponse | undefined;
    setSnapshotsResponse(value?: SnapshotsResponse): Message;

    hasChunkRequest(): boolean;
    clearChunkRequest(): void;
    getChunkRequest(): ChunkRequest | undefined;
    setChunkRequest(value?: ChunkRequest): Message;

    hasChunkResponse(): boolean;
    clearChunkResponse(): void;
    getChunkResponse(): ChunkResponse | undefined;
    setChunkResponse(value?: ChunkResponse): Message;

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
        snapshotsRequest?: SnapshotsRequest.AsObject,
        snapshotsResponse?: SnapshotsResponse.AsObject,
        chunkRequest?: ChunkRequest.AsObject,
        chunkResponse?: ChunkResponse.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        SNAPSHOTS_REQUEST = 1,
        SNAPSHOTS_RESPONSE = 2,
        CHUNK_REQUEST = 3,
        CHUNK_RESPONSE = 4,
    }

}

export class SnapshotsRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SnapshotsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SnapshotsRequest): SnapshotsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SnapshotsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SnapshotsRequest;
    static deserializeBinaryFromReader(message: SnapshotsRequest, reader: jspb.BinaryReader): SnapshotsRequest;
}

export namespace SnapshotsRequest {
    export type AsObject = {
    }
}

export class SnapshotsResponse extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): SnapshotsResponse;
    getFormat(): number;
    setFormat(value: number): SnapshotsResponse;
    getChunks(): number;
    setChunks(value: number): SnapshotsResponse;
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): SnapshotsResponse;
    getMetadata(): Uint8Array | string;
    getMetadata_asU8(): Uint8Array;
    getMetadata_asB64(): string;
    setMetadata(value: Uint8Array | string): SnapshotsResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SnapshotsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SnapshotsResponse): SnapshotsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SnapshotsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SnapshotsResponse;
    static deserializeBinaryFromReader(message: SnapshotsResponse, reader: jspb.BinaryReader): SnapshotsResponse;
}

export namespace SnapshotsResponse {
    export type AsObject = {
        height: number,
        format: number,
        chunks: number,
        hash: Uint8Array | string,
        metadata: Uint8Array | string,
    }
}

export class ChunkRequest extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): ChunkRequest;
    getFormat(): number;
    setFormat(value: number): ChunkRequest;
    getIndex(): number;
    setIndex(value: number): ChunkRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ChunkRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ChunkRequest): ChunkRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ChunkRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ChunkRequest;
    static deserializeBinaryFromReader(message: ChunkRequest, reader: jspb.BinaryReader): ChunkRequest;
}

export namespace ChunkRequest {
    export type AsObject = {
        height: number,
        format: number,
        index: number,
    }
}

export class ChunkResponse extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): ChunkResponse;
    getFormat(): number;
    setFormat(value: number): ChunkResponse;
    getIndex(): number;
    setIndex(value: number): ChunkResponse;
    getChunk(): Uint8Array | string;
    getChunk_asU8(): Uint8Array;
    getChunk_asB64(): string;
    setChunk(value: Uint8Array | string): ChunkResponse;
    getMissing(): boolean;
    setMissing(value: boolean): ChunkResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ChunkResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ChunkResponse): ChunkResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ChunkResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ChunkResponse;
    static deserializeBinaryFromReader(message: ChunkResponse, reader: jspb.BinaryReader): ChunkResponse;
}

export namespace ChunkResponse {
    export type AsObject = {
        height: number,
        format: number,
        index: number,
        chunk: Uint8Array | string,
        missing: boolean,
    }
}
