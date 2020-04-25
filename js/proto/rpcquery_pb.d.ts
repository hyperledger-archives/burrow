// package: rpcquery
// file: rpcquery.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as github_com_tendermint_tendermint_abci_types_types_pb from "./github.com/tendermint/tendermint/abci/types/types_pb";
import * as names_pb from "./names_pb";
import * as acm_pb from "./acm_pb";
import * as validator_pb from "./validator_pb";
import * as registry_pb from "./registry_pb";
import * as rpc_pb from "./rpc_pb";
import * as payload_pb from "./payload_pb";

export class StatusParam extends jspb.Message { 
    getBlocktimewithin(): string;
    setBlocktimewithin(value: string): void;

    getBlockseentimewithin(): string;
    setBlockseentimewithin(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StatusParam.AsObject;
    static toObject(includeInstance: boolean, msg: StatusParam): StatusParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StatusParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StatusParam;
    static deserializeBinaryFromReader(message: StatusParam, reader: jspb.BinaryReader): StatusParam;
}

export namespace StatusParam {
    export type AsObject = {
        blocktimewithin: string,
        blockseentimewithin: string,
    }
}

export class GetAccountParam extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetAccountParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetAccountParam): GetAccountParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetAccountParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetAccountParam;
    static deserializeBinaryFromReader(message: GetAccountParam, reader: jspb.BinaryReader): GetAccountParam;
}

export namespace GetAccountParam {
    export type AsObject = {
        address: Uint8Array | string,
    }
}

export class GetMetadataParam extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;

    getMetadatahash(): Uint8Array | string;
    getMetadatahash_asU8(): Uint8Array;
    getMetadatahash_asB64(): string;
    setMetadatahash(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetMetadataParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetMetadataParam): GetMetadataParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetMetadataParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetMetadataParam;
    static deserializeBinaryFromReader(message: GetMetadataParam, reader: jspb.BinaryReader): GetMetadataParam;
}

export namespace GetMetadataParam {
    export type AsObject = {
        address: Uint8Array | string,
        metadatahash: Uint8Array | string,
    }
}

export class MetadataResult extends jspb.Message { 
    getMetadata(): string;
    setMetadata(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MetadataResult.AsObject;
    static toObject(includeInstance: boolean, msg: MetadataResult): MetadataResult.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: MetadataResult, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MetadataResult;
    static deserializeBinaryFromReader(message: MetadataResult, reader: jspb.BinaryReader): MetadataResult;
}

export namespace MetadataResult {
    export type AsObject = {
        metadata: string,
    }
}

export class GetStorageParam extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;

    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetStorageParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetStorageParam): GetStorageParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetStorageParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetStorageParam;
    static deserializeBinaryFromReader(message: GetStorageParam, reader: jspb.BinaryReader): GetStorageParam;
}

export namespace GetStorageParam {
    export type AsObject = {
        address: Uint8Array | string,
        key: Uint8Array | string,
    }
}

export class StorageValue extends jspb.Message { 
    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StorageValue.AsObject;
    static toObject(includeInstance: boolean, msg: StorageValue): StorageValue.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: StorageValue, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StorageValue;
    static deserializeBinaryFromReader(message: StorageValue, reader: jspb.BinaryReader): StorageValue;
}

export namespace StorageValue {
    export type AsObject = {
        value: Uint8Array | string,
    }
}

