// package: filecoin.miner.pb
// file: miner.proto

import * as jspb from "google-protobuf";

export class Index extends jspb.Message {
  hasMeta(): boolean;
  clearMeta(): void;
  getMeta(): MetaIndex | undefined;
  setMeta(value?: MetaIndex): void;

  hasChain(): boolean;
  clearChain(): void;
  getChain(): ChainIndex | undefined;
  setChain(value?: ChainIndex): void;

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
    meta?: MetaIndex.AsObject,
    chain?: ChainIndex.AsObject,
  }
}

export class ChainIndex extends jspb.Message {
  getLastupdated(): number;
  setLastupdated(value: number): void;

  getPowerMap(): jspb.Map<string, Power>;
  clearPowerMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChainIndex.AsObject;
  static toObject(includeInstance: boolean, msg: ChainIndex): ChainIndex.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChainIndex, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChainIndex;
  static deserializeBinaryFromReader(message: ChainIndex, reader: jspb.BinaryReader): ChainIndex;
}

export namespace ChainIndex {
  export type AsObject = {
    lastupdated: number,
    powerMap: Array<[string, Power.AsObject]>,
  }
}

export class Power extends jspb.Message {
  getPower(): number;
  setPower(value: number): void;

  getRelative(): number;
  setRelative(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Power.AsObject;
  static toObject(includeInstance: boolean, msg: Power): Power.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Power, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Power;
  static deserializeBinaryFromReader(message: Power, reader: jspb.BinaryReader): Power;
}

export namespace Power {
  export type AsObject = {
    power: number,
    relative: number,
  }
}

export class MetaIndex extends jspb.Message {
  getOnline(): number;
  setOnline(value: number): void;

  getOffline(): number;
  setOffline(value: number): void;

  getInfoMap(): jspb.Map<string, Meta>;
  clearInfoMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MetaIndex.AsObject;
  static toObject(includeInstance: boolean, msg: MetaIndex): MetaIndex.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MetaIndex, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MetaIndex;
  static deserializeBinaryFromReader(message: MetaIndex, reader: jspb.BinaryReader): MetaIndex;
}

export namespace MetaIndex {
  export type AsObject = {
    online: number,
    offline: number,
    infoMap: Array<[string, Meta.AsObject]>,
  }
}

export class Meta extends jspb.Message {
  getLastupdated(): number;
  setLastupdated(value: number): void;

  getUseragent(): string;
  setUseragent(value: string): void;

  hasLocation(): boolean;
  clearLocation(): void;
  getLocation(): Location | undefined;
  setLocation(value?: Location): void;

  getOnline(): boolean;
  setOnline(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Meta.AsObject;
  static toObject(includeInstance: boolean, msg: Meta): Meta.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Meta, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Meta;
  static deserializeBinaryFromReader(message: Meta, reader: jspb.BinaryReader): Meta;
}

export namespace Meta {
  export type AsObject = {
    lastupdated: number,
    useragent: string,
    location?: Location.AsObject,
    online: boolean,
  }
}

export class Location extends jspb.Message {
  getCountry(): string;
  setCountry(value: string): void;

  getLongitude(): number;
  setLongitude(value: number): void;

  getLatitude(): number;
  setLatitude(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Location.AsObject;
  static toObject(includeInstance: boolean, msg: Location): Location.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Location, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Location;
  static deserializeBinaryFromReader(message: Location, reader: jspb.BinaryReader): Location;
}

export namespace Location {
  export type AsObject = {
    country: string,
    longitude: number,
    latitude: number,
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

