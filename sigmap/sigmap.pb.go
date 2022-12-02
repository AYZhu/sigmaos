// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: sigmap.proto

package sigmap

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

// A Qid is the server's unique identification for the file being
// accessed: two files on the same server hierarchy are the same if
// and only if their qids are the same.
type Tqid struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type    uint32 `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Version uint32 `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	Path    uint64 `protobuf:"varint,3,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *Tqid) Reset() {
	*x = Tqid{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tqid) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tqid) ProtoMessage() {}

func (x *Tqid) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tqid.ProtoReflect.Descriptor instead.
func (*Tqid) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{0}
}

func (x *Tqid) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *Tqid) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *Tqid) GetPath() uint64 {
	if x != nil {
		return x.Path
	}
	return 0
}

type Stat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type   uint32 `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Dev    uint32 `protobuf:"varint,2,opt,name=dev,proto3" json:"dev,omitempty"`
	Qid    *Tqid  `protobuf:"bytes,3,opt,name=qid,proto3" json:"qid,omitempty"`
	Mode   uint32 `protobuf:"varint,4,opt,name=mode,proto3" json:"mode,omitempty"`
	Atime  uint32 `protobuf:"varint,5,opt,name=atime,proto3" json:"atime,omitempty"`   // last access time in seconds
	Mtime  uint32 `protobuf:"varint,6,opt,name=mtime,proto3" json:"mtime,omitempty"`   // last modified time in seconds
	Length uint64 `protobuf:"varint,7,opt,name=length,proto3" json:"length,omitempty"` // file length in bytes
	Name   string `protobuf:"bytes,8,opt,name=name,proto3" json:"name,omitempty"`      // file name
	Uid    string `protobuf:"bytes,9,opt,name=uid,proto3" json:"uid,omitempty"`        // owner name
	Gid    string `protobuf:"bytes,10,opt,name=gid,proto3" json:"gid,omitempty"`       // group name
	Muid   string `protobuf:"bytes,11,opt,name=muid,proto3" json:"muid,omitempty"`     // name of the last user that modified the file
}

func (x *Stat) Reset() {
	*x = Stat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Stat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Stat) ProtoMessage() {}

func (x *Stat) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Stat.ProtoReflect.Descriptor instead.
func (*Stat) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{1}
}

func (x *Stat) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *Stat) GetDev() uint32 {
	if x != nil {
		return x.Dev
	}
	return 0
}

func (x *Stat) GetQid() *Tqid {
	if x != nil {
		return x.Qid
	}
	return nil
}

func (x *Stat) GetMode() uint32 {
	if x != nil {
		return x.Mode
	}
	return 0
}

func (x *Stat) GetAtime() uint32 {
	if x != nil {
		return x.Atime
	}
	return 0
}

func (x *Stat) GetMtime() uint32 {
	if x != nil {
		return x.Mtime
	}
	return 0
}

func (x *Stat) GetLength() uint64 {
	if x != nil {
		return x.Length
	}
	return 0
}

func (x *Stat) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Stat) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

func (x *Stat) GetGid() string {
	if x != nil {
		return x.Gid
	}
	return ""
}

func (x *Stat) GetMuid() string {
	if x != nil {
		return x.Muid
	}
	return ""
}

type Rattach struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Qid *Tqid `protobuf:"bytes,1,opt,name=qid,proto3" json:"qid,omitempty"`
}

func (x *Rattach) Reset() {
	*x = Rattach{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Rattach) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Rattach) ProtoMessage() {}

func (x *Rattach) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Rattach.ProtoReflect.Descriptor instead.
func (*Rattach) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{2}
}

func (x *Rattach) GetQid() *Tqid {
	if x != nil {
		return x.Qid
	}
	return nil
}

type Tinterval struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Start uint64 `protobuf:"varint,1,opt,name=start,proto3" json:"start,omitempty"`
	End   uint64 `protobuf:"varint,2,opt,name=end,proto3" json:"end,omitempty"`
}

func (x *Tinterval) Reset() {
	*x = Tinterval{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tinterval) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tinterval) ProtoMessage() {}

func (x *Tinterval) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tinterval.ProtoReflect.Descriptor instead.
func (*Tinterval) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{3}
}

func (x *Tinterval) GetStart() uint64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *Tinterval) GetEnd() uint64 {
	if x != nil {
		return x.End
	}
	return 0
}

type Tfenceid struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path     uint64 `protobuf:"varint,1,opt,name=path,proto3" json:"path,omitempty"`
	Serverid uint64 `protobuf:"varint,2,opt,name=serverid,proto3" json:"serverid,omitempty"`
}

