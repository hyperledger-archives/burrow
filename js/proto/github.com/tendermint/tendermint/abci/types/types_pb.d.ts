// package: types
// file: github.com/tendermint/tendermint/abci/types/types.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "../../../../../github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as github_com_tendermint_tendermint_libs_common_types_pb from "../../../../../github.com/tendermint/tendermint/libs/common/types_pb";

export class Request extends jspb.Message { 

    hasEcho(): boolean;
    clearEcho(): void;
    getEcho(): RequestEcho | undefined;
    setEcho(value?: RequestEcho): void;


    hasFlush(): boolean;
    clearFlush(): void;
    getFlush(): RequestFlush | undefined;
    setFlush(value?: RequestFlush): void;


    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): RequestInfo | undefined;
    setInfo(value?: RequestInfo): void;


    hasSetOption(): boolean;
    clearSetOption(): void;
    getSetOption(): RequestSetOption | undefined;
    setSetOption(value?: RequestSetOption): void;


    hasInitChain(): boolean;
    clearInitChain(): void;
    getInitChain(): RequestInitChain | undefined;
    setInitChain(value?: RequestInitChain): void;


    hasQuery(): boolean;
    clearQuery(): void;
    getQuery(): RequestQuery | undefined;
    setQuery(value?: RequestQuery): void;


    hasBeginBlock(): boolean;
    clearBeginBlock(): void;
    getBeginBlock(): RequestBeginBlock | undefined;
    setBeginBlock(value?: RequestBeginBlock): void;


    hasCheckTx(): boolean;
    clearCheckTx(): void;
    getCheckTx(): RequestCheckTx | undefined;
    setCheckTx(value?: RequestCheckTx): void;


    hasDeliverTx(): boolean;
    clearDeliverTx(): void;
    getDeliverTx(): RequestDeliverTx | undefined;
    setDeliverTx(value?: RequestDeliverTx): void;


    hasEndBlock(): boolean;
    clearEndBlock(): void;
    getEndBlock(): RequestEndBlock | undefined;
    setEndBlock(value?: RequestEndBlock): void;


    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): RequestCommit | undefined;
    setCommit(value?: RequestCommit): void;


    getValueCase(): Request.ValueCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Request.AsObject;
    static toObject(includeInstance: boolean, msg: Request): Request.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Request, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Request;
    static deserializeBinaryFromReader(message: Request, reader: jspb.BinaryReader): Request;
}

export namespace Request {
    export type AsObject = {
        echo?: RequestEcho.AsObject,
        flush?: RequestFlush.AsObject,
        info?: RequestInfo.AsObject,
        setOption?: RequestSetOption.AsObject,
        initChain?: RequestInitChain.AsObject,
        query?: RequestQuery.AsObject,
        beginBlock?: RequestBeginBlock.AsObject,
        checkTx?: RequestCheckTx.AsObject,
        deliverTx?: RequestDeliverTx.AsObject,
        endBlock?: RequestEndBlock.AsObject,
        commit?: RequestCommit.AsObject,
    }

    export enum ValueCase {
        VALUE_NOT_SET = 0,
    
    ECHO = 2,

    FLUSH = 3,

    INFO = 4,

    SET_OPTION = 5,

    INIT_CHAIN = 6,

    QUERY = 7,

    BEGIN_BLOCK = 8,

    CHECK_TX = 9,

    DELIVER_TX = 19,

    END_BLOCK = 11,

    COMMIT = 12,

    }

}

export class RequestEcho extends jspb.Message { 
    getMessage(): string;
    setMessage(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestEcho.AsObject;
    static toObject(includeInstance: boolean, msg: RequestEcho): RequestEcho.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestEcho, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestEcho;
    static deserializeBinaryFromReader(message: RequestEcho, reader: jspb.BinaryReader): RequestEcho;
}

export namespace RequestEcho {
    export type AsObject = {
        message: string,
    }
}

export class RequestFlush extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestFlush.AsObject;
    static toObject(includeInstance: boolean, msg: RequestFlush): RequestFlush.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestFlush, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestFlush;
    static deserializeBinaryFromReader(message: RequestFlush, reader: jspb.BinaryReader): RequestFlush;
}

export namespace RequestFlush {
    export type AsObject = {
    }
}

