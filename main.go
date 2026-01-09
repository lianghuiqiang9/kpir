package main

import (
	"fmt"

	"github.com/local/kvs"
	"github.com/local/utils"
)

func main() {
	kv := &utils.KV{}
	kv.Setup(1<<20, 32)
	kv.Random()
	kv.Sort()

	kvs := kvs.BBHash2KVS{} //BFFKVS{}
	db := kvs.Encode(kv)
	for i := uint64(0); i < kv.BucketCount; i++ {
		for j, key := range kv.Buckets[i].Keys {
			indexes := kvs.Index(i, key)

			rawVal := db.GetBatchEntry(indexes)

			val, flag := kvs.Decode(key, rawVal)

			val2, flag2 := kv.GetVal(i, key)

			if j%(1<<16) == 0 {
				fmt.Println("j: ", j, " key: ", key, " val: ", val, " flag: ", flag)
				fmt.Println("j: ", j, " key: ", key, " val2: ", val2, " flag2: ", flag2)
			}

		}
	}
}
