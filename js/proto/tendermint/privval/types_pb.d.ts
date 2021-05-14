// package: tendermint.privval
// file: tendermint/privval/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as tendermint_crypto_keys_pb from "../../tendermint/crypto/keys_pb";
import * as tendermint_types_types_pb from "../../tendermint/types/types_pb";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";

export class RemoteSignerError extends jspb.Message { 
    getCode(): number;
    setCode(value: number): RemoteSignerError;
    getDescription(): string;
    setDescription(value: string): RemoteSignerError;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RemoteSignerError.AsObject;
    static toObject(includeInstance: boolean, msg: RemoteSignerError): RemoteSignerError.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RemoteSignerError, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RemoteSignerError;
    static deserializeBinaryFromReader(message: RemoteSignerError, reader: jspb.BinaryReader): RemoteSignerError;
}

export namespace RemoteSignerError {
    export type AsObject = {
        code: number,
        description: string,
    }
}

export class PubKeyRequest extends jspb.Message { 
    getChainId(): string;
    setChainId(value: string): PubKeyRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PubKeyRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PubKeyRequest): PubKeyRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PubKeyRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PubKeyRequest;
    static deserializeBinaryFromReader(message: PubKeyRequest, reader: jspb.BinaryReader): PubKeyRequest;
}

export namespace PubKeyRequest {
    export type AsObject = {
        chainId: string,
    }
}

export class PubKeyResponse extends jspb.Message { 

    hasPubKey(): boolean;
    clearPubKey(): void;
    getPubKey(): tendermint_crypto_keys_pb.PublicKey | undefined;
    setPubKey(value?: tendermint_crypto_keys_pb.PublicKey): PubKeyResponse;

    hasError(): boolean;
    clearError(): void;
    getError(): RemoteSignerError | undefined;
    setError(value?: RemoteSignerError): PubKeyResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PubKeyResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PubKeyResponse): PubKeyResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PubKeyResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PubKeyResponse;
    static deserializeBinaryFromReader(message: PubKeyResponse, reader: jspb.BinaryReader): PubKeyResponse;
}

export namespace PubKeyResponse {
    export type AsObject = {
        pubKey?: tendermint_crypto_keys_pb.PublicKey.AsObject,
        error?: RemoteSignerError.AsObject,
    }
}

export class SignVoteRequest extends jspb.Message { 

    hasVote(): boolean;
    clearVote(): void;
    getVote(): tendermint_types_types_pb.Vote | undefined;
    setVote(value?: tendermint_types_types_pb.Vote): SignVoteRequest;
    getChainId(): string;
    setChainId(value: string): SignVoteRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SignVoteRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SignVoteRequest): SignVoteRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SignVoteRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SignVoteRequest;
    static deserializeBinaryFromReader(message: SignVoteRequest, reader: jspb.BinaryReader): SignVoteRequest;
}

export namespace SignVoteRequest {
    export type AsObject = {
        vote?: tendermint_types_types_pb.Vote.AsObject,
        chainId: string,
    }
}

export class SignedVoteResponse extends jspb.Message { 

    hasVote(): boolean;
    clearVote(): void;
    getVote(): tendermint_types_types_pb.Vote | undefined;
    setVote(value?: tendermint_types_types_pb.Vote): SignedVoteResponse;

    hasError(): boolean;
    clearError(): void;
    getError(): RemoteSignerError | undefined;
    setError(value?: RemoteSignerError): SignedVoteResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SignedVoteResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SignedVoteResponse): SignedVoteResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SignedVoteResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SignedVoteResponse;
    static deserializeBinaryFromReader(message: SignedVoteResponse, reader: jspb.BinaryReader): SignedVoteResponse;
}

export namespace SignedVoteResponse {
    export type AsObject = {
        vote?: tendermint_types_types_pb.Vote.AsObject,
        error?: RemoteSignerError.AsObject,
    }
}

export class SignProposalRequest extends jspb.Message { 

    hasProposal(): boolean;
    clearProposal(): void;
    getProposal(): tendermint_types_types_pb.Proposal | undefined;
    setProposal(value?: tendermint_types_types_pb.Proposal): SignProposalRequest;
    getChainId(): string;
    setChainId(value: string): SignProposalRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SignProposalRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SignProposalRequest): SignProposalRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SignProposalRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SignProposalRequest;
    static deserializeBinaryFromReader(message: SignProposalRequest, reader: jspb.BinaryReader): SignProposalRequest;
}

