package bbhash

/*
#cgo CPPFLAGS: -w
#cgo LDFLAGS: -L${SRCDIR} -lbbhash_wrapper -lstdc++
#include <stdint.h>

// 声明 C 函数
typedef struct boophf_t boophf_t; // 对应 C++ 的 boophf_t

boophf_t* bbhash_build(const uint64_t* keys, size_t n, int threads, double gamma);
uint64_t bbhash_lookup(boophf_t* handle, uint64_t key);
uint64_t bbhash_total_bits(boophf_t* handle);
void bbhash_free(boophf_t* handle);
*/
import "C"
import "unsafe"

type BBHash struct {
	handle *C.boophf_t
}

// Build 构建完美哈希函数
func New(keys []uint64) *BBHash {
	threads := 1
	gamma := 1.0
	if len(keys) == 0 {
		return nil
	}
	// 传递切片首地址
	ptr := C.bbhash_build(
		(*C.uint64_t)(unsafe.Pointer(&keys[0])),
		C.size_t(len(keys)),
		C.int(threads),
		C.double(gamma),
	)
	return &BBHash{handle: ptr}
}

// Lookup 正向映射 Key -> Index
func (b *BBHash) Lookup(key uint64) uint64 {
	return uint64(C.bbhash_lookup(b.handle, C.uint64_t(key)))
}

// TotalBitSize 获取 MPHF 内部位图和结构占用的总 Bit 数
func (b *BBHash) Bits() uint64 {
	return uint64(C.bbhash_total_bits(b.handle))
}

// Free 释放 C++ 内存
func (b *BBHash) Free() {
	if b.handle != nil {
		C.bbhash_free(b.handle)
		b.handle = nil
	}
}