func (x *Tfenceid) Reset() {
	*x = Tfenceid{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tfenceid) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tfenceid) ProtoMessage() {}

func (x *Tfenceid) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tfenceid.ProtoReflect.Descriptor instead.
func (*Tfenceid) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{4}
}

func (x *Tfenceid) GetPath() uint64 {
	if x != nil {
		return x.Path
	}
	return 0
}

func (x *Tfenceid) GetServerid() uint64 {
	if x != nil {
		return x.Serverid
	}
	return 0
}

type Tfence struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fenceid *Tfenceid `protobuf:"bytes,1,opt,name=fenceid,proto3" json:"fenceid,omitempty"`
	Epoch   uint64    `protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (x *Tfence) Reset() {
	*x = Tfence{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tfence) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tfence) ProtoMessage() {}

func (x *Tfence) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tfence.ProtoReflect.Descriptor instead.
func (*Tfence) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{5}
}

func (x *Tfence) GetFenceid() *Tfenceid {
	if x != nil {
		return x.Fenceid
	}
	return nil
}

func (x *Tfence) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

type Fcall struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type     uint32     `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Tag      uint32     `protobuf:"varint,2,opt,name=tag,proto3" json:"tag,omitempty"`
	Client   uint64     `protobuf:"varint,3,opt,name=client,proto3" json:"client,omitempty"`
	Session  uint64     `protobuf:"varint,4,opt,name=session,proto3" json:"session,omitempty"`
	Seqno    uint64     `protobuf:"varint,5,opt,name=seqno,proto3" json:"seqno,omitempty"`
	Received *Tinterval `protobuf:"bytes,6,opt,name=received,proto3" json:"received,omitempty"`
	Fence    *Tfence    `protobuf:"bytes,7,opt,name=fence,proto3" json:"fence,omitempty"`
}

func (x *Fcall) Reset() {
	*x = Fcall{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Fcall) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fcall) ProtoMessage() {}

func (x *Fcall) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fcall.ProtoReflect.Descriptor instead.
func (*Fcall) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{6}
}

func (x *Fcall) GetType() uint32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *Fcall) GetTag() uint32 {
	if x != nil {
		return x.Tag
	}
	return 0
}

func (x *Fcall) GetClient() uint64 {
	if x != nil {
		return x.Client
	}
	return 0
}

func (x *Fcall) GetSession() uint64 {
	if x != nil {
		return x.Session
	}
	return 0
}

func (x *Fcall) GetSeqno() uint64 {
	if x != nil {
		return x.Seqno
	}
	return 0
}

func (x *Fcall) GetReceived() *Tinterval {
	if x != nil {
		return x.Received
	}
	return nil
}

func (x *Fcall) GetFence() *Tfence {
	if x != nil {
		return x.Fence
	}
	return nil
}

type Twalk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fid    uint32   `protobuf:"varint,1,opt,name=fid,proto3" json:"fid,omitempty"`
	NewFid uint32   `protobuf:"varint,2,opt,name=newFid,proto3" json:"newFid,omitempty"`
	Wnames []string `protobuf:"bytes,3,rep,name=wnames,proto3" json:"wnames,omitempty"`
}

func (x *Twalk) Reset() {
	*x = Twalk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Twalk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Twalk) ProtoMessage() {}

func (x *Twalk) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Twalk.ProtoReflect.Descriptor instead.
func (*Twalk) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{7}
}

func (x *Twalk) GetFid() uint32 {
	if x != nil {
		return x.Fid
	}
	return 0
}

func (x *Twalk) GetNewFid() uint32 {
	if x != nil {
		return x.NewFid
	}
	return 0
}

func (x *Twalk) GetWnames() []string {
	if x != nil {
		return x.Wnames
	}
	return nil
}

type Rwalk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Qids []*Tqid `protobuf:"bytes,1,rep,name=qids,proto3" json:"qids,omitempty"`
}

func (x *Rwalk) Reset() {
	*x = Rwalk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Rwalk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Rwalk) ProtoMessage() {}

func (x *Rwalk) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Rwalk.ProtoReflect.Descriptor instead.
func (*Rwalk) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{8}
}

func (x *Rwalk) GetQids() []*Tqid {
	if x != nil {
		return x.Qids
	}
	return nil
}

type Tstat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fid uint32 `protobuf:"varint,1,opt,name=fid,proto3" json:"fid,omitempty"`
}

func (x *Tstat) Reset() {
	*x = Tstat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tstat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tstat) ProtoMessage() {}

func (x *Tstat) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tstat.ProtoReflect.Descriptor instead.
func (*Tstat) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{9}
}

