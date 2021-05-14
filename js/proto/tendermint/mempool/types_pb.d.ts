// package: tendermint.mempool
// file: tendermint/mempool/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class Txs extends jspb.Message { 
    clearTxsList(): void;
    getTxsList(): Array<Uint8Array | string>;
    getTxsList_asU8(): Array<Uint8Array>;
    getTxsList_asB64(): Array<string>;
    setTxsList(value: Array<Uint8Array | string>): Txs;
    addTxs(value: Uint8Array | string, index?: number): Uint8Array | string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Txs.AsObject;
    static toObject(includeInstance: boolean, msg: Txs): Txs.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Txs, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Txs;
    static deserializeBinaryFromReader(message: Txs, reader: jspb.BinaryReader): Txs;
}

export namespace Txs {
    export type AsObject = {
        txsList: Array<Uint8Array | string>,
    }
}

export class Message extends jspb.Message { 

    hasTxs(): boolean;
    clearTxs(): void;
    getTxs(): Txs | undefined;
    setTxs(value?: Txs): Message;

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
        txs?: Txs.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        TXS = 1,
    }

}
