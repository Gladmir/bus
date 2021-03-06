// Code generated by protoc-gen-go.
// source: test.proto
// DO NOT EDIT!

/*
Package bus is a generated protocol buffer package.

It is generated from these files:
	test.proto

It has these top-level messages:
	TestFrame
*/
package bus

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TestFrame_EventType int32

const (
	TestFrame_PING TestFrame_EventType = 0
	TestFrame_PONG TestFrame_EventType = 1
)

var TestFrame_EventType_name = map[int32]string{
	0: "PING",
	1: "PONG",
}
var TestFrame_EventType_value = map[string]int32{
	"PING": 0,
	"PONG": 1,
}

func (x TestFrame_EventType) String() string {
	return proto.EnumName(TestFrame_EventType_name, int32(x))
}

type TestFrame struct {
	EventType TestFrame_EventType `protobuf:"varint,1,opt,name=eventType,enum=bus.TestFrame_EventType" json:"eventType,omitempty"`
	Ping      *TestFrame_Ping     `protobuf:"bytes,2,opt,name=ping" json:"ping,omitempty"`
	Pong      *TestFrame_Pong     `protobuf:"bytes,3,opt,name=pong" json:"pong,omitempty"`
}

func (m *TestFrame) Reset()         { *m = TestFrame{} }
func (m *TestFrame) String() string { return proto.CompactTextString(m) }
func (*TestFrame) ProtoMessage()    {}

func (m *TestFrame) GetPing() *TestFrame_Ping {
	if m != nil {
		return m.Ping
	}
	return nil
}

func (m *TestFrame) GetPong() *TestFrame_Pong {
	if m != nil {
		return m.Pong
	}
	return nil
}

type TestFrame_Ping struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch" json:"epoch,omitempty"`
}

func (m *TestFrame_Ping) Reset()         { *m = TestFrame_Ping{} }
func (m *TestFrame_Ping) String() string { return proto.CompactTextString(m) }
func (*TestFrame_Ping) ProtoMessage()    {}

type TestFrame_Pong struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch" json:"epoch,omitempty"`
}

func (m *TestFrame_Pong) Reset()         { *m = TestFrame_Pong{} }
func (m *TestFrame_Pong) String() string { return proto.CompactTextString(m) }
func (*TestFrame_Pong) ProtoMessage()    {}

func init() {
	proto.RegisterEnum("bus.TestFrame_EventType", TestFrame_EventType_name, TestFrame_EventType_value)
}
