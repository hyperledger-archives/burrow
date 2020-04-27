// package: tendermint.consensus
// file: tendermint/consensus/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as tendermint_libs_bits_types_pb from "../../tendermint/libs/bits/types_pb";

export class NewRoundStep extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): NewRoundStep;
    getRound(): number;
    setRound(value: number): NewRoundStep;
    getStep(): number;
    setStep(value: number): NewRoundStep;
    getSecondsSinceStartTime(): number;
    setSecondsSinceStartTime(value: number): NewRoundStep;
    getLastCommitRound(): number;
    setLastCommitRound(value: number): NewRoundStep;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NewRoundStep.AsObject;
    static toObject(includeInstance: boolean, msg: NewRoundStep): NewRoundStep.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NewRoundStep, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NewRoundStep;
    static deserializeBinaryFromReader(message: NewRoundStep, reader: jspb.BinaryReader): NewRoundStep;
}

export namespace NewRoundStep {
    export type AsObject = {
        height: number,
        round: number,
        step: number,
        secondsSinceStartTime: number,
        lastCommitRound: number,
    }
}

export class NewValidBlock extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): NewValidBlock;
    getRound(): number;
    setRound(value: number): NewValidBlock;

    hasBlockPartSetHeader(): boolean;
    clearBlockPartSetHeader(): void;
    getBlockPartSetHeader(): tendermint_types_types_pb.PartSetHeader | undefined;
    setBlockPartSetHeader(value?: tendermint_types_types_pb.PartSetHeader): NewValidBlock;

    hasBlockParts(): boolean;
    clearBlockParts(): void;
    getBlockParts(): tendermint_libs_bits_types_pb.BitArray | undefined;
    setBlockParts(value?: tendermint_libs_bits_types_pb.BitArray): NewValidBlock;
    getIsCommit(): boolean;
    setIsCommit(value: boolean): NewValidBlock;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NewValidBlock.AsObject;
    static toObject(includeInstance: boolean, msg: NewValidBlock): NewValidBlock.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NewValidBlock, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NewValidBlock;
    static deserializeBinaryFromReader(message: NewValidBlock, reader: jspb.BinaryReader): NewValidBlock;
}

export namespace NewValidBlock {
    export type AsObject = {
        height: number,
        round: number,
        blockPartSetHeader?: tendermint_types_types_pb.PartSetHeader.AsObject,
        blockParts?: tendermint_libs_bits_types_pb.BitArray.AsObject,
        isCommit: boolean,
    }
}

export class Proposal extends jspb.Message { 

    hasProposal(): boolean;
    clearProposal(): void;
    getProposal(): tendermint_types_types_pb.Proposal | undefined;
    setProposal(value?: tendermint_types_types_pb.Proposal): Proposal;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Proposal.AsObject;
    static toObject(includeInstance: boolean, msg: Proposal): Proposal.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Proposal, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Proposal;
    static deserializeBinaryFromReader(message: Proposal, reader: jspb.BinaryReader): Proposal;
}

export namespace Proposal {
    export type AsObject = {
        proposal?: tendermint_types_types_pb.Proposal.AsObject,
    }
}

export class ProposalPOL extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): ProposalPOL;
    getProposalPolRound(): number;
    setProposalPolRound(value: number): ProposalPOL;

    hasProposalPol(): boolean;
    clearProposalPol(): void;
    getProposalPol(): tendermint_libs_bits_types_pb.BitArray | undefined;
    setProposalPol(value?: tendermint_libs_bits_types_pb.BitArray): ProposalPOL;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProposalPOL.AsObject;
    static toObject(includeInstance: boolean, msg: ProposalPOL): ProposalPOL.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProposalPOL, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProposalPOL;
    static deserializeBinaryFromReader(message: ProposalPOL, reader: jspb.BinaryReader): ProposalPOL;
}

export namespace ProposalPOL {
    export type AsObject = {
        height: number,
        proposalPolRound: number,
        proposalPol?: tendermint_libs_bits_types_pb.BitArray.AsObject,
    }
}

export class BlockPart extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): BlockPart;
    getRound(): number;
    setRound(value: number): BlockPart;

    hasPart(): boolean;
    clearPart(): void;
    getPart(): tendermint_types_types_pb.Part | undefined;
    setPart(value?: tendermint_types_types_pb.Part): BlockPart;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockPart.AsObject;
    static toObject(includeInstance: boolean, msg: BlockPart): BlockPart.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockPart, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockPart;
    static deserializeBinaryFromReader(message: BlockPart, reader: jspb.BinaryReader): BlockPart;
}

export namespace BlockPart {
    export type AsObject = {
        height: number,
        round: number,
        part?: tendermint_types_types_pb.Part.AsObject,
    }
}

export class Vote extends jspb.Message { 

    hasVote(): boolean;
    clearVote(): void;
    getVote(): tendermint_types_types_pb.Vote | undefined;
    setVote(value?: tendermint_types_types_pb.Vote): Vote;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Vote.AsObject;
    static toObject(includeInstance: boolean, msg: Vote): Vote.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Vote, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Vote;
    static deserializeBinaryFromReader(message: Vote, reader: jspb.BinaryReader): Vote;
}

export namespace Vote {
    export type AsObject = {
        vote?: tendermint_types_types_pb.Vote.AsObject,
    }
}

