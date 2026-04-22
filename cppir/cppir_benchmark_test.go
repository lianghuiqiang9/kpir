package cppir

import (
	"flag"
	"fmt"
	"testing"
	"time"

	utils "github.com/local/utils"
)

var (
	logNumEntries uint64
	bitsPerEntry  uint64
	batchSize     uint64
	batchtype     string
	pirID         string
)

func init() {
	flag.Uint64Var(&logNumEntries, "logN", 20, "Log2 of number of entries")
	flag.Uint64Var(&bitsPerEntry, "bitsPerEntry", 32, "Number of bits per entry")
	flag.Uint64Var(&batchSize, "batch", 1, "Batch size for PIR query")
	flag.StringVar(&batchtype, "type", "skip", "Batch type: skip or rewind")
	flag.StringVar(&pirID, "pirID", "singleserver", "Scheme type: piano, singleserver, singlepass")
}

// go test -bench=BenchmarkSkip -benchmem -run=none -v -args -logN=20 -bitsPerEntry=32 -batch=1 -type="skip"
// go test -bench=BenchmarkSkip -benchmem -run=none -v -args -logN=25 -bitsPerEntry=32 -batch=1 -type="skip"

// go test -bench=BenchmarkSkip -benchmem -run=none -v
func BenchmarkSkip(b *testing.B) {
	if !flag.Parsed() {
		flag.Parse()
	}

	cppir := Piano{}

	db := utils.EncodedDB{}
	numEntries := uint64(1 << logNumEntries)

	db.InitParams(numEntries, bitsPerEntry)
	cppir.InitParams(numEntries, bitsPerEntry, batchtype)
	db.Random()

	dbSizeMB := float64(db.Size()) / (1024 * 1024)

	var totalQueryTime, totalAnswerTime, totalReconTime time.Duration
	var totalQuerySize, totalAnswerSize float64
	var numQueries uint64

	// --- A. Preprocessing (GenerateHint) ---
	startHint := time.Now()
	cppir.GenerateHint(&db)
	generateHintDuration := time.Since(startHint)
	hintSizeMB := float64(cppir.Hint.Size()) / (1024 * 1024)

	fmt.Println("********* CPPIR Batch Query Start **********")

	maxQueries := cppir.Params.Q / batchSize

	for q := uint64(0); q < maxQueries; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		// --- B. Query (Client) ---
		startQuery := time.Now()
		req, state := cppir.Query(targetIndexes)
		totalQueryTime += time.Since(startQuery)
		totalQuerySize += float64(req.Size())

		// --- C. Answer (Server) ---
		startAnswer := time.Now()
		resp := cppir.Answer(&db, req)
		totalAnswerTime += time.Since(startAnswer)
		totalAnswerSize += float64(resp.Size())

		// --- D. Reconstruct & Refresh (Client) ---
		startRecon := time.Now()
		results := cppir.Reconstruct(resp, state)
		cppir.Refresh(targetIndexes, results, state)
		totalReconTime += time.Since(startRecon)

		if !db.BatchEntryEqualsData(targetIndexes, results) {
			fmt.Println("CPPIR Batch PIR Failed at query", q)
		}
		numQueries++
	}

	fmt.Printf("CPPIR Name(): %s, batchtype: %s\n", cppir.Name(), batchtype)
	fmt.Printf("CPPIR Evaluation Results (N=2^%d, w=%d, Batch=%d)\n", logNumEntries, bitsPerEntry, batchSize)
	fmt.Printf("1. Preprocessing (Hint Gen): %v (Size: %.4f MB, Offline Communication: %.4f MB), maxQuery: %d\n", generateHintDuration, hintSizeMB, dbSizeMB, maxQueries)

	avgQ := float64(totalQueryTime.Nanoseconds()) / float64(numQueries)
	avgA := float64(totalAnswerTime.Nanoseconds()) / float64(numQueries)
	avgR := float64(totalReconTime.Nanoseconds()) / float64(numQueries)

	fmt.Printf("2. Client Query (Avg):       %.4f us (Size: %.4f KBytes)\n", avgQ/1000, totalQuerySize/float64(numQueries)/1024)
	fmt.Printf("3. Server Answer (Avg):      %.4f us (Size: %.4f KBytes)\n", avgA/1000, totalAnswerSize/float64(numQueries)/1024)
	fmt.Printf("4. Client Reconstruct (Avg): %.4f us\n", avgR/1000)

}

