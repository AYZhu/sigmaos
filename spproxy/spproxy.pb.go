// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.12.4
// source: spproxy.proto

package spproxy

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

type OpenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	Mode string `protobuf:"bytes,2,opt,name=mode,proto3" json:"mode,omitempty"`
}

func (x *OpenRequest) Reset() {
	*x = OpenRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_spproxy_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OpenRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OpenRequest) ProtoMessage() {}

func (x *OpenRequest) ProtoReflect() protoreflect.Message {
	mi := &file_spproxy_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OpenRequest.ProtoReflect.Descriptor instead.
func (*OpenRequest) Descriptor() ([]byte, []int) {
	return file_spproxy_proto_rawDescGZIP(), []int{0}
}

func (x *OpenRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *OpenRequest) GetMode() string {
	if x != nil {
		return x.Mode
	}
	return ""
}

type OpenResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result int64 `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *OpenResult) Reset() {
	*x = OpenResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_spproxy_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OpenResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OpenResult) ProtoMessage() {}

func (x *OpenResult) ProtoReflect() protoreflect.Message {
	mi := &file_spproxy_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OpenResult.ProtoReflect.Descriptor instead.
func (*OpenResult) Descriptor() ([]byte, []int) {
	return file_spproxy_proto_rawDescGZIP(), []int{1}
}

func (x *OpenResult) GetResult() int64 {
	if x != nil {
		return x.Result
	}
	return 0
}

var File_spproxy_proto protoreflect.FileDescriptor

var file_spproxy_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x70, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x35, 0x0a, 0x0b, 0x4f, 0x70, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65,
	0x78, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x22, 0x24, 0x0a, 0x0a, 0x4f, 0x70, 0x65, 0x6e, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x32, 0x37, 0x0a, 0x0e,
	0x53, 0x50, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x25,
	0x0a, 0x08, 0x53, 0x50, 0x50, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x12, 0x0c, 0x2e, 0x4f, 0x70, 0x65,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0b, 0x2e, 0x4f, 0x70, 0x65, 0x6e, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x11, 0x5a, 0x0f, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x6f, 0x73,
	0x2f, 0x73, 0x70, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_spproxy_proto_rawDescOnce sync.Once
	file_spproxy_proto_rawDescData = file_spproxy_proto_rawDesc
)

func file_spproxy_proto_rawDescGZIP() []byte {
	file_spproxy_proto_rawDescOnce.Do(func() {
		file_spproxy_proto_rawDescData = protoimpl.X.CompressGZIP(file_spproxy_proto_rawDescData)
	})
	return file_spproxy_proto_rawDescData
}

var file_spproxy_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_spproxy_proto_goTypes = []interface{}{
	(*OpenRequest)(nil), // 0: OpenRequest
	(*OpenResult)(nil),  // 1: OpenResult
}
var file_spproxy_proto_depIdxs = []int32{
	0, // 0: SPProxyService.SPPProxy:input_type -> OpenRequest
	1, // 1: SPProxyService.SPPProxy:output_type -> OpenResult
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_spproxy_proto_init() }
func file_spproxy_proto_init() {
	if File_spproxy_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_spproxy_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OpenRequest); i {
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
		file_spproxy_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OpenResult); i {
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
			RawDescriptor: file_spproxy_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_spproxy_proto_goTypes,
		DependencyIndexes: file_spproxy_proto_depIdxs,
		MessageInfos:      file_spproxy_proto_msgTypes,
	}.Build()
	File_spproxy_proto = out.File
	file_spproxy_proto_rawDesc = nil
	file_spproxy_proto_goTypes = nil
	file_spproxy_proto_depIdxs = nil
}