export class HasVote extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): HasVote;
    getRound(): number;
    setRound(value: number): HasVote;
    getType(): tendermint_types_types_pb.SignedMsgType;
    setType(value: tendermint_types_types_pb.SignedMsgType): HasVote;
    getIndex(): number;
    setIndex(value: number): HasVote;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HasVote.AsObject;
    static toObject(includeInstance: boolean, msg: HasVote): HasVote.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HasVote, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HasVote;
    static deserializeBinaryFromReader(message: HasVote, reader: jspb.BinaryReader): HasVote;
}

export namespace HasVote {
    export type AsObject = {
        height: number,
        round: number,
        type: tendermint_types_types_pb.SignedMsgType,
        index: number,
    }
}

export class VoteSetMaj23 extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): VoteSetMaj23;
    getRound(): number;
    setRound(value: number): VoteSetMaj23;
    getType(): tendermint_types_types_pb.SignedMsgType;
    setType(value: tendermint_types_types_pb.SignedMsgType): VoteSetMaj23;

    hasBlockId(): boolean;
    clearBlockId(): void;
    getBlockId(): tendermint_types_types_pb.BlockID | undefined;
    setBlockId(value?: tendermint_types_types_pb.BlockID): VoteSetMaj23;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): VoteSetMaj23.AsObject;
    static toObject(includeInstance: boolean, msg: VoteSetMaj23): VoteSetMaj23.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: VoteSetMaj23, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): VoteSetMaj23;
    static deserializeBinaryFromReader(message: VoteSetMaj23, reader: jspb.BinaryReader): VoteSetMaj23;
}

export namespace VoteSetMaj23 {
    export type AsObject = {
        height: number,
        round: number,
        type: tendermint_types_types_pb.SignedMsgType,
        blockId?: tendermint_types_types_pb.BlockID.AsObject,
    }
}

export class VoteSetBits extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): VoteSetBits;
    getRound(): number;
    setRound(value: number): VoteSetBits;
    getType(): tendermint_types_types_pb.SignedMsgType;
    setType(value: tendermint_types_types_pb.SignedMsgType): VoteSetBits;

    hasBlockId(): boolean;
    clearBlockId(): void;
    getBlockId(): tendermint_types_types_pb.BlockID | undefined;
    setBlockId(value?: tendermint_types_types_pb.BlockID): VoteSetBits;

    hasVotes(): boolean;
    clearVotes(): void;
    getVotes(): tendermint_libs_bits_types_pb.BitArray | undefined;
    setVotes(value?: tendermint_libs_bits_types_pb.BitArray): VoteSetBits;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): VoteSetBits.AsObject;
    static toObject(includeInstance: boolean, msg: VoteSetBits): VoteSetBits.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: VoteSetBits, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): VoteSetBits;
    static deserializeBinaryFromReader(message: VoteSetBits, reader: jspb.BinaryReader): VoteSetBits;
}

export namespace VoteSetBits {
    export type AsObject = {
        height: number,
        round: number,
        type: tendermint_types_types_pb.SignedMsgType,
        blockId?: tendermint_types_types_pb.BlockID.AsObject,
        votes?: tendermint_libs_bits_types_pb.BitArray.AsObject,
    }
}

export class Message extends jspb.Message { 

    hasNewRoundStep(): boolean;
    clearNewRoundStep(): void;
    getNewRoundStep(): NewRoundStep | undefined;
    setNewRoundStep(value?: NewRoundStep): Message;

    hasNewValidBlock(): boolean;
    clearNewValidBlock(): void;
    getNewValidBlock(): NewValidBlock | undefined;
    setNewValidBlock(value?: NewValidBlock): Message;

    hasProposal(): boolean;
    clearProposal(): void;
    getProposal(): Proposal | undefined;
    setProposal(value?: Proposal): Message;

    hasProposalPol(): boolean;
    clearProposalPol(): void;
    getProposalPol(): ProposalPOL | undefined;
    setProposalPol(value?: ProposalPOL): Message;

    hasBlockPart(): boolean;
    clearBlockPart(): void;
    getBlockPart(): BlockPart | undefined;
    setBlockPart(value?: BlockPart): Message;

    hasVote(): boolean;
    clearVote(): void;
    getVote(): Vote | undefined;
    setVote(value?: Vote): Message;

    hasHasVote(): boolean;
    clearHasVote(): void;
    getHasVote(): HasVote | undefined;
    setHasVote(value?: HasVote): Message;

    hasVoteSetMaj23(): boolean;
    clearVoteSetMaj23(): void;
    getVoteSetMaj23(): VoteSetMaj23 | undefined;
    setVoteSetMaj23(value?: VoteSetMaj23): Message;

    hasVoteSetBits(): boolean;
    clearVoteSetBits(): void;
    getVoteSetBits(): VoteSetBits | undefined;
    setVoteSetBits(value?: VoteSetBits): Message;

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
        newRoundStep?: NewRoundStep.AsObject,
        newValidBlock?: NewValidBlock.AsObject,
        proposal?: Proposal.AsObject,
        proposalPol?: ProposalPOL.AsObject,
        blockPart?: BlockPart.AsObject,
        vote?: Vote.AsObject,
        hasVote?: HasVote.AsObject,
        voteSetMaj23?: VoteSetMaj23.AsObject,
        voteSetBits?: VoteSetBits.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        NEW_ROUND_STEP = 1,
        NEW_VALID_BLOCK = 2,
        PROPOSAL = 3,
        PROPOSAL_POL = 4,
        BLOCK_PART = 5,
        VOTE = 6,
        HAS_VOTE = 7,
        VOTE_SET_MAJ23 = 8,
        VOTE_SET_BITS = 9,
    }

}
