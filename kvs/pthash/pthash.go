package pthash

/*
#cgo CXXFLAGS: -std=c++17 -I./pthash/include -I./pthash/external/bits/external/essentials/include -I./pthash/external/mm_file/include -I./pthash/external/xxHash -I./pthash/external/bits/include

#cgo LDFLAGS: -L. -lpthash_wrapper -Wl,-rpath,.
#include <stdint.h>
#include <stdlib.h>

void* pthash_init(uint64_t* keys, size_t n);
uint64_t pthash_lookup(void* ptr, uint64_t key);
uint64_t pthash_get_bits(void* ptr);
void pthash_free(void* ptr);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type PTHash struct {
	ptr unsafe.Pointer
}

func New(keys []uint64) *PTHash {
	if len(keys) == 0 {
		return nil
	}
	cKeys := (*C.uint64_t)(unsafe.Pointer(&keys[0]))
	ptr := C.pthash_init(cKeys, C.size_t(len(keys)))

	p := &PTHash{ptr: ptr}

	runtime.SetFinalizer(p, func(obj *PTHash) {
		obj.Free()
	})
	return p
}

func (p *PTHash) Lookup(key uint64) uint64 {
	return uint64(C.pthash_lookup(p.ptr, C.uint64_t(key)))
}

func (p *PTHash) Bits() uint64 {
	if p.ptr == nil {
		return 0
	}
	return uint64(C.pthash_get_bits(p.ptr))
}

func (p *PTHash) Free() {
	if p.ptr != nil {
		C.pthash_free(p.ptr)
		p.ptr = nil
	}
}
