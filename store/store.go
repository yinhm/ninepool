// CGO_CFLAGS="-I/usr/include/rocksdb" \
// CGO_LDFLAGS="-L/usr/lib" \
// go get github.com/tecbot/gorocksdb
package store

import (
	"bytes"
	"encoding/hex"
	capn "github.com/glycerine/go-capnproto"
	"github.com/golang/glog"
	rocksdb "github.com/tecbot/gorocksdb"
	"github.com/yinhm/ninepool/proto"
	"unsafe"
	"time"
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

func (db *Store) Close() {
	db.options.Destroy()
	db.ro.Destroy()
	db.wo.Destroy()
	db.rdb.Close()
	db.rdb = nil
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
  // - unixtime: the unixtime for the key
  // +----------+----------+----------+
  // |  8bits   |  16bits  |  64bits  |
  // +----------+----------+----------+
  // |   app    |  table   | unixtime |
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

type Share struct {
	db     *Store
	prefix *Prefix
}

func NewShare(store *Store) *Share {
	prefix := NewPrefix(uint8(16), uint16(16))
	
	return &Share {
		db:     store,
		prefix: prefix,
	}
}

func (s Share) Put(pshare proto.Share) ([]byte, error) {
	s.prefix.SetUnixtime(time.Now().Unix())
	key := s.prefix.Bytes()
	buf := bytes.Buffer{}
	pshare.Segment.WriteTo(&buf)
	if err := s.db.Put(key, buf.Bytes()); err != nil {
		return key, err
	}
	return key, nil
}

func (s Share) Get(key []byte) (proto.Share, error) {
	var ps proto.Share

	value, err := s.db.Get(key)
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
