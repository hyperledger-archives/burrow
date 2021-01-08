// package: payload
// file: payload.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as permission_pb from "./permission_pb";
import * as registry_pb from "./registry_pb";
import * as spec_pb from "./spec_pb";

export class Any extends jspb.Message { 

    hasCalltx(): boolean;
    clearCalltx(): void;
    getCalltx(): CallTx | undefined;
    setCalltx(value?: CallTx): Any;


    hasSendtx(): boolean;
    clearSendtx(): void;
    getSendtx(): SendTx | undefined;
    setSendtx(value?: SendTx): Any;


    hasNametx(): boolean;
    clearNametx(): void;
    getNametx(): NameTx | undefined;
    setNametx(value?: NameTx): Any;


    hasPermstx(): boolean;
    clearPermstx(): void;
    getPermstx(): PermsTx | undefined;
    setPermstx(value?: PermsTx): Any;


    hasGovtx(): boolean;
    clearGovtx(): void;
    getGovtx(): GovTx | undefined;
    setGovtx(value?: GovTx): Any;


    hasBondtx(): boolean;
    clearBondtx(): void;
    getBondtx(): BondTx | undefined;
    setBondtx(value?: BondTx): Any;


    hasUnbondtx(): boolean;
    clearUnbondtx(): void;
    getUnbondtx(): UnbondTx | undefined;
    setUnbondtx(value?: UnbondTx): Any;


    hasBatchtx(): boolean;
    clearBatchtx(): void;
    getBatchtx(): BatchTx | undefined;
    setBatchtx(value?: BatchTx): Any;


    hasProposaltx(): boolean;
    clearProposaltx(): void;
    getProposaltx(): ProposalTx | undefined;
    setProposaltx(value?: ProposalTx): Any;


    hasIdentifytx(): boolean;
    clearIdentifytx(): void;
    getIdentifytx(): IdentifyTx | undefined;
    setIdentifytx(value?: IdentifyTx): Any;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Any.AsObject;
    static toObject(includeInstance: boolean, msg: Any): Any.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Any, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Any;
    static deserializeBinaryFromReader(message: Any, reader: jspb.BinaryReader): Any;
}

export namespace Any {
    export type AsObject = {
        calltx?: CallTx.AsObject,
        sendtx?: SendTx.AsObject,
        nametx?: NameTx.AsObject,
        permstx?: PermsTx.AsObject,
        govtx?: GovTx.AsObject,
        bondtx?: BondTx.AsObject,
        unbondtx?: UnbondTx.AsObject,
        batchtx?: BatchTx.AsObject,
        proposaltx?: ProposalTx.AsObject,
        identifytx?: IdentifyTx.AsObject,
    }
}

export class TxInput extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): TxInput;

    getAmount(): number;
    setAmount(value: number): TxInput;

    getSequence(): number;
    setSequence(value: number): TxInput;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TxInput.AsObject;
    static toObject(includeInstance: boolean, msg: TxInput): TxInput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TxInput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TxInput;
    static deserializeBinaryFromReader(message: TxInput, reader: jspb.BinaryReader): TxInput;
}

export namespace TxInput {
    export type AsObject = {
        address: Uint8Array | string,
        amount: number,
        sequence: number,
    }
}

export class TxOutput extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): TxOutput;

    getAmount(): number;
    setAmount(value: number): TxOutput;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TxOutput.AsObject;
    static toObject(includeInstance: boolean, msg: TxOutput): TxOutput.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TxOutput, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TxOutput;
    static deserializeBinaryFromReader(message: TxOutput, reader: jspb.BinaryReader): TxOutput;
}

export namespace TxOutput {
    export type AsObject = {
        address: Uint8Array | string,
        amount: number,
    }
}

export class CallTx extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): TxInput | undefined;
    setInput(value?: TxInput): CallTx;

    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): CallTx;

    getGaslimit(): number;
    setGaslimit(value: number): CallTx;

    getFee(): number;
    setFee(value: number): CallTx;

    getData(): Uint8Array | string;
    getData_asU8(): Uint8Array;
    getData_asB64(): string;
    setData(value: Uint8Array | string): CallTx;

    getWasm(): Uint8Array | string;
    getWasm_asU8(): Uint8Array;
    getWasm_asB64(): string;
    setWasm(value: Uint8Array | string): CallTx;

    clearContractmetaList(): void;
    getContractmetaList(): Array<ContractMeta>;
    setContractmetaList(value: Array<ContractMeta>): CallTx;
    addContractmeta(value?: ContractMeta, index?: number): ContractMeta;

    getGasprice(): number;
    setGasprice(value: number): CallTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CallTx.AsObject;
    static toObject(includeInstance: boolean, msg: CallTx): CallTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CallTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CallTx;
    static deserializeBinaryFromReader(message: CallTx, reader: jspb.BinaryReader): CallTx;
}

export namespace CallTx {
    export type AsObject = {
        input?: TxInput.AsObject,
        address: Uint8Array | string,
        gaslimit: number,
        fee: number,
        data: Uint8Array | string,
        wasm: Uint8Array | string,
        contractmetaList: Array<ContractMeta.AsObject>,
        gasprice: number,
    }
}

