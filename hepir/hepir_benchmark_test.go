package hepir

import (
	"flag"
	"fmt"
	"testing"
	"time"

	simplepir "github.com/local/simplepir"
	utils "github.com/local/utils"
)

var (
	logNumEntries uint64
	bitsPerVal    uint64
	batchSize     uint64
	pirID         string
)

func init() {
	flag.Uint64Var(&logNumEntries, "logN", 20, "Log2 of number of entries")
	flag.Uint64Var(&bitsPerVal, "bitsPerVal", 32, "Number of bits per entry")
	flag.Uint64Var(&batchSize, "batch", 1, "Batch size for PIR query")
	flag.StringVar(&pirID, "pirID", "simplepir", "Scheme type: simplepir, doublepir")
}

// go test -bench=BenchmarkHepir -benchmem -run=none -v -args -logN=20 -bitsPerVal=32 -batch=1 -sipirID=simplepir

// go test -bench=BenchmarkHepir -benchmem -run=none -v
func BenchmarkHepir(b *testing.B) {
	if !flag.Parsed() {
		flag.Parse()
	}

	var hepir simplepir.HEPIR

	switch pirID {
	case "doublepir":
		hepir = &simplepir.DoublePIR{}
	case "simplepir":
		hepir = &simplepir.SimplePIR{}
	default:
		b.Fatalf("error Scheme: piano, singleserver, singlepass")
	}

	numEntries := uint64(1 << logNumEntries)
	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerVal)

	hepir.InitParams(numEntries, bitsPerVal)

	var totalQueryTime, totalAnswerTime, totalReconTime time.Duration
	var totalQuerySize, totalAnswerSize float64
	var numQueries uint64 = 10

	// --- A. Setup / Preprocessing ---
	fmt.Println("Starting Setup...")
	startSetup := time.Now()
	db.Random()
	internalDB := hepir.MakeInternalDB(&db)
	serverHint, clientHint := hepir.Setup(internalDB)

	setupDuration := time.Since(startSetup)
	clientHintSizeMB := float64(clientHint.Size()) / (1024 * 1024)
	serverHintSizeMB := float64(serverHint.Size()) / (1024 * 1024)

	fmt.Println("********* HEPIR Batch Query Start **********")

	for q := uint64(0); q < numQueries; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		// --- B. Query (Client) ---
		startQuery := time.Now()
		req, clientState := hepir.Query(targetIndexes)
		totalQueryTime += time.Since(startQuery)
		totalQuerySize += float64(req.Size())

		// --- C. Answer (Server) ---
		startAnswer := time.Now()
		resp := hepir.Answer(internalDB, req, serverHint)
		totalAnswerTime += time.Since(startAnswer)
		totalAnswerSize += float64(resp.Size())

		// --- D. Reconstruct (Client) ---
		startRecon := time.Now()
		results := hepir.Reconstruct(targetIndexes, clientHint, req, resp, clientState)
		totalReconTime += time.Since(startRecon)

		if !db.BatchEntryEqualsData(targetIndexes, results) {
			fmt.Printf("HEPIR Batch PIR Failed at query %d\n", q)
		}
	}

	fmt.Printf("SIPIR Name(): %s\n", hepir.Name())
	fmt.Printf("HEPIR Evaluation (SimplePIR) Results (N=2^%d)\n", logNumEntries)
	fmt.Printf("1. Setup (Preprocessing):    %v (clinetHint: %.2f MB, serverHint: %.2f MB)\n", setupDuration, clientHintSizeMB, serverHintSizeMB)

	avgQ := float64(totalQueryTime.Nanoseconds()) / float64(numQueries)
	avgA := float64(totalAnswerTime.Nanoseconds()) / float64(numQueries)
	avgR := float64(totalReconTime.Nanoseconds()) / float64(numQueries)

	fmt.Printf("2. Client Query (Avg):       %.2f us (Up: %.2f KB)\n", avgQ/1000, totalQuerySize/float64(numQueries)/1024)
	fmt.Printf("3. Server Answer (Avg):      %.2f us (Down: %.2f KB)\n", avgA/1000, totalAnswerSize/float64(numQueries)/1024)
	fmt.Printf("4. Client Reconstruct (Avg): %.2f us\n", avgR/1000)

}