// go test -bench=BenchmarkRewind -benchmem -run=none -v
func BenchmarkRewind(b *testing.B) {
	if !flag.Parsed() {
		flag.Parse()
	}

	var cppir CPPIR

	switch pirID {
	case "piano":
		cppir = &Piano{}
	case "singleserver":
		cppir = &SingleServer{}
	case "singlepass":
		cppir = &SinglePass{}
	default:
		b.Fatalf("error Scheme: piano, singleserver, singlepass")
	}

	numEntries := uint64(1 << logNumEntries)

	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerEntry)
	cppir.InitParams(numEntries, bitsPerEntry, batchtype)
	db.Random()

	dbSizeMB := float64(db.Size()) / (1024 * 1024)

	var totalQueryTime, totalAnswerTime, totalReconTime time.Duration
	var totalQuerySize, totalAnswerSize float64
	var numQueries uint64

	// --- A. Preprocessing (GenerateHint) ---
	startHint := time.Now()
	cppir.GenerateHint(&db)
	generateHintDuration := time.Since(startHint)
	hintSizeMB := cppir.GetHintSize() / (1024 * 1024)

	fmt.Println("********* CPPIR Batch Query Start **********")

	maxQueries := cppir.GetParamQ() / batchSize

	for q := uint64(0); q < maxQueries; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		// --- B. Query (Client) ---
		startQuery := time.Now()
		req, state := cppir.QueryAndFakeRefresh(targetIndexes)
		totalQueryTime += time.Since(startQuery)
		totalQuerySize += float64(req.Size())

		// --- C. Answer (Server) ---
		startAnswer := time.Now()
		resp := cppir.Answer(&db, req)
		totalAnswerTime += time.Since(startAnswer)
		totalAnswerSize += float64(resp.Size())

		// --- D. Reconstruct & Refresh (Client) ---
		startRecon := time.Now()
		cppir.Rewind(state)
		results := cppir.ReconstructAndRefresh(resp, state, targetIndexes)
		totalReconTime += time.Since(startRecon)

		if !db.BatchEntryEqualsData(targetIndexes, results) {
			fmt.Println("CPPIR Batch PIR Failed at query", q)
		}
		numQueries++
	}

	fmt.Printf("CPPIR Name(): %s, batchtype: %s\n", cppir.Name(), batchtype)
	fmt.Printf("CPPIR Evaluation Results (N=2^%d, w=%d, Batch=%d)\n", logNumEntries, bitsPerEntry, batchSize)
	fmt.Printf("1. Preprocessing (Hint Gen): %v (Size: %.4f MB, Offline Communication: %.4f MB), maxQuery: %d\n", generateHintDuration, hintSizeMB, dbSizeMB, maxQueries)

	avgQ := float64(totalQueryTime.Nanoseconds()) / float64(numQueries)
	avgA := float64(totalAnswerTime.Nanoseconds()) / float64(numQueries)
	avgR := float64(totalReconTime.Nanoseconds()) / float64(numQueries)

	fmt.Printf("2. Client Query (Avg):       %.4f us (Size: %.4f KBytes)\n", avgQ/1000, totalQuerySize/float64(numQueries)/1024)
	fmt.Printf("3. Server Answer (Avg):      %.4f us (Size: %.4f KBytes)\n", avgA/1000, totalAnswerSize/float64(numQueries)/1024)
	fmt.Printf("4. Client Reconstruct (Avg): %.4f us\n", avgR/1000)

}
