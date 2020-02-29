// package: txs
// file: txs.proto

import * as jspb from "google-protobuf";
import * as github_com_gogo_protobuf_gogoproto_gogo_pb from "./github.com/gogo/protobuf/gogoproto/gogo_pb";
import * as crypto_pb from "./crypto_pb";

export class Envelope extends jspb.Message {
  clearSignatoriesList(): void;
  getSignatoriesList(): Array<Signatory>;
  setSignatoriesList(value: Array<Signatory>): void;
  addSignatories(value?: Signatory, index?: number): Signatory;

  getTx(): Uint8Array | string;
  getTx_asU8(): Uint8Array;
  getTx_asB64(): string;
  setTx(value: Uint8Array | string): void;

  getEnc(): Envelope.EncodingMap[keyof Envelope.EncodingMap];
  setEnc(value: Envelope.EncodingMap[keyof Envelope.EncodingMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Envelope.AsObject;
  static toObject(includeInstance: boolean, msg: Envelope): Envelope.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Envelope, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Envelope;
  static deserializeBinaryFromReader(message: Envelope, reader: jspb.BinaryReader): Envelope;
}

export namespace Envelope {
  export type AsObject = {
    signatoriesList: Array<Signatory.AsObject>,
    tx: Uint8Array | string,
    enc: Envelope.EncodingMap[keyof Envelope.EncodingMap],
  }

  export interface EncodingMap {
    JSON: 0;
    RLP: 1;
  }

  export const Encoding: EncodingMap;
}

export class Signatory extends jspb.Message {
  getAddress(): Uint8Array | string;
  getAddress_asU8(): Uint8Array;
  getAddress_asB64(): string;
  setAddress(value: Uint8Array | string): void;

  hasPublickey(): boolean;
  clearPublickey(): void;
  getPublickey(): crypto_pb.PublicKey | undefined;
  setPublickey(value?: crypto_pb.PublicKey): void;

  hasSignature(): boolean;
  clearSignature(): void;
  getSignature(): crypto_pb.Signature | undefined;
  setSignature(value?: crypto_pb.Signature): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Signatory.AsObject;
  static toObject(includeInstance: boolean, msg: Signatory): Signatory.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Signatory, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Signatory;
  static deserializeBinaryFromReader(message: Signatory, reader: jspb.BinaryReader): Signatory;
}

export namespace Signatory {
  export type AsObject = {
    address: Uint8Array | string,
    publickey?: crypto_pb.PublicKey.AsObject,
    signature?: crypto_pb.Signature.AsObject,
  }
}

export class Receipt extends jspb.Message {
  getTxtype(): number;
  setTxtype(value: number): void;

  getTxhash(): Uint8Array | string;
  getTxhash_asU8(): Uint8Array;
  getTxhash_asB64(): string;
  setTxhash(value: Uint8Array | string): void;

  getCreatescontract(): boolean;
  setCreatescontract(value: boolean): void;

  getContractaddress(): Uint8Array | string;
  getContractaddress_asU8(): Uint8Array;
  getContractaddress_asB64(): string;
  setContractaddress(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Receipt.AsObject;
  static toObject(includeInstance: boolean, msg: Receipt): Receipt.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Receipt, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Receipt;
  static deserializeBinaryFromReader(message: Receipt, reader: jspb.BinaryReader): Receipt;
}

export namespace Receipt {
  export type AsObject = {
    txtype: number,
    txhash: Uint8Array | string,
    createscontract: boolean,
    contractaddress: Uint8Array | string,
  }
}