func (x *Tstat) GetFid() uint32 {
	if x != nil {
		return x.Fid
	}
	return 0
}

type Rstat struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Size uint32 `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	Stat *Stat  `protobuf:"bytes,2,opt,name=stat,proto3" json:"stat,omitempty"`
}

func (x *Rstat) Reset() {
	*x = Rstat{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Rstat) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Rstat) ProtoMessage() {}

func (x *Rstat) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Rstat.ProtoReflect.Descriptor instead.
func (*Rstat) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{10}
}

func (x *Rstat) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *Rstat) GetStat() *Stat {
	if x != nil {
		return x.Stat
	}
	return nil
}

type TreadV struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fid     uint32 `protobuf:"varint,1,opt,name=fid,proto3" json:"fid,omitempty"`
	Offset  uint64 `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Count   uint32 `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Version uint32 `protobuf:"varint,4,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *TreadV) Reset() {
	*x = TreadV{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TreadV) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TreadV) ProtoMessage() {}

func (x *TreadV) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TreadV.ProtoReflect.Descriptor instead.
func (*TreadV) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{11}
}

func (x *TreadV) GetFid() uint32 {
	if x != nil {
		return x.Fid
	}
	return 0
}

func (x *TreadV) GetOffset() uint64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *TreadV) GetCount() uint32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *TreadV) GetVersion() uint32 {
	if x != nil {
		return x.Version
	}
	return 0
}

