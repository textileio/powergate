// package: filecoin.reputation.pb
// file: reputation.proto

import * as jspb from "google-protobuf";

export class MinerScore extends jspb.Message {
  getAddr(): string;
  setAddr(value: string): void;

  getScore(): number;
  setScore(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MinerScore.AsObject;
  static toObject(includeInstance: boolean, msg: MinerScore): MinerScore.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MinerScore, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MinerScore;
  static deserializeBinaryFromReader(message: MinerScore, reader: jspb.BinaryReader): MinerScore;
}

export namespace MinerScore {
  export type AsObject = {
    addr: string,
    score: number,
  }
}

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

export class AddSourceRequest extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getMaddr(): string;
  setMaddr(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddSourceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddSourceRequest): AddSourceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddSourceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddSourceRequest;
  static deserializeBinaryFromReader(message: AddSourceRequest, reader: jspb.BinaryReader): AddSourceRequest;
}

export namespace AddSourceRequest {
  export type AsObject = {
    id: string,
    maddr: string,
  }
}

export class AddSourceReply extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddSourceReply.AsObject;
  static toObject(includeInstance: boolean, msg: AddSourceReply): AddSourceReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddSourceReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddSourceReply;
  static deserializeBinaryFromReader(message: AddSourceReply, reader: jspb.BinaryReader): AddSourceReply;
}

export namespace AddSourceReply {
  export type AsObject = {
  }
}

export class GetTopMinersRequest extends jspb.Message {
  getLimit(): number;
  setLimit(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetTopMinersRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetTopMinersRequest): GetTopMinersRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetTopMinersRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetTopMinersRequest;
  static deserializeBinaryFromReader(message: GetTopMinersRequest, reader: jspb.BinaryReader): GetTopMinersRequest;
}

export namespace GetTopMinersRequest {
  export type AsObject = {
    limit: number,
  }
}

export class GetTopMinersReply extends jspb.Message {
  clearTopminersList(): void;
  getTopminersList(): Array<MinerScore>;
  setTopminersList(value: Array<MinerScore>): void;
  addTopminers(value?: MinerScore, index?: number): MinerScore;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetTopMinersReply.AsObject;
  static toObject(includeInstance: boolean, msg: GetTopMinersReply): GetTopMinersReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetTopMinersReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetTopMinersReply;
  static deserializeBinaryFromReader(message: GetTopMinersReply, reader: jspb.BinaryReader): GetTopMinersReply;
}

export namespace GetTopMinersReply {
  export type AsObject = {
    topminersList: Array<MinerScore.AsObject>,
  }
}

