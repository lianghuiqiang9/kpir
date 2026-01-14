package sipir

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
	sipir := Piano{}
	batchtype := "skip"

	numEntries := uint64(1 << logNumEntries)
	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerEntry)
	sipir.InitParams(numEntries, bitsPerEntry, batchtype)
	fmt.Printf("%s, initial Params: %+v\n", sipir.Name(), sipir.Params)

	if saveToDisk {
		db.LoadDB("", "")
		sipir.LoadHint("", "", &db)
	} else {
		db.Random()
		sipir.GenerateHint(&db)
	}

	fmt.Println("********* SIPIR Batch Query Start **********")

	for q := uint64(0); q < sipir.Params.Q/batchSize; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		req, state := sipir.Query(targetIndexes)

		resp := sipir.Answer(&db, req)

		results := sipir.Reconstruct(resp, state)
		sipir.Refresh(targetIndexes, results, state)

		flag := db.BatchEntryEqualsData(targetIndexes, results)
		if flag == false {
			fmt.Println("SIPIR Batch PIR Failed")
		}
	}
	fmt.Println("SIPIR Batch PIR finished successfully")
}

// go test -run TestRewind
func TestRewind(t *testing.T) {
	logNumEntries := uint64(20)
	bitsPerEntry := uint64(96) // *64 bits
	batchSize := uint64(3)
	saveToDisk := false
	batchtype := "rewind"
	sipir := SinglePass{} //Piano{} //SingleServer{} //SinglePass{}

	numEntries := uint64(1 << logNumEntries)
	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerEntry)
	sipir.InitParams(numEntries, bitsPerEntry, batchtype)
	fmt.Printf("%s, initial Params: %+v\n", sipir.Name(), sipir.Params)

	if saveToDisk {
		db.LoadDB("", "")
		sipir.LoadHint("", "", &db)
	} else {
		db.Random()
		sipir.GenerateHint(&db)
	}

	fmt.Println("********* SIPIR Batch Query Start **********")

	for q := uint64(0); q < sipir.Params.Q/batchSize; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		req, state := sipir.QueryAndFakeRefresh(targetIndexes)

		resp := sipir.Answer(&db, req)

		//rewind
		sipir.Rewind(state)
		results := sipir.ReconstructAndRefresh(resp, state, targetIndexes)

		flag := db.BatchEntryEqualsData(targetIndexes, results)
		if flag == false {
			fmt.Println("SIPIR Batch PIR Failed")
		}
	}
	fmt.Println("SIPIR Batch PIR finished successfully")
}
