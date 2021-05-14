// package: rpcevents
// file: rpcevents.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as exec_pb from "./exec_pb";

export class GetBlockRequest extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): GetBlockRequest;
    getWait(): boolean;
    setWait(value: boolean): GetBlockRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetBlockRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetBlockRequest): GetBlockRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetBlockRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetBlockRequest;
    static deserializeBinaryFromReader(message: GetBlockRequest, reader: jspb.BinaryReader): GetBlockRequest;
}

export namespace GetBlockRequest {
    export type AsObject = {
        height: number,
        wait: boolean,
    }
}

export class TxRequest extends jspb.Message { 
    getTxhash(): Uint8Array | string;
    getTxhash_asU8(): Uint8Array;
    getTxhash_asB64(): string;
    setTxhash(value: Uint8Array | string): TxRequest;
    getWait(): boolean;
    setWait(value: boolean): TxRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TxRequest.AsObject;
    static toObject(includeInstance: boolean, msg: TxRequest): TxRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TxRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TxRequest;
    static deserializeBinaryFromReader(message: TxRequest, reader: jspb.BinaryReader): TxRequest;
}

export namespace TxRequest {
    export type AsObject = {
        txhash: Uint8Array | string,
        wait: boolean,
    }
}

export class BlocksRequest extends jspb.Message { 

    hasBlockrange(): boolean;
    clearBlockrange(): void;
    getBlockrange(): BlockRange | undefined;
    setBlockrange(value?: BlockRange): BlocksRequest;
    getQuery(): string;
    setQuery(value: string): BlocksRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlocksRequest.AsObject;
    static toObject(includeInstance: boolean, msg: BlocksRequest): BlocksRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlocksRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlocksRequest;
    static deserializeBinaryFromReader(message: BlocksRequest, reader: jspb.BinaryReader): BlocksRequest;
}

export namespace BlocksRequest {
    export type AsObject = {
        blockrange?: BlockRange.AsObject,
        query: string,
    }
}

export class EventsResponse extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): EventsResponse;
    clearEventsList(): void;
    getEventsList(): Array<exec_pb.Event>;
    setEventsList(value: Array<exec_pb.Event>): EventsResponse;
    addEvents(value?: exec_pb.Event, index?: number): exec_pb.Event;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EventsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: EventsResponse): EventsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EventsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EventsResponse;
    static deserializeBinaryFromReader(message: EventsResponse, reader: jspb.BinaryReader): EventsResponse;
}

export namespace EventsResponse {
    export type AsObject = {
        height: number,
        eventsList: Array<exec_pb.Event.AsObject>,
    }
}

export class GetTxsRequest extends jspb.Message { 
    getStartheight(): number;
    setStartheight(value: number): GetTxsRequest;
    getEndheight(): number;
    setEndheight(value: number): GetTxsRequest;
    getQuery(): string;
    setQuery(value: string): GetTxsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetTxsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetTxsRequest): GetTxsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetTxsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetTxsRequest;
    static deserializeBinaryFromReader(message: GetTxsRequest, reader: jspb.BinaryReader): GetTxsRequest;
}

export namespace GetTxsRequest {
    export type AsObject = {
        startheight: number,
        endheight: number,
        query: string,
    }
}

export class GetTxsResponse extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): GetTxsResponse;
    clearTxexecutionsList(): void;
    getTxexecutionsList(): Array<exec_pb.TxExecution>;
    setTxexecutionsList(value: Array<exec_pb.TxExecution>): GetTxsResponse;
    addTxexecutions(value?: exec_pb.TxExecution, index?: number): exec_pb.TxExecution;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetTxsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetTxsResponse): GetTxsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetTxsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetTxsResponse;
    static deserializeBinaryFromReader(message: GetTxsResponse, reader: jspb.BinaryReader): GetTxsResponse;
}

export namespace GetTxsResponse {
    export type AsObject = {
        height: number,
        txexecutionsList: Array<exec_pb.TxExecution.AsObject>,
    }
}

export class Bound extends jspb.Message { 
    getType(): Bound.BoundType;
    setType(value: Bound.BoundType): Bound;
    getIndex(): number;
    setIndex(value: number): Bound;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Bound.AsObject;
    static toObject(includeInstance: boolean, msg: Bound): Bound.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Bound, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Bound;
    static deserializeBinaryFromReader(message: Bound, reader: jspb.BinaryReader): Bound;
}

export namespace Bound {
    export type AsObject = {
        type: Bound.BoundType,
        index: number,
    }

    export enum BoundType {
    ABSOLUTE = 0,
    RELATIVE = 1,
    FIRST = 2,
    LATEST = 3,
    STREAM = 4,
    }

}

export class BlockRange extends jspb.Message { 

    hasStart(): boolean;
    clearStart(): void;
    getStart(): Bound | undefined;
    setStart(value?: Bound): BlockRange;

    hasEnd(): boolean;
    clearEnd(): void;
    getEnd(): Bound | undefined;
    setEnd(value?: Bound): BlockRange;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockRange.AsObject;
    static toObject(includeInstance: boolean, msg: BlockRange): BlockRange.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockRange, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockRange;
    static deserializeBinaryFromReader(message: BlockRange, reader: jspb.BinaryReader): BlockRange;
}

export namespace BlockRange {
    export type AsObject = {
        start?: Bound.AsObject,
        end?: Bound.AsObject,
    }
}