type Twriteread struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fid  uint64 `protobuf:"varint,1,opt,name=fid,proto3" json:"fid,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Twriteread) Reset() {
	*x = Twriteread{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Twriteread) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Twriteread) ProtoMessage() {}

func (x *Twriteread) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Twriteread.ProtoReflect.Descriptor instead.
func (*Twriteread) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{12}
}

func (x *Twriteread) GetFid() uint64 {
	if x != nil {
		return x.Fid
	}
	return 0
}

func (x *Twriteread) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Rwriteread struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Rwriteread) Reset() {
	*x = Rwriteread{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sigmap_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Rwriteread) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Rwriteread) ProtoMessage() {}

func (x *Rwriteread) ProtoReflect() protoreflect.Message {
	mi := &file_sigmap_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Rwriteread.ProtoReflect.Descriptor instead.
func (*Rwriteread) Descriptor() ([]byte, []int) {
	return file_sigmap_proto_rawDescGZIP(), []int{13}
}

func (x *Rwriteread) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_sigmap_proto protoreflect.FileDescriptor

var file_sigmap_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x48,
	0x0a, 0x04, 0x54, 0x71, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0xe9, 0x01, 0x0a, 0x04, 0x53, 0x74, 0x61,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x65, 0x76, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x03, 0x64, 0x65, 0x76, 0x12, 0x17, 0x0a, 0x03, 0x71, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x54, 0x71, 0x69, 0x64, 0x52, 0x03, 0x71, 0x69, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04,
	0x6d, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x05, 0x61, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x6d, 0x74, 0x69, 0x6d, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x75, 0x69, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x67, 0x69, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6d, 0x75, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6d, 0x75, 0x69, 0x64, 0x22, 0x22, 0x0a, 0x07, 0x52, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x12,
	0x17, 0x0a, 0x03, 0x71, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x05, 0x2e, 0x54,
	0x71, 0x69, 0x64, 0x52, 0x03, 0x71, 0x69, 0x64, 0x22, 0x33, 0x0a, 0x09, 0x54, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x65,
	0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x22, 0x3a, 0x0a,
	0x08, 0x54, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1a, 0x0a,
	0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x69, 0x64, 0x22, 0x43, 0x0a, 0x06, 0x54, 0x66, 0x65,
	0x6e, 0x63, 0x65, 0x12, 0x23, 0x0a, 0x07, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x54, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x69, 0x64, 0x52,
	0x07, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x22, 0xbc,
	0x01, 0x0a, 0x05, 0x46, 0x63, 0x61, 0x6c, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x74, 0x61, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x74, 0x61, 0x67, 0x12, 0x16,
	0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x14, 0x0a, 0x05, 0x73, 0x65, 0x71, 0x6e, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x05, 0x73, 0x65, 0x71, 0x6e, 0x6f, 0x12, 0x26, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76,
	0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x54, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x76, 0x61, 0x6c, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x64, 0x12, 0x1d,
	0x0a, 0x05, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e,
	0x54, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x05, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x49, 0x0a,
	0x05, 0x54, 0x77, 0x61, 0x6c, 0x6b, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x03, 0x66, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x65, 0x77, 0x46,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x6e, 0x65, 0x77, 0x46, 0x69, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x77, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x06, 0x77, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x22, 0x0a, 0x05, 0x52, 0x77, 0x61, 0x6c,
	0x6b, 0x12, 0x19, 0x0a, 0x04, 0x71, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x05, 0x2e, 0x54, 0x71, 0x69, 0x64, 0x52, 0x04, 0x71, 0x69, 0x64, 0x73, 0x22, 0x19, 0x0a, 0x05,
	0x54, 0x73, 0x74, 0x61, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x03, 0x66, 0x69, 0x64, 0x22, 0x36, 0x0a, 0x05, 0x52, 0x73, 0x74, 0x61, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04,
	0x73, 0x69, 0x7a, 0x65, 0x12, 0x19, 0x0a, 0x04, 0x73, 0x74, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x05, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x52, 0x04, 0x73, 0x74, 0x61, 0x74, 0x22,
	0x62, 0x0a, 0x06, 0x54, 0x72, 0x65, 0x61, 0x64, 0x56, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x66, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6f,
	0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x6f, 0x66, 0x66,
	0x73, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x22, 0x32, 0x0a, 0x0a, 0x54, 0x77, 0x72, 0x69, 0x74, 0x65, 0x72, 0x65, 0x61,
	0x64, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03,
	0x66, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x20, 0x0a, 0x0a, 0x52, 0x77, 0x72, 0x69, 0x74,
	0x65, 0x72, 0x65, 0x61, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x10, 0x5a, 0x0e, 0x73, 0x69, 0x67,
	0x6d, 0x61, 0x6f, 0x73, 0x2f, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_sigmap_proto_rawDescOnce sync.Once
	file_sigmap_proto_rawDescData = file_sigmap_proto_rawDesc
)

func file_sigmap_proto_rawDescGZIP() []byte {
	file_sigmap_proto_rawDescOnce.Do(func() {
		file_sigmap_proto_rawDescData = protoimpl.X.CompressGZIP(file_sigmap_proto_rawDescData)
	})
	return file_sigmap_proto_rawDescData
}

var file_sigmap_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_sigmap_proto_goTypes = []interface{}{
	(*Tqid)(nil),       // 0: Tqid
	(*Stat)(nil),       // 1: Stat
	(*Rattach)(nil),    // 2: Rattach
	(*Tinterval)(nil),  // 3: Tinterval
	(*Tfenceid)(nil),   // 4: Tfenceid
	(*Tfence)(nil),     // 5: Tfence
	(*Fcall)(nil),      // 6: Fcall
	(*Twalk)(nil),      // 7: Twalk
	(*Rwalk)(nil),      // 8: Rwalk
	(*Tstat)(nil),      // 9: Tstat
	(*Rstat)(nil),      // 10: Rstat
	(*TreadV)(nil),     // 11: TreadV
	(*Twriteread)(nil), // 12: Twriteread
	(*Rwriteread)(nil), // 13: Rwriteread
}
var file_sigmap_proto_depIdxs = []int32{
	0, // 0: Stat.qid:type_name -> Tqid
	0, // 1: Rattach.qid:type_name -> Tqid
	4, // 2: Tfence.fenceid:type_name -> Tfenceid
	3, // 3: Fcall.received:type_name -> Tinterval
	5, // 4: Fcall.fence:type_name -> Tfence
	0, // 5: Rwalk.qids:type_name -> Tqid
	1, // 6: Rstat.stat:type_name -> Stat
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_sigmap_proto_init() }
func file_sigmap_proto_init() {
	if File_sigmap_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sigmap_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tqid); i {
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
		file_sigmap_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Stat); i {
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
		file_sigmap_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Rattach); i {
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
		file_sigmap_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tinterval); i {
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
		file_sigmap_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tfenceid); i {
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
		file_sigmap_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tfence); i {
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
		file_sigmap_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Fcall); i {
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
		file_sigmap_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Twalk); i {
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
		file_sigmap_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Rwalk); i {
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
		file_sigmap_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tstat); i {
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
		file_sigmap_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Rstat); i {
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
		file_sigmap_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TreadV); i {
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
		file_sigmap_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Twriteread); i {
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
		file_sigmap_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Rwriteread); i {
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
			RawDescriptor: file_sigmap_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_sigmap_proto_goTypes,
		DependencyIndexes: file_sigmap_proto_depIdxs,
		MessageInfos:      file_sigmap_proto_msgTypes,
	}.Build()
	File_sigmap_proto = out.File
	file_sigmap_proto_rawDesc = nil
	file_sigmap_proto_goTypes = nil
	file_sigmap_proto_depIdxs = nil
}
