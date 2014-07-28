package stratum_test

import (
	"github.com/yinhm/ninepool/stratum"
	"os"	
	"testing"
)

func TestStore(t *testing.T) {
	dbpath := os.TempDir() + "/ninepool"
	db := stratum.NewStore(dbpath)

	err := db.Put([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Errorf("put failed, %s", err)
	}

	slice, err := db.Get([]byte("key1"))
	if err != nil {
		t.Errorf("get failed, %s", err)
	}

	if string(slice.Data()) != "value1" {
		t.Errorf("value not match: %s", string(slice.Data()))
	}
	slice.Free()
}
