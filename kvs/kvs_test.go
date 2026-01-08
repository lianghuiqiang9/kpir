package kvs

import (
	"fmt"
	"math/rand/v2"
	"testing"

	xorfilter "github.com/FastFilter/xorfilter"

	"github.com/stretchr/testify/assert"

	bbhash "github.com/local/bbhash"
	consensusrecsplit "github.com/local/consensusrecsplit"
	pthash "github.com/local/pthash"
	utils "github.com/local/utils"
	bbhash2 "github.com/relab/bbhash"
)

// go test -run TestBinaryFuseBasic
func TestBinaryFuseBasic(t *testing.T) {
	NumKeys := uint64(1 << 20)
	keys := make([]uint64, NumKeys)
	for i := range keys {
		keys[i] = rand.Uint64()
	}
	filter, _ := xorfilter.NewBinaryFuse[uint32](keys)
	for _, v := range keys {
		assert.Equal(t, true, filter.Contains(v))
	}
	falsesize := 10000000
	matches := 0
	bpv := float64(len(filter.Fingerprints)) * 32.0 / float64(NumKeys)
	fmt.Println("Binary Fuse filter:")
	fmt.Println("bits per entry ", bpv)
	for i := 0; i < falsesize; i++ {
		v := rand.Uint64()
		if filter.Contains(v) {
			matches++
		}
	}
	fpp := float64(matches) * 100.0 / float64(falsesize)
	fmt.Println("false positive rate ", fpp)
}

// go test -run TestBBHash2SlowLookup
func TestBBHash2SlowLookup(t *testing.T) {
	NumKeys := uint64(1 << 20)
	keys := utils.GenerateUniqueUint64s(int(NumKeys))
	bb, _ := bbhash2.New(keys, bbhash2.Gamma(1), bbhash2.WithReverseMap())
	for _, key := range keys[:100] {
		idx := bb.Find(key)

		idx++

		//fmt.Println("idx: ", idx, " key: ", key)
	}
	keySort := make([]uint64, NumKeys+1)
	for i := uint64(1); i < NumKeys+1; i++ {
		keySort[i] = bb.Key(uint64(i))
	}

	size, _ := utils.SaveToFile(bb, "bbhash2.dat")
	fmt.Printf("BBHash saved to file, size = %d bits, radio = %f bits/key\n", size*8, float64(size)/float64(NumKeys)*8.0)
}

// go test -run TestConsensusRecSplit
func TestConsensusRecSplit(t *testing.T) {
	NumKeys := uint64(1 << 20)
	keys := utils.GenerateUniqueUint64s(int(NumKeys))

	mph := consensusrecsplit.New(keys)

	defer mph.Free()

	for _, k := range keys {
		idx := mph.Lookup(k)
		fmt.Println("idx: ", idx)
	}

	fmt.Printf("ConsensusRecSplit size = %.4f bits,  radio = %.4f bits/key\n", float64(mph.Bits()), float64(mph.Bits())/float64(NumKeys))
}

// go test -run TestBBHashFastLookup
func TestBBHashFastLookup(t *testing.T) {
	NumKeys := uint64(1 << 20)
	keys := utils.GenerateUniqueUint64s(int(NumKeys))

	bb := bbhash.New(keys)

	defer bb.Free()

	for _, k := range keys {
		idx := bb.Lookup(k)
		//fmt.Println("idx: ", idx)
		idx++
	}
	bbbits := bb.Bits()
	fmt.Printf("BBHash size = %.4f bits,  radio = %.4f bits/key\n", float64(bbbits), float64(bbbits)/float64(NumKeys))
}

