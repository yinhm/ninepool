// CGO_CFLAGS="-I/usr/include/rocksdb" \
// CGO_LDFLAGS="-L/usr/lib" \
// go get github.com/tecbot/gorocksdb
package store

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	capn "github.com/glycerine/go-capnproto"
	"github.com/golang/glog"
	rocksdb "github.com/tecbot/gorocksdb"
	"github.com/yinhm/ninepool/proto"
	"io"
	"time"
	"unsafe"
)

type Store struct {
	dbpath  string
	rdb     *rocksdb.DB
	options *rocksdb.Options
	ro      *rocksdb.ReadOptions
	wo      *rocksdb.WriteOptions
}

func NewStore(dbpath string) *Store {
	db := new(Store)
	db.dbpath = dbpath
	db.options = NewStoreOptions()
	db.initReadOptions()
	db.initWriteOptions()

	rdb, err := rocksdb.OpenDb(db.options, db.dbpath)
	if err != nil {
		db.destroyOptions()
		glog.Fatalf("Can not open db: %s", err)
	}
	db.rdb = rdb
	return db
}

func DestroyStore(dbpath string) error {
	options := NewStoreOptions()
	return rocksdb.DestroyDb(dbpath, options)
}

func NewStoreOptions() *rocksdb.Options {
	var prefix Prefix
	transform := NewFixedPrefixTransform(int(unsafe.Sizeof(prefix)))

	opts := rocksdb.NewDefaultOptions()
	opts.SetBlockCache(rocksdb.NewLRUCache(128 << 20)) // 128MB
	// Default bits_per_key is 10, which yields ~1% false positive rate.
	opts.SetFilterPolicy(rocksdb.NewBloomFilter(10))
	opts.SetPrefixExtractor(transform)
	opts.SetWriteBufferSize(16 << 20) // 8MB
	opts.SetTargetFileSizeBase(16 << 20)
	opts.SetCreateIfMissing(true)
	return opts
}

func (db *Store) initReadOptions() {
	db.ro = rocksdb.NewDefaultReadOptions()
}

func (db *Store) initWriteOptions() {
	db.wo = rocksdb.NewDefaultWriteOptions()
}

// FIXME: Close in testing cause runtime error
func (db *Store) Close() {
	db.rdb.Close()
}

func (db *Store) destroyOptions() {
	db.options.Destroy()
	db.ro.Destroy()
	db.wo.Destroy()
	db.options = nil
	db.ro = nil
	db.wo = nil
}

func (db *Store) Get(key []byte) (*rocksdb.Slice, error) {
	return db.rdb.Get(db.ro, key)
}

func (db *Store) Put(key, value []byte) error {
	return db.rdb.Put(db.wo, key, value)
}

func (db *Store) Delete(key []byte) error {
	return db.rdb.Delete(db.wo, key)
}

// https://github.com/facebook/rocksdb/wiki/Prefix-Seek-API-Changes
type FixedPrefixTransform struct {
	size int
}

func NewFixedPrefixTransform(size int) *FixedPrefixTransform {
	return &FixedPrefixTransform{
		size: size,
	}
}

func (t *FixedPrefixTransform) Transform(src []byte) []byte {
	return src[0:t.size]
}

func (t *FixedPrefixTransform) InDomain(src []byte) bool {
	return len(src) >= t.size
}

func (t *FixedPrefixTransform) InRange(src []byte) bool {
	return len(src) == t.size
}

func (t *FixedPrefixTransform) Name() string {
	return "FixedPrefixTransform"
}

// 88 bits prefix
type Prefix struct {
	// proto.Prefix are defined as following:
	// - app: app id(max 255), <16 is reserved.
	// - symbol: Predefined table id, <16 is reserved.
	// - custom: 32 bits custom, unix time or some custom id
	// +----------+----------+----------+
	// |  8bits   |  16bits  |  64bits  |
	// +----------+----------+----------+
	// |   app    |  table   |  custom  |
	// +----------+----------+----------+
	proto.Prefix
}

func NewPrefix(app uint8, table uint16) *Prefix {
	seg := capn.NewBuffer(nil)
	pp := proto.NewRootPrefix(seg)
	return &Prefix{pp}
}

func ParsePrefix() *Prefix {
	// TODO
	return nil
}

func (p *Prefix) Bytes() []byte {
	buf := bytes.Buffer{}
	p.Segment.WriteTo(&buf)
	return buf.Bytes()
}

func (p *Prefix) WriteTo(w io.Writer) (int64, error) {
	return p.Segment.WriteTo(w)
}

// prefix + unixnano form a key.
type Key struct {
	Prefix
	unixnano int64
}

func NewKey(prefix Prefix, unixnano int64) *Key {
	return &Key{
		Prefix:   prefix,
		unixnano: unixnano,
	}
}

func NewNanoKey(prefix Prefix) *Key {
	unixnano := time.Now().UnixNano()
	return NewKey(prefix, unixnano)
}

func (k *Key) Bytes() ([]byte, error) {
	buf := bytes.Buffer{}
	k.WriteTo(&buf)
	err := binary.Write(&buf, binary.LittleEndian, k.unixnano)
	return buf.Bytes(), err
}

func (k *Key) String() string {
	bytes, _ := k.Bytes()
	return string(bytes)
}

func (k *Key) Time() time.Time {
	return time.Unix(0, k.unixnano)
}

type Share struct {
	db     *Store
	prefix *Prefix
}

func NewShare(store *Store) *Share {
	prefix := NewPrefix(uint8(16), uint16(16))

	return &Share{
		db:     store,
		prefix: prefix,
	}
}

func (s Share) Put(pshare proto.Share) (*Key, error) {
	pshare.SetCreated(time.Now().Unix())
	key := NewNanoKey(*s.prefix)
	buf := bytes.Buffer{}
	pshare.Segment.WriteTo(&buf)

	kb, err := key.Bytes()
	if err != nil {
		return key, err
	}

	err = s.db.Put(kb, buf.Bytes())
	return key, err
}

func (s Share) Get(key *Key) (proto.Share, error) {
	var ps proto.Share

	bytes, err := key.Bytes()
	if err != nil {
		return ps, err
	}

	value, err := s.db.Get(bytes)
	if err != nil {
		return ps, err
	}
	defer value.Free()

	buf, _, err := capn.ReadFromMemoryZeroCopy(value.Data())
	if err != nil {
		return ps, err
	}

	ps = proto.ReadRootShare(buf)
	//log.Printf("%s", s.String(ps))
	return ps, nil
}

func (s Share) String(ps proto.Share) string {
	buf := bytes.Buffer{}
	ps.Segment.WriteTo(&buf)
	return hex.EncodeToString(buf.Bytes())
}
