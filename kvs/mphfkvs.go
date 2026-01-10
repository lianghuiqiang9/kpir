package kvs

import (
	"unsafe"

	"github.com/local/utils"
)

type KVSInterface interface {
	Name() string
	EncodeBucket(kv *utils.Bucket) utils.EncodedDB
	Encode(kv *utils.KV) utils.EncodedDB
	Lookup(i uint64, key uint64) []uint64
	Decode(key uint64, rawVal [][]uint64) ([]uint64, bool)
	Contains(outkey uint64, key uint64, db *utils.EncodedDB) ([]uint64, bool)
	Size() uint64
	Free()
}

type MPHFKVS[M MPHFInterface, T utils.Unsigned] struct {
	BatchSize    uint64
	MPH          []M
	ValueOffsets []uint32
	NameStr      string
	NewMPHFunc   func(keys []uint64) M
}

func (g *MPHFKVS[M, T]) Name() string { return g.NameStr }

func (g *MPHFKVS[M, T]) Size() uint64 {
	var total uint64
	total += uint64(unsafe.Sizeof(*g))
	for _, m := range g.MPH {
		if any(m) != nil {
			total += (m.Bits() / 8)
		}
	}
	total += uint64(len(g.ValueOffsets) * 4)
	return total
}

func (g *MPHFKVS[M, T]) EncodeBucket(bucket *utils.Bucket) utils.EncodedDB {
	W := bucket.Uint64PerVal
	totalKeys := bucket.TotalKeys

	NumEntries := utils.NextPerfectSquare(totalKeys)
	dbData := make([]uint64, NumEntries*W)

	g.BatchSize = 1
	g.ValueOffsets = []uint32{0, uint32(totalKeys)}
	g.MPH = make([]M, 1)

	g.MPH[0] = g.EncodeSingleBucket(bucket, dbData, 0)

	return utils.EncodedDB{
		Data:           dbData,
		NumEntries:     NumEntries,
		Uint64PerEntry: uint64(W),
		BitsPerEntry:   uint64(W) * 64,
	}
}

func (g *MPHFKVS[M, T]) EncodeSingleBucket(bucket *utils.Bucket, dbData []uint64, startEntryOffset uint64) M {
	W := bucket.Uint64PerVal
	mph := g.NewMPHFunc(bucket.Keys)
	baseOffset := startEntryOffset * W

	process := func(k uint64, v []uint64) {
		idx := mph.Lookup(k)
		if idx == ^uint64(0) {
			idx = 0
		}

		val := utils.KVSFingerPrint[T](k, v)
		targetOffset := baseOffset + idx*W
		copy(dbData[targetOffset:targetOffset+W], val)
	}

	if bucket.IsSort {
		for _, k := range bucket.Keys {
			v, _ := bucket.GetValInterpolation(k)
			process(k, v)
		}
	} else {
		Map := utils.MakeMap(bucket.Keys, bucket.Values)
		for _, k := range bucket.Keys {
			process(k, Map[k])
		}
	}
	return mph
}

func (g *MPHFKVS[M, T]) Encode(kv *utils.KV) utils.EncodedDB {
	count := kv.BucketCount
	W := kv.Uint64PerVal
	NumEntries := utils.NextPerfectSquare(kv.TotalKeys)
	dbData := make([]uint64, NumEntries*W)

	g.BatchSize = 1
	g.MPH = make([]M, count)
	g.ValueOffsets = make([]uint32, count+1)

	var currentArrayOffset uint64 = 0
	for i := uint64(0); i < count; i++ {
		g.ValueOffsets[i] = uint32(currentArrayOffset)
		bucket := kv.Buckets[i]
		if len(bucket.Keys) == 0 {
			continue
		}

		g.MPH[i] = g.EncodeSingleBucket(bucket, dbData, currentArrayOffset)
		currentArrayOffset += bucket.TotalKeys
	}
	g.ValueOffsets[count] = uint32(currentArrayOffset)

	return utils.EncodedDB{
		Data: dbData, NumEntries: NumEntries, Uint64PerEntry: uint64(W), BitsPerEntry: uint64(W) * 64,
	}
}

func (g *MPHFKVS[M, T]) Lookup(i uint64, key uint64) []uint64 {
	baseOffset := uint64(g.ValueOffsets[i])
	idx := g.MPH[i].Lookup(key)
	if idx == ^uint64(0) {
		return []uint64{baseOffset}
	}
	return []uint64{baseOffset + idx}
}

func (g *MPHFKVS[M, T]) Decode(key uint64, rawVal [][]uint64) ([]uint64, bool) {
	return utils.KVSFingerPrintInv[T](key, rawVal[0])
}

func (g *MPHFKVS[M, T]) Free() {
	for _, m := range g.MPH {
		if any(m) != nil {
			m.Free()
		}
	}
}

func (g *MPHFKVS[M, T]) Contains(outkey uint64, key uint64, db *utils.EncodedDB) ([]uint64, bool) {
	indexes := g.Lookup(outkey, key)
	rawVal := db.GetBatchEntry(indexes)
	return g.Decode(key, rawVal)
}

func (g *MPHFKVS[M, T]) GetBatchSize() uint64 {
	return g.BatchSize
}
