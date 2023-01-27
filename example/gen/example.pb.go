// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.5
// source: github.com/lastbackend/toolkit/example/apis/example.proto

package servicepb

import (
	ptypes "github.com/lastbackend/toolkit/example/gen/ptypes"
	_ "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options"
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

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_lastbackend_toolkit_example_apis_example_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_lastbackend_toolkit_example_apis_example_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescGZIP(), []int{0}
}

var File_github_com_lastbackend_toolkit_example_apis_example_proto protoreflect.FileDescriptor

var file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDesc = []byte{
	0x0a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73,
	0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x65, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x6c, 0x61, 0x73,
	0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x1a, 0x53, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73,
	0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x74, 0x6f, 0x6f, 0x6c,
	0x6b, 0x69, 0x74, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74, 0x2f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f,
	0x6f, 0x6c, 0x6b, 0x69, 0x74, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x73, 0x2f, 0x70, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x32, 0x6a, 0x0a, 0x07, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x5f, 0x0a, 0x0a,
	0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x12, 0x26, 0x2e, 0x6c, 0x61, 0x73,
	0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x2e, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x27, 0x2e, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64,
	0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f,
	0x72, 0x6c, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0xbb, 0x01,
	0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73,
	0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74,
	0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x3b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x62, 0x9a, 0xb5, 0x18, 0x16, 0x0a, 0x02,
	0xb8, 0x17, 0x12, 0x10, 0x2f, 0x75, 0x73, 0x72, 0x2f, 0x62, 0x69, 0x6e, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0xa2, 0xb5, 0x18, 0x3e, 0x0a, 0x3c, 0x0a, 0x3a, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x2f, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70,
	0x6c, 0x65, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x5f, 0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x8a, 0xb5, 0x18, 0x21, 0x0a, 0x1f, 0x0a, 0x05, 0x70, 0x67,
	0x73, 0x71, 0x6c, 0x12, 0x16, 0x0a, 0x0d, 0x70, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x5f,
	0x67, 0x6f, 0x72, 0x6d, 0x12, 0x05, 0x70, 0x67, 0x73, 0x71, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescOnce sync.Once
	file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescData = file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDesc
)

func file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescGZIP() []byte {
	file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescOnce.Do(func() {
		file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescData)
	})
	return file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDescData
}

var file_github_com_lastbackend_toolkit_example_apis_example_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_lastbackend_toolkit_example_apis_example_proto_goTypes = []interface{}{
	(*Empty)(nil),                     // 0: lastbackend.example.Empty
	(*ptypes.HelloWorldRequest)(nil),  // 1: lastbackend.example.HelloWorldRequest
	(*ptypes.HelloWorldResponse)(nil), // 2: lastbackend.example.HelloWorldResponse
}
var file_github_com_lastbackend_toolkit_example_apis_example_proto_depIdxs = []int32{
	1, // 0: lastbackend.example.Example.HelloWorld:input_type -> lastbackend.example.HelloWorldRequest
	2, // 1: lastbackend.example.Example.HelloWorld:output_type -> lastbackend.example.HelloWorldResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_lastbackend_toolkit_example_apis_example_proto_init() }
func file_github_com_lastbackend_toolkit_example_apis_example_proto_init() {
	if File_github_com_lastbackend_toolkit_example_apis_example_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_lastbackend_toolkit_example_apis_example_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
			RawDescriptor: file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_lastbackend_toolkit_example_apis_example_proto_goTypes,
		DependencyIndexes: file_github_com_lastbackend_toolkit_example_apis_example_proto_depIdxs,
		MessageInfos:      file_github_com_lastbackend_toolkit_example_apis_example_proto_msgTypes,
	}.Build()
	File_github_com_lastbackend_toolkit_example_apis_example_proto = out.File
	file_github_com_lastbackend_toolkit_example_apis_example_proto_rawDesc = nil
	file_github_com_lastbackend_toolkit_example_apis_example_proto_goTypes = nil
	file_github_com_lastbackend_toolkit_example_apis_example_proto_depIdxs = nil
}
