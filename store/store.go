// CGO_CFLAGS="-I/usr/include/rocksdb" \
// CGO_LDFLAGS="-L/usr/lib" \
// go get github.com/tecbot/gorocksdb
package store

import (
	"github.com/golang/glog"
	rocksdb "github.com/tecbot/gorocksdb"
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

// 88 bites prefix
type Prefix struct {
	// Instance app id(max 255), <16 is reserved.
	app uint8
	// Predefined table id, <16 is reserved.
	table    uint16
	unixtime int64 // seconds
}

func NewSharePrefix() *Prefix {
	return &Prefix{
		app:   uint8(16),
		table: uint16(16),
	}
}

func ParsePrefix() *Prefix {
	// TODO
	return nil
}
