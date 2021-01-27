// package: rpc
// file: rpc.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as tendermint_pb from "./tendermint_pb";
import * as validator_pb from "./validator_pb";
import * as bcm_pb from "./bcm_pb";

export class ResultStatus extends jspb.Message { 
    getChainid(): string;
    setChainid(value: string): ResultStatus;

    getRunid(): string;
    setRunid(value: string): ResultStatus;

    getBurrowversion(): string;
    setBurrowversion(value: string): ResultStatus;

    getGenesishash(): Uint8Array | string;
    getGenesishash_asU8(): Uint8Array;
    getGenesishash_asB64(): string;
    setGenesishash(value: Uint8Array | string): ResultStatus;


    hasNodeinfo(): boolean;
    clearNodeinfo(): void;
    getNodeinfo(): tendermint_pb.NodeInfo | undefined;
    setNodeinfo(value?: tendermint_pb.NodeInfo): ResultStatus;


    hasSyncinfo(): boolean;
    clearSyncinfo(): void;
    getSyncinfo(): bcm_pb.SyncInfo | undefined;
    setSyncinfo(value?: bcm_pb.SyncInfo): ResultStatus;

    getCatchingup(): boolean;
    setCatchingup(value: boolean): ResultStatus;


    hasValidatorinfo(): boolean;
    clearValidatorinfo(): void;
    getValidatorinfo(): validator_pb.Validator | undefined;
    setValidatorinfo(value?: validator_pb.Validator): ResultStatus;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResultStatus.AsObject;
    static toObject(includeInstance: boolean, msg: ResultStatus): ResultStatus.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResultStatus, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResultStatus;
    static deserializeBinaryFromReader(message: ResultStatus, reader: jspb.BinaryReader): ResultStatus;
}

export namespace ResultStatus {
    export type AsObject = {
        chainid: string,
        runid: string,
        burrowversion: string,
        genesishash: Uint8Array | string,
        nodeinfo?: tendermint_pb.NodeInfo.AsObject,
        syncinfo?: bcm_pb.SyncInfo.AsObject,
        catchingup: boolean,
        validatorinfo?: validator_pb.Validator.AsObject,
    }
}
