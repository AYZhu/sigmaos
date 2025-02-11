// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: hotel/proto/rec.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	proto "sigmaos/tracing/proto"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RecRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Require           string                   `protobuf:"bytes,1,opt,name=require,proto3" json:"require,omitempty"`
	Lat               float64                  `protobuf:"fixed64,2,opt,name=lat,proto3" json:"lat,omitempty"`
	Lon               float64                  `protobuf:"fixed64,3,opt,name=lon,proto3" json:"lon,omitempty"`
	SpanContextConfig *proto.SpanContextConfig `protobuf:"bytes,4,opt,name=spanContextConfig,proto3" json:"spanContextConfig,omitempty"`
}

func (x *RecRequest) Reset() {
	*x = RecRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hotel_proto_rec_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecRequest) ProtoMessage() {}

func (x *RecRequest) ProtoReflect() protoreflect.Message {
	mi := &file_hotel_proto_rec_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecRequest.ProtoReflect.Descriptor instead.
func (*RecRequest) Descriptor() ([]byte, []int) {
	return file_hotel_proto_rec_proto_rawDescGZIP(), []int{0}
}

func (x *RecRequest) GetRequire() string {
	if x != nil {
		return x.Require
	}
	return ""
}

func (x *RecRequest) GetLat() float64 {
	if x != nil {
		return x.Lat
	}
	return 0
}

func (x *RecRequest) GetLon() float64 {
	if x != nil {
		return x.Lon
	}
	return 0
}

func (x *RecRequest) GetSpanContextConfig() *proto.SpanContextConfig {
	if x != nil {
		return x.SpanContextConfig
	}
	return nil
}

type RecResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HotelIds []string `protobuf:"bytes,1,rep,name=hotelIds,proto3" json:"hotelIds,omitempty"`
}

func (x *RecResult) Reset() {
	*x = RecResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_hotel_proto_rec_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecResult) ProtoMessage() {}

func (x *RecResult) ProtoReflect() protoreflect.Message {
	mi := &file_hotel_proto_rec_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecResult.ProtoReflect.Descriptor instead.
func (*RecResult) Descriptor() ([]byte, []int) {
	return file_hotel_proto_rec_proto_rawDescGZIP(), []int{1}
}

func (x *RecResult) GetHotelIds() []string {
	if x != nil {
		return x.HotelIds
	}
	return nil
}

var File_hotel_proto_rec_proto protoreflect.FileDescriptor

var file_hotel_proto_rec_proto_rawDesc = []byte{
	0x0a, 0x15, 0x68, 0x6f, 0x74, 0x65, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x65,
	0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x72, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8c, 0x01, 0x0a, 0x0a, 0x52, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x6c, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x6c, 0x61, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x6c, 0x6f,
	0x6e, 0x12, 0x40, 0x0a, 0x11, 0x73, 0x70, 0x61, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x53,
	0x70, 0x61, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x52, 0x11, 0x73, 0x70, 0x61, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x22, 0x27, 0x0a, 0x09, 0x52, 0x65, 0x63, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x74, 0x65, 0x6c, 0x49, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x74, 0x65, 0x6c, 0x49, 0x64, 0x73, 0x32, 0x29, 0x0a, 0x03,
	0x52, 0x65, 0x63, 0x12, 0x22, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x73, 0x12, 0x0b,
	0x2e, 0x52, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x52, 0x65,
	0x63, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x15, 0x5a, 0x13, 0x73, 0x69, 0x67, 0x6d, 0x61,
	0x6f, 0x73, 0x2f, 0x68, 0x6f, 0x74, 0x65, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_hotel_proto_rec_proto_rawDescOnce sync.Once
	file_hotel_proto_rec_proto_rawDescData = file_hotel_proto_rec_proto_rawDesc
)

func file_hotel_proto_rec_proto_rawDescGZIP() []byte {
	file_hotel_proto_rec_proto_rawDescOnce.Do(func() {
		file_hotel_proto_rec_proto_rawDescData = protoimpl.X.CompressGZIP(file_hotel_proto_rec_proto_rawDescData)
	})
	return file_hotel_proto_rec_proto_rawDescData
}

var file_hotel_proto_rec_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_hotel_proto_rec_proto_goTypes = []interface{}{
	(*RecRequest)(nil),              // 0: RecRequest
	(*RecResult)(nil),               // 1: RecResult
	(*proto.SpanContextConfig)(nil), // 2: SpanContextConfig
}
var file_hotel_proto_rec_proto_depIdxs = []int32{
	2, // 0: RecRequest.spanContextConfig:type_name -> SpanContextConfig
	0, // 1: Rec.GetRecs:input_type -> RecRequest
	1, // 2: Rec.GetRecs:output_type -> RecResult
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_hotel_proto_rec_proto_init() }
func file_hotel_proto_rec_proto_init() {
	if File_hotel_proto_rec_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_hotel_proto_rec_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecRequest); i {
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
		file_hotel_proto_rec_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecResult); i {
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
			RawDescriptor: file_hotel_proto_rec_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_hotel_proto_rec_proto_goTypes,
		DependencyIndexes: file_hotel_proto_rec_proto_depIdxs,
		MessageInfos:      file_hotel_proto_rec_proto_msgTypes,
	}.Build()
	File_hotel_proto_rec_proto = out.File
	file_hotel_proto_rec_proto_rawDesc = nil
	file_hotel_proto_rec_proto_goTypes = nil
	file_hotel_proto_rec_proto_depIdxs = nil
}
