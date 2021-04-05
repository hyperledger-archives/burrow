// package: encoding
// file: encoding.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";

export class TestMessage extends jspb.Message { 
    getType(): number;
    setType(value: number): TestMessage;

    getAmount(): number;
    setAmount(value: number): TestMessage;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TestMessage.AsObject;
    static toObject(includeInstance: boolean, msg: TestMessage): TestMessage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TestMessage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TestMessage;
    static deserializeBinaryFromReader(message: TestMessage, reader: jspb.BinaryReader): TestMessage;
}

export namespace TestMessage {
    export type AsObject = {
        type: number,
        amount: number,
    }
}
