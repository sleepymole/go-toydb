// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: storage/storagepb/storage.proto

package storagepb

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

// EngineStatus contains the current status of the storage engine.
type EngineStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name is the name of the engine.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Keys is the number of live keys in the engine.
	Keys int64 `protobuf:"varint,2,opt,name=keys,proto3" json:"keys,omitempty"`
	// Size is the logical size of live key/values pairs.
	Size int64 `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	// TotalDiskSize is the on-disk size of all data, live and garbage.
	TotalDiskSize int64 `protobuf:"varint,4,opt,name=total_disk_size,json=totalDiskSize,proto3" json:"total_disk_size,omitempty"`
	// LiveDiskSize is the on-disk size of live data.
	LiveDiskSize int64 `protobuf:"varint,5,opt,name=live_disk_size,json=liveDiskSize,proto3" json:"live_disk_size,omitempty"`
	// GarbageDiskSize is the on-disk size of garbage data.
	GarbageDiskSize int64 `protobuf:"varint,6,opt,name=garbage_disk_size,json=garbageDiskSize,proto3" json:"garbage_disk_size,omitempty"`
}

func (x *EngineStatus) Reset() {
	*x = EngineStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storagepb_storage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EngineStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EngineStatus) ProtoMessage() {}

func (x *EngineStatus) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storagepb_storage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EngineStatus.ProtoReflect.Descriptor instead.
func (*EngineStatus) Descriptor() ([]byte, []int) {
	return file_storage_storagepb_storage_proto_rawDescGZIP(), []int{0}
}

func (x *EngineStatus) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *EngineStatus) GetKeys() int64 {
	if x != nil {
		return x.Keys
	}
	return 0
}

func (x *EngineStatus) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *EngineStatus) GetTotalDiskSize() int64 {
	if x != nil {
		return x.TotalDiskSize
	}
	return 0
}

func (x *EngineStatus) GetLiveDiskSize() int64 {
	if x != nil {
		return x.LiveDiskSize
	}
	return 0
}

func (x *EngineStatus) GetGarbageDiskSize() int64 {
	if x != nil {
		return x.GarbageDiskSize
	}
	return 0
}

// MVCCStatus contains the current status of the MVCC engine.
type MVCCStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Versions is the number of MVCC versions (i.e. read-write transactions).
	Versions int64 `protobuf:"varint,1,opt,name=versions,proto3" json:"versions,omitempty"`
	// ActiveTxns is the number of active transactions.
	ActiveTxns int64 `protobuf:"varint,2,opt,name=active_txns,json=activeTxns,proto3" json:"active_txns,omitempty"`
	// Storage is the underlying storage engine status.
	Storage *EngineStatus `protobuf:"bytes,3,opt,name=storage,proto3" json:"storage,omitempty"`
}

func (x *MVCCStatus) Reset() {
	*x = MVCCStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storagepb_storage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MVCCStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MVCCStatus) ProtoMessage() {}

func (x *MVCCStatus) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storagepb_storage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MVCCStatus.ProtoReflect.Descriptor instead.
func (*MVCCStatus) Descriptor() ([]byte, []int) {
	return file_storage_storagepb_storage_proto_rawDescGZIP(), []int{1}
}

func (x *MVCCStatus) GetVersions() int64 {
	if x != nil {
		return x.Versions
	}
	return 0
}

func (x *MVCCStatus) GetActiveTxns() int64 {
	if x != nil {
		return x.ActiveTxns
	}
	return 0
}

func (x *MVCCStatus) GetStorage() *EngineStatus {
	if x != nil {
		return x.Storage
	}
	return nil
}

// MVCCTxnState is the MVCC transaction state, which determines its write version
// and isolation. It is separate from the transaction itself to allow it to be
// passed around independently of the engine.
type MVCCTxnState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Version is the version this transaction is running at. Only one read-write
	// transaction can run at a given version, since this identifies the its write.
	Version uint64 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	// ReadOnly is true if this transaction is read-only.
	ReadOnly bool `protobuf:"varint,2,opt,name=read_only,json=readOnly,proto3" json:"read_only,omitempty"`
	// ActiveTxns is the set of concurrent active (uncommitted) transactions, as of
	// the start of this transaction. Their writes should be invisible to this
	// transaction even if they're writting at a lower version, since they're not
	// committed yet.
	ActiveTxns []uint64 `protobuf:"varint,3,rep,packed,name=active_txns,json=activeTxns,proto3" json:"active_txns,omitempty"`
}

func (x *MVCCTxnState) Reset() {
	*x = MVCCTxnState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storagepb_storage_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MVCCTxnState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MVCCTxnState) ProtoMessage() {}

func (x *MVCCTxnState) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storagepb_storage_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MVCCTxnState.ProtoReflect.Descriptor instead.
func (*MVCCTxnState) Descriptor() ([]byte, []int) {
	return file_storage_storagepb_storage_proto_rawDescGZIP(), []int{2}
}

func (x *MVCCTxnState) GetVersion() uint64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *MVCCTxnState) GetReadOnly() bool {
	if x != nil {
		return x.ReadOnly
	}
	return false
}

func (x *MVCCTxnState) GetActiveTxns() []uint64 {
	if x != nil {
		return x.ActiveTxns
	}
	return nil
}

var File_storage_storagepb_storage_proto protoreflect.FileDescriptor

var file_storage_storagepb_storage_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x70, 0x62, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0c, 0x74, 0x6f, 0x79, 0x64, 0x62, 0x2e, 0x6d, 0x76, 0x63, 0x63, 0x70, 0x62, 0x22,
	0xc4, 0x01, 0x0a, 0x0c, 0x45, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x26, 0x0a, 0x0f,
	0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x44, 0x69, 0x73, 0x6b,
	0x53, 0x69, 0x7a, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x6c, 0x69, 0x76, 0x65, 0x5f, 0x64, 0x69, 0x73,
	0x6b, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x6c, 0x69,
	0x76, 0x65, 0x44, 0x69, 0x73, 0x6b, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x2a, 0x0a, 0x11, 0x67, 0x61,
	0x72, 0x62, 0x61, 0x67, 0x65, 0x5f, 0x64, 0x69, 0x73, 0x6b, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x67, 0x61, 0x72, 0x62, 0x61, 0x67, 0x65, 0x44, 0x69,
	0x73, 0x6b, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x7f, 0x0a, 0x0a, 0x4d, 0x56, 0x43, 0x43, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x74, 0x78, 0x6e, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x54, 0x78, 0x6e,
	0x73, 0x12, 0x34, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x74, 0x6f, 0x79, 0x64, 0x62, 0x2e, 0x6d, 0x76, 0x63, 0x63, 0x70,
	0x62, 0x2e, 0x45, 0x6e, 0x67, 0x69, 0x6e, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x07,
	0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x22, 0x66, 0x0a, 0x0c, 0x4d, 0x56, 0x43, 0x43, 0x54,
	0x78, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x72, 0x65, 0x61, 0x64, 0x5f, 0x6f, 0x6e, 0x6c, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x61, 0x64, 0x4f, 0x6e, 0x6c, 0x79, 0x12, 0x1f,
	0x0a, 0x0b, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x74, 0x78, 0x6e, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x04, 0x52, 0x0a, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x54, 0x78, 0x6e, 0x73, 0x42,
	0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6c,
	0x65, 0x65, 0x70, 0x79, 0x6d, 0x6f, 0x6c, 0x65, 0x2f, 0x67, 0x6f, 0x2d, 0x74, 0x6f, 0x79, 0x64,
	0x62, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_storage_storagepb_storage_proto_rawDescOnce sync.Once
	file_storage_storagepb_storage_proto_rawDescData = file_storage_storagepb_storage_proto_rawDesc
)

func file_storage_storagepb_storage_proto_rawDescGZIP() []byte {
	file_storage_storagepb_storage_proto_rawDescOnce.Do(func() {
		file_storage_storagepb_storage_proto_rawDescData = protoimpl.X.CompressGZIP(file_storage_storagepb_storage_proto_rawDescData)
	})
	return file_storage_storagepb_storage_proto_rawDescData
}

var file_storage_storagepb_storage_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_storage_storagepb_storage_proto_goTypes = []interface{}{
	(*EngineStatus)(nil), // 0: toydb.mvccpb.EngineStatus
	(*MVCCStatus)(nil),   // 1: toydb.mvccpb.MVCCStatus
	(*MVCCTxnState)(nil), // 2: toydb.mvccpb.MVCCTxnState
}
var file_storage_storagepb_storage_proto_depIdxs = []int32{
	0, // 0: toydb.mvccpb.MVCCStatus.storage:type_name -> toydb.mvccpb.EngineStatus
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_storage_storagepb_storage_proto_init() }
func file_storage_storagepb_storage_proto_init() {
	if File_storage_storagepb_storage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_storage_storagepb_storage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EngineStatus); i {
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
		file_storage_storagepb_storage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MVCCStatus); i {
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
		file_storage_storagepb_storage_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MVCCTxnState); i {
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
			RawDescriptor: file_storage_storagepb_storage_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_storage_storagepb_storage_proto_goTypes,
		DependencyIndexes: file_storage_storagepb_storage_proto_depIdxs,
		MessageInfos:      file_storage_storagepb_storage_proto_msgTypes,
	}.Build()
	File_storage_storagepb_storage_proto = out.File
	file_storage_storagepb_storage_proto_rawDesc = nil
	file_storage_storagepb_storage_proto_goTypes = nil
	file_storage_storagepb_storage_proto_depIdxs = nil
}