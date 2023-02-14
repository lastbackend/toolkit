// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: github.com/lastbackend/toolkit/examples/wss/apis/server.proto

package serverpb

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	gen "github.com/lastbackend/toolkit/examples/helloworld/gen"
	_ "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type SubscribeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SubscribeRequest) Reset() {
	*x = SubscribeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeRequest) ProtoMessage() {}

func (x *SubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeRequest.ProtoReflect.Descriptor instead.
func (*SubscribeRequest) Descriptor() ([]byte, []int) {
	return file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescGZIP(), []int{0}
}

type SubscribeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SubscribeResponse) Reset() {
	*x = SubscribeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeResponse) ProtoMessage() {}

func (x *SubscribeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeResponse.ProtoReflect.Descriptor instead.
func (*SubscribeResponse) Descriptor() ([]byte, []int) {
	return file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescGZIP(), []int{1}
}

var File_github_com_lastbackend_toolkit_examples_wss_apis_server_proto protoreflect.FileDescriptor

var file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDesc = []byte{
	0x0a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73,
	0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x77, 0x73, 0x73, 0x2f, 0x61, 0x70,
	0x69, 0x73, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67,
	0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x53, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74,
	0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65,
	0x6e, 0x2d, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69,
	0x74, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x48, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74, 0x2f, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x2f,
	0x61, 0x70, 0x69, 0x73, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x12, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69,
	0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x13, 0x0a, 0x11, 0x53, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xdd,
	0x01, 0x0a, 0x06, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x12, 0x53, 0x0a, 0x09, 0x53, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x12, 0x19, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1a, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x53, 0x75, 0x62, 0x73,
	0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x0f, 0x8a,
	0xa6, 0x1d, 0x0b, 0x12, 0x09, 0x0a, 0x07, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x7e,
	0x0a, 0x0a, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x12, 0x18, 0x2e, 0x68,
	0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x2e, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f,
	0x72, 0x6c, 0x64, 0x2e, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x3e,
	0x8a, 0xa6, 0x1d, 0x2c, 0x0a, 0x2a, 0x0a, 0x0a, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72,
	0x6c, 0x64, 0x12, 0x1c, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x2e,
	0x47, 0x72, 0x65, 0x65, 0x74, 0x65, 0x72, 0x2f, 0x53, 0x61, 0x79, 0x48, 0x65, 0x6c, 0x6c, 0x6f,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x08, 0x22, 0x06, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x42, 0xf7,
	0x02, 0x5a, 0x3f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61,
	0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69,
	0x74, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2f, 0x77, 0x73, 0x73, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x3b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x70, 0x62, 0x92, 0x41, 0x98, 0x02, 0x12, 0x5d, 0x0a, 0x16, 0x57, 0x65, 0x62, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x20, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x20, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x22, 0x3e, 0x0a, 0x0c, 0x4c, 0x61, 0x73, 0x74, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64,
	0x12, 0x17, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61,
	0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x1a, 0x15, 0x74, 0x65, 0x61, 0x6d, 0x73,
	0x40, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2e, 0x63, 0x6f, 0x6d,
	0x32, 0x03, 0x31, 0x2e, 0x30, 0x1a, 0x0f, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x2e, 0x63, 0x6f, 0x6d, 0x2a, 0x02, 0x01, 0x02, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70,
	0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x52, 0x7e,
	0x0a, 0x03, 0x35, 0x30, 0x30, 0x12, 0x77, 0x0a, 0x15, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x20, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x5e,
	0x0a, 0x5c, 0x40, 0x01, 0x4a, 0x54, 0x7b, 0x22, 0x63, 0x6f, 0x64, 0x65, 0x22, 0x3a, 0x20, 0x35,
	0x30, 0x30, 0x2c, 0x20, 0x22, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x3a, 0x20, 0x22, 0x49,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x20, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x22, 0x2c, 0x20, 0x22, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22,
	0x3a, 0x20, 0x22, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x20, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x20, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x7d, 0x9a, 0x02, 0x01, 0x06, 0x9a, 0xb5,
	0x18, 0x16, 0x0a, 0x02, 0x90, 0x3f, 0x12, 0x10, 0x2f, 0x75, 0x73, 0x72, 0x2f, 0x62, 0x69, 0x6e,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescOnce sync.Once
	file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescData = file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDesc
)

func file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescGZIP() []byte {
	file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescOnce.Do(func() {
		file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescData)
	})
	return file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDescData
}

var file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_goTypes = []interface{}{
	(*SubscribeRequest)(nil),  // 0: gateway.SubscribeRequest
	(*SubscribeResponse)(nil), // 1: gateway.SubscribeResponse
	(*gen.HelloRequest)(nil),  // 2: helloworld.HelloRequest
	(*gen.HelloReply)(nil),    // 3: helloworld.HelloReply
}
var file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_depIdxs = []int32{
	0, // 0: gateway.Router.Subscribe:input_type -> gateway.SubscribeRequest
	2, // 1: gateway.Router.HelloWorld:input_type -> helloworld.HelloRequest
	1, // 2: gateway.Router.Subscribe:output_type -> gateway.SubscribeResponse
	3, // 3: gateway.Router.HelloWorld:output_type -> helloworld.HelloReply
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_init() }
func file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_init() {
	if File_github_com_lastbackend_toolkit_examples_wss_apis_server_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_goTypes,
		DependencyIndexes: file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_depIdxs,
		MessageInfos:      file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_msgTypes,
	}.Build()
	File_github_com_lastbackend_toolkit_examples_wss_apis_server_proto = out.File
	file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_rawDesc = nil
	file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_goTypes = nil
	file_github_com_lastbackend_toolkit_examples_wss_apis_server_proto_depIdxs = nil
}
