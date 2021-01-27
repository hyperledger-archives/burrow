// package: tendermint.types
// file: tendermint/types/canonical.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class CanonicalBlockID extends jspb.Message { 
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): CanonicalBlockID;


    hasPartSetHeader(): boolean;
    clearPartSetHeader(): void;
    getPartSetHeader(): CanonicalPartSetHeader | undefined;
    setPartSetHeader(value?: CanonicalPartSetHeader): CanonicalBlockID;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CanonicalBlockID.AsObject;
    static toObject(includeInstance: boolean, msg: CanonicalBlockID): CanonicalBlockID.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CanonicalBlockID, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CanonicalBlockID;
    static deserializeBinaryFromReader(message: CanonicalBlockID, reader: jspb.BinaryReader): CanonicalBlockID;
}

export namespace CanonicalBlockID {
    export type AsObject = {
        hash: Uint8Array | string,
        partSetHeader?: CanonicalPartSetHeader.AsObject,
    }
}

export class CanonicalPartSetHeader extends jspb.Message { 
    getTotal(): number;
    setTotal(value: number): CanonicalPartSetHeader;

    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): CanonicalPartSetHeader;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CanonicalPartSetHeader.AsObject;
    static toObject(includeInstance: boolean, msg: CanonicalPartSetHeader): CanonicalPartSetHeader.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CanonicalPartSetHeader, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CanonicalPartSetHeader;
    static deserializeBinaryFromReader(message: CanonicalPartSetHeader, reader: jspb.BinaryReader): CanonicalPartSetHeader;
}

export namespace CanonicalPartSetHeader {
    export type AsObject = {
        total: number,
        hash: Uint8Array | string,
    }
}

export class CanonicalProposal extends jspb.Message { 
    getType(): tendermint_types_types_pb.SignedMsgType;
    setType(value: tendermint_types_types_pb.SignedMsgType): CanonicalProposal;

    getHeight(): number;
    setHeight(value: number): CanonicalProposal;

    getRound(): number;
    setRound(value: number): CanonicalProposal;

    getPolRound(): number;
    setPolRound(value: number): CanonicalProposal;


    hasBlockId(): boolean;
    clearBlockId(): void;
    getBlockId(): CanonicalBlockID | undefined;
    setBlockId(value?: CanonicalBlockID): CanonicalProposal;


    hasTimestamp(): boolean;
    clearTimestamp(): void;
    getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): CanonicalProposal;

    getChainId(): string;
    setChainId(value: string): CanonicalProposal;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CanonicalProposal.AsObject;
    static toObject(includeInstance: boolean, msg: CanonicalProposal): CanonicalProposal.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CanonicalProposal, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CanonicalProposal;
    static deserializeBinaryFromReader(message: CanonicalProposal, reader: jspb.BinaryReader): CanonicalProposal;
}

export namespace CanonicalProposal {
    export type AsObject = {
        type: tendermint_types_types_pb.SignedMsgType,
        height: number,
        round: number,
        polRound: number,
        blockId?: CanonicalBlockID.AsObject,
        timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        chainId: string,
    }
}

export class CanonicalVote extends jspb.Message { 
    getType(): tendermint_types_types_pb.SignedMsgType;
    setType(value: tendermint_types_types_pb.SignedMsgType): CanonicalVote;

    getHeight(): number;
    setHeight(value: number): CanonicalVote;

    getRound(): number;
    setRound(value: number): CanonicalVote;


    hasBlockId(): boolean;
    clearBlockId(): void;
    getBlockId(): CanonicalBlockID | undefined;
    setBlockId(value?: CanonicalBlockID): CanonicalVote;


    hasTimestamp(): boolean;
    clearTimestamp(): void;
    getTimestamp(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTimestamp(value?: google_protobuf_timestamp_pb.Timestamp): CanonicalVote;

    getChainId(): string;
    setChainId(value: string): CanonicalVote;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CanonicalVote.AsObject;
    static toObject(includeInstance: boolean, msg: CanonicalVote): CanonicalVote.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CanonicalVote, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CanonicalVote;
    static deserializeBinaryFromReader(message: CanonicalVote, reader: jspb.BinaryReader): CanonicalVote;
}

export namespace CanonicalVote {
    export type AsObject = {
        type: tendermint_types_types_pb.SignedMsgType,
        height: number,
        round: number,
        blockId?: CanonicalBlockID.AsObject,
        timestamp?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        chainId: string,
    }
}
