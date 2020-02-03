// package: filecoin.ask.pb
// file: ask.proto

import * as jspb from "google-protobuf";

export class Query extends jspb.Message {
  getMaxprice(): number;
  setMaxprice(value: number): void;

  getPiecesize(): number;
  setPiecesize(value: number): void;

  getLimit(): number;
  setLimit(value: number): void;

  getOffset(): number;
  setOffset(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Query.AsObject;
  static toObject(includeInstance: boolean, msg: Query): Query.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Query, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Query;
  static deserializeBinaryFromReader(message: Query, reader: jspb.BinaryReader): Query;
}

export namespace Query {
  export type AsObject = {
    maxprice: number,
    piecesize: number,
    limit: number,
    offset: number,
  }
}

export class StorageAsk extends jspb.Message {
  getPrice(): number;
  setPrice(value: number): void;

  getMinpiecesize(): number;
  setMinpiecesize(value: number): void;

  getMiner(): string;
  setMiner(value: string): void;

  getTimestamp(): number;
  setTimestamp(value: number): void;

  getExpiry(): number;
  setExpiry(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StorageAsk.AsObject;
  static toObject(includeInstance: boolean, msg: StorageAsk): StorageAsk.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StorageAsk, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StorageAsk;
  static deserializeBinaryFromReader(message: StorageAsk, reader: jspb.BinaryReader): StorageAsk;
}

export namespace StorageAsk {
  export type AsObject = {
    price: number,
    minpiecesize: number,
    miner: string,
    timestamp: number,
    expiry: number,
  }
}

export class Index extends jspb.Message {
  getLastupdated(): number;
  setLastupdated(value: number): void;

  getStoragemedianprice(): number;
  setStoragemedianprice(value: number): void;

  getStorageMap(): jspb.Map<string, StorageAsk>;
  clearStorageMap(): void;
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
    lastupdated: number,
    storagemedianprice: number,
    storageMap: Array<[string, StorageAsk.AsObject]>,
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

export class QueryRequest extends jspb.Message {
  hasQuery(): boolean;
  clearQuery(): void;
  getQuery(): Query | undefined;
  setQuery(value?: Query): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): QueryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: QueryRequest): QueryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: QueryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): QueryRequest;
  static deserializeBinaryFromReader(message: QueryRequest, reader: jspb.BinaryReader): QueryRequest;
}

export namespace QueryRequest {
  export type AsObject = {
    query?: Query.AsObject,
  }
}

export class QueryReply extends jspb.Message {
  clearAsksList(): void;
  getAsksList(): Array<StorageAsk>;
  setAsksList(value: Array<StorageAsk>): void;
  addAsks(value?: StorageAsk, index?: number): StorageAsk;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): QueryReply.AsObject;
  static toObject(includeInstance: boolean, msg: QueryReply): QueryReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: QueryReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): QueryReply;
  static deserializeBinaryFromReader(message: QueryReply, reader: jspb.BinaryReader): QueryReply;
}

export namespace QueryReply {
  export type AsObject = {
    asksList: Array<StorageAsk.AsObject>,
  }
}

