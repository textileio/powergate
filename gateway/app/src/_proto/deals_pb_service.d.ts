// package: filecoin.deals.pb
// file: deals.proto

import * as deals_pb from "./deals_pb";
import {grpc} from "@improbable-eng/grpc-web";

type APIStore = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: true;
  readonly responseStream: false;
  readonly requestType: typeof deals_pb.StoreRequest;
  readonly responseType: typeof deals_pb.StoreReply;
};

type APIWatch = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof deals_pb.WatchRequest;
  readonly responseType: typeof deals_pb.WatchReply;
};

export class API {
  static readonly serviceName: string;
  static readonly Store: APIStore;
  static readonly Watch: APIWatch;
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
  store(metadata?: grpc.Metadata): RequestStream<deals_pb.StoreRequest>;
  watch(requestMessage: deals_pb.WatchRequest, metadata?: grpc.Metadata): ResponseStream<deals_pb.WatchReply>;
}

