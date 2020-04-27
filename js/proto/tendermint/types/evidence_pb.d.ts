// package: tendermint.types
// file: tendermint/types/evidence.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as tendermint_types_validator_pb from "../../tendermint/types/validator_pb";

export class Evidence extends jspb.Message { 

    hasDuplicateVoteEvidence(): boolean;
    clearDuplicateVoteEvidence(): void;
    getDuplicateVoteEvidence(): DuplicateVoteEvidence | undefined;
    setDuplicateVoteEvidence(value?: DuplicateVoteEvidence): Evidence;

    hasLightClientAttackEvidence(): boolean;
    clearLightClientAttackEvidence(): void;
    getLightClientAttackEvidence(): LightClientAttackEvidence | undefined;
    setLightClientAttackEvidence(value?: LightClientAttackEvidence): Evidence;

    getSumCase(): Evidence.SumCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Evidence.AsObject;
    static toObject(includeInstance: boolean, msg: Evidence): Evidence.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Evidence, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Evidence;
    static deserializeBinaryFromReader(message: Evidence, reader: jspb.BinaryReader): Evidence;
}

export namespace Evidence {
    export type AsObject = {
        duplicateVoteEvidence?: DuplicateVoteEvidence.AsObject,
        lightClientAttackEvidence?: LightClientAttackEvidence.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        DUPLICATE_VOTE_EVIDENCE = 1,
        LIGHT_CLIENT_ATTACK_EVIDENCE = 2,
    }

}

export class DuplicateVoteEvidence extends jspb.Message { 

    hasVoteA(): boolean;
    clearVoteA(): void;
    getVoteA(): tendermint_types_types_pb.Vote | undefined;
    setVoteA(value?: tendermint_types_types_pb.Vote): DuplicateVoteEvidence;

    hasVoteB(): boolean;
    clearVoteB(): void;
    getVoteB(): tendermint_types_types_pb.Vote | undefined;
    setVoteB(value?: tendermint_types_types_pb.Vote): DuplicateVoteEvidence;
    getTotalVotingPower(): number;
    setTotalVotingPower(value: number): DuplicateVoteEvidence;
    getValidatorPower(): number;
    setValidatorPower(value: number): DuplicateVoteEvidence;

    hasTimestamp(): boolean;
    clearTimestamp(): void;
    getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): DuplicateVoteEvidence;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DuplicateVoteEvidence.AsObject;
    static toObject(includeInstance: boolean, msg: DuplicateVoteEvidence): DuplicateVoteEvidence.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DuplicateVoteEvidence, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DuplicateVoteEvidence;
    static deserializeBinaryFromReader(message: DuplicateVoteEvidence, reader: jspb.BinaryReader): DuplicateVoteEvidence;
}

export namespace DuplicateVoteEvidence {
    export type AsObject = {
        voteA?: tendermint_types_types_pb.Vote.AsObject,
        voteB?: tendermint_types_types_pb.Vote.AsObject,
        totalVotingPower: number,
        validatorPower: number,
        timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class LightClientAttackEvidence extends jspb.Message { 

    hasConflictingBlock(): boolean;
    clearConflictingBlock(): void;
    getConflictingBlock(): tendermint_types_types_pb.LightBlock | undefined;
    setConflictingBlock(value?: tendermint_types_types_pb.LightBlock): LightClientAttackEvidence;
    getCommonHeight(): number;
    setCommonHeight(value: number): LightClientAttackEvidence;
    clearByzantineValidatorsList(): void;
    getByzantineValidatorsList(): Array<tendermint_types_validator_pb.Validator>;
    setByzantineValidatorsList(value: Array<tendermint_types_validator_pb.Validator>): LightClientAttackEvidence;
    addByzantineValidators(value?: tendermint_types_validator_pb.Validator, index?: number): tendermint_types_validator_pb.Validator;
    getTotalVotingPower(): number;
    setTotalVotingPower(value: number): LightClientAttackEvidence;

    hasTimestamp(): boolean;
    clearTimestamp(): void;
    getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): LightClientAttackEvidence;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LightClientAttackEvidence.AsObject;
    static toObject(includeInstance: boolean, msg: LightClientAttackEvidence): LightClientAttackEvidence.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LightClientAttackEvidence, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LightClientAttackEvidence;
    static deserializeBinaryFromReader(message: LightClientAttackEvidence, reader: jspb.BinaryReader): LightClientAttackEvidence;
}

export namespace LightClientAttackEvidence {
    export type AsObject = {
        conflictingBlock?: tendermint_types_types_pb.LightBlock.AsObject,
        commonHeight: number,
        byzantineValidatorsList: Array<tendermint_types_validator_pb.Validator.AsObject>,
        totalVotingPower: number,
        timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class EvidenceList extends jspb.Message { 
    clearEvidenceList(): void;
    getEvidenceList(): Array<Evidence>;
    setEvidenceList(value: Array<Evidence>): EvidenceList;
    addEvidence(value?: Evidence, index?: number): Evidence;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EvidenceList.AsObject;
    static toObject(includeInstance: boolean, msg: EvidenceList): EvidenceList.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EvidenceList, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EvidenceList;
    static deserializeBinaryFromReader(message: EvidenceList, reader: jspb.BinaryReader): EvidenceList;
}

export namespace EvidenceList {
    export type AsObject = {
        evidenceList: Array<Evidence.AsObject>,
    }
}
