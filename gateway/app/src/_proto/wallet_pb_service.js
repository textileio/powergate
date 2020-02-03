/* eslint-disable */
// package: filecoin.wallet.pb
// file: wallet.proto

var wallet_pb = require("./wallet_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var API = (function () {
  function API() {}
  API.serviceName = "filecoin.wallet.pb.API";
  return API;
}());

API.NewWallet = {
  methodName: "NewWallet",
  service: API,
  requestStream: false,
  responseStream: false,
  requestType: wallet_pb.NewWalletRequest,
  responseType: wallet_pb.NewWalletReply
};

API.WalletBalance = {
  methodName: "WalletBalance",
  service: API,
  requestStream: false,
  responseStream: false,
  requestType: wallet_pb.WalletBalanceRequest,
  responseType: wallet_pb.WalletBalanceReply
};

exports.API = API;

function APIClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

APIClient.prototype.newWallet = function newWallet(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(API.NewWallet, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

APIClient.prototype.walletBalance = function walletBalance(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(API.WalletBalance, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.APIClient = APIClient;

