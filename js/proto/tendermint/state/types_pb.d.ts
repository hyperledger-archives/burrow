// package: tendermint.state
// file: tendermint/state/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as tendermint_abci_types_pb from "../../tendermint/abci/types_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as tendermint_types_validator_pb from "../../tendermint/types/validator_pb";
import * as tendermint_types_params_pb from "../../tendermint/types/params_pb";
import * as tendermint_version_types_pb from "../../tendermint/version/types_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class ABCIResponses extends jspb.Message { 
    clearDeliverTxsList(): void;
    getDeliverTxsList(): Array<tendermint_abci_types_pb.ResponseDeliverTx>;
    setDeliverTxsList(value: Array<tendermint_abci_types_pb.ResponseDeliverTx>): ABCIResponses;
    addDeliverTxs(value?: tendermint_abci_types_pb.ResponseDeliverTx, index?: number): tendermint_abci_types_pb.ResponseDeliverTx;


    hasEndBlock(): boolean;
    clearEndBlock(): void;
    getEndBlock(): tendermint_abci_types_pb.ResponseEndBlock | undefined;
    setEndBlock(value?: tendermint_abci_types_pb.ResponseEndBlock): ABCIResponses;


    hasBeginBlock(): boolean;
    clearBeginBlock(): void;
    getBeginBlock(): tendermint_abci_types_pb.ResponseBeginBlock | undefined;
    setBeginBlock(value?: tendermint_abci_types_pb.ResponseBeginBlock): ABCIResponses;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ABCIResponses.AsObject;
    static toObject(includeInstance: boolean, msg: ABCIResponses): ABCIResponses.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ABCIResponses, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ABCIResponses;
    static deserializeBinaryFromReader(message: ABCIResponses, reader: jspb.BinaryReader): ABCIResponses;
}

export namespace ABCIResponses {
    export type AsObject = {
        deliverTxsList: Array<tendermint_abci_types_pb.ResponseDeliverTx.AsObject>,
        endBlock?: tendermint_abci_types_pb.ResponseEndBlock.AsObject,
        beginBlock?: tendermint_abci_types_pb.ResponseBeginBlock.AsObject,
    }
}

export class ValidatorsInfo extends jspb.Message { 

    hasValidatorSet(): boolean;
    clearValidatorSet(): void;
    getValidatorSet(): tendermint_types_validator_pb.ValidatorSet | undefined;
    setValidatorSet(value?: tendermint_types_validator_pb.ValidatorSet): ValidatorsInfo;

    getLastHeightChanged(): number;
    setLastHeightChanged(value: number): ValidatorsInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ValidatorsInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ValidatorsInfo): ValidatorsInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ValidatorsInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ValidatorsInfo;
    static deserializeBinaryFromReader(message: ValidatorsInfo, reader: jspb.BinaryReader): ValidatorsInfo;
}

export namespace ValidatorsInfo {
    export type AsObject = {
        validatorSet?: tendermint_types_validator_pb.ValidatorSet.AsObject,
        lastHeightChanged: number,
    }
}

export class ConsensusParamsInfo extends jspb.Message { 

    hasConsensusParams(): boolean;
    clearConsensusParams(): void;
    getConsensusParams(): tendermint_types_params_pb.ConsensusParams | undefined;
    setConsensusParams(value?: tendermint_types_params_pb.ConsensusParams): ConsensusParamsInfo;

    getLastHeightChanged(): number;
    setLastHeightChanged(value: number): ConsensusParamsInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ConsensusParamsInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ConsensusParamsInfo): ConsensusParamsInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ConsensusParamsInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ConsensusParamsInfo;
    static deserializeBinaryFromReader(message: ConsensusParamsInfo, reader: jspb.BinaryReader): ConsensusParamsInfo;
}

export namespace ConsensusParamsInfo {
    export type AsObject = {
        consensusParams?: tendermint_types_params_pb.ConsensusParams.AsObject,
        lastHeightChanged: number,
    }
}

export class Version extends jspb.Message { 

