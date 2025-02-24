// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        (unknown)
// source: teleport/devicetrust/v1/device_source.proto

package devicetrustv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Origin of a device.
type DeviceOrigin int32

const (
	// Unspecified or absent origin.
	DeviceOrigin_DEVICE_ORIGIN_UNSPECIFIED DeviceOrigin = 0
	// Devices originated from direct API usage.
	DeviceOrigin_DEVICE_ORIGIN_API DeviceOrigin = 1
	// Devices originated from Jamf sync.
	DeviceOrigin_DEVICE_ORIGIN_JAMF DeviceOrigin = 2
	// Source originated from Microsoft Intune sync.
	DeviceOrigin_DEVICE_ORIGIN_INTUNE DeviceOrigin = 3
)

// Enum value maps for DeviceOrigin.
var (
	DeviceOrigin_name = map[int32]string{
		0: "DEVICE_ORIGIN_UNSPECIFIED",
		1: "DEVICE_ORIGIN_API",
		2: "DEVICE_ORIGIN_JAMF",
		3: "DEVICE_ORIGIN_INTUNE",
	}
	DeviceOrigin_value = map[string]int32{
		"DEVICE_ORIGIN_UNSPECIFIED": 0,
		"DEVICE_ORIGIN_API":         1,
		"DEVICE_ORIGIN_JAMF":        2,
		"DEVICE_ORIGIN_INTUNE":      3,
	}
)

func (x DeviceOrigin) Enum() *DeviceOrigin {
	p := new(DeviceOrigin)
	*p = x
	return p
}

func (x DeviceOrigin) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeviceOrigin) Descriptor() protoreflect.EnumDescriptor {
	return file_teleport_devicetrust_v1_device_source_proto_enumTypes[0].Descriptor()
}

func (DeviceOrigin) Type() protoreflect.EnumType {
	return &file_teleport_devicetrust_v1_device_source_proto_enumTypes[0]
}

func (x DeviceOrigin) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeviceOrigin.Descriptor instead.
func (DeviceOrigin) EnumDescriptor() ([]byte, []int) {
	return file_teleport_devicetrust_v1_device_source_proto_rawDescGZIP(), []int{0}
}

// Source of device, for devices that are managed by external systems
// (for example, MDMs).
type DeviceSource struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Name of the source.
	// Matches the name of the corresponding MDM service, if applicable.
	// Readonly.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Origin of the source.
	// Readonly.
	Origin        DeviceOrigin `protobuf:"varint,2,opt,name=origin,proto3,enum=teleport.devicetrust.v1.DeviceOrigin" json:"origin,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeviceSource) Reset() {
	*x = DeviceSource{}
	mi := &file_teleport_devicetrust_v1_device_source_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeviceSource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceSource) ProtoMessage() {}

func (x *DeviceSource) ProtoReflect() protoreflect.Message {
	mi := &file_teleport_devicetrust_v1_device_source_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceSource.ProtoReflect.Descriptor instead.
func (*DeviceSource) Descriptor() ([]byte, []int) {
	return file_teleport_devicetrust_v1_device_source_proto_rawDescGZIP(), []int{0}
}

func (x *DeviceSource) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DeviceSource) GetOrigin() DeviceOrigin {
	if x != nil {
		return x.Origin
	}
	return DeviceOrigin_DEVICE_ORIGIN_UNSPECIFIED
}

var File_teleport_devicetrust_v1_device_source_proto protoreflect.FileDescriptor

var file_teleport_devicetrust_v1_device_source_proto_rawDesc = string([]byte{
	0x0a, 0x2b, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x74, 0x72, 0x75, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x74,
	0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x74, 0x72,
	0x75, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x22, 0x61, 0x0a, 0x0c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3d, 0x0a, 0x06, 0x6f, 0x72,
	0x69, 0x67, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x74, 0x65, 0x6c,
	0x65, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x74, 0x72, 0x75, 0x73,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x72, 0x69, 0x67, 0x69,
	0x6e, 0x52, 0x06, 0x6f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x2a, 0x76, 0x0a, 0x0c, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x4f, 0x72, 0x69, 0x67, 0x69, 0x6e, 0x12, 0x1d, 0x0a, 0x19, 0x44, 0x45, 0x56,
	0x49, 0x43, 0x45, 0x5f, 0x4f, 0x52, 0x49, 0x47, 0x49, 0x4e, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45,
	0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x44, 0x45, 0x56, 0x49,
	0x43, 0x45, 0x5f, 0x4f, 0x52, 0x49, 0x47, 0x49, 0x4e, 0x5f, 0x41, 0x50, 0x49, 0x10, 0x01, 0x12,
	0x16, 0x0a, 0x12, 0x44, 0x45, 0x56, 0x49, 0x43, 0x45, 0x5f, 0x4f, 0x52, 0x49, 0x47, 0x49, 0x4e,
	0x5f, 0x4a, 0x41, 0x4d, 0x46, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x44, 0x45, 0x56, 0x49, 0x43,
	0x45, 0x5f, 0x4f, 0x52, 0x49, 0x47, 0x49, 0x4e, 0x5f, 0x49, 0x4e, 0x54, 0x55, 0x4e, 0x45, 0x10,
	0x03, 0x42, 0x5a, 0x5a, 0x58, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x67, 0x72, 0x61, 0x76, 0x69, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x2f, 0x74, 0x65,
	0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x74, 0x72, 0x75, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x3b,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x74, 0x72, 0x75, 0x73, 0x74, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_teleport_devicetrust_v1_device_source_proto_rawDescOnce sync.Once
	file_teleport_devicetrust_v1_device_source_proto_rawDescData []byte
)

func file_teleport_devicetrust_v1_device_source_proto_rawDescGZIP() []byte {
	file_teleport_devicetrust_v1_device_source_proto_rawDescOnce.Do(func() {
		file_teleport_devicetrust_v1_device_source_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_teleport_devicetrust_v1_device_source_proto_rawDesc), len(file_teleport_devicetrust_v1_device_source_proto_rawDesc)))
	})
	return file_teleport_devicetrust_v1_device_source_proto_rawDescData
}

var file_teleport_devicetrust_v1_device_source_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_teleport_devicetrust_v1_device_source_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_teleport_devicetrust_v1_device_source_proto_goTypes = []any{
	(DeviceOrigin)(0),    // 0: teleport.devicetrust.v1.DeviceOrigin
	(*DeviceSource)(nil), // 1: teleport.devicetrust.v1.DeviceSource
}
var file_teleport_devicetrust_v1_device_source_proto_depIdxs = []int32{
	0, // 0: teleport.devicetrust.v1.DeviceSource.origin:type_name -> teleport.devicetrust.v1.DeviceOrigin
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_teleport_devicetrust_v1_device_source_proto_init() }
func file_teleport_devicetrust_v1_device_source_proto_init() {
	if File_teleport_devicetrust_v1_device_source_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_teleport_devicetrust_v1_device_source_proto_rawDesc), len(file_teleport_devicetrust_v1_device_source_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_teleport_devicetrust_v1_device_source_proto_goTypes,
		DependencyIndexes: file_teleport_devicetrust_v1_device_source_proto_depIdxs,
		EnumInfos:         file_teleport_devicetrust_v1_device_source_proto_enumTypes,
		MessageInfos:      file_teleport_devicetrust_v1_device_source_proto_msgTypes,
	}.Build()
	File_teleport_devicetrust_v1_device_source_proto = out.File
	file_teleport_devicetrust_v1_device_source_proto_goTypes = nil
	file_teleport_devicetrust_v1_device_source_proto_depIdxs = nil
}
