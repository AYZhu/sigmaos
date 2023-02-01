// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: kernelsrv/proto/kernelsrv.proto

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

type BootRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Args []string `protobuf:"bytes,2,rep,name=args,proto3" json:"args,omitempty"`
}

func (x *BootRequest) Reset() {
	*x = BootRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BootRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BootRequest) ProtoMessage() {}

func (x *BootRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BootRequest.ProtoReflect.Descriptor instead.
func (*BootRequest) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{0}
}

func (x *BootRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *BootRequest) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

type BootResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PidStr string `protobuf:"bytes,1,opt,name=pidStr,proto3" json:"pidStr,omitempty"`
}

func (x *BootResult) Reset() {
	*x = BootResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BootResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BootResult) ProtoMessage() {}

func (x *BootResult) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BootResult.ProtoReflect.Descriptor instead.
func (*BootResult) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{1}
}

func (x *BootResult) GetPidStr() string {
	if x != nil {
		return x.PidStr
	}
	return ""
}

type SetCPUSharesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PidStr string `protobuf:"bytes,1,opt,name=PidStr,proto3" json:"PidStr,omitempty"`
	Shares int64  `protobuf:"varint,2,opt,name=Shares,proto3" json:"Shares,omitempty"`
}

func (x *SetCPUSharesRequest) Reset() {
	*x = SetCPUSharesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetCPUSharesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetCPUSharesRequest) ProtoMessage() {}

func (x *SetCPUSharesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetCPUSharesRequest.ProtoReflect.Descriptor instead.
func (*SetCPUSharesRequest) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{2}
}

func (x *SetCPUSharesRequest) GetPidStr() string {
	if x != nil {
		return x.PidStr
	}
	return ""
}

func (x *SetCPUSharesRequest) GetShares() int64 {
	if x != nil {
		return x.Shares
	}
	return 0
}

type SetCPUSharesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetCPUSharesResponse) Reset() {
	*x = SetCPUSharesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetCPUSharesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetCPUSharesResponse) ProtoMessage() {}

func (x *SetCPUSharesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetCPUSharesResponse.ProtoReflect.Descriptor instead.
func (*SetCPUSharesResponse) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{3}
}

type GetCPUUtilRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PidStr string `protobuf:"bytes,1,opt,name=PidStr,proto3" json:"PidStr,omitempty"`
}

func (x *GetCPUUtilRequest) Reset() {
	*x = GetCPUUtilRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCPUUtilRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCPUUtilRequest) ProtoMessage() {}

func (x *GetCPUUtilRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCPUUtilRequest.ProtoReflect.Descriptor instead.
func (*GetCPUUtilRequest) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{4}
}

func (x *GetCPUUtilRequest) GetPidStr() string {
	if x != nil {
		return x.PidStr
	}
	return ""
}

type GetCPUUtilResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Util float64 `protobuf:"fixed64,1,opt,name=util,proto3" json:"util,omitempty"`
}

func (x *GetCPUUtilResponse) Reset() {
	*x = GetCPUUtilResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCPUUtilResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCPUUtilResponse) ProtoMessage() {}

func (x *GetCPUUtilResponse) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCPUUtilResponse.ProtoReflect.Descriptor instead.
func (*GetCPUUtilResponse) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{5}
}

func (x *GetCPUUtilResponse) GetUtil() float64 {
	if x != nil {
		return x.Util
	}
	return 0
}

type ShutdownRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ShutdownRequest) Reset() {
	*x = ShutdownRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShutdownRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShutdownRequest) ProtoMessage() {}

func (x *ShutdownRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShutdownRequest.ProtoReflect.Descriptor instead.
func (*ShutdownRequest) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{6}
}

type ShutdownResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ShutdownResult) Reset() {
	*x = ShutdownResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShutdownResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShutdownResult) ProtoMessage() {}

func (x *ShutdownResult) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShutdownResult.ProtoReflect.Descriptor instead.
func (*ShutdownResult) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{7}
}

type KillRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *KillRequest) Reset() {
	*x = KillRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KillRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KillRequest) ProtoMessage() {}

func (x *KillRequest) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KillRequest.ProtoReflect.Descriptor instead.
func (*KillRequest) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{8}
}

func (x *KillRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type KillResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *KillResult) Reset() {
	*x = KillResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KillResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KillResult) ProtoMessage() {}

