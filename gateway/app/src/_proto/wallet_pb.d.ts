// package: filecoin.wallet.pb
// file: wallet.proto

import * as jspb from "google-protobuf";

export class NewWalletRequest extends jspb.Message {
  getTyp(): string;
  setTyp(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NewWalletRequest.AsObject;
  static toObject(includeInstance: boolean, msg: NewWalletRequest): NewWalletRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: NewWalletRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NewWalletRequest;
  static deserializeBinaryFromReader(message: NewWalletRequest, reader: jspb.BinaryReader): NewWalletRequest;
}

export namespace NewWalletRequest {
  export type AsObject = {
    typ: string,
  }
}

export class NewWalletReply extends jspb.Message {
  getAddress(): string;
  setAddress(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NewWalletReply.AsObject;
  static toObject(includeInstance: boolean, msg: NewWalletReply): NewWalletReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: NewWalletReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NewWalletReply;
  static deserializeBinaryFromReader(message: NewWalletReply, reader: jspb.BinaryReader): NewWalletReply;
}

export namespace NewWalletReply {
  export type AsObject = {
    address: string,
  }
}

export class WalletBalanceRequest extends jspb.Message {
  getAddress(): string;
  setAddress(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WalletBalanceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: WalletBalanceRequest): WalletBalanceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WalletBalanceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WalletBalanceRequest;
  static deserializeBinaryFromReader(message: WalletBalanceRequest, reader: jspb.BinaryReader): WalletBalanceRequest;
}

export namespace WalletBalanceRequest {
  export type AsObject = {
    address: string,
  }
}

export class WalletBalanceReply extends jspb.Message {
  getBalance(): number;
  setBalance(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WalletBalanceReply.AsObject;
  static toObject(includeInstance: boolean, msg: WalletBalanceReply): WalletBalanceReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WalletBalanceReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WalletBalanceReply;
  static deserializeBinaryFromReader(message: WalletBalanceReply, reader: jspb.BinaryReader): WalletBalanceReply;
}

export namespace WalletBalanceReply {
  export type AsObject = {
    balance: number,
  }
}

