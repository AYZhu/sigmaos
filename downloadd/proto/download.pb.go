// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.12.4
// source: download.proto

package proto

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

type DownloadLibRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NamedPath  string `protobuf:"bytes,1,opt,name=namedPath,proto3" json:"namedPath,omitempty"`
	Realm      string `protobuf:"bytes,2,opt,name=realm,proto3" json:"realm,omitempty"`
	CopyFolder bool   `protobuf:"varint,3,opt,name=copyFolder,proto3" json:"copyFolder,omitempty"`
}

func (x *DownloadLibRequest) Reset() {
	*x = DownloadLibRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_download_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadLibRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadLibRequest) ProtoMessage() {}

func (x *DownloadLibRequest) ProtoReflect() protoreflect.Message {
	mi := &file_download_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadLibRequest.ProtoReflect.Descriptor instead.
func (*DownloadLibRequest) Descriptor() ([]byte, []int) {
	return file_download_proto_rawDescGZIP(), []int{0}
}

func (x *DownloadLibRequest) GetNamedPath() string {
	if x != nil {
		return x.NamedPath
	}
	return ""
}

func (x *DownloadLibRequest) GetRealm() string {
	if x != nil {
		return x.Realm
	}
	return ""
}

func (x *DownloadLibRequest) GetCopyFolder() bool {
	if x != nil {
		return x.CopyFolder
	}
	return false
}

type DownloadLibResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TmpPath string `protobuf:"bytes,1,opt,name=tmpPath,proto3" json:"tmpPath,omitempty"`
}

func (x *DownloadLibResponse) Reset() {
	*x = DownloadLibResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_download_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadLibResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadLibResponse) ProtoMessage() {}

func (x *DownloadLibResponse) ProtoReflect() protoreflect.Message {
	mi := &file_download_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadLibResponse.ProtoReflect.Descriptor instead.
func (*DownloadLibResponse) Descriptor() ([]byte, []int) {
	return file_download_proto_rawDescGZIP(), []int{1}
}

func (x *DownloadLibResponse) GetTmpPath() string {
	if x != nil {
		return x.TmpPath
	}
	return ""
}

var File_download_proto protoreflect.FileDescriptor

var file_download_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x68, 0x0a, 0x12, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x62, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x64, 0x50,
	0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x64,
	0x50, 0x61, 0x74, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f,
	0x70, 0x79, 0x46, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a,
	0x63, 0x6f, 0x70, 0x79, 0x46, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x22, 0x2f, 0x0a, 0x13, 0x44, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x6d, 0x70, 0x50, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x74, 0x6d, 0x70, 0x50, 0x61, 0x74, 0x68, 0x32, 0x45, 0x0a, 0x09, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x64, 0x12, 0x38, 0x0a, 0x0b, 0x44, 0x6f, 0x77, 0x6e,
	0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x62, 0x12, 0x13, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f,
	0x61, 0x64, 0x4c, 0x69, 0x62, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x62, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x19, 0x5a, 0x17, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x6f, 0x73, 0x2f, 0x64, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_download_proto_rawDescOnce sync.Once
	file_download_proto_rawDescData = file_download_proto_rawDesc
)

func file_download_proto_rawDescGZIP() []byte {
	file_download_proto_rawDescOnce.Do(func() {
		file_download_proto_rawDescData = protoimpl.X.CompressGZIP(file_download_proto_rawDescData)
	})
	return file_download_proto_rawDescData
}

var file_download_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_download_proto_goTypes = []interface{}{
	(*DownloadLibRequest)(nil),  // 0: DownloadLibRequest
	(*DownloadLibResponse)(nil), // 1: DownloadLibResponse
}
var file_download_proto_depIdxs = []int32{
	0, // 0: Downloadd.DownloadLib:input_type -> DownloadLibRequest
	1, // 1: Downloadd.DownloadLib:output_type -> DownloadLibResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_download_proto_init() }
func file_download_proto_init() {
	if File_download_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_download_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadLibRequest); i {
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
		file_download_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadLibResponse); i {
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
			RawDescriptor: file_download_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_download_proto_goTypes,
		DependencyIndexes: file_download_proto_depIdxs,
		MessageInfos:      file_download_proto_msgTypes,
	}.Build()
	File_download_proto = out.File
	file_download_proto_rawDesc = nil
	file_download_proto_goTypes = nil
	file_download_proto_depIdxs = nil
}
