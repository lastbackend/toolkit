// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: protoc-gen-engine/engine/options/annotations.proto

package annotations

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Cache struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Plugin string `protobuf:"bytes,1,opt,name=plugin,proto3" json:"plugin,omitempty"`
	Prefix string `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *Cache) Reset() {
	*x = Cache{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cache) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cache) ProtoMessage() {}

func (x *Cache) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cache.ProtoReflect.Descriptor instead.
func (*Cache) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{0}
}

func (x *Cache) GetPlugin() string {
	if x != nil {
		return x.Plugin
	}
	return ""
}

func (x *Cache) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

type Storage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Plugin string `protobuf:"bytes,1,opt,name=plugin,proto3" json:"plugin,omitempty"`
	Prefix string `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *Storage) Reset() {
	*x = Storage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Storage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Storage) ProtoMessage() {}

func (x *Storage) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Storage.ProtoReflect.Descriptor instead.
func (*Storage) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{1}
}

func (x *Storage) GetPlugin() string {
	if x != nil {
		return x.Plugin
	}
	return ""
}

func (x *Storage) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

type Broker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Plugin string `protobuf:"bytes,1,opt,name=plugin,proto3" json:"plugin,omitempty"`
	Prefix string `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *Broker) Reset() {
	*x = Broker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Broker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Broker) ProtoMessage() {}

func (x *Broker) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Broker.ProtoReflect.Descriptor instead.
func (*Broker) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{2}
}

func (x *Broker) GetPlugin() string {
	if x != nil {
		return x.Plugin
	}
	return ""
}

func (x *Broker) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

type Plugins struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Storage map[string]*Storage `protobuf:"bytes,1,rep,name=storage,proto3" json:"storage,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Cache   map[string]*Cache   `protobuf:"bytes,2,rep,name=cache,proto3" json:"cache,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Broker  map[string]*Broker  `protobuf:"bytes,3,rep,name=broker,proto3" json:"broker,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Plugins) Reset() {
	*x = Plugins{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plugins) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plugins) ProtoMessage() {}

func (x *Plugins) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Plugins.ProtoReflect.Descriptor instead.
func (*Plugins) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{3}
}

func (x *Plugins) GetStorage() map[string]*Storage {
	if x != nil {
		return x.Storage
	}
	return nil
}

func (x *Plugins) GetCache() map[string]*Cache {
	if x != nil {
		return x.Cache
	}
	return nil
}

func (x *Plugins) GetBroker() map[string]*Broker {
	if x != nil {
		return x.Broker
	}
	return nil
}

type Client struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Service string `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	Package string `protobuf:"bytes,2,opt,name=package,proto3" json:"package,omitempty"`
}

func (x *Client) Reset() {
	*x = Client{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Client) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Client) ProtoMessage() {}

func (x *Client) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Client.ProtoReflect.Descriptor instead.
func (*Client) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{4}
}

func (x *Client) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *Client) GetPackage() string {
	if x != nil {
		return x.Package
	}
	return ""
}

type Clients struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Client []*Client `protobuf:"bytes,1,rep,name=client,proto3" json:"client,omitempty"`
}

func (x *Clients) Reset() {
	*x = Clients{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Clients) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Clients) ProtoMessage() {}

func (x *Clients) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Clients.ProtoReflect.Descriptor instead.
func (*Clients) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{5}
}

func (x *Clients) GetClient() []*Client {
	if x != nil {
		return x.Client
	}
	return nil
}

type DockerfileSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Package         string   `protobuf:"bytes,1,opt,name=package,proto3" json:"package,omitempty"`
	Expose          []int32  `protobuf:"varint,2,rep,packed,name=expose,proto3" json:"expose,omitempty"`
	Commands        []string `protobuf:"bytes,3,rep,name=commands,proto3" json:"commands,omitempty"`
	RewriteIfExists bool     `protobuf:"varint,4,opt,name=rewrite_if_exists,json=rewriteIfExists,proto3" json:"rewrite_if_exists,omitempty"`
}

func (x *DockerfileSpec) Reset() {
	*x = DockerfileSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DockerfileSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DockerfileSpec) ProtoMessage() {}

