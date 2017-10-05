// Code generated by protoc-gen-gogo.
// source: cockroach/pkg/ccl/utilccl/licenseccl/license.proto
// DO NOT EDIT!

/*
	Package licenseccl is a generated protocol buffer package.

	It is generated from these files:
		cockroach/pkg/ccl/utilccl/licenseccl/license.proto

	It has these top-level messages:
		License
*/
package licenseccl

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import github_com_cockroachdb_cockroach_pkg_util_uuid "github.com/cockroachdb/cockroach/pkg/util/uuid"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type License_Type int32

const (
	License_NonCommercial License_Type = 0
	License_Enterprise    License_Type = 1
	License_Evaluation    License_Type = 2
)

var License_Type_name = map[int32]string{
	0: "NonCommercial",
	1: "Enterprise",
	2: "Evaluation",
}
var License_Type_value = map[string]int32{
	"NonCommercial": 0,
	"Enterprise":    1,
	"Evaluation":    2,
}

func (x License_Type) String() string {
	return proto.EnumName(License_Type_name, int32(x))
}
func (License_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorLicense, []int{0, 0} }

type License struct {
	ClusterID         []github_com_cockroachdb_cockroach_pkg_util_uuid.UUID `protobuf:"bytes,1,rep,name=cluster_id,json=clusterId,customtype=github.com/cockroachdb/cockroach/pkg/util/uuid.UUID" json:"cluster_id"`
	ValidUntilUnixSec int64                                                 `protobuf:"varint,2,opt,name=valid_until_unix_sec,json=validUntilUnixSec,proto3" json:"valid_until_unix_sec,omitempty"`
	Type              License_Type                                          `protobuf:"varint,3,opt,name=type,proto3,enum=cockroach.ccl.utilccl.licenseccl.License_Type" json:"type,omitempty"`
	OrganizationName  string                                                `protobuf:"bytes,4,opt,name=organization_name,json=organizationName,proto3" json:"organization_name,omitempty"`
}

func (m *License) Reset()                    { *m = License{} }
func (m *License) String() string            { return proto.CompactTextString(m) }
func (*License) ProtoMessage()               {}
func (*License) Descriptor() ([]byte, []int) { return fileDescriptorLicense, []int{0} }

func init() {
	proto.RegisterType((*License)(nil), "cockroach.ccl.utilccl.licenseccl.License")
	proto.RegisterEnum("cockroach.ccl.utilccl.licenseccl.License_Type", License_Type_name, License_Type_value)
}
func (m *License) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *License) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ClusterID) > 0 {
		for _, msg := range m.ClusterID {
			dAtA[i] = 0xa
			i++
			i = encodeVarintLicense(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.ValidUntilUnixSec != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintLicense(dAtA, i, uint64(m.ValidUntilUnixSec))
	}
	if m.Type != 0 {
		dAtA[i] = 0x18
		i++
		i = encodeVarintLicense(dAtA, i, uint64(m.Type))
	}
	if len(m.OrganizationName) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintLicense(dAtA, i, uint64(len(m.OrganizationName)))
		i += copy(dAtA[i:], m.OrganizationName)
	}
	return i, nil
}