export class RequestInfo extends jspb.Message { 
    getVersion(): string;
    setVersion(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestInfo.AsObject;
    static toObject(includeInstance: boolean, msg: RequestInfo): RequestInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestInfo;
    static deserializeBinaryFromReader(message: RequestInfo, reader: jspb.BinaryReader): RequestInfo;
}

export namespace RequestInfo {
    export type AsObject = {
        version: string,
    }
}

export class RequestSetOption extends jspb.Message { 
    getKey(): string;
    setKey(value: string): void;

    getValue(): string;
    setValue(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestSetOption.AsObject;
    static toObject(includeInstance: boolean, msg: RequestSetOption): RequestSetOption.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestSetOption, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestSetOption;
    static deserializeBinaryFromReader(message: RequestSetOption, reader: jspb.BinaryReader): RequestSetOption;
}

export namespace RequestSetOption {
    export type AsObject = {
        key: string,
        value: string,
    }
}

export class RequestInitChain extends jspb.Message { 

    hasTime(): boolean;
    clearTime(): void;
    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

    getChainId(): string;
    setChainId(value: string): void;


    hasConsensusParams(): boolean;
    clearConsensusParams(): void;
    getConsensusParams(): ConsensusParams | undefined;
    setConsensusParams(value?: ConsensusParams): void;

    clearValidatorsList(): void;
    getValidatorsList(): Array<ValidatorUpdate>;
    setValidatorsList(value: Array<ValidatorUpdate>): void;
    addValidators(value?: ValidatorUpdate, index?: number): ValidatorUpdate;

    getAppStateBytes(): Uint8Array | string;
    getAppStateBytes_asU8(): Uint8Array;
    getAppStateBytes_asB64(): string;
    setAppStateBytes(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestInitChain.AsObject;
    static toObject(includeInstance: boolean, msg: RequestInitChain): RequestInitChain.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestInitChain, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestInitChain;
    static deserializeBinaryFromReader(message: RequestInitChain, reader: jspb.BinaryReader): RequestInitChain;
}

export namespace RequestInitChain {
    export type AsObject = {
        time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        chainId: string,
        consensusParams?: ConsensusParams.AsObject,
        validatorsList: Array<ValidatorUpdate.AsObject>,
        appStateBytes: Uint8Array | string,
    }
}

export class RequestQuery extends jspb.Message { 
    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): void;

    getPath(): string;
    setPath(value: string): void;

    getHeight(): number;
    setHeight(value: number): void;

    getProve(): boolean;
    setProve(value: boolean): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestQuery.AsObject;
    static toObject(includeInstance: boolean, msg: RequestQuery): RequestQuery.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestQuery, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestQuery;
    static deserializeBinaryFromReader(message: RequestQuery, reader: jspb.BinaryReader): RequestQuery;
}

export namespace RequestQuery {
    export type AsObject = {
        data: Uint8Array | string,
        path: string,
        height: number,
        prove: boolean,
    }
}

export class RequestBeginBlock extends jspb.Message { 
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): void;


    hasHeader(): boolean;
    clearHeader(): void;
    getHeader(): Header | undefined;
    setHeader(value?: Header): void;


    hasLastCommitInfo(): boolean;
    clearLastCommitInfo(): void;
    getLastCommitInfo(): LastCommitInfo | undefined;
    setLastCommitInfo(value?: LastCommitInfo): void;

