//
//  Generated code. Do not modify.
//  source: proto/common/headers.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// Common header fields for authentication and client identification
/// These headers are used across all API requests for authentication and tracking
class RequestHeaders extends $pb.GeneratedMessage {
  factory RequestHeaders({
    $core.String? authorization,
    $core.String? xClientId,
    $core.String? xClientVersion,
    $core.String? xUserId,
    $core.String? xUserPerms,
  }) {
    final $result = create();
    if (authorization != null) {
      $result.authorization = authorization;
    }
    if (xClientId != null) {
      $result.xClientId = xClientId;
    }
    if (xClientVersion != null) {
      $result.xClientVersion = xClientVersion;
    }
    if (xUserId != null) {
      $result.xUserId = xUserId;
    }
    if (xUserPerms != null) {
      $result.xUserPerms = xUserPerms;
    }
    return $result;
  }
  RequestHeaders._() : super();
  factory RequestHeaders.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RequestHeaders.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'RequestHeaders', package: const $pb.PackageName(_omitMessageNames ? '' : 'v1'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'authorization')
    ..aOS(2, _omitFieldNames ? '' : 'xClientId')
    ..aOS(3, _omitFieldNames ? '' : 'xClientVersion')
    ..aOS(4, _omitFieldNames ? '' : 'xUserId')
    ..aOS(5, _omitFieldNames ? '' : 'xUserPerms')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RequestHeaders clone() => RequestHeaders()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RequestHeaders copyWith(void Function(RequestHeaders) updates) => super.copyWith((message) => updates(message as RequestHeaders)) as RequestHeaders;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RequestHeaders create() => RequestHeaders._();
  RequestHeaders createEmptyInstance() => create();
  static $pb.PbList<RequestHeaders> createRepeated() => $pb.PbList<RequestHeaders>();
  @$core.pragma('dart2js:noInline')
  static RequestHeaders getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RequestHeaders>(create);
  static RequestHeaders? _defaultInstance;

  /// JWT authentication token (Bearer token)
  /// Should be provided in Authorization header as "Bearer <token>"
  @$pb.TagNumber(1)
  $core.String get authorization => $_getSZ(0);
  @$pb.TagNumber(1)
  set authorization($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAuthorization() => $_has(0);
  @$pb.TagNumber(1)
  void clearAuthorization() => clearField(1);

  /// Client identifier for tracking and analytics
  /// Used to identify the calling application or service
  @$pb.TagNumber(2)
  $core.String get xClientId => $_getSZ(1);
  @$pb.TagNumber(2)
  set xClientId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasXClientId() => $_has(1);
  @$pb.TagNumber(2)
  void clearXClientId() => clearField(2);

  /// Client version for compatibility and feature tracking
  /// Used to track client versions and handle backward compatibility
  @$pb.TagNumber(3)
  $core.String get xClientVersion => $_getSZ(2);
  @$pb.TagNumber(3)
  set xClientVersion($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasXClientVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearXClientVersion() => clearField(3);

  /// Authenticated user's ID from external auth service
  @$pb.TagNumber(4)
  $core.String get xUserId => $_getSZ(3);
  @$pb.TagNumber(4)
  set xUserId($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasXUserId() => $_has(3);
  @$pb.TagNumber(4)
  void clearXUserId() => clearField(4);

  /// Comma-separated permission scopes granted to this user
  @$pb.TagNumber(5)
  $core.String get xUserPerms => $_getSZ(4);
  @$pb.TagNumber(5)
  set xUserPerms($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasXUserPerms() => $_has(4);
  @$pb.TagNumber(5)
  void clearXUserPerms() => clearField(5);
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