func (x *KillResult) ProtoReflect() protoreflect.Message {
	mi := &file_kernelsrv_proto_kernelsrv_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KillResult.ProtoReflect.Descriptor instead.
func (*KillResult) Descriptor() ([]byte, []int) {
	return file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP(), []int{9}
}

var File_kernelsrv_proto_kernelsrv_proto protoreflect.FileDescriptor

var file_kernelsrv_proto_kernelsrv_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x73, 0x72, 0x76, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x73, 0x72, 0x76, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x35, 0x0a, 0x0b, 0x42, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x22, 0x24, 0x0a, 0x0a, 0x42, 0x6f, 0x6f, 0x74,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x69, 0x64, 0x53, 0x74, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x69, 0x64, 0x53, 0x74, 0x72, 0x22, 0x45,
	0x0a, 0x13, 0x53, 0x65, 0x74, 0x43, 0x50, 0x55, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x69, 0x64, 0x53, 0x74, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x50, 0x69, 0x64, 0x53, 0x74, 0x72, 0x12, 0x16, 0x0a,
	0x06, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x53,
	0x68, 0x61, 0x72, 0x65, 0x73, 0x22, 0x16, 0x0a, 0x14, 0x53, 0x65, 0x74, 0x43, 0x50, 0x55, 0x53,
	0x68, 0x61, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b, 0x0a,
	0x11, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x55, 0x74, 0x69, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x69, 0x64, 0x53, 0x74, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x50, 0x69, 0x64, 0x53, 0x74, 0x72, 0x22, 0x28, 0x0a, 0x12, 0x47, 0x65,
	0x74, 0x43, 0x50, 0x55, 0x55, 0x74, 0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x75, 0x74, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04,
	0x75, 0x74, 0x69, 0x6c, 0x22, 0x11, 0x0a, 0x0f, 0x53, 0x68, 0x75, 0x74, 0x64, 0x6f, 0x77, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x10, 0x0a, 0x0e, 0x53, 0x68, 0x75, 0x74, 0x64,
	0x6f, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x21, 0x0a, 0x0b, 0x4b, 0x69, 0x6c,
	0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x0c, 0x0a, 0x0a,
	0x4b, 0x69, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x19, 0x5a, 0x17, 0x73, 0x69,
	0x67, 0x6d, 0x61, 0x6f, 0x73, 0x2f, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x73, 0x72, 0x76, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kernelsrv_proto_kernelsrv_proto_rawDescOnce sync.Once
	file_kernelsrv_proto_kernelsrv_proto_rawDescData = file_kernelsrv_proto_kernelsrv_proto_rawDesc
)

func file_kernelsrv_proto_kernelsrv_proto_rawDescGZIP() []byte {
	file_kernelsrv_proto_kernelsrv_proto_rawDescOnce.Do(func() {
		file_kernelsrv_proto_kernelsrv_proto_rawDescData = protoimpl.X.CompressGZIP(file_kernelsrv_proto_kernelsrv_proto_rawDescData)
	})
	return file_kernelsrv_proto_kernelsrv_proto_rawDescData
}

var file_kernelsrv_proto_kernelsrv_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_kernelsrv_proto_kernelsrv_proto_goTypes = []interface{}{
	(*BootRequest)(nil),          // 0: BootRequest
	(*BootResult)(nil),           // 1: BootResult
	(*SetCPUSharesRequest)(nil),  // 2: SetCPUSharesRequest
	(*SetCPUSharesResponse)(nil), // 3: SetCPUSharesResponse
	(*GetCPUUtilRequest)(nil),    // 4: GetCPUUtilRequest
	(*GetCPUUtilResponse)(nil),   // 5: GetCPUUtilResponse
	(*ShutdownRequest)(nil),      // 6: ShutdownRequest
	(*ShutdownResult)(nil),       // 7: ShutdownResult
	(*KillRequest)(nil),          // 8: KillRequest
	(*KillResult)(nil),           // 9: KillResult
}
var file_kernelsrv_proto_kernelsrv_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kernelsrv_proto_kernelsrv_proto_init() }
func file_kernelsrv_proto_kernelsrv_proto_init() {
	if File_kernelsrv_proto_kernelsrv_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BootRequest); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BootResult); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetCPUSharesRequest); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetCPUSharesResponse); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCPUUtilRequest); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCPUUtilResponse); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShutdownRequest); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShutdownResult); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KillRequest); i {
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
		file_kernelsrv_proto_kernelsrv_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KillResult); i {
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
			RawDescriptor: file_kernelsrv_proto_kernelsrv_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kernelsrv_proto_kernelsrv_proto_goTypes,
		DependencyIndexes: file_kernelsrv_proto_kernelsrv_proto_depIdxs,
		MessageInfos:      file_kernelsrv_proto_kernelsrv_proto_msgTypes,
	}.Build()
	File_kernelsrv_proto_kernelsrv_proto = out.File
	file_kernelsrv_proto_kernelsrv_proto_rawDesc = nil
	file_kernelsrv_proto_kernelsrv_proto_goTypes = nil
	file_kernelsrv_proto_kernelsrv_proto_depIdxs = nil
}