    clearByzantineValidatorsList(): void;
    getByzantineValidatorsList(): Array<Evidence>;
    setByzantineValidatorsList(value: Array<Evidence>): void;
    addByzantineValidators(value?: Evidence, index?: number): Evidence;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestBeginBlock.AsObject;
    static toObject(includeInstance: boolean, msg: RequestBeginBlock): RequestBeginBlock.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestBeginBlock, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestBeginBlock;
    static deserializeBinaryFromReader(message: RequestBeginBlock, reader: jspb.BinaryReader): RequestBeginBlock;
}

export namespace RequestBeginBlock {
    export type AsObject = {
        hash: Uint8Array | string,
        header?: Header.AsObject,
        lastCommitInfo?: LastCommitInfo.AsObject,
        byzantineValidatorsList: Array<Evidence.AsObject>,
    }
}

export class RequestCheckTx extends jspb.Message { 
    getTx(): Uint8Array | string;
    getTx_asU8(): Uint8Array;
    getTx_asB64(): string;
    setTx(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestCheckTx.AsObject;
    static toObject(includeInstance: boolean, msg: RequestCheckTx): RequestCheckTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestCheckTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestCheckTx;
    static deserializeBinaryFromReader(message: RequestCheckTx, reader: jspb.BinaryReader): RequestCheckTx;
}

export namespace RequestCheckTx {
    export type AsObject = {
        tx: Uint8Array | string,
    }
}

export class RequestDeliverTx extends jspb.Message { 
    getTx(): Uint8Array | string;
    getTx_asU8(): Uint8Array;
    getTx_asB64(): string;
    setTx(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestDeliverTx.AsObject;
    static toObject(includeInstance: boolean, msg: RequestDeliverTx): RequestDeliverTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestDeliverTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestDeliverTx;
    static deserializeBinaryFromReader(message: RequestDeliverTx, reader: jspb.BinaryReader): RequestDeliverTx;
}

export namespace RequestDeliverTx {
    export type AsObject = {
        tx: Uint8Array | string,
    }
}

export class RequestEndBlock extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestEndBlock.AsObject;
    static toObject(includeInstance: boolean, msg: RequestEndBlock): RequestEndBlock.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestEndBlock, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestEndBlock;
    static deserializeBinaryFromReader(message: RequestEndBlock, reader: jspb.BinaryReader): RequestEndBlock;
}

export namespace RequestEndBlock {
    export type AsObject = {
        height: number,
    }
}

export class RequestCommit extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestCommit.AsObject;
    static toObject(includeInstance: boolean, msg: RequestCommit): RequestCommit.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RequestCommit, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestCommit;
    static deserializeBinaryFromReader(message: RequestCommit, reader: jspb.BinaryReader): RequestCommit;
}

export namespace RequestCommit {
    export type AsObject = {
    }
}

export class Response extends jspb.Message { 

    hasException(): boolean;
    clearException(): void;
    getException(): ResponseException | undefined;
    setException(value?: ResponseException): void;


    hasEcho(): boolean;
    clearEcho(): void;
    getEcho(): ResponseEcho | undefined;
    setEcho(value?: ResponseEcho): void;


    hasFlush(): boolean;
    clearFlush(): void;
    getFlush(): ResponseFlush | undefined;
    setFlush(value?: ResponseFlush): void;


    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): ResponseInfo | undefined;
    setInfo(value?: ResponseInfo): void;


    hasSetOption(): boolean;
    clearSetOption(): void;
    getSetOption(): ResponseSetOption | undefined;
    setSetOption(value?: ResponseSetOption): void;


    hasInitChain(): boolean;
    clearInitChain(): void;
    getInitChain(): ResponseInitChain | undefined;
    setInitChain(value?: ResponseInitChain): void;


    hasQuery(): boolean;
    clearQuery(): void;
    getQuery(): ResponseQuery | undefined;
    setQuery(value?: ResponseQuery): void;


    hasBeginBlock(): boolean;
    clearBeginBlock(): void;
    getBeginBlock(): ResponseBeginBlock | undefined;
    setBeginBlock(value?: ResponseBeginBlock): void;


    hasCheckTx(): boolean;
    clearCheckTx(): void;
    getCheckTx(): ResponseCheckTx | undefined;
    setCheckTx(value?: ResponseCheckTx): void;


    hasDeliverTx(): boolean;
    clearDeliverTx(): void;
    getDeliverTx(): ResponseDeliverTx | undefined;
    setDeliverTx(value?: ResponseDeliverTx): void;


    hasEndBlock(): boolean;
    clearEndBlock(): void;
    getEndBlock(): ResponseEndBlock | undefined;
    setEndBlock(value?: ResponseEndBlock): void;


    hasCommit(): boolean;
    clearCommit(): void;
    getCommit(): ResponseCommit | undefined;
    setCommit(value?: ResponseCommit): void;


    getValueCase(): Response.ValueCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Response.AsObject;
    static toObject(includeInstance: boolean, msg: Response): Response.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Response, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Response;
    static deserializeBinaryFromReader(message: Response, reader: jspb.BinaryReader): Response;
}

