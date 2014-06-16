package main

/*

#cgo CFLAGS: -I./lib
#cgo LDFLAGS: ./lib/libmultihashing.a

#include <stdint.h>
#include <stdlib.h>
#include <multihashing.h>
#include <x11.h>

*/
import "C"
import "unsafe"
import "fmt"

func main() {
	cs := C.CString("XYZ")
	output := C.multihash_x11(cs, 3)
	fmt.Println(C.GoString(output))
	C.free(unsafe.Pointer(cs))
}
