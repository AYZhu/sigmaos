// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.12.4
// source: socialnetwork/proto/mol.proto

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

type MoLRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *MoLRequest) Reset() {
	*x = MoLRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_socialnetwork_proto_mol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MoLRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MoLRequest) ProtoMessage() {}

func (x *MoLRequest) ProtoReflect() protoreflect.Message {
	mi := &file_socialnetwork_proto_mol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MoLRequest.ProtoReflect.Descriptor instead.
func (*MoLRequest) Descriptor() ([]byte, []int) {
	return file_socialnetwork_proto_mol_proto_rawDescGZIP(), []int{0}
}

func (x *MoLRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type MoLResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Meaning float32 `protobuf:"fixed32,1,opt,name=meaning,proto3" json:"meaning,omitempty"`
}

func (x *MoLResult) Reset() {
	*x = MoLResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_socialnetwork_proto_mol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MoLResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MoLResult) ProtoMessage() {}

func (x *MoLResult) ProtoReflect() protoreflect.Message {
	mi := &file_socialnetwork_proto_mol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MoLResult.ProtoReflect.Descriptor instead.
func (*MoLResult) Descriptor() ([]byte, []int) {
	return file_socialnetwork_proto_mol_proto_rawDescGZIP(), []int{1}
}

func (x *MoLResult) GetMeaning() float32 {
	if x != nil {
		return x.Meaning
	}
	return 0
}

var File_socialnetwork_proto_mol_proto protoreflect.FileDescriptor

var file_socialnetwork_proto_mol_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x73, 0x6f, 0x63, 0x69, 0x61, 0x6c, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x20, 0x0a, 0x0a, 0x4d, 0x6f, 0x4c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x22, 0x25, 0x0a, 0x09, 0x4d, 0x6f, 0x4c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x6d, 0x65, 0x61, 0x6e, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52,
	0x07, 0x6d, 0x65, 0x61, 0x6e, 0x69, 0x6e, 0x67, 0x32, 0x37, 0x0a, 0x0d, 0x4d, 0x65, 0x61, 0x6e,
	0x69, 0x6e, 0x67, 0x4f, 0x66, 0x4c, 0x69, 0x66, 0x65, 0x12, 0x26, 0x0a, 0x0b, 0x46, 0x69, 0x6e,
	0x64, 0x4d, 0x65, 0x61, 0x6e, 0x69, 0x6e, 0x67, 0x12, 0x0b, 0x2e, 0x4d, 0x6f, 0x4c, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x4d, 0x6f, 0x4c, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x42, 0x1d, 0x5a, 0x1b, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x6f, 0x73, 0x2f, 0x73, 0x6f, 0x63,
	0x69, 0x61, 0x6c, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_socialnetwork_proto_mol_proto_rawDescOnce sync.Once
	file_socialnetwork_proto_mol_proto_rawDescData = file_socialnetwork_proto_mol_proto_rawDesc
)

func file_socialnetwork_proto_mol_proto_rawDescGZIP() []byte {
	file_socialnetwork_proto_mol_proto_rawDescOnce.Do(func() {
		file_socialnetwork_proto_mol_proto_rawDescData = protoimpl.X.CompressGZIP(file_socialnetwork_proto_mol_proto_rawDescData)
	})
	return file_socialnetwork_proto_mol_proto_rawDescData
}

var file_socialnetwork_proto_mol_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_socialnetwork_proto_mol_proto_goTypes = []interface{}{
	(*MoLRequest)(nil), // 0: MoLRequest
	(*MoLResult)(nil),  // 1: MoLResult
}
var file_socialnetwork_proto_mol_proto_depIdxs = []int32{
	0, // 0: MeaningOfLife.FindMeaning:input_type -> MoLRequest
	1, // 1: MeaningOfLife.FindMeaning:output_type -> MoLResult
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_socialnetwork_proto_mol_proto_init() }
func file_socialnetwork_proto_mol_proto_init() {
	if File_socialnetwork_proto_mol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_socialnetwork_proto_mol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MoLRequest); i {
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
		file_socialnetwork_proto_mol_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MoLResult); i {
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
			RawDescriptor: file_socialnetwork_proto_mol_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_socialnetwork_proto_mol_proto_goTypes,
		DependencyIndexes: file_socialnetwork_proto_mol_proto_depIdxs,
		MessageInfos:      file_socialnetwork_proto_mol_proto_msgTypes,
	}.Build()
	File_socialnetwork_proto_mol_proto = out.File
	file_socialnetwork_proto_mol_proto_rawDesc = nil
	file_socialnetwork_proto_mol_proto_goTypes = nil
	file_socialnetwork_proto_mol_proto_depIdxs = nil
}
