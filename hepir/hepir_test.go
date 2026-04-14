package hepir

import (
	"fmt"
	"testing"

	"github.com/local/simplepir"

	"github.com/local/utils"
)

// go test -run TestHepir
func TestHepir(t *testing.T) {
	logNumEntries := uint64(20)
	bitsPerEntry := uint64(32)
	batchSize := uint64(1)
	saveToDisk := false
	hepir := simplepir.DoublePIR{} //simplepir.DoublePIR{} //simplepir.SimplePIR{}

	numEntries := uint64(1 << logNumEntries)
	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerEntry)

	hepir.InitParams(numEntries, bitsPerEntry)
	internalDB := &simplepir.InternalDB{}
	serverHint := &simplepir.State{}
	clientHint := &simplepir.State{}
	if saveToDisk {
		db.LoadDB("", "")
		internalDB = hepir.MakeInternalDB(&db)
		serverHint, clientHint = hepir.LoadHint("", "", internalDB)
	} else {
		db.Random()
		internalDB = hepir.MakeInternalDB(&db)
		*serverHint, *clientHint = hepir.Setup(internalDB)
	}

	fmt.Println("********* HEPIR Batch Query Start **********")

	for q := uint64(0); q < 10; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		//req, clientState := hepir.Query(targetIndexes)
		EncZero, clientState := hepir.QueryOffline(batchSize)
		req := hepir.QueryOnline(targetIndexes, EncZero)

		resp := hepir.Answer(internalDB, req, *serverHint)

		results := hepir.Reconstruct(targetIndexes, *clientHint, req, resp, clientState)

		flag := db.BatchEntryEqualsData(targetIndexes, results)
		if flag == false {
			fmt.Println("HEPIR Batch PIR Failed")
		}
	}
	fmt.Println("HEPIR Batch PIR finished successfully")
}

// multiply is multiply indexes one h1