func (x *DockerfileSpec) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DockerfileSpec.ProtoReflect.Descriptor instead.
func (*DockerfileSpec) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{6}
}

func (x *DockerfileSpec) GetPackage() string {
	if x != nil {
		return x.Package
	}
	return ""
}

func (x *DockerfileSpec) GetExpose() []int32 {
	if x != nil {
		return x.Expose
	}
	return nil
}

func (x *DockerfileSpec) GetCommands() []string {
	if x != nil {
		return x.Commands
	}
	return nil
}

func (x *DockerfileSpec) GetRewriteIfExists() bool {
	if x != nil {
		return x.RewriteIfExists
	}
	return false
}

type TestSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mockery *MockeryTestsSpec `protobuf:"bytes,1,opt,name=mockery,proto3" json:"mockery,omitempty"`
}

func (x *TestSpec) Reset() {
	*x = TestSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestSpec) ProtoMessage() {}

func (x *TestSpec) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestSpec.ProtoReflect.Descriptor instead.
func (*TestSpec) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{7}
}

func (x *TestSpec) GetMockery() *MockeryTestsSpec {
	if x != nil {
		return x.Mockery
	}
	return nil
}

type MockeryTestsSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Package string `protobuf:"bytes,1,opt,name=package,proto3" json:"package,omitempty"`
}

func (x *MockeryTestsSpec) Reset() {
	*x = MockeryTestsSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MockeryTestsSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MockeryTestsSpec) ProtoMessage() {}

func (x *MockeryTestsSpec) ProtoReflect() protoreflect.Message {
	mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MockeryTestsSpec.ProtoReflect.Descriptor instead.
func (*MockeryTestsSpec) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{8}
}

func (x *MockeryTestsSpec) GetPackage() string {
	if x != nil {
		return x.Package
	}
	return ""
}

var file_protoc_gen_engine_engine_options_annotations_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*Plugins)(nil),
		Field:         50001,
		Name:          "engine.plugins",
		Tag:           "bytes,50001,opt,name=plugins",
		Filename:      "protoc-gen-engine/engine/options/annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*Clients)(nil),
		Field:         50002,
		Name:          "engine.clients",
		Tag:           "bytes,50002,opt,name=clients",
		Filename:      "protoc-gen-engine/engine/options/annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FileOptions)(nil),
		ExtensionType: (*DockerfileSpec)(nil),
		Field:         50003,
		Name:          "engine.dockerfile_spec",
		Tag:           "bytes,50003,opt,name=dockerfile_spec",
		Filename:      "protoc-gen-engine/engine/options/annotations.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FileOptions)(nil),
		ExtensionType: (*TestSpec)(nil),
		Field:         50004,
		Name:          "engine.tests_mockery_spec",
		Tag:           "bytes,50004,opt,name=tests_mockery_spec",
		Filename:      "protoc-gen-engine/engine/options/annotations.proto",
	},
}

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional engine.Plugins plugins = 50001;
	E_Plugins = &file_protoc_gen_engine_engine_options_annotations_proto_extTypes[0]
	// optional engine.Clients clients = 50002;
	E_Clients = &file_protoc_gen_engine_engine_options_annotations_proto_extTypes[1]
)

// Extension fields to descriptorpb.FileOptions.
var (
	// optional engine.DockerfileSpec dockerfile_spec = 50003;
	E_DockerfileSpec = &file_protoc_gen_engine_engine_options_annotations_proto_extTypes[2]
	// optional engine.TestSpec tests_mockery_spec = 50004;
	E_TestsMockerySpec = &file_protoc_gen_engine_engine_options_annotations_proto_extTypes[3]
)

var File_protoc_gen_engine_engine_options_annotations_proto protoreflect.FileDescriptor

