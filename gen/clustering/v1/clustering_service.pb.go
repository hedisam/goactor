// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: clustering/v1/clustering_service.proto

package clusteringv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SpawnRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ActorSignature string `protobuf:"bytes,1,opt,name=actor_signature,json=actorSignature,proto3" json:"actor_signature,omitempty"`
	ActorData      []byte `protobuf:"bytes,2,opt,name=actor_data,json=actorData,proto3" json:"actor_data,omitempty"`
}

func (x *SpawnRequest) Reset() {
	*x = SpawnRequest{}
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SpawnRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpawnRequest) ProtoMessage() {}

func (x *SpawnRequest) ProtoReflect() protoreflect.Message {
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpawnRequest.ProtoReflect.Descriptor instead.
func (*SpawnRequest) Descriptor() ([]byte, []int) {
	return file_clustering_v1_clustering_service_proto_rawDescGZIP(), []int{0}
}

func (x *SpawnRequest) GetActorSignature() string {
	if x != nil {
		return x.ActorSignature
	}
	return ""
}

func (x *SpawnRequest) GetActorData() []byte {
	if x != nil {
		return x.ActorData
	}
	return nil
}

type SpawnResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ref string `protobuf:"bytes,1,opt,name=ref,proto3" json:"ref,omitempty"`
}

func (x *SpawnResponse) Reset() {
	*x = SpawnResponse{}
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SpawnResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SpawnResponse) ProtoMessage() {}

func (x *SpawnResponse) ProtoReflect() protoreflect.Message {
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SpawnResponse.ProtoReflect.Descriptor instead.
func (*SpawnResponse) Descriptor() ([]byte, []int) {
	return file_clustering_v1_clustering_service_proto_rawDescGZIP(), []int{1}
}

func (x *SpawnResponse) GetRef() string {
	if x != nil {
		return x.Ref
	}
	return ""
}

type SendRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ref     string   `protobuf:"bytes,2,opt,name=ref,proto3" json:"ref,omitempty"`
	Message *Message `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *SendRequest) Reset() {
	*x = SendRequest{}
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SendRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendRequest) ProtoMessage() {}

func (x *SendRequest) ProtoReflect() protoreflect.Message {
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendRequest.ProtoReflect.Descriptor instead.
func (*SendRequest) Descriptor() ([]byte, []int) {
	return file_clustering_v1_clustering_service_proto_rawDescGZIP(), []int{2}
}

func (x *SendRequest) GetRef() string {
	if x != nil {
		return x.Ref
	}
	return ""
}

func (x *SendRequest) GetMessage() *Message {
	if x != nil {
		return x.Message
	}
	return nil
}

type SendResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SendResponse) Reset() {
	*x = SendResponse{}
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SendResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendResponse) ProtoMessage() {}

func (x *SendResponse) ProtoReflect() protoreflect.Message {
	mi := &file_clustering_v1_clustering_service_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendResponse.ProtoReflect.Descriptor instead.
func (*SendResponse) Descriptor() ([]byte, []int) {
	return file_clustering_v1_clustering_service_proto_rawDescGZIP(), []int{3}
}

var File_clustering_v1_clustering_service_proto protoreflect.FileDescriptor

var file_clustering_v1_clustering_service_proto_rawDesc = []byte{
	0x0a, 0x26, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65,
	0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x1a, 0x1e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72,
	0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x56, 0x0a, 0x0c, 0x53, 0x70, 0x61, 0x77, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0e, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x12, 0x1d, 0x0a, 0x0a, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x44, 0x61, 0x74, 0x61, 0x22,
	0x21, 0x0a, 0x0d, 0x53, 0x70, 0x61, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x72,
	0x65, 0x66, 0x22, 0x51, 0x0a, 0x0b, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x72, 0x65, 0x66, 0x12, 0x30, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e,
	0x67, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x0e, 0x0a, 0x0c, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x92, 0x01, 0x0a, 0x0b, 0x4e, 0x6f, 0x64, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x42, 0x0a, 0x05, 0x53, 0x70, 0x61, 0x77, 0x6e, 0x12, 0x1b,
	0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x70, 0x61, 0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x63, 0x6c,
	0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x70, 0x61, 0x77,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3f, 0x0a, 0x04, 0x53, 0x65, 0x6e,
	0x64, 0x12, 0x1a, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65,
	0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xc2, 0x01, 0x0a, 0x11, 0x63,
	0x6f, 0x6d, 0x2e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31,
	0x42, 0x16, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x65, 0x64, 0x69, 0x73, 0x61, 0x6d, 0x2f, 0x67,
	0x6f, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x3b,
	0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43,
	0x58, 0x58, 0xaa, 0x02, 0x0d, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2e,
	0x56, 0x31, 0xca, 0x02, 0x0d, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x5c,
	0x56, 0x31, 0xe2, 0x02, 0x19, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x5c,
	0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x0e, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x3a, 0x3a, 0x56, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_clustering_v1_clustering_service_proto_rawDescOnce sync.Once
	file_clustering_v1_clustering_service_proto_rawDescData = file_clustering_v1_clustering_service_proto_rawDesc
)

func file_clustering_v1_clustering_service_proto_rawDescGZIP() []byte {
	file_clustering_v1_clustering_service_proto_rawDescOnce.Do(func() {
		file_clustering_v1_clustering_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_clustering_v1_clustering_service_proto_rawDescData)
	})
	return file_clustering_v1_clustering_service_proto_rawDescData
}

var file_clustering_v1_clustering_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_clustering_v1_clustering_service_proto_goTypes = []any{
	(*SpawnRequest)(nil),  // 0: clustering.v1.SpawnRequest
	(*SpawnResponse)(nil), // 1: clustering.v1.SpawnResponse
	(*SendRequest)(nil),   // 2: clustering.v1.SendRequest
	(*SendResponse)(nil),  // 3: clustering.v1.SendResponse
	(*Message)(nil),       // 4: clustering.v1.Message
}
var file_clustering_v1_clustering_service_proto_depIdxs = []int32{
	4, // 0: clustering.v1.SendRequest.message:type_name -> clustering.v1.Message
	0, // 1: clustering.v1.NodeService.Spawn:input_type -> clustering.v1.SpawnRequest
	2, // 2: clustering.v1.NodeService.Send:input_type -> clustering.v1.SendRequest
	1, // 3: clustering.v1.NodeService.Spawn:output_type -> clustering.v1.SpawnResponse
	3, // 4: clustering.v1.NodeService.Send:output_type -> clustering.v1.SendResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_clustering_v1_clustering_service_proto_init() }
func file_clustering_v1_clustering_service_proto_init() {
	if File_clustering_v1_clustering_service_proto != nil {
		return
	}
	file_clustering_v1_clustering_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_clustering_v1_clustering_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_clustering_v1_clustering_service_proto_goTypes,
		DependencyIndexes: file_clustering_v1_clustering_service_proto_depIdxs,
		MessageInfos:      file_clustering_v1_clustering_service_proto_msgTypes,
	}.Build()
	File_clustering_v1_clustering_service_proto = out.File
	file_clustering_v1_clustering_service_proto_rawDesc = nil
	file_clustering_v1_clustering_service_proto_goTypes = nil
	file_clustering_v1_clustering_service_proto_depIdxs = nil
}