    hasConsensus(): boolean;
    clearConsensus(): void;
    getConsensus(): tendermint_version_types_pb.Consensus | undefined;
    setConsensus(value?: tendermint_version_types_pb.Consensus): Version;

    getSoftware(): string;
    setSoftware(value: string): Version;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Version.AsObject;
    static toObject(includeInstance: boolean, msg: Version): Version.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Version, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Version;
    static deserializeBinaryFromReader(message: Version, reader: jspb.BinaryReader): Version;
}

export namespace Version {
    export type AsObject = {
        consensus?: tendermint_version_types_pb.Consensus.AsObject,
        software: string,
    }
}

export class State extends jspb.Message { 

    hasVersion(): boolean;
    clearVersion(): void;
    getVersion(): Version | undefined;
    setVersion(value?: Version): State;

    getChainId(): string;
    setChainId(value: string): State;

    getInitialHeight(): number;
    setInitialHeight(value: number): State;

    getLastBlockHeight(): number;
    setLastBlockHeight(value: number): State;


    hasLastBlockId(): boolean;
    clearLastBlockId(): void;
    getLastBlockId(): tendermint_types_types_pb.BlockID | undefined;
    setLastBlockId(value?: tendermint_types_types_pb.BlockID): State;


    hasLastBlockTime(): boolean;
    clearLastBlockTime(): void;
    getLastBlockTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setLastBlockTime(value?: google_protobuf_timestamp_pb.Timestamp): State;


    hasNextValidators(): boolean;
    clearNextValidators(): void;
    getNextValidators(): tendermint_types_validator_pb.ValidatorSet | undefined;
    setNextValidators(value?: tendermint_types_validator_pb.ValidatorSet): State;


    hasValidators(): boolean;
    clearValidators(): void;
    getValidators(): tendermint_types_validator_pb.ValidatorSet | undefined;
    setValidators(value?: tendermint_types_validator_pb.ValidatorSet): State;


    hasLastValidators(): boolean;
    clearLastValidators(): void;
    getLastValidators(): tendermint_types_validator_pb.ValidatorSet | undefined;
    setLastValidators(value?: tendermint_types_validator_pb.ValidatorSet): State;

    getLastHeightValidatorsChanged(): number;
    setLastHeightValidatorsChanged(value: number): State;


    hasConsensusParams(): boolean;
    clearConsensusParams(): void;
    getConsensusParams(): tendermint_types_params_pb.ConsensusParams | undefined;
    setConsensusParams(value?: tendermint_types_params_pb.ConsensusParams): State;

    getLastHeightConsensusParamsChanged(): number;
    setLastHeightConsensusParamsChanged(value: number): State;

    getLastResultsHash(): Uint8Array | string;
    getLastResultsHash_asU8(): Uint8Array;
    getLastResultsHash_asB64(): string;
    setLastResultsHash(value: Uint8Array | string): State;

    getAppHash(): Uint8Array | string;
    getAppHash_asU8(): Uint8Array;
    getAppHash_asB64(): string;
    setAppHash(value: Uint8Array | string): State;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): State.AsObject;
    static toObject(includeInstance: boolean, msg: State): State.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: State, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): State;
    static deserializeBinaryFromReader(message: State, reader: jspb.BinaryReader): State;
}

export namespace State {
    export type AsObject = {
        version?: Version.AsObject,
        chainId: string,
        initialHeight: number,
        lastBlockHeight: number,
        lastBlockId?: tendermint_types_types_pb.BlockID.AsObject,
        lastBlockTime?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        nextValidators?: tendermint_types_validator_pb.ValidatorSet.AsObject,
        validators?: tendermint_types_validator_pb.ValidatorSet.AsObject,
        lastValidators?: tendermint_types_validator_pb.ValidatorSet.AsObject,
        lastHeightValidatorsChanged: number,
        consensusParams?: tendermint_types_params_pb.ConsensusParams.AsObject,
        lastHeightConsensusParamsChanged: number,
        lastResultsHash: Uint8Array | string,
        appHash: Uint8Array | string,
    }
}