export namespace Response {
    export type AsObject = {
        exception?: ResponseException.AsObject,
        echo?: ResponseEcho.AsObject,
        flush?: ResponseFlush.AsObject,
        info?: ResponseInfo.AsObject,
        setOption?: ResponseSetOption.AsObject,
        initChain?: ResponseInitChain.AsObject,
        query?: ResponseQuery.AsObject,
        beginBlock?: ResponseBeginBlock.AsObject,
        checkTx?: ResponseCheckTx.AsObject,
        deliverTx?: ResponseDeliverTx.AsObject,
        endBlock?: ResponseEndBlock.AsObject,
        commit?: ResponseCommit.AsObject,
    }

    export enum ValueCase {
        VALUE_NOT_SET = 0,
    
    EXCEPTION = 1,

    ECHO = 2,

    FLUSH = 3,

    INFO = 4,

    SET_OPTION = 5,

    INIT_CHAIN = 6,

    QUERY = 7,

    BEGIN_BLOCK = 8,

    CHECK_TX = 9,

    DELIVER_TX = 10,

    END_BLOCK = 11,

    COMMIT = 12,

    }

}

export class ResponseException extends jspb.Message { 
    getError(): string;
    setError(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseException.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseException): ResponseException.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseException, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseException;
    static deserializeBinaryFromReader(message: ResponseException, reader: jspb.BinaryReader): ResponseException;
}

export namespace ResponseException {
    export type AsObject = {
        error: string,
    }
}

export class ResponseEcho extends jspb.Message { 
    getMessage(): string;
    setMessage(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseEcho.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseEcho): ResponseEcho.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseEcho, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseEcho;
    static deserializeBinaryFromReader(message: ResponseEcho, reader: jspb.BinaryReader): ResponseEcho;
}

export namespace ResponseEcho {
    export type AsObject = {
        message: string,
    }
}

export class ResponseFlush extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseFlush.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseFlush): ResponseFlush.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseFlush, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseFlush;
    static deserializeBinaryFromReader(message: ResponseFlush, reader: jspb.BinaryReader): ResponseFlush;
}

export namespace ResponseFlush {
    export type AsObject = {
    }
}

export class ResponseInfo extends jspb.Message { 
    getData(): string;
    setData(value: string): void;

    getVersion(): string;
    setVersion(value: string): void;

    getLastBlockHeight(): number;
    setLastBlockHeight(value: number): void;

    getLastBlockAppHash(): Uint8Array | string;
    getLastBlockAppHash_asU8(): Uint8Array;
    getLastBlockAppHash_asB64(): string;
    setLastBlockAppHash(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseInfo): ResponseInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseInfo;
    static deserializeBinaryFromReader(message: ResponseInfo, reader: jspb.BinaryReader): ResponseInfo;
}

export namespace ResponseInfo {
    export type AsObject = {
        data: string,
        version: string,
        lastBlockHeight: number,
        lastBlockAppHash: Uint8Array | string,
    }
}

export class ResponseSetOption extends jspb.Message { 
    getCode(): number;
    setCode(value: number): void;

    getLog(): string;
    setLog(value: string): void;

    getInfo(): string;
    setInfo(value: string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseSetOption.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseSetOption): ResponseSetOption.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseSetOption, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseSetOption;
    static deserializeBinaryFromReader(message: ResponseSetOption, reader: jspb.BinaryReader): ResponseSetOption;
}

export namespace ResponseSetOption {
    export type AsObject = {
        code: number,
        log: string,
        info: string,
    }
}

export class ResponseInitChain extends jspb.Message { 

    hasConsensusParams(): boolean;
    clearConsensusParams(): void;
    getConsensusParams(): ConsensusParams | undefined;
    setConsensusParams(value?: ConsensusParams): void;

    clearValidatorsList(): void;
    getValidatorsList(): Array<ValidatorUpdate>;
    setValidatorsList(value: Array<ValidatorUpdate>): void;
    addValidators(value?: ValidatorUpdate, index?: number): ValidatorUpdate;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseInitChain.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseInitChain): ResponseInitChain.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseInitChain, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseInitChain;
    static deserializeBinaryFromReader(message: ResponseInitChain, reader: jspb.BinaryReader): ResponseInitChain;
}

