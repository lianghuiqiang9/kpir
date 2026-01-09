package kvs

import (
	"unsafe"

	utils "github.com/local/utils"
	bbhash2 "github.com/relab/bbhash"
)

type TBBHash2 uint32

type BBHash2KVS struct {
	BatchSize    uint64
	BBHash2      []*bbhash2.BBHash2
	ValueOffsets []uint32

	//BucketCount  uint64
	//Uint64PerVal uint64
}

func (bbhash2kvs *BBHash2KVS) Name() string {
	return "bbhash2kvs"
}

func (bbhash2kvs *BBHash2KVS) Size() uint64 {
	var total uint64
	total += uint64(unsafe.Sizeof(*bbhash2kvs))

	for i, b := range bbhash2kvs.BBHash2 {
		if b != nil {
			totalkeys := float64(bbhash2kvs.ValueOffsets[i+1] - bbhash2kvs.ValueOffsets[i])
			total += uint64(b.BitsPerKey() * totalkeys / 8)
		}
	}

	total += uint64(len(bbhash2kvs.ValueOffsets) * 4)

	return total
}

func (bbhash2kvs *BBHash2KVS) EncodeSingleBucket(bucket *utils.Bucket, dbData []uint64, startEntryOffset uint64) *bbhash2.BBHash2 {
	W := bucket.Uint64PerVal
	totalKeys := bucket.TotalKeys

	bb, _ := bbhash2.New(bucket.Keys, bbhash2.Gamma(1), bbhash2.WithReverseMap())

	baseOffset := startEntryOffset * W

	if bucket.IsSort {
		for i := uint64(0); i < totalKeys; i++ {
			k := bb.Key(i + 1)
			v, _ := bucket.GetValInterpolation(k)
			val := utils.KVSFingerPrint[TBBHash2](k, v)
			targetOffset := baseOffset + i*W
			copy(dbData[targetOffset:targetOffset+W], val)
		}
	} else {
		Map := utils.MakeMap(bucket.Keys, bucket.Values)
		for i := uint64(0); i < totalKeys; i++ {
			k := bb.Key(i + 1)
			v, _ := Map[k]
			val := utils.KVSFingerPrint[TBBHash2](k, v)
			targetOffset := baseOffset + i*W
			copy(dbData[targetOffset:targetOffset+W], val)
		}
	}
	return bb
}

func (bbhash2kvs *BBHash2KVS) EncodeBucket(bucket *utils.Bucket) utils.EncodedDB {
	W := bucket.Uint64PerVal
	totalKeys := bucket.TotalKeys
	NumEntries := utils.NextPerfectSquare(totalKeys)
	dbData := make([]uint64, NumEntries*W)

	bbhash2kvs.BatchSize = 1
	bbhash2kvs.ValueOffsets = []uint32{0, uint32(totalKeys)}

	bb := bbhash2kvs.EncodeSingleBucket(bucket, dbData, 0)
	bbhash2kvs.BBHash2 = []*bbhash2.BBHash2{bb}

	return utils.EncodedDB{
		Data:           dbData,
		NumEntries:     NumEntries,
		Uint64PerEntry: uint64(W),
		BitsPerEntry:   uint64(W) * 64,
	}
}

func (bbhash2kvs *BBHash2KVS) Encode(kv *utils.KV) utils.EncodedDB {
	count := kv.BucketCount
	W := kv.Uint64PerVal
	NumEntries := utils.NextPerfectSquare(kv.TotalKeys)
	dbData := make([]uint64, NumEntries*W)

	bbhash2kvs.BatchSize = 1
	bbhash2kvs.BBHash2 = make([]*bbhash2.BBHash2, count)
	bbhash2kvs.ValueOffsets = make([]uint32, count+1)

	var currentArrayOffset uint64 = 0
	for i := uint64(0); i < count; i++ {
		bbhash2kvs.ValueOffsets[i] = uint32(currentArrayOffset)
		bucket := kv.Buckets[i]
		if len(bucket.Keys) == 0 {
			continue
		}

		bb := bbhash2kvs.EncodeSingleBucket(bucket, dbData, currentArrayOffset)
		bbhash2kvs.BBHash2[i] = bb

		currentArrayOffset += bucket.TotalKeys
	}
	bbhash2kvs.ValueOffsets[count] = uint32(currentArrayOffset)

	return utils.EncodedDB{
		Data:           dbData,
		NumEntries:     NumEntries,
		Uint64PerEntry: uint64(W),
		BitsPerEntry:   uint64(W) * 64,
	}
}

func (bbhash2kvs *BBHash2KVS) Index(i uint64, key uint64) []uint64 {
	offset := uint64(bbhash2kvs.ValueOffsets[i])

	// Find
	idx := bbhash2kvs.BBHash2[i].Find(key)

	if idx == 0 {
		return []uint64{offset}
	}

	idx -= 1 // slight different with bbhash, the bbhash2 is [1,N], bbhash is [0,N-1]
	return []uint64{offset + idx}
}

func (bbhash2kvs *BBHash2KVS) Decode(key uint64, rawVal [][]uint64) ([]uint64, bool) {
	v := rawVal[0]
	return utils.KVSFingerPrintInv[TBBHash2](key, v)
}

func (bbhash2kvs *BBHash2KVS) Contains(outkey uint64, key uint64, db *utils.EncodedDB) ([]uint64, bool) {
	indexes := bbhash2kvs.Index(outkey, key)
	rawVal := db.GetBatchEntry(indexes)
	return bbhash2kvs.Decode(key, rawVal)
}

func (bbhash2kvs *BBHash2KVS) Free() {

}

func (bbhash2kvs *BBHash2KVS) GetBatchSize() uint64 {
	return bbhash2kvs.BatchSize
}
