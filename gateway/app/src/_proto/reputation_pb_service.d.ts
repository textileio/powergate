// package: filecoin.reputation.pb
// file: reputation.proto

import * as reputation_pb from "./reputation_pb";
import {grpc} from "@improbable-eng/grpc-web";

type APIAddSource = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof reputation_pb.AddSourceRequest;
  readonly responseType: typeof reputation_pb.AddSourceReply;
};

type APIGetTopMiners = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof reputation_pb.GetTopMinersRequest;
  readonly responseType: typeof reputation_pb.GetTopMinersReply;
};

export class API {
  static readonly serviceName: string;
  static readonly AddSource: APIAddSource;
  static readonly GetTopMiners: APIGetTopMiners;
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
  addSource(
    requestMessage: reputation_pb.AddSourceRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: reputation_pb.AddSourceReply|null) => void
  ): UnaryResponse;
  addSource(
    requestMessage: reputation_pb.AddSourceRequest,
    callback: (error: ServiceError|null, responseMessage: reputation_pb.AddSourceReply|null) => void
  ): UnaryResponse;
  getTopMiners(
    requestMessage: reputation_pb.GetTopMinersRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: reputation_pb.GetTopMinersReply|null) => void
  ): UnaryResponse;
  getTopMiners(
    requestMessage: reputation_pb.GetTopMinersRequest,
    callback: (error: ServiceError|null, responseMessage: reputation_pb.GetTopMinersReply|null) => void
  ): UnaryResponse;
}

