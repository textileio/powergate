// package: filecoin.ask.pb
// file: ask.proto

import * as ask_pb from "./ask_pb";
import {grpc} from "@improbable-eng/grpc-web";

type APIGet = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof ask_pb.GetRequest;
  readonly responseType: typeof ask_pb.GetReply;
};

type APIQuery = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof ask_pb.QueryRequest;
  readonly responseType: typeof ask_pb.QueryReply;
};

export class API {
  static readonly serviceName: string;
  static readonly Get: APIGet;
  static readonly Query: APIQuery;
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
  get(
    requestMessage: ask_pb.GetRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: ask_pb.GetReply|null) => void
  ): UnaryResponse;
  get(
    requestMessage: ask_pb.GetRequest,
    callback: (error: ServiceError|null, responseMessage: ask_pb.GetReply|null) => void
  ): UnaryResponse;
  query(
    requestMessage: ask_pb.QueryRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: ask_pb.QueryReply|null) => void
  ): UnaryResponse;
  query(
    requestMessage: ask_pb.QueryRequest,
    callback: (error: ServiceError|null, responseMessage: ask_pb.QueryReply|null) => void
  ): UnaryResponse;
}

