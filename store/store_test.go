package store_test

import (
	capn "github.com/glycerine/go-capnproto"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yinhm/ninepool/store"
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	dbpath := os.TempDir() + "/ninepool"

	Convey("When put then get value, it should be equal", t, func() {
		store.DestroyStore(dbpath)
		db := store.NewStore(dbpath)

		err := db.Put([]byte("key1"), []byte("value1"))
		So(err, ShouldBeNil)

		slice, err := db.Get([]byte("key1"))
		So(err, ShouldBeNil)

		So(string(slice.Data()), ShouldEqual, "value1")
		slice.Free()
	})
}

func TestShare(t *testing.T) {
	dbpath := os.TempDir() + "/ninepool"

	Convey("Subject: share put then get", t, func() {
		store.DestroyStore(dbpath)
		// db := store.NewStore(dbpath)
		// prefix := store.NewSharePrefix()
		s := capn.NewBuffer(nil)
		share := store.NewRootShare(s)
		share.SetUsername("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda")
		share.SetJobId("31")
		share.SetPool("1")
		share.SetHeader("0000000007a4a0e8212730fa0a832e89cb3d571445ca9f52b8eed811108cf3a4")
		share.SetDiff(0.1)
		share.SetIsBlock(false)
		share.SetAccepted(false)
		share.SetExtraNonce1("580000000002")
		share.SetExtraNonce2("0000")
		share.SetNtime("53b98c13")
		share.SetNonce("e20f3e56")
		So(share.Ntime(), ShouldEqual, "53b98c13")
	})
}
