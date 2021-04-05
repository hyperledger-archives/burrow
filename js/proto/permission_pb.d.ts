// package: permission
// file: permission.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "./gogoproto/gogo_pb";

export class AccountPermissions extends jspb.Message { 

    hasBase(): boolean;
    clearBase(): void;
    getBase(): BasePermissions | undefined;
    setBase(value?: BasePermissions): AccountPermissions;

    clearRolesList(): void;
    getRolesList(): Array<string>;
    setRolesList(value: Array<string>): AccountPermissions;
    addRoles(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AccountPermissions.AsObject;
    static toObject(includeInstance: boolean, msg: AccountPermissions): AccountPermissions.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AccountPermissions, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AccountPermissions;
    static deserializeBinaryFromReader(message: AccountPermissions, reader: jspb.BinaryReader): AccountPermissions;
}

export namespace AccountPermissions {
    export type AsObject = {
        base?: BasePermissions.AsObject,
        rolesList: Array<string>,
    }
}

export class BasePermissions extends jspb.Message { 

    hasPerms(): boolean;
    clearPerms(): void;
    getPerms(): number | undefined;
    setPerms(value: number): BasePermissions;


    hasSetbit(): boolean;
    clearSetbit(): void;
    getSetbit(): number | undefined;
    setSetbit(value: number): BasePermissions;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): BasePermissions.AsObject;
    static toObject(includeInstance: boolean, msg: BasePermissions): BasePermissions.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: BasePermissions, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): BasePermissions;
    static deserializeBinaryFromReader(message: BasePermissions, reader: jspb.BinaryReader): BasePermissions;
}

export namespace BasePermissions {
    export type AsObject = {
        perms?: number,
        setbit?: number,
    }
}

export class PermArgs extends jspb.Message { 

    hasAction(): boolean;
    clearAction(): void;
    getAction(): number | undefined;
    setAction(value: number): PermArgs;


    hasTarget(): boolean;
    clearTarget(): void;
    getTarget(): Uint8Array | string;
    getTarget_asU8(): Uint8Array;
    getTarget_asB64(): string;
    setTarget(value: Uint8Array | string): PermArgs;


    hasPermission(): boolean;
    clearPermission(): void;
    getPermission(): number | undefined;
    setPermission(value: number): PermArgs;


    hasRole(): boolean;
    clearRole(): void;
    getRole(): string | undefined;
    setRole(value: string): PermArgs;


    hasValue(): boolean;
    clearValue(): void;
    getValue(): boolean | undefined;
    setValue(value: boolean): PermArgs;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PermArgs.AsObject;
    static toObject(includeInstance: boolean, msg: PermArgs): PermArgs.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PermArgs, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PermArgs;
    static deserializeBinaryFromReader(message: PermArgs, reader: jspb.BinaryReader): PermArgs;
}

export namespace PermArgs {
    export type AsObject = {
        action?: number,
        target: Uint8Array | string,
        permission?: number,
        role?: string,
        value?: boolean,
    }
}
