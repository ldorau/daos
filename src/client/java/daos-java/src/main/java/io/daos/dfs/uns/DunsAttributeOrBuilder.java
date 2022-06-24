// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: DunsAttribute.proto

package io.daos.dfs.uns;

public interface DunsAttributeOrBuilder extends
    // @@protoc_insertion_point(interface_extends:uns.DunsAttribute)
    com.google.protobuf.MessageOrBuilder {

  /**
   * <code>string poolId = 1;</code>
   * @return The poolId.
   */
  java.lang.String getPoolId();
  /**
   * <code>string poolId = 1;</code>
   * @return The bytes for poolId.
   */
  com.google.protobuf.ByteString
      getPoolIdBytes();
  /**
   * <code>string contId = 2;</code>
   * @return The contId.
   */
  java.lang.String getContId();
  /**
   * <code>string contId = 2;</code>
   * @return The bytes for contId.
   */
  com.google.protobuf.ByteString
      getContIdBytes();

  /**
   * <code>.uns.Layout layout_type = 3;</code>
   * @return The enum numeric value on the wire for layoutType.
   */
  int getLayoutTypeValue();
  /**
   * <code>.uns.Layout layout_type = 3;</code>
   * @return The layoutType.
   */
  io.daos.dfs.uns.Layout getLayoutType();

  /**
   * <code>string object_type = 4;</code>
   * @return The objectType.
   */
  java.lang.String getObjectType();
  /**
   * <code>string object_type = 4;</code>
   * @return The bytes for objectType.
   */
  com.google.protobuf.ByteString
      getObjectTypeBytes();

  /**
   * <code>uint64 chunk_size = 5;</code>
   * @return The chunkSize.
   */
  long getChunkSize();

  /**
   * <code>string rel_path = 6;</code>
   * @return The relPath.
   */
  java.lang.String getRelPath();
  /**
   * <code>string rel_path = 6;</code>
   * @return The bytes for relPath.
   */
  com.google.protobuf.ByteString
      getRelPathBytes();

  /**
   * <code>bool on_lustre = 7;</code>
   * @return The onLustre.
   */
  boolean getOnLustre();

  /**
   * <code>bool no_prefix = 9;</code>
   * @return The noPrefix.
   */
  boolean getNoPrefix();

  /**
   * <code>uint32 flags = 10;</code>
   * @return The flags.
   */
  int getFlags();
}