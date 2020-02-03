// package: filecoin.slashing.pb
// file: slashing.proto

import * as jspb from "google-protobuf";

export class Index extends jspb.Message {
  getTipsetkey(): string;
  setTipsetkey(value: string): void;

  getMinersMap(): jspb.Map<string, Slashes>;
  clearMinersMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Index.AsObject;
  static toObject(includeInstance: boolean, msg: Index): Index.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Index, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Index;
  static deserializeBinaryFromReader(message: Index, reader: jspb.BinaryReader): Index;
}

export namespace Index {
  export type AsObject = {
    tipsetkey: string,
    minersMap: Array<[string, Slashes.AsObject]>,
  }
}

export class Slashes extends jspb.Message {
  clearEpochsList(): void;
  getEpochsList(): Array<number>;
  setEpochsList(value: Array<number>): void;
  addEpochs(value: number, index?: number): number;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Slashes.AsObject;
  static toObject(includeInstance: boolean, msg: Slashes): Slashes.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Slashes, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Slashes;
  static deserializeBinaryFromReader(message: Slashes, reader: jspb.BinaryReader): Slashes;
}

export namespace Slashes {
  export type AsObject = {
    epochsList: Array<number>,
  }
}

export class GetRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetRequest): GetRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetRequest;
  static deserializeBinaryFromReader(message: GetRequest, reader: jspb.BinaryReader): GetRequest;
}

export namespace GetRequest {
  export type AsObject = {
  }
}

export class GetReply extends jspb.Message {
  hasIndex(): boolean;
  clearIndex(): void;
  getIndex(): Index | undefined;
  setIndex(value?: Index): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetReply.AsObject;
  static toObject(includeInstance: boolean, msg: GetReply): GetReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetReply;
  static deserializeBinaryFromReader(message: GetReply, reader: jspb.BinaryReader): GetReply;
}

export namespace GetReply {
  export type AsObject = {
    index?: Index.AsObject,
  }
}

