// package: filecoin.fpa.pb
// file: fpa.proto

import * as jspb from "google-protobuf";

export class AddCidRequest extends jspb.Message {
  getCid(): string;
  setCid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddCidRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddCidRequest): AddCidRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddCidRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddCidRequest;
  static deserializeBinaryFromReader(message: AddCidRequest, reader: jspb.BinaryReader): AddCidRequest;
}

export namespace AddCidRequest {
  export type AsObject = {
    cid: string,
  }
}

export class AddCidReply extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddCidReply.AsObject;
  static toObject(includeInstance: boolean, msg: AddCidReply): AddCidReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddCidReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddCidReply;
  static deserializeBinaryFromReader(message: AddCidReply, reader: jspb.BinaryReader): AddCidReply;
}

export namespace AddCidReply {
  export type AsObject = {
  }
}

export class AddFileRequest extends jspb.Message {
  getChunk(): Uint8Array | string;
  getChunk_asU8(): Uint8Array;
  getChunk_asB64(): string;
  setChunk(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddFileRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddFileRequest): AddFileRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddFileRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddFileRequest;
  static deserializeBinaryFromReader(message: AddFileRequest, reader: jspb.BinaryReader): AddFileRequest;
}

export namespace AddFileRequest {
  export type AsObject = {
    chunk: Uint8Array | string,
  }
}

export class AddFileReply extends jspb.Message {
  getCid(): string;
  setCid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddFileReply.AsObject;
  static toObject(includeInstance: boolean, msg: AddFileReply): AddFileReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddFileReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddFileReply;
  static deserializeBinaryFromReader(message: AddFileReply, reader: jspb.BinaryReader): AddFileReply;
}

export namespace AddFileReply {
  export type AsObject = {
    cid: string,
  }
}

export class GetRequest extends jspb.Message {
  getCid(): string;
  setCid(value: string): void;

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
    cid: string,
  }
}

export class GetReply extends jspb.Message {
  getChunk(): Uint8Array | string;
  getChunk_asU8(): Uint8Array;
  getChunk_asB64(): string;
  setChunk(value: Uint8Array | string): void;

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
    chunk: Uint8Array | string,
  }
}

export class CreateRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateRequest): CreateRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateRequest;
  static deserializeBinaryFromReader(message: CreateRequest, reader: jspb.BinaryReader): CreateRequest;
}

export namespace CreateRequest {
  export type AsObject = {
  }
}

export class CreateReply extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getAddress(): string;
  setAddress(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateReply.AsObject;
  static toObject(includeInstance: boolean, msg: CreateReply): CreateReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateReply;
  static deserializeBinaryFromReader(message: CreateReply, reader: jspb.BinaryReader): CreateReply;
}

export namespace CreateReply {
  export type AsObject = {
    id: string,
    address: string,
  }
}

export class ShowRequest extends jspb.Message {
  getCid(): string;
  setCid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ShowRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ShowRequest): ShowRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ShowRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ShowRequest;
  static deserializeBinaryFromReader(message: ShowRequest, reader: jspb.BinaryReader): ShowRequest;
}

export namespace ShowRequest {
  export type AsObject = {
    cid: string,
  }
}

export class InfoRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): InfoRequest.AsObject;
  static toObject(includeInstance: boolean, msg: InfoRequest): InfoRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: InfoRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): InfoRequest;
  static deserializeBinaryFromReader(message: InfoRequest, reader: jspb.BinaryReader): InfoRequest;
}

export namespace InfoRequest {
  export type AsObject = {
  }
}

export class InfoReply extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  hasWallet(): boolean;
  clearWallet(): void;
  getWallet(): WalletInfo | undefined;
  setWallet(value?: WalletInfo): void;

  clearPinsList(): void;
  getPinsList(): Array<string>;
  setPinsList(value: Array<string>): void;
  addPins(value: string, index?: number): string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): InfoReply.AsObject;
  static toObject(includeInstance: boolean, msg: InfoReply): InfoReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: InfoReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): InfoReply;
  static deserializeBinaryFromReader(message: InfoReply, reader: jspb.BinaryReader): InfoReply;
}

export namespace InfoReply {
  export type AsObject = {
    id: string,
    wallet?: WalletInfo.AsObject,
    pinsList: Array<string>,
  }
}

export class WalletInfo extends jspb.Message {
  getAddress(): string;
  setAddress(value: string): void;