export namespace ResponseInitChain {
    export type AsObject = {
        consensusParams?: ConsensusParams.AsObject,
        validatorsList: Array<ValidatorUpdate.AsObject>,
    }
}

export class ResponseQuery extends jspb.Message { 
    getCode(): number;
    setCode(value: number): void;

    getLog(): string;
    setLog(value: string): void;

    getInfo(): string;
    setInfo(value: string): void;

    getIndex(): number;
    setIndex(value: number): void;

    getKey(): Uint8Array | string;
    getKey_asU8(): Uint8Array;
    getKey_asB64(): string;
    setKey(value: Uint8Array | string): void;

    getValue(): Uint8Array | string;
    getValue_asU8(): Uint8Array;
    getValue_asB64(): string;
    setValue(value: Uint8Array | string): void;

    getProof(): Uint8Array | string;
    getProof_asU8(): Uint8Array;
    getProof_asB64(): string;
    setProof(value: Uint8Array | string): void;

    getHeight(): number;
    setHeight(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseQuery.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseQuery): ResponseQuery.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseQuery, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseQuery;
    static deserializeBinaryFromReader(message: ResponseQuery, reader: jspb.BinaryReader): ResponseQuery;
}

export namespace ResponseQuery {
    export type AsObject = {
        code: number,
        log: string,
        info: string,
        index: number,
        key: Uint8Array | string,
        value: Uint8Array | string,
        proof: Uint8Array | string,
        height: number,
    }
}

export class ResponseBeginBlock extends jspb.Message { 
    clearTagsList(): void;
    getTagsList(): Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>;
    setTagsList(value: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>): void;
    addTags(value?: github_com_tendermint_tendermint_libs_common_types_pb.KVPair, index?: number): github_com_tendermint_tendermint_libs_common_types_pb.KVPair;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseBeginBlock.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseBeginBlock): ResponseBeginBlock.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseBeginBlock, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseBeginBlock;
    static deserializeBinaryFromReader(message: ResponseBeginBlock, reader: jspb.BinaryReader): ResponseBeginBlock;
}

export namespace ResponseBeginBlock {
    export type AsObject = {
        tagsList: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair.AsObject>,
    }
}

export class ResponseCheckTx extends jspb.Message { 
    getCode(): number;
    setCode(value: number): void;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): void;

    getLog(): string;
    setLog(value: string): void;

    getInfo(): string;
    setInfo(value: string): void;

    getGasWanted(): number;
    setGasWanted(value: number): void;

    getGasUsed(): number;
    setGasUsed(value: number): void;

    clearTagsList(): void;
    getTagsList(): Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>;
    setTagsList(value: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>): void;
    addTags(value?: github_com_tendermint_tendermint_libs_common_types_pb.KVPair, index?: number): github_com_tendermint_tendermint_libs_common_types_pb.KVPair;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseCheckTx.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseCheckTx): ResponseCheckTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseCheckTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseCheckTx;
    static deserializeBinaryFromReader(message: ResponseCheckTx, reader: jspb.BinaryReader): ResponseCheckTx;
}

export namespace ResponseCheckTx {
    export type AsObject = {
        code: number,
        data: Uint8Array | string,
        log: string,
        info: string,
        gasWanted: number,
        gasUsed: number,
        tagsList: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair.AsObject>,
    }
}

export class ResponseDeliverTx extends jspb.Message { 
    getCode(): number;
    setCode(value: number): void;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): void;

    getLog(): string;
    setLog(value: string): void;

    getInfo(): string;
    setInfo(value: string): void;

    getGasWanted(): number;
    setGasWanted(value: number): void;

    getGasUsed(): number;
    setGasUsed(value: number): void;

    clearTagsList(): void;
    getTagsList(): Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>;
    setTagsList(value: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>): void;
    addTags(value?: github_com_tendermint_tendermint_libs_common_types_pb.KVPair, index?: number): github_com_tendermint_tendermint_libs_common_types_pb.KVPair;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseDeliverTx.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseDeliverTx): ResponseDeliverTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseDeliverTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseDeliverTx;
    static deserializeBinaryFromReader(message: ResponseDeliverTx, reader: jspb.BinaryReader): ResponseDeliverTx;
}

