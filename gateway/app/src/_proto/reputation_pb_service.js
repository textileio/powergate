/* eslint-disable */
// package: filecoin.reputation.pb
// file: reputation.proto

var reputation_pb = require("./reputation_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var API = (function () {
  function API() {}
  API.serviceName = "filecoin.reputation.pb.API";
  return API;
}());

API.AddSource = {
  methodName: "AddSource",
  service: API,
  requestStream: false,
  responseStream: false,
  requestType: reputation_pb.AddSourceRequest,
  responseType: reputation_pb.AddSourceReply
};

API.GetTopMiners = {
  methodName: "GetTopMiners",
  service: API,
  requestStream: false,
  responseStream: false,
  requestType: reputation_pb.GetTopMinersRequest,
  responseType: reputation_pb.GetTopMinersReply
};

exports.API = API;

function APIClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

APIClient.prototype.addSource = function addSource(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(API.AddSource, {
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

APIClient.prototype.getTopMiners = function getTopMiners(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(API.GetTopMiners, {
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