export class ContractMeta extends jspb.Message { 
    getCodehash(): Uint8Array | string;
    getCodehash_asU8(): Uint8Array;
    getCodehash_asB64(): string;
    setCodehash(value: Uint8Array | string): ContractMeta;

    getMeta(): string;
    setMeta(value: string): ContractMeta;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ContractMeta.AsObject;
    static toObject(includeInstance: boolean, msg: ContractMeta): ContractMeta.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ContractMeta, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ContractMeta;
    static deserializeBinaryFromReader(message: ContractMeta, reader: jspb.BinaryReader): ContractMeta;
}

export namespace ContractMeta {
    export type AsObject = {
        codehash: Uint8Array | string,
        meta: string,
    }
}

export class SendTx extends jspb.Message { 
    clearInputsList(): void;
    getInputsList(): Array<TxInput>;
    setInputsList(value: Array<TxInput>): SendTx;
    addInputs(value?: TxInput, index?: number): TxInput;

    clearOutputsList(): void;
    getOutputsList(): Array<TxOutput>;
    setOutputsList(value: Array<TxOutput>): SendTx;
    addOutputs(value?: TxOutput, index?: number): TxOutput;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SendTx.AsObject;
    static toObject(includeInstance: boolean, msg: SendTx): SendTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SendTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SendTx;
    static deserializeBinaryFromReader(message: SendTx, reader: jspb.BinaryReader): SendTx;
}

export namespace SendTx {
    export type AsObject = {
        inputsList: Array<TxInput.AsObject>,
        outputsList: Array<TxOutput.AsObject>,
    }
}

export class PermsTx extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): TxInput | undefined;
    setInput(value?: TxInput): PermsTx;


    hasPermargs(): boolean;
    clearPermargs(): void;
    getPermargs(): permission_pb.PermArgs | undefined;
    setPermargs(value?: permission_pb.PermArgs): PermsTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PermsTx.AsObject;
    static toObject(includeInstance: boolean, msg: PermsTx): PermsTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PermsTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PermsTx;
    static deserializeBinaryFromReader(message: PermsTx, reader: jspb.BinaryReader): PermsTx;
}

export namespace PermsTx {
    export type AsObject = {
        input?: TxInput.AsObject,
        permargs?: permission_pb.PermArgs.AsObject,
    }
}

export class NameTx extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): TxInput | undefined;
    setInput(value?: TxInput): NameTx;

    getName(): string;
    setName(value: string): NameTx;

    getData(): string;
    setData(value: string): NameTx;

    getFee(): number;
    setFee(value: number): NameTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): NameTx.AsObject;
    static toObject(includeInstance: boolean, msg: NameTx): NameTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: NameTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): NameTx;
    static deserializeBinaryFromReader(message: NameTx, reader: jspb.BinaryReader): NameTx;
}

export namespace NameTx {
    export type AsObject = {
        input?: TxInput.AsObject,
        name: string,
        data: string,
        fee: number,
    }
}

export class BondTx extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): TxInput | undefined;
    setInput(value?: TxInput): BondTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BondTx.AsObject;
    static toObject(includeInstance: boolean, msg: BondTx): BondTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BondTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BondTx;
    static deserializeBinaryFromReader(message: BondTx, reader: jspb.BinaryReader): BondTx;
}

export namespace BondTx {
    export type AsObject = {
        input?: TxInput.AsObject,
    }
}

export class UnbondTx extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): TxInput | undefined;
    setInput(value?: TxInput): UnbondTx;


    hasOutput(): boolean;
    clearOutput(): void;
    getOutput(): TxOutput | undefined;
    setOutput(value?: TxOutput): UnbondTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UnbondTx.AsObject;
    static toObject(includeInstance: boolean, msg: UnbondTx): UnbondTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UnbondTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UnbondTx;
    static deserializeBinaryFromReader(message: UnbondTx, reader: jspb.BinaryReader): UnbondTx;
}

export namespace UnbondTx {
    export type AsObject = {
        input?: TxInput.AsObject,
        output?: TxOutput.AsObject,
    }
}

export class GovTx extends jspb.Message { 
    clearInputsList(): void;
    getInputsList(): Array<TxInput>;
    setInputsList(value: Array<TxInput>): GovTx;
    addInputs(value?: TxInput, index?: number): TxInput;

    clearAccountupdatesList(): void;
    getAccountupdatesList(): Array<spec_pb.TemplateAccount>;
    setAccountupdatesList(value: Array<spec_pb.TemplateAccount>): GovTx;
    addAccountupdates(value?: spec_pb.TemplateAccount, index?: number): spec_pb.TemplateAccount;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GovTx.AsObject;
    static toObject(includeInstance: boolean, msg: GovTx): GovTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GovTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GovTx;
    static deserializeBinaryFromReader(message: GovTx, reader: jspb.BinaryReader): GovTx;
}

