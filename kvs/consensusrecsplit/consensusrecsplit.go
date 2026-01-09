package consensusrecsplit

/*
#cgo CXXFLAGS: -std=c++20
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/include
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/extlib/ips2ra/include
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/extlib/util/include
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/extlib/util/extern/pasta-bit-vector/include
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/extlib/tlx
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/build/_deps/pasta_utils-src/include
#cgo CXXFLAGS: -I${SRCDIR}/ConsensusRecSplit/extlib/fips/include
#cgo CXXFLAGS: -I${SRCDIR}

#cgo LDFLAGS: -L${SRCDIR} -lconsensusrecsplit_wrapper -lstdc++

#include <stdint.h>
#include <stdlib.h>
#include <stddef.h>

// C wrapper
typedef struct MPHWrapper MPHWrapper;

MPHWrapper* create_mph(const uint64_t* keys, size_t num_keys);
uint64_t query_mph(MPHWrapper* wrapper, uint64_t key);
size_t mph_bits(MPHWrapper* wrapper);
void free_mph(MPHWrapper* wrapper);
*/
import "C"

import (
	"unsafe"
)

type MPH struct {
	ptr *C.MPHWrapper
}

func New(keys []uint64) *MPH {
	n := len(keys)

	cKeys := (*C.uint64_t)(C.malloc(C.size_t(n) * 8))
	defer C.free(unsafe.Pointer(cKeys))

	for i, k := range keys {
		ptr := (*C.uint64_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cKeys)) + uintptr(i)*8))
		*ptr = C.uint64_t(k)
	}

	p := C.create_mph(cKeys, C.size_t(n))
	return &MPH{ptr: p}
}

func (m *MPH) Lookup(key uint64) uint64 {
	return uint64(C.query_mph(m.ptr, C.uint64_t(key)))
}

func (m *MPH) Bits() uint64 {
	return uint64(C.mph_bits(m.ptr))
}

func (m *MPH) Free() {
	if m.ptr != nil {
		C.free_mph(m.ptr)
		m.ptr = nil
	}
}
