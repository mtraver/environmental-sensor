// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.13.0
// source: configpb/config.proto

package configpb

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

type Job_Operation int32

const (
	Job_INVALID  Job_Operation = 0
	Job_SETUP    Job_Operation = 1
	Job_SENSE    Job_Operation = 2
	Job_SHUTDOWN Job_Operation = 3
)

// Enum value maps for Job_Operation.
var (
	Job_Operation_name = map[int32]string{
		0: "INVALID",
		1: "SETUP",
		2: "SENSE",
		3: "SHUTDOWN",
	}
	Job_Operation_value = map[string]int32{
		"INVALID":  0,
		"SETUP":    1,
		"SENSE":    2,
		"SHUTDOWN": 3,
	}
)

func (x Job_Operation) Enum() *Job_Operation {
	p := new(Job_Operation)
	*p = x
	return p
}

func (x Job_Operation) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Job_Operation) Descriptor() protoreflect.EnumDescriptor {
	return file_configpb_config_proto_enumTypes[0].Descriptor()
}

func (Job_Operation) Type() protoreflect.EnumType {
	return &file_configpb_config_proto_enumTypes[0]
}

func (x Job_Operation) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Job_Operation.Descriptor instead.
func (Job_Operation) EnumDescriptor() ([]byte, []int) {
	return file_configpb_config_proto_rawDescGZIP(), []int{1, 0}
}