export class ListAccountsParam extends jspb.Message { 
    getQuery(): string;
    setQuery(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListAccountsParam.AsObject;
    static toObject(includeInstance: boolean, msg: ListAccountsParam): ListAccountsParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListAccountsParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListAccountsParam;
    static deserializeBinaryFromReader(message: ListAccountsParam, reader: jspb.BinaryReader): ListAccountsParam;
}

export namespace ListAccountsParam {
    export type AsObject = {
        query: string,
    }
}

export class GetNameParam extends jspb.Message { 
    getName(): string;
    setName(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetNameParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetNameParam): GetNameParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetNameParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetNameParam;
    static deserializeBinaryFromReader(message: GetNameParam, reader: jspb.BinaryReader): GetNameParam;
}

export namespace GetNameParam {
    export type AsObject = {
        name: string,
    }
}

export class ListNamesParam extends jspb.Message { 
    getQuery(): string;
    setQuery(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListNamesParam.AsObject;
    static toObject(includeInstance: boolean, msg: ListNamesParam): ListNamesParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListNamesParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListNamesParam;
    static deserializeBinaryFromReader(message: ListNamesParam, reader: jspb.BinaryReader): ListNamesParam;
}

export namespace ListNamesParam {
    export type AsObject = {
        query: string,
    }
}

export class GetNetworkRegistryParam extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetNetworkRegistryParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetNetworkRegistryParam): GetNetworkRegistryParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetNetworkRegistryParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetNetworkRegistryParam;
    static deserializeBinaryFromReader(message: GetNetworkRegistryParam, reader: jspb.BinaryReader): GetNetworkRegistryParam;
}

export namespace GetNetworkRegistryParam {
    export type AsObject = {
    }
}

export class GetValidatorSetParam extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetValidatorSetParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetValidatorSetParam): GetValidatorSetParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetValidatorSetParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetValidatorSetParam;
    static deserializeBinaryFromReader(message: GetValidatorSetParam, reader: jspb.BinaryReader): GetValidatorSetParam;
}

export namespace GetValidatorSetParam {
    export type AsObject = {
    }
}

export class GetValidatorSetHistoryParam extends jspb.Message { 
    getIncludeprevious(): number;
    setIncludeprevious(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetValidatorSetHistoryParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetValidatorSetHistoryParam): GetValidatorSetHistoryParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetValidatorSetHistoryParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetValidatorSetHistoryParam;
    static deserializeBinaryFromReader(message: GetValidatorSetHistoryParam, reader: jspb.BinaryReader): GetValidatorSetHistoryParam;
}

export namespace GetValidatorSetHistoryParam {
    export type AsObject = {
        includeprevious: number,
    }
}

export class NetworkRegistry extends jspb.Message { 
    clearSetList(): void;
    getSetList(): Array<RegisteredValidator>;
    setSetList(value: Array<RegisteredValidator>): void;
    addSet(value?: RegisteredValidator, index?: number): RegisteredValidator;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NetworkRegistry.AsObject;
    static toObject(includeInstance: boolean, msg: NetworkRegistry): NetworkRegistry.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NetworkRegistry, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NetworkRegistry;
    static deserializeBinaryFromReader(message: NetworkRegistry, reader: jspb.BinaryReader): NetworkRegistry;
}

export namespace NetworkRegistry {
    export type AsObject = {
        setList: Array<RegisteredValidator.AsObject>,
    }
}

export class RegisteredValidator extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;


    hasNode(): boolean;
    clearNode(): void;
    getNode(): registry_pb.NodeIdentity | undefined;
    setNode(value?: registry_pb.NodeIdentity): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RegisteredValidator.AsObject;
    static toObject(includeInstance: boolean, msg: RegisteredValidator): RegisteredValidator.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RegisteredValidator, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RegisteredValidator;
    static deserializeBinaryFromReader(message: RegisteredValidator, reader: jspb.BinaryReader): RegisteredValidator;
}

export namespace RegisteredValidator {
    export type AsObject = {
        address: Uint8Array | string,
        node?: registry_pb.NodeIdentity.AsObject,
    }
}

export class ValidatorSetHistory extends jspb.Message { 
    clearHistoryList(): void;
    getHistoryList(): Array<ValidatorSet>;
    setHistoryList(value: Array<ValidatorSet>): void;
    addHistory(value?: ValidatorSet, index?: number): ValidatorSet;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ValidatorSetHistory.AsObject;
    static toObject(includeInstance: boolean, msg: ValidatorSetHistory): ValidatorSetHistory.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ValidatorSetHistory, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ValidatorSetHistory;
    static deserializeBinaryFromReader(message: ValidatorSetHistory, reader: jspb.BinaryReader): ValidatorSetHistory;
}

