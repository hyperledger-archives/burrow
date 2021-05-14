// package: tendermint.consensus
// file: tendermint/consensus/wal.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
import * as tendermint_consensus_types_pb from "../../tendermint/consensus/types_pb";
import * as tendermint_types_events_pb from "../../tendermint/types/events_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class MsgInfo extends jspb.Message { 

    hasMsg(): boolean;
    clearMsg(): void;
    getMsg(): tendermint_consensus_types_pb.Message | undefined;
    setMsg(value?: tendermint_consensus_types_pb.Message): MsgInfo;
    getPeerId(): string;
    setPeerId(value: string): MsgInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MsgInfo.AsObject;
    static toObject(includeInstance: boolean, msg: MsgInfo): MsgInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: MsgInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MsgInfo;
    static deserializeBinaryFromReader(message: MsgInfo, reader: jspb.BinaryReader): MsgInfo;
}

export namespace MsgInfo {
    export type AsObject = {
        msg?: tendermint_consensus_types_pb.Message.AsObject,
        peerId: string,
    }
}

export class TimeoutInfo extends jspb.Message { 

    hasDuration(): boolean;
    clearDuration(): void;
    getDuration(): google_protobuf_duration_pb.Duration | undefined;
    setDuration(value?: google_protobuf_duration_pb.Duration): TimeoutInfo;
    getHeight(): number;
    setHeight(value: number): TimeoutInfo;
    getRound(): number;
    setRound(value: number): TimeoutInfo;
    getStep(): number;
    setStep(value: number): TimeoutInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TimeoutInfo.AsObject;
    static toObject(includeInstance: boolean, msg: TimeoutInfo): TimeoutInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TimeoutInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TimeoutInfo;
    static deserializeBinaryFromReader(message: TimeoutInfo, reader: jspb.BinaryReader): TimeoutInfo;
}

export namespace TimeoutInfo {
    export type AsObject = {
        duration?: google_protobuf_duration_pb.Duration.AsObject,
        height: number,
        round: number,
        step: number,
    }
}

export class EndHeight extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): EndHeight;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EndHeight.AsObject;
    static toObject(includeInstance: boolean, msg: EndHeight): EndHeight.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EndHeight, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EndHeight;
    static deserializeBinaryFromReader(message: EndHeight, reader: jspb.BinaryReader): EndHeight;
}

export namespace EndHeight {
    export type AsObject = {
        height: number,
    }
}

export class WALMessage extends jspb.Message { 

    hasEventDataRoundState(): boolean;
    clearEventDataRoundState(): void;
    getEventDataRoundState(): tendermint_types_events_pb.EventDataRoundState | undefined;
    setEventDataRoundState(value?: tendermint_types_events_pb.EventDataRoundState): WALMessage;

    hasMsgInfo(): boolean;
    clearMsgInfo(): void;
    getMsgInfo(): MsgInfo | undefined;
    setMsgInfo(value?: MsgInfo): WALMessage;

    hasTimeoutInfo(): boolean;
    clearTimeoutInfo(): void;
    getTimeoutInfo(): TimeoutInfo | undefined;
    setTimeoutInfo(value?: TimeoutInfo): WALMessage;

    hasEndHeight(): boolean;
    clearEndHeight(): void;
    getEndHeight(): EndHeight | undefined;
    setEndHeight(value?: EndHeight): WALMessage;

    getSumCase(): WALMessage.SumCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WALMessage.AsObject;
    static toObject(includeInstance: boolean, msg: WALMessage): WALMessage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WALMessage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WALMessage;
    static deserializeBinaryFromReader(message: WALMessage, reader: jspb.BinaryReader): WALMessage;
}

export namespace WALMessage {
    export type AsObject = {
        eventDataRoundState?: tendermint_types_events_pb.EventDataRoundState.AsObject,
        msgInfo?: MsgInfo.AsObject,
        timeoutInfo?: TimeoutInfo.AsObject,
        endHeight?: EndHeight.AsObject,
    }

    export enum SumCase {
        SUM_NOT_SET = 0,
        EVENT_DATA_ROUND_STATE = 1,
        MSG_INFO = 2,
        TIMEOUT_INFO = 3,
        END_HEIGHT = 4,
    }

}

export class TimedWALMessage extends jspb.Message { 

    hasTime(): boolean;
    clearTime(): void;
    getTime(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setTime(value?: google_protobuf_timestamp_pb.Timestamp): TimedWALMessage;

    hasMsg(): boolean;
    clearMsg(): void;
    getMsg(): WALMessage | undefined;
    setMsg(value?: WALMessage): TimedWALMessage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TimedWALMessage.AsObject;
    static toObject(includeInstance: boolean, msg: TimedWALMessage): TimedWALMessage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TimedWALMessage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TimedWALMessage;
    static deserializeBinaryFromReader(message: TimedWALMessage, reader: jspb.BinaryReader): TimedWALMessage;
}

export namespace TimedWALMessage {
    export type AsObject = {
        time?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        msg?: WALMessage.AsObject,
    }
}
