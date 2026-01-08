package main

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/local/hepir/simplepir"
	"github.com/local/kvs"
	"github.com/local/sipir"
	"github.com/local/utils"
)

// go test kpir_test.go -v -run TestKeywordSipirRewind
func TestKeywordSipirRewind(t *testing.T) {
	logNumsKeys := uint64(20)
	bitsPerVal := uint64(32)

	//saveToDisk := false
	batchtype := "rewind"
	sipir := sipir.SingleServer{} //Piano{} //SingleServer{} //SinglePass{}
	kvs := kvs.NewPTHashKVS()     //BFFKVS{} // NewConsensusRecSplitKVS() //NewPTHashKVS() //NewBBHashKVS()
	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()

	kv.Sort()
	db := kvs.Encode(kv)

	sipir.InitParams(db.NumEntries, db.Uint64PerEntry, batchtype)
	sipir.GenerateHint(&db)

	batchSize := kvs.BatchSize
	fmt.Println("batchSize: ", batchSize, " sipir.Params.Q: ", sipir.Params.Q)

	for q := uint64(0); q < sipir.Params.Q/batchSize; q++ {

		outkey := rand.Uint64() % kv.BucketCount
		innkeyIndex := rand.Uint64() % kv.Buckets[outkey].TotalKeys
		innkey := kv.Buckets[outkey].Keys[innkeyIndex]

		targetIndexes := kvs.Index(outkey, innkey)
		//fmt.Println("outkey: ", outkey, " innkey: ", innkey, " targetIndexes: ", targetIndexes)

		// Batch IndexPIR Start

		req, state := sipir.QueryAndFakeRefresh(targetIndexes)
		//fmt.Printf("Sent Batch Query: %d indexes. Message Size: %.2f bytes\n", batchSize, req.(sipirMessage).Size())

		resp := sipir.Answer(&db, req)

		//rewind
		sipir.Rewind(state)

		results := sipir.ReconstructAndRefresh(resp, state, targetIndexes)
		// Batch IndexPIR End

		val, _ := kvs.Decode(innkey, results)

		_, march := kv.GetValAndComp(outkey, innkey, val)
		if !march {
			fmt.Println("Keyword SIPIR Failed")
		}

	}
	fmt.Println("Keyword SIPIR finished successfully")
}

// go test kpir_test.go -v -run TestKeywordSipirSkip
func TestKeywordSipirSkip(t *testing.T) {
	logNumsKeys := uint64(20)
	bitsPerVal := uint64(32)

	//saveToDisk := false
	batchtype := "skip"
	sipir := sipir.Piano{}
	kvs := kvs.NewPTHashKVS() //BFFKVS{} // NewConsensusRecSplitKVS() //NewPTHashKVS() //NewBBHashKVS()
	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()

	kv.Sort()
	db := kvs.Encode(kv)

	sipir.InitParams(db.NumEntries, db.Uint64PerEntry, batchtype)
	sipir.GenerateHint(&db)

	batchSize := kvs.BatchSize
	fmt.Println("batchSize: ", batchSize, " sipir.Params.Q: ", sipir.Params.Q)

	for q := uint64(0); q < sipir.Params.Q/batchSize; q++ {
		outkey := rand.Uint64() % kv.BucketCount
		innkeyIndex := rand.Uint64() % kv.Buckets[outkey].TotalKeys
		innkey := kv.Buckets[outkey].Keys[innkeyIndex]

		targetIndexes := kvs.Index(outkey, innkey)
		//fmt.Println("targetIndexes: ", targetIndexes)

		// Batch IndexPIR Start

		req, state := sipir.Query(targetIndexes)
		//fmt.Printf("Sent Batch Query: %d indexes. Message Size: %.2f bytes\n", batchSize, req.(sipirMessage).Size())

		resp := sipir.Answer(&db, req)

		results := sipir.Reconstruct(resp, state)

		sipir.Refresh(targetIndexes, results, state)

		// Batch IndexPIR End

		val, _ := kvs.Decode(innkey, results)

		_, march := kv.GetValAndComp(outkey, innkey, val)
		if !march {
			fmt.Println("Keyword SIPIR Failed")
		}
	}

	fmt.Println("Keyword SIPIR finished successfully")

}

// go test kpir_test.go -v -run TestKeywordHepir
func TestKeywordHepir(t *testing.T) {
	logNumsKeys := uint64(20)
	bitsPerVal := uint64(32)

	//saveToDisk := false
	hepir := simplepir.DoublePIR{} //simplepir.DoublePIR{} //simplepir.SimplePIR{}
	kvs := kvs.NewPTHashKVS()      //BFFKVS{} // NewConsensusRecSplitKVS() //NewPTHashKVS() //NewBBHashKVS()
	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()

	kv.Sort()
	db := kvs.Encode(kv)

	hepir.InitParams(db.NumEntries, db.Uint64PerEntry)
	internalDB := hepir.MakeInternalDB(&db)
	serverHint, clientHint := hepir.Setup(internalDB)

	batchSize := kvs.BatchSize
	fmt.Println("batchSize: ", batchSize)

	for q := uint64(0); q < 100; q++ {

		outkey, innkey := kv.GenRandomKey()

		targetIndexes := kvs.Index(outkey, innkey)
		//fmt.Println("targetIndexes: ", targetIndexes)

		// Batch IndexPIR Start

		req, clientState := hepir.Query(targetIndexes)
		//clientState, EncZero := hepir.QueryOffline(int(batchSize))
		//req := hepir.QueryOnline(targetIndexes, EncZero)

		resp := hepir.Answer(internalDB, req, serverHint)

		results := hepir.Reconstruct(targetIndexes, clientHint, req, resp, clientState)

		// Batch IndexPIR End

		val, _ := kvs.Decode(innkey, results)
		//fmt.Println("val: ", val)
		_, march := kv.GetValAndComp(outkey, innkey, val)
		if !march {
			fmt.Println("Keyword HEPIR Failed")
		}

	}
	fmt.Println("Keyword HEPIR finished successfully")

}
