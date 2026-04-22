package cppir

import (
	"fmt"
	"testing"

	utils "github.com/local/utils"
)

// go test -run TestSkip
func TestSkip(t *testing.T) {
	logNumEntries := uint64(20)
	bitsPerEntry := uint64(32)
	batchSize := uint64(3)
	saveToDisk := false
	cppir := Piano{}
	batchtype := "skip"

	numEntries := uint64(1 << logNumEntries)
	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerEntry)
	cppir.InitParams(numEntries, bitsPerEntry, batchtype)
	fmt.Printf("%s, initial Params: %+v\n", cppir.Name(), cppir.Params)

	if saveToDisk {
		db.LoadDB("", "")
		cppir.LoadHint("", "", &db)
	} else {
		db.Random()
		cppir.GenerateHint(&db)
	}

	fmt.Println("********* CPPIR Batch Query Start **********")

	for q := uint64(0); q < cppir.Params.Q/batchSize; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		req, state := cppir.Query(targetIndexes)

		resp := cppir.Answer(&db, req)

		results := cppir.Reconstruct(resp, state)
		cppir.Refresh(targetIndexes, results, state)

		flag := db.BatchEntryEqualsData(targetIndexes, results)
		if flag == false {
			fmt.Println("CPPIR Batch PIR Failed")
		}
	}
	fmt.Println("CPPIR Batch PIR finished successfully")
}

// go test -run TestRewind
func TestRewind(t *testing.T) {
	logNumEntries := uint64(20)
	bitsPerEntry := uint64(96) // *64 bits
	batchSize := uint64(3)
	saveToDisk := false
	batchtype := "rewind"
	cppir := SinglePass{} //Piano{} //SingleServer{} //SinglePass{}

	numEntries := uint64(1 << logNumEntries)
	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerEntry)
	cppir.InitParams(numEntries, bitsPerEntry, batchtype)
	fmt.Printf("%s, initial Params: %+v\n", cppir.Name(), cppir.Params)

	if saveToDisk {
		db.LoadDB("", "")
		cppir.LoadHint("", "", &db)
	} else {
		db.Random()
		cppir.GenerateHint(&db)
	}

	fmt.Println("********* CPPIR Batch Query Start **********")

	for q := uint64(0); q < cppir.Params.Q/batchSize; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		req, state := cppir.QueryAndFakeRefresh(targetIndexes)

		resp := cppir.Answer(&db, req)

		//rewind
		cppir.Rewind(state)
		results := cppir.ReconstructAndRefresh(resp, state, targetIndexes)

		flag := db.BatchEntryEqualsData(targetIndexes, results)
		if flag == false {
			fmt.Println("CPPIR Batch PIR Failed")
		}
	}
	fmt.Println("CPPIR Batch PIR finished successfully")
}
