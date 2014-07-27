// CGO_CFLAGS="-I/usr/include/rocksdb" \
// CGO_LDFLAGS="-L/usr/lib" \
// go get github.com/tecbot/gorocksdb
package stratum

import (
	"github.com/golang/glog"
	"os"
	rocksdb "github.com/tecbot/gorocksdb"
)

type Store struct {
	*rocksdb.DB
}

func NewStore() *Store {
	dbName := os.TempDir() + "/ninepool"
	options := rocksdb.NewDefaultOptions()
	options.SetCreateIfMissing(true)
	db, err := rocksdb.OpenDb(options, dbName)
	if err != nil {
		glog.Fatalf("Can not open db: %s", err)
	}
	return &Store{db}
}

func (db *Store) put(key, value []byte) error {
	o := rocksdb.NewDefaultWriteOptions()
	return db.Put(o, key, value)
}