export namespace GovTx {
    export type AsObject = {
        inputsList: Array<TxInput.AsObject>,
        accountupdatesList: Array<spec_pb.TemplateAccount.AsObject>,
    }
}

export class ProposalTx extends jspb.Message { 

    hasInput(): boolean;
    clearInput(): void;
    getInput(): TxInput | undefined;
    setInput(value?: TxInput): ProposalTx;

    getVotingweight(): number;
    setVotingweight(value: number): ProposalTx;

    getProposalhash(): Uint8Array | string;
    getProposalhash_asU8(): Uint8Array;
    getProposalhash_asB64(): string;
    setProposalhash(value: Uint8Array | string): ProposalTx;


    hasProposal(): boolean;
    clearProposal(): void;
    getProposal(): Proposal | undefined;
    setProposal(value?: Proposal): ProposalTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProposalTx.AsObject;
    static toObject(includeInstance: boolean, msg: ProposalTx): ProposalTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProposalTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProposalTx;
    static deserializeBinaryFromReader(message: ProposalTx, reader: jspb.BinaryReader): ProposalTx;
}

export namespace ProposalTx {
    export type AsObject = {
        input?: TxInput.AsObject,
        votingweight: number,
        proposalhash: Uint8Array | string,
        proposal?: Proposal.AsObject,
    }
}

export class IdentifyTx extends jspb.Message { 
    clearInputsList(): void;
    getInputsList(): Array<TxInput>;
    setInputsList(value: Array<TxInput>): IdentifyTx;
    addInputs(value?: TxInput, index?: number): TxInput;


    hasNode(): boolean;
    clearNode(): void;
    getNode(): registry_pb.NodeIdentity | undefined;
    setNode(value?: registry_pb.NodeIdentity): IdentifyTx;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): IdentifyTx.AsObject;
    static toObject(includeInstance: boolean, msg: IdentifyTx): IdentifyTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: IdentifyTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): IdentifyTx;
    static deserializeBinaryFromReader(message: IdentifyTx, reader: jspb.BinaryReader): IdentifyTx;
}

export namespace IdentifyTx {
    export type AsObject = {
        inputsList: Array<TxInput.AsObject>,
        node?: registry_pb.NodeIdentity.AsObject,
    }
}

export class BatchTx extends jspb.Message { 
    clearInputsList(): void;
    getInputsList(): Array<TxInput>;
    setInputsList(value: Array<TxInput>): BatchTx;
    addInputs(value?: TxInput, index?: number): TxInput;

    clearTxsList(): void;
    getTxsList(): Array<Any>;
    setTxsList(value: Array<Any>): BatchTx;
    addTxs(value?: Any, index?: number): Any;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BatchTx.AsObject;
    static toObject(includeInstance: boolean, msg: BatchTx): BatchTx.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BatchTx, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BatchTx;
    static deserializeBinaryFromReader(message: BatchTx, reader: jspb.BinaryReader): BatchTx;
}

export namespace BatchTx {
    export type AsObject = {
        inputsList: Array<TxInput.AsObject>,
        txsList: Array<Any.AsObject>,
    }
}

export class Vote extends jspb.Message { 
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): Vote;

    getVotingweight(): number;
    setVotingweight(value: number): Vote;


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
        address: Uint8Array | string,
        votingweight: number,
    }
}

export class Proposal extends jspb.Message { 
    getName(): string;
    setName(value: string): Proposal;

    getDescription(): string;
    setDescription(value: string): Proposal;


    hasBatchtx(): boolean;
    clearBatchtx(): void;
    getBatchtx(): BatchTx | undefined;
    setBatchtx(value?: BatchTx): Proposal;


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
        name: string,
        description: string,
        batchtx?: BatchTx.AsObject,
    }
}

export class Ballot extends jspb.Message { 

    hasProposal(): boolean;
    clearProposal(): void;
    getProposal(): Proposal | undefined;
    setProposal(value?: Proposal): Ballot;

    getFinalizingtx(): Uint8Array | string;
    getFinalizingtx_asU8(): Uint8Array;
    getFinalizingtx_asB64(): string;
    setFinalizingtx(value: Uint8Array | string): Ballot;

    getProposalstate(): Ballot.ProposalState;
    setProposalstate(value: Ballot.ProposalState): Ballot;

    clearVotesList(): void;
    getVotesList(): Array<Vote>;
    setVotesList(value: Array<Vote>): Ballot;
    addVotes(value?: Vote, index?: number): Vote;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Ballot.AsObject;
    static toObject(includeInstance: boolean, msg: Ballot): Ballot.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Ballot, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Ballot;
    static deserializeBinaryFromReader(message: Ballot, reader: jspb.BinaryReader): Ballot;
}

export namespace Ballot {
    export type AsObject = {
        proposal?: Proposal.AsObject,
        finalizingtx: Uint8Array | string,
        proposalstate: Ballot.ProposalState,
        votesList: Array<Vote.AsObject>,
    }

    export enum ProposalState {
    PROPOSED = 0,
    EXECUTED = 1,
    FAILED = 2,
    }

}
