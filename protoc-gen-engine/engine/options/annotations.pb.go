// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.14.0
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

type Client struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Plugin string `protobuf:"bytes,1,opt,name=plugin,proto3" json:"plugin,omitempty"`
	Prefix string `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *Client) Reset() {
	*x = Client{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Client) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Client) ProtoMessage() {}

func (x *Client) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Client.ProtoReflect.Descriptor instead.
func (*Client) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{3}
}

func (x *Client) GetPlugin() string {
	if x != nil {
		return x.Plugin
	}
	return ""
}

func (x *Client) GetPrefix() string {
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
	Client  map[string]*Client  `protobuf:"bytes,4,rep,name=client,proto3" json:"client,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Plugins) Reset() {
	*x = Plugins{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plugins) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plugins) ProtoMessage() {}

func (x *Plugins) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use Plugins.ProtoReflect.Descriptor instead.
func (*Plugins) Descriptor() ([]byte, []int) {
	return file_protoc_gen_engine_engine_options_annotations_proto_rawDescGZIP(), []int{4}
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

func (x *Plugins) GetClient() map[string]*Client {
	if x != nil {
		return x.Client
	}
	return nil
}

var file_protoc_gen_engine_engine_options_annotations_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*Plugins)(nil),
		Field:         50001,
		Name:          "custom_options.plugins",
		Tag:           "bytes,50001,opt,name=plugins",
		Filename:      "protoc-gen-engine/engine/options/annotations.proto",
	},
}

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional custom_options.Plugins plugins = 50001;
	E_Plugins = &file_protoc_gen_engine_engine_options_annotations_proto_extTypes[0]
)

var File_protoc_gen_engine_engine_options_annotations_proto protoreflect.FileDescriptor

var file_protoc_gen_engine_engine_options_annotations_proto_rawDesc = []byte{
	0x0a, 0x32, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x65, 0x6e, 0x67,
	0x69, 0x6e, 0x65, 0x2f, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x37, 0x0a, 0x05, 0x43, 0x61, 0x63, 0x68, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69,
	0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22,
	0x39, 0x0a, 0x07, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22, 0x38, 0x0a, 0x06, 0x42, 0x72,
	0x6f, 0x6b, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06,
	0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72,
	0x65, 0x66, 0x69, 0x78, 0x22, 0x38, 0x0a, 0x06, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x22, 0xc9,
	0x04, 0x0a, 0x07, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12, 0x3e, 0x0a, 0x07, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x63, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x50, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x38, 0x0a, 0x05, 0x63, 0x61,
	0x63, 0x68, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x2e, 0x43, 0x61, 0x63, 0x68, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x63,
	0x61, 0x63, 0x68, 0x65, 0x12, 0x3b, 0x0a, 0x06, 0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x42, 0x72,
	0x6f, 0x6b, 0x65, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x62, 0x72, 0x6f, 0x6b, 0x65,
	0x72, 0x12, 0x3b, 0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x23, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x1a, 0x53,
	0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x2d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x1a, 0x4f, 0x0a, 0x0a, 0x43, 0x61, 0x63, 0x68, 0x65, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x2b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x15, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x43, 0x61, 0x63, 0x68, 0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x1a, 0x51, 0x0a, 0x0b, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x42, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x51, 0x0a, 0x0b, 0x43, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d,
	0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x3a, 0x54, 0x0a, 0x07, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd1, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x52, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73,
	0x42, 0x4c, 0x5a, 0x4a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c,
	0x61, 0x73, 0x74, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x65, 0x6e, 0x67, 0x69, 0x6e,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x65, 0x6e, 0x67,
	0x69, 0x6e, 0x65, 0x2f, 0x65, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x3b, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_protoc_gen_engine_engine_options_annotations_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_protoc_gen_engine_engine_options_annotations_proto_goTypes = []interface{}{
	(*Cache)(nil),                       // 0: custom_options.Cache
	(*Storage)(nil),                     // 1: custom_options.Storage
	(*Broker)(nil),                      // 2: custom_options.Broker
	(*Client)(nil),                      // 3: custom_options.Client
	(*Plugins)(nil),                     // 4: custom_options.Plugins
	nil,                                 // 5: custom_options.Plugins.StorageEntry
	nil,                                 // 6: custom_options.Plugins.CacheEntry
	nil,                                 // 7: custom_options.Plugins.BrokerEntry
	nil,                                 // 8: custom_options.Plugins.ClientEntry
	(*descriptorpb.ServiceOptions)(nil), // 9: google.protobuf.ServiceOptions
}
var file_protoc_gen_engine_engine_options_annotations_proto_depIdxs = []int32{
	5,  // 0: custom_options.Plugins.storage:type_name -> custom_options.Plugins.StorageEntry
	6,  // 1: custom_options.Plugins.cache:type_name -> custom_options.Plugins.CacheEntry
	7,  // 2: custom_options.Plugins.broker:type_name -> custom_options.Plugins.BrokerEntry
	8,  // 3: custom_options.Plugins.client:type_name -> custom_options.Plugins.ClientEntry
	1,  // 4: custom_options.Plugins.StorageEntry.value:type_name -> custom_options.Storage
	0,  // 5: custom_options.Plugins.CacheEntry.value:type_name -> custom_options.Cache
	2,  // 6: custom_options.Plugins.BrokerEntry.value:type_name -> custom_options.Broker
	3,  // 7: custom_options.Plugins.ClientEntry.value:type_name -> custom_options.Client
	9,  // 8: custom_options.plugins:extendee -> google.protobuf.ServiceOptions
	4,  // 9: custom_options.plugins:type_name -> custom_options.Plugins
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	9,  // [9:10] is the sub-list for extension type_name
	8,  // [8:9] is the sub-list for extension extendee
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
		file_protoc_gen_engine_engine_options_annotations_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protoc_gen_engine_engine_options_annotations_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 1,
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
