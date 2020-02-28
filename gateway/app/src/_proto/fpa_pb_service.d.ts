// package: filecoin.fpa.pb
// file: fpa.proto

import * as fpa_pb from "./fpa_pb";
import {grpc} from "@improbable-eng/grpc-web";

type APIAddCid = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof fpa_pb.AddCidRequest;
  readonly responseType: typeof fpa_pb.AddCidReply;
};

type APIAddFile = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: true;
  readonly responseStream: false;
  readonly requestType: typeof fpa_pb.AddFileRequest;
  readonly responseType: typeof fpa_pb.AddFileReply;
};

type APIGet = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof fpa_pb.GetRequest;
  readonly responseType: typeof fpa_pb.GetReply;
};

type APICreate = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof fpa_pb.CreateRequest;
  readonly responseType: typeof fpa_pb.CreateReply;
};

type APIInfo = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof fpa_pb.InfoRequest;
  readonly responseType: typeof fpa_pb.InfoReply;
};

type APIShow = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof fpa_pb.ShowRequest;
  readonly responseType: typeof fpa_pb.ShowReply;
};

export class API {
  static readonly serviceName: string;
  static readonly AddCid: APIAddCid;
  static readonly AddFile: APIAddFile;
  static readonly Get: APIGet;
  static readonly Create: APICreate;
  static readonly Info: APIInfo;
  static readonly Show: APIShow;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class APIClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  addCid(
    requestMessage: fpa_pb.AddCidRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.AddCidReply|null) => void
  ): UnaryResponse;
  addCid(
    requestMessage: fpa_pb.AddCidRequest,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.AddCidReply|null) => void
  ): UnaryResponse;
  addFile(metadata?: grpc.Metadata): RequestStream<fpa_pb.AddFileRequest>;
  get(requestMessage: fpa_pb.GetRequest, metadata?: grpc.Metadata): ResponseStream<fpa_pb.GetReply>;
  create(
    requestMessage: fpa_pb.CreateRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.CreateReply|null) => void
  ): UnaryResponse;
  create(
    requestMessage: fpa_pb.CreateRequest,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.CreateReply|null) => void
  ): UnaryResponse;
  info(
    requestMessage: fpa_pb.InfoRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.InfoReply|null) => void
  ): UnaryResponse;
  info(
    requestMessage: fpa_pb.InfoRequest,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.InfoReply|null) => void
  ): UnaryResponse;
  show(
    requestMessage: fpa_pb.ShowRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.ShowReply|null) => void
  ): UnaryResponse;
  show(
    requestMessage: fpa_pb.ShowRequest,
    callback: (error: ServiceError|null, responseMessage: fpa_pb.ShowReply|null) => void
  ): UnaryResponse;
}

