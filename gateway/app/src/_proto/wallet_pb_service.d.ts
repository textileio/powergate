// package: filecoin.wallet.pb
// file: wallet.proto

import * as wallet_pb from "./wallet_pb";
import {grpc} from "@improbable-eng/grpc-web";

type APINewWallet = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof wallet_pb.NewWalletRequest;
  readonly responseType: typeof wallet_pb.NewWalletReply;
};

type APIWalletBalance = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof wallet_pb.WalletBalanceRequest;
  readonly responseType: typeof wallet_pb.WalletBalanceReply;
};

export class API {
  static readonly serviceName: string;
  static readonly NewWallet: APINewWallet;
  static readonly WalletBalance: APIWalletBalance;
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
  newWallet(
    requestMessage: wallet_pb.NewWalletRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: wallet_pb.NewWalletReply|null) => void
  ): UnaryResponse;
  newWallet(
    requestMessage: wallet_pb.NewWalletRequest,
    callback: (error: ServiceError|null, responseMessage: wallet_pb.NewWalletReply|null) => void
  ): UnaryResponse;
  walletBalance(
    requestMessage: wallet_pb.WalletBalanceRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: wallet_pb.WalletBalanceReply|null) => void
  ): UnaryResponse;
  walletBalance(
    requestMessage: wallet_pb.WalletBalanceRequest,
    callback: (error: ServiceError|null, responseMessage: wallet_pb.WalletBalanceReply|null) => void
  ): UnaryResponse;
}

