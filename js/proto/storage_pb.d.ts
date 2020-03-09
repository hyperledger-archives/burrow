// package: storage
// file: storage.proto

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";

export class CommitID extends jspb.Message {
  getVersion(): number;
  setVersion(value: number): void;

  getHash(): Uint8Array | string;
  getHash_asU8(): Uint8Array;
  getHash_asB64(): string;
  setHash(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CommitID.AsObject;
  static toObject(includeInstance: boolean, msg: CommitID): CommitID.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CommitID, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CommitID;
  static deserializeBinaryFromReader(message: CommitID, reader: jspb.BinaryReader): CommitID;
}

export namespace CommitID {
  export type AsObject = {
    version: number,
    hash: Uint8Array | string,
  }
}

