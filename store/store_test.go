package store_test

import (
	"bytes"
	capn "github.com/glycerine/go-capnproto"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yinhm/ninepool/store"
	"github.com/yinhm/ninepool/proto"
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
	dbpath := os.TempDir() + "/TestShare"

	Convey("Subject: share", t, func() {
		store.DestroyStore(dbpath)
		db := store.NewStore(dbpath)

		s := capn.NewBuffer(nil)
		pshare := proto.NewRootShare(s)
		pshare.SetUsername("1PJ1DVi5n6T4NisfnVbYmL17a4WNfaFsda")
		pshare.SetJobId("31")
		pshare.SetPool("1")
		pshare.SetHeader("0000000007a4a0e8212730fa0a832e89cb3d571445ca9f52b8eed811108cf3a4")
		pshare.SetDiff(0.1)
		pshare.SetIsBlock(false)
		pshare.SetAccepted(false)
		pshare.SetExtraNonce1("580000000002")
		pshare.SetExtraNonce2("0000")
		pshare.SetNtime("53b98c13")
		pshare.SetNonce("e20f3e56")
		So(pshare.Ntime(), ShouldEqual, "53b98c13")

		buf := bytes.Buffer{}
		s.WriteTo(&buf)
		s2, err := capn.ReadFromStream(&buf, nil)
		So(err, ShouldBeNil)
		ps := proto.ReadRootShare(s2)
		So(ps.Ntime(), ShouldEqual, "53b98c13")

		Convey("When put/get, it should be the same", func() {
			share := store.NewShare(db)
			key, err := share.Put(pshare)
			So(err, ShouldBeNil)

			ps2, err := share.Get(key)
			So(err, ShouldBeNil)
			So(ps2.Ntime(), ShouldEqual, "53b98c13")
		})
	})
}
