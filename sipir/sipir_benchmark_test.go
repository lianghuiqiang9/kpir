package sipir

import (
	"flag"
	"fmt"
	"testing"
	"time"

	utils "github.com/local/utils"
)

var (
	logNumEntries uint64
	bitsPerVal    uint64
	batchSize     uint64
	batchtype     string
	sipirID       string
)

func init() {
	flag.Uint64Var(&logNumEntries, "logN", 20, "Log2 of number of entries")
	flag.Uint64Var(&bitsPerVal, "bitsPerVal", 32, "Number of bits per value")
	flag.Uint64Var(&batchSize, "batch", 1, "Batch size for PIR query")
	flag.StringVar(&batchtype, "type", "skip", "Batch type: skip or rewind")
	flag.StringVar(&sipirID, "sipirID", "singleserver", "Scheme type: piano, singleserver, singlepass")
}

// go test -bench=BenchmarkSkip -benchmem -run=none -v -args -logN=20 -perEntry=1 -batch=1 -type="skip"
// go test -bench=BenchmarkSkip -benchmem -run=none -v -args -logN=25 -perEntry=1 -batch=1 -type="skip"

// go test -bench=BenchmarkSkip -benchmem -run=none -v
func BenchmarkSkip(b *testing.B) {
	if !flag.Parsed() {
		flag.Parse()
	}

	sipir := Piano{}

	db := utils.EncodedDB{}
	numEntries := uint64(1 << logNumEntries)

	db.InitParams(numEntries, bitsPerVal)
	sipir.InitParams(numEntries, bitsPerVal, batchtype)
	db.Random()

	var totalQueryTime, totalAnswerTime, totalReconTime time.Duration
	var totalQuerySize, totalAnswerSize float64
	var numQueries uint64

	// --- A. Preprocessing (GenerateHint) ---
	startHint := time.Now()
	sipir.GenerateHint(&db)
	generateHintDuration := time.Since(startHint)
	hintSizeMB := float64(sipir.Hint.Size()) / (1024 * 1024)

	fmt.Println("********* SIPIR Batch Query Start **********")

	maxQueries := sipir.Params.Q / batchSize

	for q := uint64(0); q < maxQueries; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		// --- B. Query (Client) ---
		startQuery := time.Now()
		req, state := sipir.Query(targetIndexes)
		totalQueryTime += time.Since(startQuery)
		totalQuerySize += float64(req.Size())

		// --- C. Answer (Server) ---
		startAnswer := time.Now()
		resp := sipir.Answer(&db, req)
		totalAnswerTime += time.Since(startAnswer)
		totalAnswerSize += float64(resp.Size())

		// --- D. Reconstruct & Refresh (Client) ---
		startRecon := time.Now()
		results := sipir.Reconstruct(resp, state)
		sipir.Refresh(targetIndexes, results, state)
		totalReconTime += time.Since(startRecon)

		if !db.BatchEntryEqualsData(targetIndexes, results) {
			fmt.Println("SIPIR Batch PIR Failed at query", q)
		}
		numQueries++
	}

	fmt.Printf("SIPIR Name(): %s, batchtype: %s\n", sipir.Name(), batchtype)
	fmt.Printf("SIPIR Evaluation Results (N=2^%d, w=%d, Batch=%d)\n", logNumEntries, bitsPerVal, batchSize)
	fmt.Printf("1. Preprocessing (Hint Gen): %v (Size: %.4f MB), maxQuery: %d\n", generateHintDuration, hintSizeMB, maxQueries)

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

	var sipir SIPIR

	switch sipirID {
	case "piano":
		sipir = &Piano{}
	case "singleserver":
		sipir = &SingleServer{}
	case "singlepass":
		sipir = &SinglePass{}
	default:
		b.Fatalf("error Scheme: piano, singleserver, singlepass")
	}

	numEntries := uint64(1 << logNumEntries)

	db := utils.EncodedDB{}
	db.InitParams(numEntries, bitsPerVal)
	sipir.InitParams(numEntries, bitsPerVal, batchtype)
	db.Random()

	var totalQueryTime, totalAnswerTime, totalReconTime time.Duration
	var totalQuerySize, totalAnswerSize float64
	var numQueries uint64

	// --- A. Preprocessing (GenerateHint) ---
	startHint := time.Now()
	sipir.GenerateHint(&db)
	generateHintDuration := time.Since(startHint)
	hintSizeMB := sipir.GetHintSize() / (1024 * 1024)

	fmt.Println("********* SIPIR Batch Query Start **********")

	maxQueries := sipir.GetParamQ() / batchSize

	for q := uint64(0); q < maxQueries; q++ {
		targetIndexes := utils.GenRandomIndexes(batchSize, numEntries)

		// --- B. Query (Client) ---
		startQuery := time.Now()
		req, state := sipir.QueryAndFakeRefresh(targetIndexes)
		totalQueryTime += time.Since(startQuery)
		totalQuerySize += float64(req.Size())

		// --- C. Answer (Server) ---
		startAnswer := time.Now()
		resp := sipir.Answer(&db, req)
		totalAnswerTime += time.Since(startAnswer)
		totalAnswerSize += float64(resp.Size())

		// --- D. Reconstruct & Refresh (Client) ---
		startRecon := time.Now()
		sipir.Rewind(state)
		results := sipir.ReconstructAndRefresh(resp, state, targetIndexes)
		totalReconTime += time.Since(startRecon)

		if !db.BatchEntryEqualsData(targetIndexes, results) {
			fmt.Println("SIPIR Batch PIR Failed at query", q)
		}
		numQueries++
	}

	fmt.Printf("SIPIR Name(): %s, batchtype: %s\n", sipir.Name(), batchtype)
	fmt.Printf("SIPIR Evaluation Results (N=2^%d, w=%d, Batch=%d)\n", logNumEntries, bitsPerVal, batchSize)
	fmt.Printf("1. Preprocessing (Hint Gen): %v (Size: %.4f MB), maxQuery: %d\n", generateHintDuration, hintSizeMB, maxQueries)

	avgQ := float64(totalQueryTime.Nanoseconds()) / float64(numQueries)
	avgA := float64(totalAnswerTime.Nanoseconds()) / float64(numQueries)
	avgR := float64(totalReconTime.Nanoseconds()) / float64(numQueries)

	fmt.Printf("2. Client Query (Avg):       %.4f us (Size: %.4f KBytes)\n", avgQ/1000, totalQuerySize/float64(numQueries)/1024)
	fmt.Printf("3. Server Answer (Avg):      %.4f us (Size: %.4f KBytes)\n", avgA/1000, totalAnswerSize/float64(numQueries)/1024)
	fmt.Printf("4. Client Reconstruct (Avg): %.4f us\n", avgR/1000)

}
