// package: tendermint.types
// file: tendermint/types/events.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class EventDataRoundState extends jspb.Message { 
    getHeight(): number;
    setHeight(value: number): EventDataRoundState;

    getRound(): number;
    setRound(value: number): EventDataRoundState;

    getStep(): string;
    setStep(value: string): EventDataRoundState;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EventDataRoundState.AsObject;
    static toObject(includeInstance: boolean, msg: EventDataRoundState): EventDataRoundState.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EventDataRoundState, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EventDataRoundState;
    static deserializeBinaryFromReader(message: EventDataRoundState, reader: jspb.BinaryReader): EventDataRoundState;
}

export namespace EventDataRoundState {
    export type AsObject = {
        height: number,
        round: number,
        step: string,
    }
}
