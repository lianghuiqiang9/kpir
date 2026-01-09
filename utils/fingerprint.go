package utils

import (
	"bytes"
	"encoding/gob"
	"sort"
	"unsafe"
)

type Unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

func hash64(x uint64) uint64 {
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x = x ^ (x >> 31)
	return x
}

func hashSlice(x uint64, length int) []uint64 {
	results := make([]uint64, length)

	state := x ^ 0x9e3779b97f4a7c15

	for i := 0; i < length; i++ {
		state += 0x9e3779b97f4a7c15
		results[i] = hash64(state)
	}
	return results
}

func xorInplace(a, b []uint64) {

	for i := 0; i < len(a); i++ {
		a[i] ^= b[i]
	}
}

func bitsWidth[T Unsigned]() uint64 {
	var zero T
	return uint64(unsafe.Sizeof(zero)) * 8
}

func KVSFingerPrint[T Unsigned](k uint64, v []uint64) []uint64 {
	fp := uint64(T(hash64(k))) // hash(k)
	out := hashSlice(k, len(v))
	xorInplace(out, v) // v ^ hash2(k)

	fpSize := bitsWidth[T]()
	fpInHigh := fp << (64 - fpSize)

	mask := uint64(0xFFFFFFFFFFFFFFFF >> fpSize)
	lastIdx := len(out) - 1
	out[lastIdx] = (out[lastIdx] & mask) | fpInHigh // v ^ hash2(k) | hash(k)

	return out
}

func KVSFingerPrintInv[T Unsigned](k uint64, v []uint64) ([]uint64, bool) {
	fpSize := bitsWidth[T]()
	lastIdx := len(v) - 1

	extractedFp := v[lastIdx] >> (64 - fpSize)

	expectedFp := uint64(T(hash64(k)))
	if extractedFp != expectedFp {
		return nil, false
	}

	out := hashSlice(k, len(v))

	xorInplace(out, v)

	mask := uint64(0xFFFFFFFFFFFFFFFF >> fpSize)
	out[lastIdx] &= mask

	return out, true
}

func Sort(keys []uint64, vals []uint64) ([]uint64, []uint64) {
	n := len(keys)
	w := len(vals) / n

	type pair struct {
		key uint64
		idx int
	}
	pairs := make([]pair, n)
	for i := range keys {
		pairs[i] = pair{keys[i], i}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	newKeys := make([]uint64, n)
	newValues := make([]uint64, len(vals))
	for i := 0; i < n; i++ {
		newKeys[i] = pairs[i].key
		oldIdx := pairs[i].idx
		copy(newValues[i*w:(i+1)*w], vals[oldIdx*w:(oldIdx+1)*w])
	}
	return newKeys, newValues
}

func GetValInterpolation(keys []uint64, vals []uint64, uint64PerVal int, key uint64) ([]uint64, bool) {
	low, high := 0, len(keys)-1

	for low <= high && key >= keys[low] && key <= keys[high] {
		if low == high {
			if keys[low] == key {
				break
			}
			return nil, false
		}

		pos := low + int(float64(key-keys[low])/float64(keys[high]-keys[low])*float64(high-low))

		if keys[pos] < key {
			low = pos + 1
		} else if keys[pos] > key {
			high = pos - 1
		} else {
			start := pos * uint64PerVal
			return vals[start : start+uint64PerVal], true
		}
	}
	return nil, false
}

func MakeMap(keys []uint64, vals []uint64) map[uint64][]uint64 {
	if len(keys) == 0 {
		return make(map[uint64][]uint64)
	}

	W := len(vals) / len(keys)

	res := make(map[uint64][]uint64, len(keys))

	for i := 0; i < len(keys); i++ {
		res[keys[i]] = vals[W*i : W*i+W]
	}

	return res
}

func GetSerializedSize(m interface{}) int {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return 0
	}
	return buf.Len()
}
