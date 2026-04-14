package kvs

import (
	"math/bits"
	"unsafe"

	xorfilter "github.com/FastFilter/xorfilter"
	utils "github.com/local/utils"
)

type TBFF uint32

/*
type BFFKVSHint struct {
	Count              uint64
	SegmentLength      []uint32
	SegmentCountLength []uint32
	ValueOffsets       []uint32
}
*/

type BFFKVS struct {
	//Filters      []xorfilter.BinaryFuse[T]
	//Hint         *BFFKVSHint
	BatchSize          uint64
	SegmentLength      []uint32
	SegmentCountLength []uint32
	ValueOffsets       []uint32

	//BucketCount  uint64
	//Uint64PerVal uint64
}

func (bffkvs *BFFKVS) Name() string {
	return "bffkvs"
}

func (bffkvs *BFFKVS) Size() uint64 {
	if bffkvs == nil {
		return 0
	}
	var total uint64

	total += uint64(unsafe.Sizeof(*bffkvs))

	total += uint64(len(bffkvs.SegmentLength) * 4)

	total += uint64(len(bffkvs.SegmentCountLength) * 4)

	total += uint64(len(bffkvs.ValueOffsets) * 4)

	return total
}

func (bffkvs *BFFKVS) EncodeBucket(bucket *utils.Bucket) utils.EncodedDB {

	var b xorfilter.BinaryFuseBuilder
	filter, err := xorfilter.BuildBinaryFuseBucket[TBFF](&b, bucket.Keys, bucket.Values, bucket.IsSort)
	if err != nil {
		return utils.EncodedDB{}
	}

	W := bucket.Uint64PerVal

	bffkvs.BatchSize = 3
	bffkvs.SegmentLength = []uint32{filter.SegmentLength}
	bffkvs.SegmentCountLength = []uint32{filter.SegmentCountLength}
	bffkvs.ValueOffsets = []uint32{0, filter.ArrayLength}

	//bffkvs.BucketCount = 1
	//bffkvs.Uint64PerVal = W

	NumEntries := uint32(utils.NextPerfectSquare(uint64(filter.ArrayLength)))
	dbData := make([]uint64, uint64(NumEntries)*W)
	copy(dbData[0:], filter.Fingerprints)

	//:= make([]uint64, uint64(filter.ArrayLength)*uint64(W))
	//copy(dbData[0:], filter.Fingerprints)

	//bffkvs.Filters[0].Fingerprints = nil

	return utils.EncodedDB{
		Data:           dbData,
		NumEntries:     uint64(NumEntries),
		Uint64PerEntry: uint64(W),
		BitsPerEntry:   uint64(W) * 64,
	}
}

func (bffkvs *BFFKVS) Encode(kv *utils.KV) utils.EncodedDB {
	count := kv.BucketCount
	W := kv.Uint64PerVal

	//bffkvs.BucketCount = count
	//bffkvs.Uint64PerVal = W

	filterV := make([]xorfilter.BinaryFuse[TBFF], count)

	bffkvs.BatchSize = 3
	bffkvs.SegmentLength = make([]uint32, count)
	bffkvs.SegmentCountLength = make([]uint32, count)
	bffkvs.ValueOffsets = make([]uint32, count+1)

	var currentArrayOffset uint32 = 0

	for i := uint64(0); i < count; i++ {
		bffkvs.ValueOffsets[i] = currentArrayOffset

		bucket := kv.Buckets[i]
		if len(bucket.Keys) == 0 {
			continue
		}

		var b xorfilter.BinaryFuseBuilder
		filter, err := xorfilter.BuildBinaryFuseBucket[TBFF](&b, bucket.Keys, bucket.Values, bucket.IsSort)

		if err == nil {
			filterV[i] = filter
			currentArrayOffset += filter.ArrayLength
		}
		bffkvs.SegmentLength[i] = filter.SegmentLength
		bffkvs.SegmentCountLength[i] = filter.SegmentCountLength
	}
	bffkvs.ValueOffsets[count] = currentArrayOffset
	NumEntries := uint32(utils.NextPerfectSquare(uint64(currentArrayOffset))) // not only the square, the bff will 1.125 increase the kv size.
	dbData := make([]uint64, uint64(NumEntries)*W)

	for i := uint64(0); i < count; i++ {
		if len(filterV[i].Fingerprints) == 0 {
			continue
		}

		start := uint64(bffkvs.ValueOffsets[i]) * W
		copy(dbData[start:], filterV[i].Fingerprints)

		//bffkvs.Filters[i].Fingerprints = nil
	}

	return utils.EncodedDB{
		Data:           dbData,
		NumEntries:     uint64(NumEntries),
		Uint64PerEntry: uint64(W),
		BitsPerEntry:   uint64(W) * 64,
	}
}

func (hint *BFFKVS) Lookup(i uint64, key uint64) []uint64 {

	SegmentCountLength := hint.SegmentCountLength[i]
	SegmentLength := hint.SegmentLength[i]
	offset := uint64(hint.ValueOffsets[i])

	// filter.getHashFromHash(hash)
	SegmentLengthMask := SegmentLength - 1
	hash := key
	hi, _ := bits.Mul64(hash, uint64(SegmentCountLength))
	h0 := uint32(hi)
	h1 := h0 + SegmentLength
	h2 := h1 + SegmentLength
	h1 ^= uint32(hash>>18) & SegmentLengthMask
	h2 ^= uint32(hash) & SegmentLengthMask

	return []uint64{uint64(h0) + offset, uint64(h1) + offset, uint64(h2) + offset}

}

func (bffkvs *BFFKVS) Decode(key uint64, rawVal [][]uint64) ([]uint64, bool) {
	v0, v1, v2 := rawVal[0], rawVal[1], rawVal[2]
	W := len(v0)

	buf := make([]uint64, W)

	for i := 0; i < W; i++ {
		buf[i] = v0[i] ^ v1[i] ^ v2[i]
	}

	return xorfilter.KVSFingerPrintInv[TBFF](key, buf)
}

func (bffkvs *BFFKVS) Contains(outkey uint64, key uint64, db *utils.EncodedDB) ([]uint64, bool) {
	indexes := bffkvs.Lookup(outkey, key)
	rawVal := db.GetBatchEntry(indexes)
	return bffkvs.Decode(key, rawVal)
}

func (bffkvs *BFFKVS) Free() {

}

func (bffkvs *BFFKVS) GetBatchSize() uint64 {
	return bffkvs.BatchSize
}