var file_protoc_gen_engine_engine_options_annotations_proto_rawDesc = []byte{
	0x0a, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x65, 0x6e, 0x67,
	0x69, 0x6e, 0x65, 0x2f, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x1a, 0x20, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x37,
	0x0a, 0x05, 0x43, 0x61, 0x63, 0x68, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12,
	0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22, 0x39, 0x0a, 0x07, 0x53, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72,
	0x65, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66,
	0x69, 0x78, 0x22, 0x38, 0x0a, 0x06, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22, 0x89, 0x03, 0x0a,
	0x07, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12, 0x36, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x65, 0x6e, 0x67, 0x69,
	0x6e, 0x65, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x12, 0x30, 0x0a, 0x05, 0x63, 0x61, 0x63, 0x68, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73,
	0x2e, 0x43, 0x61, 0x63, 0x68, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x63, 0x61, 0x63,
	0x68, 0x65, 0x12, 0x33, 0x0a, 0x06, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x50, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x73, 0x2e, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x06, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x1a, 0x4b, 0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x25, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x1a, 0x47, 0x0a, 0x0a, 0x43, 0x61, 0x63, 0x68, 0x65, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x23, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x43, 0x61, 0x63,
	0x68, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x49, 0x0a,
	0x0b, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x24,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x3c, 0x0a, 0x06, 0x43, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x22, 0x31, 0x0a, 0x07, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x73, 0x12, 0x26, 0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0e, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x52, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x22, 0x8a, 0x01, 0x0a, 0x0e, 0x44, 0x6f,
	0x63, 0x6b, 0x65, 0x72, 0x66, 0x69, 0x6c, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x05, 0x52, 0x06, 0x65, 0x78, 0x70, 0x6f, 0x73, 0x65, 0x12, 0x1a,
	0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12, 0x2a, 0x0a, 0x11, 0x72, 0x65,
	0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x69, 0x66, 0x5f, 0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0f, 0x72, 0x65, 0x77, 0x72, 0x69, 0x74, 0x65, 0x49, 0x66,
	0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0x3e, 0x0a, 0x08, 0x54, 0x65, 0x73, 0x74, 0x53, 0x70,
	0x65, 0x63, 0x12, 0x32, 0x0a, 0x07, 0x6d, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x4d, 0x6f, 0x63,
	0x6b, 0x65, 0x72, 0x79, 0x54, 0x65, 0x73, 0x74, 0x73, 0x53, 0x70, 0x65, 0x63, 0x52, 0x07, 0x6d,
	0x6f, 0x63, 0x6b, 0x65, 0x72, 0x79, 0x22, 0x2c, 0x0a, 0x10, 0x4d, 0x6f, 0x63, 0x6b, 0x65, 0x72,
	0x79, 0x54, 0x65, 0x73, 0x74, 0x73, 0x53, 0x70, 0x65, 0x63, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x63,
	0x6b, 0x61, 0x67, 0x65, 0x3a, 0x4c, 0x0a, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12,
	0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x18, 0xd1, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x52, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x3a, 0x4c, 0x0a, 0x07, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x1f, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd2,
	0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x07, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73,
	0x3a, 0x5f, 0x0a, 0x0f, 0x64, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x73,
	0x70, 0x65, 0x63, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0xd3, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x65, 0x6e, 0x67, 0x69,
	0x6e, 0x65, 0x2e, 0x44, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x66, 0x69, 0x6c, 0x65, 0x53, 0x70, 0x65,
	0x63, 0x52, 0x0e, 0x64, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x66, 0x69, 0x6c, 0x65, 0x53, 0x70, 0x65,
	0x63, 0x3a, 0x5e, 0x0a, 0x12, 0x74, 0x65, 0x73, 0x74, 0x73, 0x5f, 0x6d, 0x6f, 0x63, 0x6b, 0x65,
	0x72, 0x79, 0x5f, 0x73, 0x70, 0x65, 0x63, 0x12, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd4, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e,
	0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x53, 0x70, 0x65, 0x63, 0x52,
	0x10, 0x74, 0x65, 0x73, 0x74, 0x73, 0x4d, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x79, 0x53, 0x70, 0x65,
	0x63, 0x42, 0x4c, 0x5a, 0x4a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6c, 0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x65, 0x6e, 0x67, 0x69,
	0x6e, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x65, 0x6e,
	0x67, 0x69, 0x6e, 0x65, 0x2f, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x3b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protoc_gen_engine_engine_options_annotations_proto_rawDescOnce sync.Once
	file_protoc_gen_engine_engine_options_annotations_proto_rawDescData = file_protoc_gen_engine_engine_options_annotations_proto_rawDesc
)

