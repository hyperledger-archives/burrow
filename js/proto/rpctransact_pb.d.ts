// package: rpctransact
// file: rpctransact.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as exec_pb from "./exec_pb";
import * as payload_pb from "./payload_pb";
import * as txs_pb from "./txs_pb";

export class CallCodeParam extends jspb.Message { 
    getFromaddress(): Uint8Array | string;
    getFromaddress_asU8(): Uint8Array;
    getFromaddress_asB64(): string;
    setFromaddress(value: Uint8Array | string): void;

    getCode(): Uint8Array | string;
    getCode_asU8(): Uint8Array;
    getCode_asB64(): string;
    setCode(value: Uint8Array | string): void;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CallCodeParam.AsObject;
    static toObject(includeInstance: boolean, msg: CallCodeParam): CallCodeParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CallCodeParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CallCodeParam;
    static deserializeBinaryFromReader(message: CallCodeParam, reader: jspb.BinaryReader): CallCodeParam;
}

export namespace CallCodeParam {
    export type AsObject = {
        fromaddress: Uint8Array | string,
        code: Uint8Array | string,
        data: Uint8Array | string,
    }
}

export class TxEnvelope extends jspb.Message { 

    hasEnvelope(): boolean;
    clearEnvelope(): void;
    getEnvelope(): txs_pb.Envelope | undefined;
    setEnvelope(value?: txs_pb.Envelope): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TxEnvelope.AsObject;
    static toObject(includeInstance: boolean, msg: TxEnvelope): TxEnvelope.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TxEnvelope, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TxEnvelope;
    static deserializeBinaryFromReader(message: TxEnvelope, reader: jspb.BinaryReader): TxEnvelope;
}

export namespace TxEnvelope {
    export type AsObject = {
        envelope?: txs_pb.Envelope.AsObject,
    }
}

export class TxEnvelopeParam extends jspb.Message { 

    hasEnvelope(): boolean;
    clearEnvelope(): void;
    getEnvelope(): txs_pb.Envelope | undefined;
    setEnvelope(value?: txs_pb.Envelope): void;


    hasPayload(): boolean;
    clearPayload(): void;
    getPayload(): payload_pb.Any | undefined;
    setPayload(value?: payload_pb.Any): void;


    hasTimeout(): boolean;
    clearTimeout(): void;
    getTimeout(): google_protobuf_duration_pb.Duration | undefined;
    setTimeout(value?: google_protobuf_duration_pb.Duration): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TxEnvelopeParam.AsObject;
    static toObject(includeInstance: boolean, msg: TxEnvelopeParam): TxEnvelopeParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TxEnvelopeParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TxEnvelopeParam;
    static deserializeBinaryFromReader(message: TxEnvelopeParam, reader: jspb.BinaryReader): TxEnvelopeParam;
}

export namespace TxEnvelopeParam {
    export type AsObject = {
        envelope?: txs_pb.Envelope.AsObject,
        payload?: payload_pb.Any.AsObject,
        timeout?: google_protobuf_duration_pb.Duration.AsObject,
    }
}