// Config configures the iotcorelogger program.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Path to a file containing a JSON-encoded Device struct.
	// See github.com/mtraver/iotcore.
	DeviceFilePath string `protobuf:"bytes,1,opt,name=device_file_path,json=deviceFilePath,proto3" json:"device_file_path,omitempty"`
	// Path to a set of trustworthy CA certs.
	// Download Google's from https://pki.google.com/roots.pem.
	CaCertsPath      string   `protobuf:"bytes,2,opt,name=ca_certs_path,json=caCertsPath,proto3" json:"ca_certs_path,omitempty"`
	SupportedSensors []string `protobuf:"bytes,3,rep,name=supported_sensors,json=supportedSensors,proto3" json:"supported_sensors,omitempty"`
	Jobs             []*Job   `protobuf:"bytes,4,rep,name=jobs,proto3" json:"jobs,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configpb_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_configpb_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_configpb_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetDeviceFilePath() string {
	if x != nil {
		return x.DeviceFilePath
	}
	return ""
}

func (x *Config) GetCaCertsPath() string {
	if x != nil {
		return x.CaCertsPath
	}
	return ""
}

func (x *Config) GetSupportedSensors() []string {
	if x != nil {
		return x.SupportedSensors
	}
	return nil
}

func (x *Config) GetJobs() []*Job {
	if x != nil {
		return x.Jobs
	}
	return nil
}

type Job struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Cron spec that specifies when this job should run.
	Cronspec  string        `protobuf:"bytes,1,opt,name=cronspec,proto3" json:"cronspec,omitempty"`
	Operation Job_Operation `protobuf:"varint,2,opt,name=operation,proto3,enum=config.Job_Operation" json:"operation,omitempty"`
	// Sensors are processed in the order given.
	Sensors []string `protobuf:"bytes,3,rep,name=sensors,proto3" json:"sensors,omitempty"`
}

func (x *Job) Reset() {
	*x = Job{}
	if protoimpl.UnsafeEnabled {
		mi := &file_configpb_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Job) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Job) ProtoMessage() {}

func (x *Job) ProtoReflect() protoreflect.Message {
	mi := &file_configpb_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Job.ProtoReflect.Descriptor instead.
func (*Job) Descriptor() ([]byte, []int) {
	return file_configpb_config_proto_rawDescGZIP(), []int{1}
}

func (x *Job) GetCronspec() string {
	if x != nil {
		return x.Cronspec
	}
	return ""
}

func (x *Job) GetOperation() Job_Operation {
	if x != nil {
		return x.Operation
	}
	return Job_INVALID
}

func (x *Job) GetSensors() []string {
	if x != nil {
		return x.Sensors
	}
	return nil
}

var File_configpb_config_proto protoreflect.FileDescriptor

var file_configpb_config_proto_rawDesc = []byte{
	0x0a, 0x15, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22,
	0xa4, 0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x28, 0x0a, 0x10, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x46, 0x69, 0x6c, 0x65,
	0x50, 0x61, 0x74, 0x68, 0x12, 0x22, 0x0a, 0x0d, 0x63, 0x61, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x73,
	0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x61, 0x43,
	0x65, 0x72, 0x74, 0x73, 0x50, 0x61, 0x74, 0x68, 0x12, 0x2b, 0x0a, 0x11, 0x73, 0x75, 0x70, 0x70,
	0x6f, 0x72, 0x74, 0x65, 0x64, 0x5f, 0x73, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x10, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x53, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x73, 0x12, 0x1f, 0x0a, 0x04, 0x6a, 0x6f, 0x62, 0x73, 0x18, 0x04, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4a, 0x6f, 0x62,
	0x52, 0x04, 0x6a, 0x6f, 0x62, 0x73, 0x22, 0xae, 0x01, 0x0a, 0x03, 0x4a, 0x6f, 0x62, 0x12, 0x1a,
	0x0a, 0x08, 0x63, 0x72, 0x6f, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x63, 0x72, 0x6f, 0x6e, 0x73, 0x70, 0x65, 0x63, 0x12, 0x33, 0x0a, 0x09, 0x6f, 0x70,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4a, 0x6f, 0x62, 0x2e, 0x4f, 0x70, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x07, 0x73, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x73, 0x22, 0x3c, 0x0a, 0x09, 0x4f, 0x70, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49,
	0x44, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x53, 0x45, 0x54, 0x55, 0x50, 0x10, 0x01, 0x12, 0x09,
	0x0a, 0x05, 0x53, 0x45, 0x4e, 0x53, 0x45, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x48, 0x55,
	0x54, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x03, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x74, 0x72, 0x61, 0x76, 0x65, 0x72, 0x2f, 0x65, 0x6e,
	0x76, 0x69, 0x72, 0x6f, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x6c, 0x2d, 0x73, 0x65, 0x6e, 0x73,
	0x6f, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_configpb_config_proto_rawDescOnce sync.Once
	file_configpb_config_proto_rawDescData = file_configpb_config_proto_rawDesc
)

func file_configpb_config_proto_rawDescGZIP() []byte {
	file_configpb_config_proto_rawDescOnce.Do(func() {
		file_configpb_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_configpb_config_proto_rawDescData)
	})
	return file_configpb_config_proto_rawDescData
}

var file_configpb_config_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_configpb_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_configpb_config_proto_goTypes = []interface{}{
	(Job_Operation)(0), // 0: config.Job.Operation
	(*Config)(nil),     // 1: config.Config
	(*Job)(nil),        // 2: config.Job
}
var file_configpb_config_proto_depIdxs = []int32{
	2, // 0: config.Config.jobs:type_name -> config.Job
	0, // 1: config.Job.operation:type_name -> config.Job.Operation
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_configpb_config_proto_init() }
func file_configpb_config_proto_init() {
	if File_configpb_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_configpb_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_configpb_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Job); i {
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
			RawDescriptor: file_configpb_config_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_configpb_config_proto_goTypes,
		DependencyIndexes: file_configpb_config_proto_depIdxs,
		EnumInfos:         file_configpb_config_proto_enumTypes,
		MessageInfos:      file_configpb_config_proto_msgTypes,
	}.Build()
	File_configpb_config_proto = out.File
	file_configpb_config_proto_rawDesc = nil
	file_configpb_config_proto_goTypes = nil
	file_configpb_config_proto_depIdxs = nil
}
