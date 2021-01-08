// package: tendermint.rpc.grpc
// file: tendermint/rpc/grpc/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as tendermint_abci_types_pb from "../../../tendermint/abci/types_pb";

export class RequestPing extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestPing.AsObject;
    static toObject(includeInstance: boolean, msg: RequestPing): RequestPing.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestPing, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestPing;
    static deserializeBinaryFromReader(message: RequestPing, reader: jspb.BinaryReader): RequestPing;
}

export namespace RequestPing {
    export type AsObject = {
    }
}

export class RequestBroadcastTx extends jspb.Message { 
    getTx(): Uint8Array | string;
    getTx_asU8(): Uint8Array;
    getTx_asB64(): string;
    setTx(value: Uint8Array | string): RequestBroadcastTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestBroadcastTx.AsObject;
    static toObject(includeInstance: boolean, msg: RequestBroadcastTx): RequestBroadcastTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestBroadcastTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestBroadcastTx;
    static deserializeBinaryFromReader(message: RequestBroadcastTx, reader: jspb.BinaryReader): RequestBroadcastTx;
}

export namespace RequestBroadcastTx {
    export type AsObject = {
        tx: Uint8Array | string,
    }
}

export class ResponsePing extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponsePing.AsObject;
    static toObject(includeInstance: boolean, msg: ResponsePing): ResponsePing.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponsePing, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponsePing;
    static deserializeBinaryFromReader(message: ResponsePing, reader: jspb.BinaryReader): ResponsePing;
}

export namespace ResponsePing {
    export type AsObject = {
    }
}

export class ResponseBroadcastTx extends jspb.Message { 

    hasCheckTx(): boolean;
    clearCheckTx(): void;
    getCheckTx(): tendermint_abci_types_pb.ResponseCheckTx | undefined;
    setCheckTx(value?: tendermint_abci_types_pb.ResponseCheckTx): ResponseBroadcastTx;


    hasDeliverTx(): boolean;
    clearDeliverTx(): void;
    getDeliverTx(): tendermint_abci_types_pb.ResponseDeliverTx | undefined;
    setDeliverTx(value?: tendermint_abci_types_pb.ResponseDeliverTx): ResponseBroadcastTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseBroadcastTx.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseBroadcastTx): ResponseBroadcastTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseBroadcastTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseBroadcastTx;
    static deserializeBinaryFromReader(message: ResponseBroadcastTx, reader: jspb.BinaryReader): ResponseBroadcastTx;
}

export namespace ResponseBroadcastTx {
    export type AsObject = {
        checkTx?: tendermint_abci_types_pb.ResponseCheckTx.AsObject,
        deliverTx?: tendermint_abci_types_pb.ResponseDeliverTx.AsObject,
    }
}