export namespace ResponseDeliverTx {
    export type AsObject = {
        code: number,
        data: Uint8Array | string,
        log: string,
        info: string,
        gasWanted: number,
        gasUsed: number,
        tagsList: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair.AsObject>,
    }
}

export class ResponseEndBlock extends jspb.Message { 
    clearValidatorUpdatesList(): void;
    getValidatorUpdatesList(): Array<ValidatorUpdate>;
    setValidatorUpdatesList(value: Array<ValidatorUpdate>): void;
    addValidatorUpdates(value?: ValidatorUpdate, index?: number): ValidatorUpdate;


    hasConsensusParamUpdates(): boolean;
    clearConsensusParamUpdates(): void;
    getConsensusParamUpdates(): ConsensusParams | undefined;
    setConsensusParamUpdates(value?: ConsensusParams): void;

    clearTagsList(): void;
    getTagsList(): Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>;
    setTagsList(value: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair>): void;
    addTags(value?: github_com_tendermint_tendermint_libs_common_types_pb.KVPair, index?: number): github_com_tendermint_tendermint_libs_common_types_pb.KVPair;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseEndBlock.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseEndBlock): ResponseEndBlock.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseEndBlock, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseEndBlock;
    static deserializeBinaryFromReader(message: ResponseEndBlock, reader: jspb.BinaryReader): ResponseEndBlock;
}

export namespace ResponseEndBlock {
    export type AsObject = {
        validatorUpdatesList: Array<ValidatorUpdate.AsObject>,
        consensusParamUpdates?: ConsensusParams.AsObject,
        tagsList: Array<github_com_tendermint_tendermint_libs_common_types_pb.KVPair.AsObject>,
    }
}

export class ResponseCommit extends jspb.Message { 
    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ResponseCommit.AsObject;
    static toObject(includeInstance: boolean, msg: ResponseCommit): ResponseCommit.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ResponseCommit, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ResponseCommit;
    static deserializeBinaryFromReader(message: ResponseCommit, reader: jspb.BinaryReader): ResponseCommit;
}

export namespace ResponseCommit {
    export type AsObject = {
        data: Uint8Array | string,
    }
}

export class ConsensusParams extends jspb.Message { 

    hasBlockSize(): boolean;
    clearBlockSize(): void;
    getBlockSize(): BlockSize | undefined;
    setBlockSize(value?: BlockSize): void;


    hasTxSize(): boolean;
    clearTxSize(): void;
    getTxSize(): TxSize | undefined;
    setTxSize(value?: TxSize): void;


    hasBlockGossip(): boolean;
    clearBlockGossip(): void;
    getBlockGossip(): BlockGossip | undefined;
    setBlockGossip(value?: BlockGossip): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ConsensusParams.AsObject;
    static toObject(includeInstance: boolean, msg: ConsensusParams): ConsensusParams.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ConsensusParams, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ConsensusParams;
    static deserializeBinaryFromReader(message: ConsensusParams, reader: jspb.BinaryReader): ConsensusParams;
}

export namespace ConsensusParams {
    export type AsObject = {
        blockSize?: BlockSize.AsObject,
        txSize?: TxSize.AsObject,
        blockGossip?: BlockGossip.AsObject,
    }
}

export class BlockSize extends jspb.Message { 
    getMaxBytes(): number;
    setMaxBytes(value: number): void;

    getMaxGas(): number;
    setMaxGas(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockSize.AsObject;
    static toObject(includeInstance: boolean, msg: BlockSize): BlockSize.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockSize, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockSize;
    static deserializeBinaryFromReader(message: BlockSize, reader: jspb.BinaryReader): BlockSize;
}

export namespace BlockSize {
    export type AsObject = {
        maxBytes: number,
        maxGas: number,
    }
}

export class TxSize extends jspb.Message { 
    getMaxBytes(): number;
    setMaxBytes(value: number): void;

