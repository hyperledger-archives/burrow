// package: dump
// file: dump.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as acm_pb from "./acm_pb";
import * as exec_pb from "./exec_pb";
import * as names_pb from "./names_pb";

export class Storage extends jspb.Message { 
    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): void;

    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Storage.AsObject;
    static toObject(includeInstance: boolean, msg: Storage): Storage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Storage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Storage;
    static deserializeBinaryFromReader(message: Storage, reader: jspb.BinaryReader): Storage;
}

export namespace Storage {
    export type AsObject = {
        key: Uint8Array | string,
        value: Uint8Array | string,
    }
}

export class AccountStorage extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;

    clearStorageList(): void;
    getStorageList(): Array<Storage>;
    setStorageList(value: Array<Storage>): void;
    addStorage(value?: Storage, index?: number): Storage;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AccountStorage.AsObject;
    static toObject(includeInstance: boolean, msg: AccountStorage): AccountStorage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AccountStorage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AccountStorage;
    static deserializeBinaryFromReader(message: AccountStorage, reader: jspb.BinaryReader): AccountStorage;
}

export namespace AccountStorage {
    export type AsObject = {
        address: Uint8Array | string,
        storageList: Array<Storage.AsObject>,
    }
}

export class EVMEvent extends jspb.Message { 
    getChainid(): string;
    setChainid(value: string): void;

    getIndex(): number;
    setIndex(value: number): void;


    hasTime(): boolean;
    clearTime(): void;
    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): void;


    hasEvent(): boolean;
    clearEvent(): void;
    getEvent(): exec_pb.LogEvent | undefined;
    setEvent(value?: exec_pb.LogEvent): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EVMEvent.AsObject;
    static toObject(includeInstance: boolean, msg: EVMEvent): EVMEvent.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EVMEvent, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EVMEvent;
    static deserializeBinaryFromReader(message: EVMEvent, reader: jspb.BinaryReader): EVMEvent;
}

export namespace EVMEvent {
    export type AsObject = {
        chainid: string,
        index: number,
        time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        event?: exec_pb.LogEvent.AsObject,
    }
}

export class Dump extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): void;


    hasAccount(): boolean;
    clearAccount(): void;
    getAccount(): acm_pb.Account | undefined;
    setAccount(value?: acm_pb.Account): void;


    hasAccountstorage(): boolean;
    clearAccountstorage(): void;
    getAccountstorage(): AccountStorage | undefined;
    setAccountstorage(value?: AccountStorage): void;


    hasEvmevent(): boolean;
    clearEvmevent(): void;
    getEvmevent(): EVMEvent | undefined;
    setEvmevent(value?: EVMEvent): void;


    hasName(): boolean;
    clearName(): void;
    getName(): names_pb.Entry | undefined;
    setName(value?: names_pb.Entry): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Dump.AsObject;
    static toObject(includeInstance: boolean, msg: Dump): Dump.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Dump, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Dump;
    static deserializeBinaryFromReader(message: Dump, reader: jspb.BinaryReader): Dump;
}

export namespace Dump {
    export type AsObject = {
        height: number,
        account?: acm_pb.Account.AsObject,
        accountstorage?: AccountStorage.AsObject,
        evmevent?: EVMEvent.AsObject,
        name?: names_pb.Entry.AsObject,
    }
}
