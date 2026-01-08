package kvs

import utils "github.com/local/utils"

type KVS interface {
	Name() string
	Size() uint64
	Free()

	Encode(kv *utils.KV) utils.EncodedDB

	Index(bucketID uint64, key uint64) []uint64

	Decode(key uint64, rawVal [][]uint64) ([]uint64, bool)

	GetBatchSize() uint64
}
