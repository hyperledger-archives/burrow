// package: tendermint.blockchain
// file: tendermint/blockchain/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as tendermint_types_block_pb from "../../tendermint/types/block_pb";

export class BlockRequest extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): BlockRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockRequest.AsObject;
    static toObject(includeInstance: boolean, msg: BlockRequest): BlockRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockRequest;
    static deserializeBinaryFromReader(message: BlockRequest, reader: jspb.BinaryReader): BlockRequest;
}

export namespace BlockRequest {
    export type AsObject = {
        height: number,
    }
}

export class NoBlockResponse extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): NoBlockResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NoBlockResponse.AsObject;
    static toObject(includeInstance: boolean, msg: NoBlockResponse): NoBlockResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NoBlockResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NoBlockResponse;
    static deserializeBinaryFromReader(message: NoBlockResponse, reader: jspb.BinaryReader): NoBlockResponse;
}

export namespace NoBlockResponse {
    export type AsObject = {
        height: number,
    }
}

export class BlockResponse extends jspb.Message { 

    hasBlock(): boolean;
    clearBlock(): void;
    getBlock(): tendermint_types_block_pb.Block | undefined;
    setBlock(value?: tendermint_types_block_pb.Block): BlockResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockResponse.AsObject;
    static toObject(includeInstance: boolean, msg: BlockResponse): BlockResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockResponse;
    static deserializeBinaryFromReader(message: BlockResponse, reader: jspb.BinaryReader): BlockResponse;
}

export namespace BlockResponse {
    export type AsObject = {
        block?: tendermint_types_block_pb.Block.AsObject,
    }
}

export class StatusRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StatusRequest.AsObject;
    static toObject(includeInstance: boolean, msg: StatusRequest): StatusRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StatusRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StatusRequest;
    static deserializeBinaryFromReader(message: StatusRequest, reader: jspb.BinaryReader): StatusRequest;
}

export namespace StatusRequest {
    export type AsObject = {
    }
}

export class StatusResponse extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): StatusResponse;

    getBase(): number;
    setBase(value: number): StatusResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StatusResponse.AsObject;
    static toObject(includeInstance: boolean, msg: StatusResponse): StatusResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StatusResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StatusResponse;
    static deserializeBinaryFromReader(message: StatusResponse, reader: jspb.BinaryReader): StatusResponse;
}

export namespace StatusResponse {
    export type AsObject = {
        height: number,
        base: number,
    }
}

export class Message extends jspb.Message { 

    hasBlockRequest(): boolean;
    clearBlockRequest(): void;
    getBlockRequest(): BlockRequest | undefined;
    setBlockRequest(value?: BlockRequest): Message;


    hasNoBlockResponse(): boolean;
    clearNoBlockResponse(): void;
    getNoBlockResponse(): NoBlockResponse | undefined;
    setNoBlockResponse(value?: NoBlockResponse): Message;


    hasBlockResponse(): boolean;
    clearBlockResponse(): void;
    getBlockResponse(): BlockResponse | undefined;
    setBlockResponse(value?: BlockResponse): Message;


    hasStatusRequest(): boolean;
    clearStatusRequest(): void;
    getStatusRequest(): StatusRequest | undefined;
    setStatusRequest(value?: StatusRequest): Message;


    hasStatusResponse(): boolean;
    clearStatusResponse(): void;
    getStatusResponse(): StatusResponse | undefined;
    setStatusResponse(value?: StatusResponse): Message;


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
        blockRequest?: BlockRequest.AsObject,
        noBlockResponse?: NoBlockResponse.AsObject,
        blockResponse?: BlockResponse.AsObject,
        statusRequest?: StatusRequest.AsObject,
        statusResponse?: StatusResponse.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
    
    BLOCK_REQUEST = 1,

    NO_BLOCK_RESPONSE = 2,

    BLOCK_RESPONSE = 3,

    STATUS_REQUEST = 4,

    STATUS_RESPONSE = 5,

    }

}