func file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP() []byte {
	file_protoc_gen_engine_engine_options_annotations_proto_rawDescOnce.Do(func() {
		file_protoc_gen_engine_engine_options_annotations_proto_rawDescData = protoimpl.X.CompressGZIP(file_protoc_gen_engine_engine_options_annotations_proto_rawDescData)
	})
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescData
}

var file_protoc_gen_engine_engine_options_annotations_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_protoc_gen_engine_engine_options_annotations_proto_goTypes = []interface{}{
	(*Cache)(nil),                       // 0: engine.Cache
	(*Storage)(nil),                     // 1: engine.Storage
	(*Broker)(nil),                      // 2: engine.Broker
	(*Plugins)(nil),                     // 3: engine.Plugins
	(*Client)(nil),                      // 4: engine.Client
	(*Clients)(nil),                     // 5: engine.Clients
	(*DockerfileSpec)(nil),              // 6: engine.DockerfileSpec
	(*TestSpec)(nil),                    // 7: engine.TestSpec
	(*MockeryTestsSpec)(nil),            // 8: engine.MockeryTestsSpec
	nil,                                 // 9: engine.Plugins.StorageEntry
	nil,                                 // 10: engine.Plugins.CacheEntry
	nil,                                 // 11: engine.Plugins.BrokerEntry
	(*descriptorpb.ServiceOptions)(nil), // 12: google.protobuf.ServiceOptions
	(*descriptorpb.FileOptions)(nil),    // 13: google.protobuf.FileOptions
}
var file_protoc_gen_engine_engine_options_annotations_proto_depIdxs = []int32{
	9,  // 0: engine.Plugins.storage:type_name -> engine.Plugins.StorageEntry
	10, // 1: engine.Plugins.cache:type_name -> engine.Plugins.CacheEntry
	11, // 2: engine.Plugins.broker:type_name -> engine.Plugins.BrokerEntry
	4,  // 3: engine.Clients.client:type_name -> engine.Client
	8,  // 4: engine.TestSpec.mockery:type_name -> engine.MockeryTestsSpec
	1,  // 5: engine.Plugins.StorageEntry.value:type_name -> engine.Storage
	0,  // 6: engine.Plugins.CacheEntry.value:type_name -> engine.Cache
	2,  // 7: engine.Plugins.BrokerEntry.value:type_name -> engine.Broker
	12, // 8: engine.plugins:extendee -> google.protobuf.ServiceOptions
	12, // 9: engine.clients:extendee -> google.protobuf.ServiceOptions
	13, // 10: engine.dockerfile_spec:extendee -> google.protobuf.FileOptions
	13, // 11: engine.tests_mockery_spec:extendee -> google.protobuf.FileOptions
	3,  // 12: engine.plugins:type_name -> engine.Plugins
	5,  // 13: engine.clients:type_name -> engine.Clients
	6,  // 14: engine.dockerfile_spec:type_name -> engine.DockerfileSpec
	7,  // 15: engine.tests_mockery_spec:type_name -> engine.TestSpec
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	12, // [12:16] is the sub-list for extension type_name
	8,  // [8:12] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_protoc_gen_engine_engine_options_annotations_proto_init() }
func file_protoc_gen_engine_engine_options_annotations_proto_init() {
	if File_protoc_gen_engine_engine_options_annotations_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Cache); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Storage); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Broker); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Plugins); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Client); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Clients); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DockerfileSpec); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestSpec); i {
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MockeryTestsSpec); i {
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
			RawDescriptor: file_protoc_gen_engine_engine_options_annotations_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 4,
			NumServices:   0,
		},
		GoTypes:           file_protoc_gen_engine_engine_options_annotations_proto_goTypes,
		DependencyIndexes: file_protoc_gen_engine_engine_options_annotations_proto_depIdxs,
		MessageInfos:      file_protoc_gen_engine_engine_options_annotations_proto_msgTypes,
		ExtensionInfos:    file_protoc_gen_engine_engine_options_annotations_proto_extTypes,
	}.Build()
	File_protoc_gen_engine_engine_options_annotations_proto = out.File
	file_protoc_gen_engine_engine_options_annotations_proto_rawDesc = nil
	file_protoc_gen_engine_engine_options_annotations_proto_goTypes = nil
	file_protoc_gen_engine_engine_options_annotations_proto_depIdxs = nil
}