  getBalance(): number;
  setBalance(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WalletInfo.AsObject;
  static toObject(includeInstance: boolean, msg: WalletInfo): WalletInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WalletInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WalletInfo;
  static deserializeBinaryFromReader(message: WalletInfo, reader: jspb.BinaryReader): WalletInfo;
}

export namespace WalletInfo {
  export type AsObject = {
    address: string,
    balance: number,
  }
}

export class ShowReply extends jspb.Message {
  getCid(): string;
  setCid(value: string): void;

  getCreated(): number;
  setCreated(value: number): void;

  hasHot(): boolean;
  clearHot(): void;
  getHot(): ShowReply.HotInfo | undefined;
  setHot(value?: ShowReply.HotInfo): void;

  hasCold(): boolean;
  clearCold(): void;
  getCold(): ShowReply.ColdInfo | undefined;
  setCold(value?: ShowReply.ColdInfo): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ShowReply.AsObject;
  static toObject(includeInstance: boolean, msg: ShowReply): ShowReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ShowReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ShowReply;
  static deserializeBinaryFromReader(message: ShowReply, reader: jspb.BinaryReader): ShowReply;
}

export namespace ShowReply {
  export type AsObject = {
    cid: string,
    created: number,
    hot?: ShowReply.HotInfo.AsObject,
    cold?: ShowReply.ColdInfo.AsObject,
  }

  export class HotInfo extends jspb.Message {
    getSize(): number;
    setSize(value: number): void;

    hasIpfs(): boolean;
    clearIpfs(): void;
    getIpfs(): ShowReply.IpfsHotInfo | undefined;
    setIpfs(value?: ShowReply.IpfsHotInfo): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HotInfo.AsObject;
    static toObject(includeInstance: boolean, msg: HotInfo): HotInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HotInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HotInfo;
    static deserializeBinaryFromReader(message: HotInfo, reader: jspb.BinaryReader): HotInfo;
  }

  export namespace HotInfo {
    export type AsObject = {
      size: number,
      ipfs?: ShowReply.IpfsHotInfo.AsObject,
    }
  }

  export class IpfsHotInfo extends jspb.Message {
    getCreated(): number;
    setCreated(value: number): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): IpfsHotInfo.AsObject;
    static toObject(includeInstance: boolean, msg: IpfsHotInfo): IpfsHotInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: IpfsHotInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): IpfsHotInfo;
    static deserializeBinaryFromReader(message: IpfsHotInfo, reader: jspb.BinaryReader): IpfsHotInfo;
  }

  export namespace IpfsHotInfo {
    export type AsObject = {
      created: number,
    }
  }

  export class ColdInfo extends jspb.Message {
    hasFilecoin(): boolean;
    clearFilecoin(): void;
    getFilecoin(): ShowReply.FilInfo | undefined;
    setFilecoin(value?: ShowReply.FilInfo): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ColdInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ColdInfo): ColdInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ColdInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ColdInfo;
    static deserializeBinaryFromReader(message: ColdInfo, reader: jspb.BinaryReader): ColdInfo;
  }

  export namespace ColdInfo {
    export type AsObject = {
      filecoin?: ShowReply.FilInfo.AsObject,
    }
  }

  export class FilInfo extends jspb.Message {
    getPayloadcid(): string;
    setPayloadcid(value: string): void;

    getDuration(): number;
    setDuration(value: number): void;

    clearProposalsList(): void;
    getProposalsList(): Array<ShowReply.FilStorage>;
    setProposalsList(value: Array<ShowReply.FilStorage>): void;
    addProposals(value?: ShowReply.FilStorage, index?: number): ShowReply.FilStorage;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FilInfo.AsObject;
    static toObject(includeInstance: boolean, msg: FilInfo): FilInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FilInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FilInfo;
    static deserializeBinaryFromReader(message: FilInfo, reader: jspb.BinaryReader): FilInfo;
  }

  export namespace FilInfo {
    export type AsObject = {
      payloadcid: string,
      duration: number,
      proposalsList: Array<ShowReply.FilStorage.AsObject>,
    }
  }

  export class FilStorage extends jspb.Message {
    getProposalcid(): string;
    setProposalcid(value: string): void;

    getFailed(): boolean;
    setFailed(value: boolean): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FilStorage.AsObject;
    static toObject(includeInstance: boolean, msg: FilStorage): FilStorage.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FilStorage, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FilStorage;
    static deserializeBinaryFromReader(message: FilStorage, reader: jspb.BinaryReader): FilStorage;
  }

  export namespace FilStorage {
    export type AsObject = {
      proposalcid: string,
      failed: boolean,
    }
  }
}