    getMaxGas(): number;
    setMaxGas(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TxSize.AsObject;
    static toObject(includeInstance: boolean, msg: TxSize): TxSize.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TxSize, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TxSize;
    static deserializeBinaryFromReader(message: TxSize, reader: jspb.BinaryReader): TxSize;
}

export namespace TxSize {
    export type AsObject = {
        maxBytes: number,
        maxGas: number,
    }
}

export class BlockGossip extends jspb.Message { 
    getBlockPartSizeBytes(): number;
    setBlockPartSizeBytes(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockGossip.AsObject;
    static toObject(includeInstance: boolean, msg: BlockGossip): BlockGossip.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockGossip, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockGossip;
    static deserializeBinaryFromReader(message: BlockGossip, reader: jspb.BinaryReader): BlockGossip;
}

export namespace BlockGossip {
    export type AsObject = {
        blockPartSizeBytes: number,
    }
}

export class LastCommitInfo extends jspb.Message { 
    getRound(): number;
    setRound(value: number): void;

    clearVotesList(): void;
    getVotesList(): Array<VoteInfo>;
    setVotesList(value: Array<VoteInfo>): void;
    addVotes(value?: VoteInfo, index?: number): VoteInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LastCommitInfo.AsObject;
    static toObject(includeInstance: boolean, msg: LastCommitInfo): LastCommitInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LastCommitInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LastCommitInfo;
    static deserializeBinaryFromReader(message: LastCommitInfo, reader: jspb.BinaryReader): LastCommitInfo;
}

export namespace LastCommitInfo {
    export type AsObject = {
        round: number,
        votesList: Array<VoteInfo.AsObject>,
    }
}

export class Header extends jspb.Message { 
    getChainId(): string;
    setChainId(value: string): void;

    getHeight(): number;
    setHeight(value: number): void;


    hasTime(): boolean;
    clearTime(): void;
    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

    getNumTxs(): number;
    setNumTxs(value: number): void;

    getTotalTxs(): number;
    setTotalTxs(value: number): void;


    hasLastBlockId(): boolean;
    clearLastBlockId(): void;
    getLastBlockId(): BlockID | undefined;
    setLastBlockId(value?: BlockID): void;

    getLastCommitHash(): Uint8Array | string;
    getLastCommitHash_asU8(): Uint8Array;
    getLastCommitHash_asB64(): string;
    setLastCommitHash(value: Uint8Array | string): void;

    getDataHash(): Uint8Array | string;
    getDataHash_asU8(): Uint8Array;
    getDataHash_asB64(): string;
    setDataHash(value: Uint8Array | string): void;

    getValidatorsHash(): Uint8Array | string;
    getValidatorsHash_asU8(): Uint8Array;
    getValidatorsHash_asB64(): string;
    setValidatorsHash(value: Uint8Array | string): void;

    getNextValidatorsHash(): Uint8Array | string;
    getNextValidatorsHash_asU8(): Uint8Array;
    getNextValidatorsHash_asB64(): string;
    setNextValidatorsHash(value: Uint8Array | string): void;

    getConsensusHash(): Uint8Array | string;
    getConsensusHash_asU8(): Uint8Array;
    getConsensusHash_asB64(): string;
    setConsensusHash(value: Uint8Array | string): void;

    getAppHash(): Uint8Array | string;
    getAppHash_asU8(): Uint8Array;
    getAppHash_asB64(): string;
    setAppHash(value: Uint8Array | string): void;

    getLastResultsHash(): Uint8Array | string;
    getLastResultsHash_asU8(): Uint8Array;
    getLastResultsHash_asB64(): string;
    setLastResultsHash(value: Uint8Array | string): void;

    getEvidenceHash(): Uint8Array | string;
    getEvidenceHash_asU8(): Uint8Array;
    getEvidenceHash_asB64(): string;
    setEvidenceHash(value: Uint8Array | string): void;

    getProposerAddress(): Uint8Array | string;
    getProposerAddress_asU8(): Uint8Array;
    getProposerAddress_asB64(): string;
    setProposerAddress(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Header.AsObject;
    static toObject(includeInstance: boolean, msg: Header): Header.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Header, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Header;
    static deserializeBinaryFromReader(message: Header, reader: jspb.BinaryReader): Header;
}

export namespace Header {
    export type AsObject = {
        chainId: string,
        height: number,
        time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        numTxs: number,
        totalTxs: number,
        lastBlockId?: BlockID.AsObject,
        lastCommitHash: Uint8Array | string,
        dataHash: Uint8Array | string,
        validatorsHash: Uint8Array | string,
        nextValidatorsHash: Uint8Array | string,
        consensusHash: Uint8Array | string,
        appHash: Uint8Array | string,
        lastResultsHash: Uint8Array | string,
        evidenceHash: Uint8Array | string,
        proposerAddress: Uint8Array | string,
    }
}

export class BlockID extends jspb.Message { 
    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): void;