func encodeFixed64License(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32License(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintLicense(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *License) Size() (n int) {
	var l int
	_ = l
	if len(m.ClusterID) > 0 {
		for _, e := range m.ClusterID {
			l = e.Size()
			n += 1 + l + sovLicense(uint64(l))
		}
	}
	if m.ValidUntilUnixSec != 0 {
		n += 1 + sovLicense(uint64(m.ValidUntilUnixSec))
	}
	if m.Type != 0 {
		n += 1 + sovLicense(uint64(m.Type))
	}
	l = len(m.OrganizationName)
	if l > 0 {
		n += 1 + l + sovLicense(uint64(l))
	}
	return n
}

func sovLicense(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozLicense(x uint64) (n int) {
	return sovLicense(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *License) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLicense
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: License: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: License: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClusterID", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthLicense
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v github_com_cockroachdb_cockroach_pkg_util_uuid.UUID
			m.ClusterID = append(m.ClusterID, v)
			if err := m.ClusterID[len(m.ClusterID)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidUntilUnixSec", wireType)
			}
			m.ValidUntilUnixSec = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ValidUntilUnixSec |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= (License_Type(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrganizationName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthLicense
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OrganizationName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLicense(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLicense
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLicense(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLicense
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLicense
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthLicense
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowLicense
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipLicense(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthLicense = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLicense   = fmt.Errorf("proto: integer overflow")
)

func init() {
	proto.RegisterFile("cockroach/pkg/ccl/utilccl/licenseccl/license.proto", fileDescriptorLicense)
}

var fileDescriptorLicense = []byte{
	// 360 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0x41, 0x8b, 0xda, 0x40,
	0x14, 0xc7, 0x33, 0x2a, 0x2d, 0x0e, 0xad, 0xc4, 0xe0, 0x21, 0xf4, 0x10, 0x83, 0xf4, 0x10, 0x28,
	0xcc, 0x80, 0x9e, 0x7a, 0x55, 0x0b, 0x15, 0x8a, 0x87, 0xb4, 0xb9, 0xf4, 0x12, 0xc6, 0xc9, 0x10,
	0x07, 0x27, 0x33, 0x21, 0x99, 0x11, 0xed, 0xa7, 0xe8, 0xc7, 0xf2, 0xb6, 0x7b, 0x5c, 0xf6, 0x20,
	0xbb, 0xd9, 0x2f, 0xb2, 0x24, 0x8a, 0xba, 0xa7, 0x3d, 0xcd, 0x1b, 0xfe, 0xff, 0xdf, 0x7b, 0xef,
	0xff, 0xe0, 0x98, 0x2a, 0xba, 0x29, 0x14, 0xa1, 0x6b, 0x9c, 0x6f, 0x52, 0x4c, 0xa9, 0xc0, 0x46,
	0x73, 0x51, 0xbf, 0x82, 0x53, 0x26, 0x4b, 0x76, 0x53, 0xa2, 0xbc, 0x50, 0x5a, 0x39, 0xfe, 0x85,
	0x41, 0x94, 0x0a, 0x74, 0xf6, 0xa3, 0xab, 0xff, 0xcb, 0x20, 0x55, 0xa9, 0x6a, 0xcc, 0xb8, 0xae,
	0x4e, 0xdc, 0xe8, 0xae, 0x05, 0x3f, 0xfe, 0x3a, 0x99, 0x9c, 0x14, 0x42, 0x2a, 0x4c, 0xa9, 0x59,
	0x11, 0xf3, 0xc4, 0x05, 0x7e, 0x3b, 0xf8, 0x34, 0xfd, 0x79, 0x38, 0x0e, 0xad, 0xc7, 0xe3, 0x70,
	0x92, 0x72, 0xbd, 0x36, 0x2b, 0x44, 0x55, 0x86, 0x2f, 0xa3, 0x92, 0x15, 0x7e, 0xbb, 0x6a, 0x3d,
	0x16, 0x1b, 0xc3, 0x13, 0x14, 0x45, 0x8b, 0x79, 0x75, 0x1c, 0x76, 0x67, 0xa7, 0x86, 0x8b, 0x79,
	0xd8, 0x3d, 0xf7, 0x5e, 0x24, 0x0e, 0x86, 0x83, 0x2d, 0x11, 0x3c, 0x89, 0x8d, 0xd4, 0x5c, 0xc4,
	0x46, 0xf2, 0x5d, 0x5c, 0x32, 0xea, 0xb6, 0x7c, 0x10, 0xb4, 0xc3, 0x7e, 0xa3, 0x45, 0xb5, 0x14,
	0x49, 0xbe, 0xfb, 0xcd, 0xa8, 0x33, 0x85, 0x1d, 0xbd, 0xcf, 0x99, 0xdb, 0xf6, 0x41, 0xd0, 0x1b,
	0x23, 0xf4, 0x5e, 0x58, 0x74, 0x8e, 0x84, 0xfe, 0xec, 0x73, 0x16, 0x36, 0xac, 0xf3, 0x0d, 0xf6,
	0x55, 0x91, 0x12, 0xc9, 0xff, 0x11, 0xcd, 0x95, 0x8c, 0x25, 0xc9, 0x98, 0xdb, 0xf1, 0x41, 0xd0,
	0x0d, 0xed, 0x5b, 0x61, 0x49, 0x32, 0x36, 0xfa, 0x0e, 0x3b, 0x35, 0xea, 0xf4, 0xe1, 0xe7, 0xa5,
	0x92, 0x33, 0x95, 0x65, 0xac, 0xa0, 0x9c, 0x08, 0xdb, 0x72, 0x7a, 0x10, 0xfe, 0x90, 0x9a, 0x15,
	0x79, 0xc1, 0x4b, 0x66, 0x83, 0xe6, 0xbf, 0x25, 0xc2, 0x34, 0xb0, 0xdd, 0x9a, 0x7e, 0x3d, 0x3c,
	0x7b, 0xd6, 0xa1, 0xf2, 0xc0, 0x7d, 0xe5, 0x81, 0x87, 0xca, 0x03, 0x4f, 0x95, 0x07, 0xfe, 0xbf,
	0x78, 0xd6, 0x5f, 0x78, 0x5d, 0x70, 0xf5, 0xa1, 0x39, 0xff, 0xe4, 0x35, 0x00, 0x00, 0xff, 0xff,
	0xf9, 0x65, 0xe1, 0x02, 0xec, 0x01, 0x00, 0x00,
}