export namespace ValidatorSetHistory {
    export type AsObject = {
        historyList: Array<ValidatorSet.AsObject>,
    }
}

export class ValidatorSet extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): void;

    clearSetList(): void;
    getSetList(): Array<validator_pb.Validator>;
    setSetList(value: Array<validator_pb.Validator>): void;
    addSet(value?: validator_pb.Validator, index?: number): validator_pb.Validator;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ValidatorSet.AsObject;
    static toObject(includeInstance: boolean, msg: ValidatorSet): ValidatorSet.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ValidatorSet, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ValidatorSet;
    static deserializeBinaryFromReader(message: ValidatorSet, reader: jspb.BinaryReader): ValidatorSet;
}

export namespace ValidatorSet {
    export type AsObject = {
        height: number,
        setList: Array<validator_pb.Validator.AsObject>,
    }
}

export class GetProposalParam extends jspb.Message { 
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetProposalParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetProposalParam): GetProposalParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetProposalParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetProposalParam;
    static deserializeBinaryFromReader(message: GetProposalParam, reader: jspb.BinaryReader): GetProposalParam;
}

export namespace GetProposalParam {
    export type AsObject = {
        hash: Uint8Array | string,
    }
}

export class ListProposalsParam extends jspb.Message { 
    getProposed(): boolean;
    setProposed(value: boolean): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListProposalsParam.AsObject;
    static toObject(includeInstance: boolean, msg: ListProposalsParam): ListProposalsParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListProposalsParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListProposalsParam;
    static deserializeBinaryFromReader(message: ListProposalsParam, reader: jspb.BinaryReader): ListProposalsParam;
}

export namespace ListProposalsParam {
    export type AsObject = {
        proposed: boolean,
    }
}

export class ProposalResult extends jspb.Message { 
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): void;


    hasBallot(): boolean;
    clearBallot(): void;
    getBallot(): payload_pb.Ballot | undefined;
    setBallot(value?: payload_pb.Ballot): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProposalResult.AsObject;
    static toObject(includeInstance: boolean, msg: ProposalResult): ProposalResult.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProposalResult, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProposalResult;
    static deserializeBinaryFromReader(message: ProposalResult, reader: jspb.BinaryReader): ProposalResult;
}

export namespace ProposalResult {
    export type AsObject = {
        hash: Uint8Array | string,
        ballot?: payload_pb.Ballot.AsObject,
    }
}

export class GetStatsParam extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetStatsParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetStatsParam): GetStatsParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetStatsParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetStatsParam;
    static deserializeBinaryFromReader(message: GetStatsParam, reader: jspb.BinaryReader): GetStatsParam;
}

export namespace GetStatsParam {
    export type AsObject = {
    }
}

export class Stats extends jspb.Message { 
    getAccountswithcode(): number;
    setAccountswithcode(value: number): void;

    getAccountswithoutcode(): number;
    setAccountswithoutcode(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Stats.AsObject;
    static toObject(includeInstance: boolean, msg: Stats): Stats.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Stats, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Stats;
    static deserializeBinaryFromReader(message: Stats, reader: jspb.BinaryReader): Stats;
}

export namespace Stats {
    export type AsObject = {
        accountswithcode: number,
        accountswithoutcode: number,
    }
}

export class GetBlockParam extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetBlockParam.AsObject;
    static toObject(includeInstance: boolean, msg: GetBlockParam): GetBlockParam.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetBlockParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetBlockParam;
    static deserializeBinaryFromReader(message: GetBlockParam, reader: jspb.BinaryReader): GetBlockParam;
}

export namespace GetBlockParam {
    export type AsObject = {
        height: number,
    }
}
