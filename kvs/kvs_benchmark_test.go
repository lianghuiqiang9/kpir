package kvs

import (
	"flag"
	"fmt"
	"testing"
	"time"

	utils "github.com/local/utils"
)

var (
	logNumsKeys uint64
	bitsPerVal  uint64
	kvsID       string
)

func init() {
	flag.Uint64Var(&logNumsKeys, "logN", 20, "Log2 of number of entries")
	flag.Uint64Var(&bitsPerVal, "bitsPerVal", 32, "Number of bits per entry")
	flag.StringVar(&kvsID, "kvsID", "pthashkvs", "KVS type: pthashkvs, bbhashkvs, bbhash2kvs, consensusrecsplitkvs, bffkvs")
}

// go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=20 -bitsPerVal=32 -kvsID=bffkvs
// go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=20 -bitsPerVal=32 -kvsID=pthashkvs
// go test -bench=BenchmarkKVStore -benchmem -run=none -v
func BenchmarkKVStore(b *testing.B) {
	if !flag.Parsed() {
		flag.Parse()
	}

	var kvs KVS

	switch kvsID {
	case "pthashkvs":
		kvs = NewPTHashKVS()
	case "bbhashkvs":
		kvs = NewBBHashKVS()
	case "bbhash2kvs":
		kvs = &BBHash2KVS{}
	case "bffkvs":
		kvs = &BFFKVS{}
	case "consensusrecsplitkvs":
		kvs = NewConsensusRecSplitKVS()
	default:
		kvs = NewPTHashKVS()
	}

	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()
	defer kvs.Free()

	startEncode := time.Now()
	kv.Sort()
	sortDuration := time.Since(startEncode)

	startEncode = time.Now()
	db := kvs.Encode(kv)
	encodeDuration := time.Since(startEncode)

	// --- 3.  Index (Online - Mapping) ---
	testBucket := uint64(0)
	testKey := kv.Buckets[testBucket].Keys[0]

	b.Run("KVS_Index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = kvs.Lookup(testBucket, testKey)
		}
	})

	// --- 4.  Decode (Online - Reconstruction) ---
	indexes := kvs.Lookup(testBucket, testKey)
	rawVal := db.GetBatchEntry(indexes)

	b.Run("KVS_Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = kvs.Decode(testKey, rawVal)
		}
	})

	pass := true
	for i := uint64(0); i < kv.BucketCount; i++ {
		for _, key := range kv.Buckets[i].Keys[:1000] {
			idx := kvs.Lookup(i, key)
			rv := db.GetBatchEntry(idx)
			val, _ := kvs.Decode(key, rv)
			_, flag := kv.GetValAndComp(i, key, val)
			if !flag {
				pass = false
				break
			}
		}
	}

	fmt.Printf("KVStore Benchmark %s (N=2^%d, bitsPerVal=%dbits)\n", kvs.Name(), logNumsKeys, bitsPerVal)
	fmt.Printf("Sort Time:   %v\n", sortDuration)
	fmt.Printf("Encode Time:   %v\n", encodeDuration)
	fmt.Printf("KVS Hint Size: %.2f KB\n", float64(kvs.Size())/1024)
	fmt.Printf("KVS Ratio : %.3f \n", float64(db.NumEntries)/float64(uint64(1<<logNumsKeys)))
	fmt.Printf("Correctness:   %v\n", pass)
}