// go test -run TestPTHash
func TestPTHash(t *testing.T) {
	NumKeys := uint64(1 << 20)
	keys := utils.GenerateUniqueUint64s(int(NumKeys))

	bb := pthash.New(keys)

	defer bb.Free()

	for _, k := range keys {
		idx := bb.Lookup(k)
		//fmt.Println("idx: ", idx)
		idx++
	}
	bbbits := bb.Bits()
	fmt.Printf("BBHash size = %.4f bits,  radio = %.4f bits/key\n", float64(bbbits), float64(bbbits)/float64(NumKeys))
}

// go test -run TestBucketStore
func TestBucketStore(t *testing.T) {
	bucket := &utils.Bucket{}
	bucket.Setup(1<<20, 160)
	bucket.Random()

	// 每桶只看前 3 条数据
	bucket.Print(3)
	val, flag := bucket.GetVal(14310880194869718394)
	fmt.Println("", val, flag)

	kvs := BBHash2KVS{} //BFFKVS{} // NewConsensusRecSplitKVS() //NewPTHashKVS() //NewBBHashKVS()
	//PTHashKVS{} //BFFKVS{} BBHashKVS{} BBHash2KVS{} PTHashKVS{} ConsensusRecSplitKVS{}
	defer kvs.Free()

	bucket.Sort()
	db := kvs.EncodeBucket(bucket)
	bucket.Print(3)

	for _, key := range bucket.Keys {
		indexes := kvs.Index(0, key)

		rawVal := db.GetBatchEntry(indexes)

		val, flag := kvs.Decode(key, rawVal)

		val2, flag2 := bucket.GetVal(key)
		assert.Equal(t, true, flag)
		assert.Equal(t, true, flag2)
		assert.Equal(t, true, val[0] == val2[0])
		//fmt.Println("key: ", key, " val: ", val, " flag: ", flag)
		//fmt.Println("key: ", key, " val2: ", val2, " flag2: ", flag2)
	}

	falsesize := 1000
	matches := 0
	fmt.Println("KVS:")
	for i := 0; i < falsesize; i++ {
		k := rand.Uint64()
		_, flag := kvs.Contains(0, k, &db)
		if flag {

			matches++
		}
	}
	fpp := float64(matches) * 100.0 / float64(falsesize)
	fmt.Println("false positive rate ", fpp)
}

// go test -run TestKVStore
func TestKVStore(t *testing.T) {

	kv := &utils.KV{}
	kv.Setup(1<<20, 32)
	kv.Random()
	fmt.Println("Random done")
	kv.Print(3)

	kvs := NewPTHashKVS() //BFFKVS{} // NewConsensusRecSplitKVS() //NewPTHashKVS() //NewBBHashKVS()
	//PTHashKVS{} //BFFKVS{} BBHashKVS{} BBHash2KVS{} PTHashKVS{} ConsensusRecSplitKVS{}
	defer kvs.Free()

	kv.Sort()
	db := kvs.Encode(kv)
	fmt.Println("Encode done")

	fmt.Println("Sort done")

	for i := uint64(0); i < kv.BucketCount; i++ {
		for _, key := range kv.Buckets[i].Keys[:1000] {
			indexes := kvs.Index(i, key)

			rawVal := db.GetBatchEntry(indexes)

			val, flag := kvs.Decode(key, rawVal)

			val2, flag2 := kv.GetVal(i, key)
			assert.Equal(t, true, flag)
			assert.Equal(t, true, flag2)
			assert.Equal(t, true, val[0] == val2[0])
			//fmt.Println("key: ", key, " val: ", val, " flag: ", flag)
			//fmt.Println("key: ", key, " val2: ", val2, " flag2: ", flag2)

		}
	}

	falsesize := 1000
	matches := 0
	fmt.Println("KVS:")
	for i := uint64(0); i < kv.BucketCount; i++ {
		for j := 0; j < falsesize; j++ {
			k := rand.Uint64()
			_, flag := kvs.Contains(i, k, &db)
			if flag {

				matches++
			}
		}
	}
	fpp := float64(matches) * 100.0 / float64(falsesize)
	fmt.Println("false positive rate ", fpp)
}