export namespace SignProposalRequest {
    export type AsObject = {
        proposal?: tendermint_types_types_pb.Proposal.AsObject,
        chainId: string,
    }
}

export class SignedProposalResponse extends jspb.Message { 

    hasProposal(): boolean;
    clearProposal(): void;
    getProposal(): tendermint_types_types_pb.Proposal | undefined;
    setProposal(value?: tendermint_types_types_pb.Proposal): SignedProposalResponse;

    hasError(): boolean;
    clearError(): void;
    getError(): RemoteSignerError | undefined;
    setError(value?: RemoteSignerError): SignedProposalResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SignedProposalResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SignedProposalResponse): SignedProposalResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SignedProposalResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SignedProposalResponse;
    static deserializeBinaryFromReader(message: SignedProposalResponse, reader: jspb.BinaryReader): SignedProposalResponse;
}

export namespace SignedProposalResponse {
    export type AsObject = {
        proposal?: tendermint_types_types_pb.Proposal.AsObject,
        error?: RemoteSignerError.AsObject,
    }
}

export class PingRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PingRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PingRequest): PingRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PingRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PingRequest;
    static deserializeBinaryFromReader(message: PingRequest, reader: jspb.BinaryReader): PingRequest;
}

export namespace PingRequest {
    export type AsObject = {
    }
}

export class PingResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PingResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PingResponse): PingResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PingResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PingResponse;
    static deserializeBinaryFromReader(message: PingResponse, reader: jspb.BinaryReader): PingResponse;
}

export namespace PingResponse {
    export type AsObject = {
    }
}

export class Message extends jspb.Message { 

    hasPubKeyRequest(): boolean;
    clearPubKeyRequest(): void;
    getPubKeyRequest(): PubKeyRequest | undefined;
    setPubKeyRequest(value?: PubKeyRequest): Message;

    hasPubKeyResponse(): boolean;
    clearPubKeyResponse(): void;
    getPubKeyResponse(): PubKeyResponse | undefined;
    setPubKeyResponse(value?: PubKeyResponse): Message;

    hasSignVoteRequest(): boolean;
    clearSignVoteRequest(): void;
    getSignVoteRequest(): SignVoteRequest | undefined;
    setSignVoteRequest(value?: SignVoteRequest): Message;

    hasSignedVoteResponse(): boolean;
    clearSignedVoteResponse(): void;
    getSignedVoteResponse(): SignedVoteResponse | undefined;
    setSignedVoteResponse(value?: SignedVoteResponse): Message;

    hasSignProposalRequest(): boolean;
    clearSignProposalRequest(): void;
    getSignProposalRequest(): SignProposalRequest | undefined;
    setSignProposalRequest(value?: SignProposalRequest): Message;

    hasSignedProposalResponse(): boolean;
    clearSignedProposalResponse(): void;
    getSignedProposalResponse(): SignedProposalResponse | undefined;
    setSignedProposalResponse(value?: SignedProposalResponse): Message;

    hasPingRequest(): boolean;
    clearPingRequest(): void;
    getPingRequest(): PingRequest | undefined;
    setPingRequest(value?: PingRequest): Message;

    hasPingResponse(): boolean;
    clearPingResponse(): void;
    getPingResponse(): PingResponse | undefined;
    setPingResponse(value?: PingResponse): Message;

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
        pubKeyRequest?: PubKeyRequest.AsObject,
        pubKeyResponse?: PubKeyResponse.AsObject,
        signVoteRequest?: SignVoteRequest.AsObject,
        signedVoteResponse?: SignedVoteResponse.AsObject,
        signProposalRequest?: SignProposalRequest.AsObject,
        signedProposalResponse?: SignedProposalResponse.AsObject,
        pingRequest?: PingRequest.AsObject,
        pingResponse?: PingResponse.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        PUB_KEY_REQUEST = 1,
        PUB_KEY_RESPONSE = 2,
        SIGN_VOTE_REQUEST = 3,
        SIGNED_VOTE_RESPONSE = 4,
        SIGN_PROPOSAL_REQUEST = 5,
        SIGNED_PROPOSAL_RESPONSE = 6,
        PING_REQUEST = 7,
        PING_RESPONSE = 8,
    }

}

export enum Errors {
    ERRORS_UNKNOWN = 0,
    ERRORS_UNEXPECTED_RESPONSE = 1,
    ERRORS_NO_CONNECTION = 2,
    ERRORS_CONNECTION_TIMEOUT = 3,
    ERRORS_READ_TIMEOUT = 4,
    ERRORS_WRITE_TIMEOUT = 5,
}
