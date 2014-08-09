package proto

// AUTO GENERATED - DO NOT EDIT

import (
	C "github.com/glycerine/go-capnproto"
	"math"
	"unsafe"
)

type Prefix C.Struct

func NewPrefix(s *C.Segment) Prefix      { return Prefix(s.NewStruct(16, 0)) }
func NewRootPrefix(s *C.Segment) Prefix  { return Prefix(s.NewRootStruct(16, 0)) }
func ReadRootPrefix(s *C.Segment) Prefix { return Prefix(s.Root(0).ToStruct()) }
func (s Prefix) App() uint8              { return C.Struct(s).Get8(0) }
func (s Prefix) SetApp(v uint8)          { C.Struct(s).Set8(0, v) }
func (s Prefix) Table() uint16           { return C.Struct(s).Get16(2) }
func (s Prefix) SetTable(v uint16)       { C.Struct(s).Set16(2, v) }
func (s Prefix) Unixtime() int64         { return int64(C.Struct(s).Get64(8)) }
func (s Prefix) SetUnixtime(v int64)     { C.Struct(s).Set64(8, uint64(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s *Prefix) MarshalJSON() (bs []byte, err error) {
	return
}

type Prefix_List C.PointerList

func NewPrefixList(s *C.Segment, sz int) Prefix_List {
	return Prefix_List(s.NewCompositeList(16, 0, sz))
}
func (s Prefix_List) Len() int        { return C.PointerList(s).Len() }
func (s Prefix_List) At(i int) Prefix { return Prefix(C.PointerList(s).At(i).ToStruct()) }
func (s Prefix_List) ToArray() []Prefix {
	return *(*[]Prefix)(unsafe.Pointer(C.PointerList(s).ToArray()))
}

type Share C.Struct

func NewShare(s *C.Segment) Share       { return Share(s.NewStruct(16, 8)) }
func NewRootShare(s *C.Segment) Share   { return Share(s.NewRootStruct(16, 8)) }
func ReadRootShare(s *C.Segment) Share  { return Share(s.Root(0).ToStruct()) }
func (s Share) Username() string        { return C.Struct(s).GetObject(0).ToText() }
func (s Share) SetUsername(v string)    { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s Share) JobId() string           { return C.Struct(s).GetObject(1).ToText() }
func (s Share) SetJobId(v string)       { C.Struct(s).SetObject(1, s.Segment.NewText(v)) }
func (s Share) Pool() string            { return C.Struct(s).GetObject(2).ToText() }
func (s Share) SetPool(v string)        { C.Struct(s).SetObject(2, s.Segment.NewText(v)) }
func (s Share) Header() string          { return C.Struct(s).GetObject(3).ToText() }
func (s Share) SetHeader(v string)      { C.Struct(s).SetObject(3, s.Segment.NewText(v)) }
func (s Share) Diff() float64           { return math.Float64frombits(C.Struct(s).Get64(0)) }
func (s Share) SetDiff(v float64)       { C.Struct(s).Set64(0, math.Float64bits(v)) }
func (s Share) IsBlock() bool           { return C.Struct(s).Get1(64) }
func (s Share) SetIsBlock(v bool)       { C.Struct(s).Set1(64, v) }
func (s Share) Accepted() bool          { return C.Struct(s).Get1(65) }
func (s Share) SetAccepted(v bool)      { C.Struct(s).Set1(65, v) }
func (s Share) ExtraNonce1() string     { return C.Struct(s).GetObject(4).ToText() }
func (s Share) SetExtraNonce1(v string) { C.Struct(s).SetObject(4, s.Segment.NewText(v)) }
func (s Share) ExtraNonce2() string     { return C.Struct(s).GetObject(5).ToText() }
func (s Share) SetExtraNonce2(v string) { C.Struct(s).SetObject(5, s.Segment.NewText(v)) }
func (s Share) Ntime() string           { return C.Struct(s).GetObject(6).ToText() }
func (s Share) SetNtime(v string)       { C.Struct(s).SetObject(6, s.Segment.NewText(v)) }
func (s Share) Nonce() string           { return C.Struct(s).GetObject(7).ToText() }
func (s Share) SetNonce(v string)       { C.Struct(s).SetObject(7, s.Segment.NewText(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s *Share) MarshalJSON() (bs []byte, err error) {
	return
}

type Share_List C.PointerList

func NewShareList(s *C.Segment, sz int) Share_List { return Share_List(s.NewCompositeList(16, 8, sz)) }
func (s Share_List) Len() int                      { return C.PointerList(s).Len() }
func (s Share_List) At(i int) Share                { return Share(C.PointerList(s).At(i).ToStruct()) }
func (s Share_List) ToArray() []Share              { return *(*[]Share)(unsafe.Pointer(C.PointerList(s).ToArray())) }
