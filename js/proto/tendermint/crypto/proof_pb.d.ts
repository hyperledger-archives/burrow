// package: tendermint.crypto
// file: tendermint/crypto/proof.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";

export class Proof extends jspb.Message { 
    getTotal(): number;
    setTotal(value: number): Proof;

    getIndex(): number;
    setIndex(value: number): Proof;

    getLeafHash(): Uint8Array | string;
    getLeafHash_asU8(): Uint8Array;
    getLeafHash_asB64(): string;
    setLeafHash(value: Uint8Array | string): Proof;

    clearAuntsList(): void;
    getAuntsList(): Array<Uint8Array | string>;
    getAuntsList_asU8(): Array<Uint8Array>;
    getAuntsList_asB64(): Array<string>;
    setAuntsList(value: Array<Uint8Array | string>): Proof;
    addAunts(value: Uint8Array | string, index?: number): Uint8Array | string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Proof.AsObject;
    static toObject(includeInstance: boolean, msg: Proof): Proof.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Proof, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Proof;
    static deserializeBinaryFromReader(message: Proof, reader: jspb.BinaryReader): Proof;
}

export namespace Proof {
    export type AsObject = {
        total: number,
        index: number,
        leafHash: Uint8Array | string,
        auntsList: Array<Uint8Array | string>,
    }
}

export class ValueOp extends jspb.Message { 
    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): ValueOp;


    hasProof(): boolean;
    clearProof(): void;
    getProof(): Proof | undefined;
    setProof(value?: Proof): ValueOp;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ValueOp.AsObject;
    static toObject(includeInstance: boolean, msg: ValueOp): ValueOp.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ValueOp, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ValueOp;
    static deserializeBinaryFromReader(message: ValueOp, reader: jspb.BinaryReader): ValueOp;
}

export namespace ValueOp {
    export type AsObject = {
        key: Uint8Array | string,
        proof?: Proof.AsObject,
    }
}

export class DominoOp extends jspb.Message { 
    getKey(): string;
    setKey(value: string): DominoOp;

    getInput(): string;
    setInput(value: string): DominoOp;

    getOutput(): string;
    setOutput(value: string): DominoOp;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DominoOp.AsObject;
    static toObject(includeInstance: boolean, msg: DominoOp): DominoOp.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DominoOp, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DominoOp;
    static deserializeBinaryFromReader(message: DominoOp, reader: jspb.BinaryReader): DominoOp;
}

export namespace DominoOp {
    export type AsObject = {
        key: string,
        input: string,
        output: string,
    }
}

export class ProofOp extends jspb.Message { 
    getType(): string;
    setType(value: string): ProofOp;

    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): ProofOp;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): ProofOp;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProofOp.AsObject;
    static toObject(includeInstance: boolean, msg: ProofOp): ProofOp.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProofOp, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProofOp;
    static deserializeBinaryFromReader(message: ProofOp, reader: jspb.BinaryReader): ProofOp;
}

export namespace ProofOp {
    export type AsObject = {
        type: string,
        key: Uint8Array | string,
        data: Uint8Array | string,
    }
}

export class ProofOps extends jspb.Message { 
    clearOpsList(): void;
    getOpsList(): Array<ProofOp>;
    setOpsList(value: Array<ProofOp>): ProofOps;
    addOps(value?: ProofOp, index?: number): ProofOp;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProofOps.AsObject;
    static toObject(includeInstance: boolean, msg: ProofOps): ProofOps.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProofOps, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProofOps;
    static deserializeBinaryFromReader(message: ProofOps, reader: jspb.BinaryReader): ProofOps;
}

export namespace ProofOps {
    export type AsObject = {
        opsList: Array<ProofOp.AsObject>,
    }
}
