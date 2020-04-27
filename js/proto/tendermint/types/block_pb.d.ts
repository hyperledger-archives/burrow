// package: tendermint.types
// file: tendermint/types/block.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as tendermint_types_evidence_pb from "../../tendermint/types/evidence_pb";

export class Block extends jspb.Message { 

    hasHeader(): boolean;
    clearHeader(): void;
    getHeader(): tendermint_types_types_pb.Header | undefined;
    setHeader(value?: tendermint_types_types_pb.Header): Block;

    hasData(): boolean;
    clearData(): void;
    getData(): tendermint_types_types_pb.Data | undefined;
    setData(value?: tendermint_types_types_pb.Data): Block;

    hasEvidence(): boolean;
    clearEvidence(): void;
    getEvidence(): tendermint_types_evidence_pb.EvidenceList | undefined;
    setEvidence(value?: tendermint_types_evidence_pb.EvidenceList): Block;

    hasLastCommit(): boolean;
    clearLastCommit(): void;
    getLastCommit(): tendermint_types_types_pb.Commit | undefined;
    setLastCommit(value?: tendermint_types_types_pb.Commit): Block;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Block.AsObject;
    static toObject(includeInstance: boolean, msg: Block): Block.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Block, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Block;
    static deserializeBinaryFromReader(message: Block, reader: jspb.BinaryReader): Block;
}

export namespace Block {
    export type AsObject = {
        header?: tendermint_types_types_pb.Header.AsObject,
        data?: tendermint_types_types_pb.Data.AsObject,
        evidence?: tendermint_types_evidence_pb.EvidenceList.AsObject,
        lastCommit?: tendermint_types_types_pb.Commit.AsObject,
    }
}