    hasPartsHeader(): boolean;
    clearPartsHeader(): void;
    getPartsHeader(): PartSetHeader | undefined;
    setPartsHeader(value?: PartSetHeader): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BlockID.AsObject;
    static toObject(includeInstance: boolean, msg: BlockID): BlockID.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BlockID, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BlockID;
    static deserializeBinaryFromReader(message: BlockID, reader: jspb.BinaryReader): BlockID;
}

export namespace BlockID {
    export type AsObject = {
        hash: Uint8Array | string,
        partsHeader?: PartSetHeader.AsObject,
    }
}

export class PartSetHeader extends jspb.Message { 
    getTotal(): number;
    setTotal(value: number): void;

    getHash(): Uint8Array | string;
    getHash_asU8(): Uint8Array;
    getHash_asB64(): string;
    setHash(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PartSetHeader.AsObject;
    static toObject(includeInstance: boolean, msg: PartSetHeader): PartSetHeader.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PartSetHeader, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PartSetHeader;
    static deserializeBinaryFromReader(message: PartSetHeader, reader: jspb.BinaryReader): PartSetHeader;
}

export namespace PartSetHeader {
    export type AsObject = {
        total: number,
        hash: Uint8Array | string,
    }
}

export class Validator extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): void;

    getPower(): number;
    setPower(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Validator.AsObject;
    static toObject(includeInstance: boolean, msg: Validator): Validator.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Validator, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Validator;
    static deserializeBinaryFromReader(message: Validator, reader: jspb.BinaryReader): Validator;
}

export namespace Validator {
    export type AsObject = {
        address: Uint8Array | string,
        power: number,
    }
}

export class ValidatorUpdate extends jspb.Message { 

    hasPubKey(): boolean;
    clearPubKey(): void;
    getPubKey(): PubKey | undefined;
    setPubKey(value?: PubKey): void;

    getPower(): number;
    setPower(value: number): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ValidatorUpdate.AsObject;
    static toObject(includeInstance: boolean, msg: ValidatorUpdate): ValidatorUpdate.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ValidatorUpdate, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ValidatorUpdate;
    static deserializeBinaryFromReader(message: ValidatorUpdate, reader: jspb.BinaryReader): ValidatorUpdate;
}

export namespace ValidatorUpdate {
    export type AsObject = {
        pubKey?: PubKey.AsObject,
        power: number,
    }
}

export class VoteInfo extends jspb.Message { 

    hasValidator(): boolean;
    clearValidator(): void;
    getValidator(): Validator | undefined;
    setValidator(value?: Validator): void;

    getSignedLastBlock(): boolean;
    setSignedLastBlock(value: boolean): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): VoteInfo.AsObject;
    static toObject(includeInstance: boolean, msg: VoteInfo): VoteInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: VoteInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): VoteInfo;
    static deserializeBinaryFromReader(message: VoteInfo, reader: jspb.BinaryReader): VoteInfo;
}

export namespace VoteInfo {
    export type AsObject = {
        validator?: Validator.AsObject,
        signedLastBlock: boolean,
    }
}

export class PubKey extends jspb.Message { 
    getType(): string;
    setType(value: string): void;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PubKey.AsObject;
    static toObject(includeInstance: boolean, msg: PubKey): PubKey.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PubKey, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PubKey;
    static deserializeBinaryFromReader(message: PubKey, reader: jspb.BinaryReader): PubKey;
}

export namespace PubKey {
    export type AsObject = {
        type: string,
        data: Uint8Array | string,
    }
}

export class Evidence extends jspb.Message { 
    getType(): string;
    setType(value: string): void;


    hasValidator(): boolean;
    clearValidator(): void;
    getValidator(): Validator | undefined;
    setValidator(value?: Validator): void;

    getHeight(): number;
    setHeight(value: number): void;


    hasTime(): boolean;
    clearTime(): void;
    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): void;

    getTotalVotingPower(): number;
    setTotalVotingPower(value: number): void;


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
        type: string,
        validator?: Validator.AsObject,
        height: number,
        time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        totalVotingPower: number,
    }
}
