package stratum_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/yinhm/ninepool/stratum"
	"os"	
	"testing"
)

func TestStore(t *testing.T) {
	dbpath := os.TempDir() + "/ninepool"

	Convey("When put then get value, it should be equal", t, func() {
		stratum.DestroyStore(dbpath)
		db := stratum.NewStore(dbpath)

		err := db.Put([]byte("key1"), []byte("value1"))
		So(err, ShouldBeNil)

		slice, err := db.Get([]byte("key1"))
		So(err, ShouldBeNil)

		So(string(slice.Data()), ShouldEqual, "value1")
		slice.Free()
	})
}
