/* eslint-disable */
// package: filecoin.slashing.pb
// file: slashing.proto

var slashing_pb = require("./slashing_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var API = (function () {
  function API() {}
  API.serviceName = "filecoin.slashing.pb.API";
  return API;
}());

API.Get = {
  methodName: "Get",
  service: API,
  requestStream: false,
  responseStream: false,
  requestType: slashing_pb.GetRequest,
  responseType: slashing_pb.GetReply
};

exports.API = API;

function APIClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

APIClient.prototype.get = function get(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(API.Get, {
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

