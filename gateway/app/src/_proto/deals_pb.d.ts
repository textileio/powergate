// package: filecoin.deals.pb
// file: deals.proto

import * as jspb from "google-protobuf";

export class DealConfig extends jspb.Message {
  getMiner(): string;
  setMiner(value: string): void;

  getEpochprice(): number;
  setEpochprice(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DealConfig.AsObject;
  static toObject(includeInstance: boolean, msg: DealConfig): DealConfig.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DealConfig, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DealConfig;
  static deserializeBinaryFromReader(message: DealConfig, reader: jspb.BinaryReader): DealConfig;
}

export namespace DealConfig {
  export type AsObject = {
    miner: string,
    epochprice: number,
  }
}

export class DealInfo extends jspb.Message {
  getProposalcid(): string;
  setProposalcid(value: string): void;

  getStateid(): number;
  setStateid(value: number): void;

  getStatename(): string;
  setStatename(value: string): void;

  getMiner(): string;
  setMiner(value: string): void;

  getPieceref(): Uint8Array | string;
  getPieceref_asU8(): Uint8Array;
  getPieceref_asB64(): string;
  setPieceref(value: Uint8Array | string): void;

  getSize(): number;
  setSize(value: number): void;

  getPriceperepoch(): number;
  setPriceperepoch(value: number): void;

  getDuration(): number;
  setDuration(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DealInfo.AsObject;
  static toObject(includeInstance: boolean, msg: DealInfo): DealInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DealInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DealInfo;
  static deserializeBinaryFromReader(message: DealInfo, reader: jspb.BinaryReader): DealInfo;
}

export namespace DealInfo {
  export type AsObject = {
    proposalcid: string,
    stateid: number,
    statename: string,
    miner: string,
    pieceref: Uint8Array | string,
    size: number,
    priceperepoch: number,
    duration: number,
  }
}

export class StoreParams extends jspb.Message {
  getAddress(): string;
  setAddress(value: string): void;

  clearDealconfigsList(): void;
  getDealconfigsList(): Array<DealConfig>;
  setDealconfigsList(value: Array<DealConfig>): void;
  addDealconfigs(value?: DealConfig, index?: number): DealConfig;

  getDuration(): number;
  setDuration(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreParams.AsObject;
  static toObject(includeInstance: boolean, msg: StoreParams): StoreParams.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StoreParams, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreParams;
  static deserializeBinaryFromReader(message: StoreParams, reader: jspb.BinaryReader): StoreParams;
}

export namespace StoreParams {
  export type AsObject = {
    address: string,
    dealconfigsList: Array<DealConfig.AsObject>,
    duration: number,
  }
}

export class StoreRequest extends jspb.Message {
  hasStoreparams(): boolean;
  clearStoreparams(): void;
  getStoreparams(): StoreParams | undefined;
  setStoreparams(value?: StoreParams): void;

  hasChunk(): boolean;
  clearChunk(): void;
  getChunk(): Uint8Array | string;
  getChunk_asU8(): Uint8Array;
  getChunk_asB64(): string;
  setChunk(value: Uint8Array | string): void;

  getPayloadCase(): StoreRequest.PayloadCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreRequest.AsObject;
  static toObject(includeInstance: boolean, msg: StoreRequest): StoreRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StoreRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreRequest;
  static deserializeBinaryFromReader(message: StoreRequest, reader: jspb.BinaryReader): StoreRequest;
}

export namespace StoreRequest {
  export type AsObject = {
    storeparams?: StoreParams.AsObject,
    chunk: Uint8Array | string,
  }

  export enum PayloadCase {
    PAYLOAD_NOT_SET = 0,
    STOREPARAMS = 1,
    CHUNK = 2,
  }
}

export class StoreReply extends jspb.Message {
  clearCidsList(): void;
  getCidsList(): Array<string>;
  setCidsList(value: Array<string>): void;
  addCids(value: string, index?: number): string;

  clearFaileddealsList(): void;
  getFaileddealsList(): Array<DealConfig>;
  setFaileddealsList(value: Array<DealConfig>): void;
  addFaileddeals(value?: DealConfig, index?: number): DealConfig;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StoreReply.AsObject;
  static toObject(includeInstance: boolean, msg: StoreReply): StoreReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: StoreReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StoreReply;
  static deserializeBinaryFromReader(message: StoreReply, reader: jspb.BinaryReader): StoreReply;
}

export namespace StoreReply {
  export type AsObject = {
    cidsList: Array<string>,
    faileddealsList: Array<DealConfig.AsObject>,
  }
}

export class WatchRequest extends jspb.Message {
  clearProposalsList(): void;
  getProposalsList(): Array<string>;
  setProposalsList(value: Array<string>): void;
  addProposals(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WatchRequest.AsObject;
  static toObject(includeInstance: boolean, msg: WatchRequest): WatchRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WatchRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WatchRequest;
  static deserializeBinaryFromReader(message: WatchRequest, reader: jspb.BinaryReader): WatchRequest;
}

export namespace WatchRequest {
  export type AsObject = {
    proposalsList: Array<string>,
  }
}

export class WatchReply extends jspb.Message {
  hasDealinfo(): boolean;
  clearDealinfo(): void;
  getDealinfo(): DealInfo | undefined;
  setDealinfo(value?: DealInfo): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WatchReply.AsObject;
  static toObject(includeInstance: boolean, msg: WatchReply): WatchReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WatchReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WatchReply;
  static deserializeBinaryFromReader(message: WatchReply, reader: jspb.BinaryReader): WatchReply;
}

export namespace WatchReply {
  export type AsObject = {
    dealinfo?: DealInfo.AsObject,
  }
}

