// package: spec
// file: spec.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";
import * as crypto_pb from "./crypto_pb";
import * as balance_pb from "./balance_pb";

export class TemplateAccount extends jspb.Message { 
    getName(): string;
    setName(value: string): TemplateAccount;
    getAddress(): Uint8Array | string;
    getAddress_asU8(): Uint8Array;
    getAddress_asB64(): string;
    setAddress(value: Uint8Array | string): TemplateAccount;

    hasPublickey(): boolean;
    clearPublickey(): void;
    getPublickey(): crypto_pb.PublicKey | undefined;
    setPublickey(value?: crypto_pb.PublicKey): TemplateAccount;
    clearAmountsList(): void;
    getAmountsList(): Array<balance_pb.Balance>;
    setAmountsList(value: Array<balance_pb.Balance>): TemplateAccount;
    addAmounts(value?: balance_pb.Balance, index?: number): balance_pb.Balance;
    clearPermissionsList(): void;
    getPermissionsList(): Array<string>;
    setPermissionsList(value: Array<string>): TemplateAccount;
    addPermissions(value: string, index?: number): string;
    clearRolesList(): void;
    getRolesList(): Array<string>;
    setRolesList(value: Array<string>): TemplateAccount;
    addRoles(value: string, index?: number): string;
    getCode(): Uint8Array | string;
    getCode_asU8(): Uint8Array;
    getCode_asB64(): string;
    setCode(value: Uint8Array | string): TemplateAccount;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TemplateAccount.AsObject;
    static toObject(includeInstance: boolean, msg: TemplateAccount): TemplateAccount.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TemplateAccount, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TemplateAccount;
    static deserializeBinaryFromReader(message: TemplateAccount, reader: jspb.BinaryReader): TemplateAccount;
}

export namespace TemplateAccount {
    export type AsObject = {
        name: string,
        address: Uint8Array | string,
        publickey?: crypto_pb.PublicKey.AsObject,
        amountsList: Array<balance_pb.Balance.AsObject>,
        permissionsList: Array<string>,
        rolesList: Array<string>,
        code: Uint8Array | string,
    }
}
