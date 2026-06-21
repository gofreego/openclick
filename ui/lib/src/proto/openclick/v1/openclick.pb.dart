//
//  Generated code. Do not modify.
//  source: proto/openclick/v1/openclick.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../common/ping.pb.dart' as $0;

class BaseServiceApi {
  $pb.RpcClient _client;
  BaseServiceApi(this._client);

  $async.Future<$0.PingResponse> ping($pb.ClientContext? ctx, $0.PingRequest request) =>
    _client.invoke<$0.PingResponse>(ctx, 'BaseService', 'Ping', request, $0.PingResponse())
  ;
}

