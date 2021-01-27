// package: tendermint.store
// file: tendermint/store/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class BlockStoreState extends jspb.Message { 
    getBase(): number;
    setBase(value: number): BlockStoreState;

    getHeight(): number;
    setHeight(value: number): BlockStoreState;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockStoreState.AsObject;
    static toObject(includeInstance: boolean, msg: BlockStoreState): BlockStoreState.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockStoreState, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockStoreState;
    static deserializeBinaryFromReader(message: BlockStoreState, reader: jspb.BinaryReader): BlockStoreState;
}

export namespace BlockStoreState {
    export type AsObject = {
        base: number,
        height: number,
    }
}
