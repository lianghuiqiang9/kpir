package main

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/local/hepir/simplepir"
	KVS "github.com/local/kvs"
	"github.com/local/sipir"
	"github.com/local/utils"
)

var (
	logNumsKeys uint64
	bitsPerVal  uint64
	kvsID       string
	batchtype   string
	pirID       string
)

func init() {
	flag.Uint64Var(&logNumsKeys, "logN", 20, "Log2 of number of entries")
	flag.Uint64Var(&bitsPerVal, "bitsPerVal", 32, "Number of bits per value")
	flag.StringVar(&kvsID, "kvsID", "pthashkvs", "KVS type: pthashkvs, bbhashkvs, bbhash2kvs, consensusrecsplitkvs, bffkvs")
	flag.StringVar(&batchtype, "type", "skip", "Batch type: skip or rewind")
	flag.StringVar(&pirID, "pirID", "singleserver", "Scheme type: piano, singleserver, singlepass, simplepir, doublepir")
}

// go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -args -logN 20 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "singleserver" -type "rewind"
// go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v
func BenchmarkKeywordSipirRewind(b *testing.B) {

	var pir sipir.SIPIR

	switch pirID {
	case "piano":
		pir = &sipir.Piano{}
	case "singleserver":
		pir = &sipir.SingleServer{}
	case "singlepass":
		pir = &sipir.SinglePass{}
	default:
		b.Fatalf("error Scheme: piano, singleserver, singlepass")
	}

	var kvs KVS.KVS

	switch kvsID {
	case "pthashkvs":
		kvs = KVS.NewPTHashKVS()
	case "bbhashkvs":
		kvs = KVS.NewBBHashKVS()
	case "bbhash2kvs":
		kvs = &KVS.BBHash2KVS{}
	case "bffkvs":
		kvs = &KVS.BFFKVS{}
	case "consensusrecsplitkvs":
		kvs = KVS.NewConsensusRecSplitKVS()
	default:
		kvs = KVS.NewPTHashKVS()
	}
	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()
	kv.Sort()

	start := time.Now()
	db := kvs.Encode(kv)
	encodeDuration := time.Since(start)
	start = time.Now()
	pir.InitParams(db.NumEntries, db.BitsPerEntry, batchtype)
	pir.GenerateHint(&db)
	genHintDuration := time.Since(start)

	batchSize := kvs.GetBatchSize()

	var tMapping, tQuery, tServer, tRewindRecon, tDecode time.Duration
	var upMsgSize, downMsgSize float64
	num := 0.0
	maxQueries := pir.GetParamQ() / batchSize

	fmt.Println("********* Keyword SIPIR (Rewind) Benchmark Start **********")

	for q := uint64(0); q < maxQueries; q++ {

		outkey, innkey := kv.GenRandomKey()
		// --- A. Keyword Mapping (Client) ---
		start := time.Now()
		targetIndexes := kvs.Lookup(outkey, innkey)
		tMapping += time.Since(start)

		// --- B. Query Gen (Client) ---
		start = time.Now()
		req, state := pir.QueryAndFakeRefresh(targetIndexes)
		tQuery += time.Since(start)
		upMsgSize += float64(req.Size())

		// --- C. Answer (Server) ---
		start = time.Now()
		resp := pir.Answer(&db, req)
		tServer += time.Since(start)
		downMsgSize += float64(resp.Size())

		// --- D. Rewind & Reconstruct (Client) ---
		start = time.Now()
		pir.Rewind(state)
		results := pir.ReconstructAndRefresh(resp, state, targetIndexes)
		tRewindRecon += time.Since(start)

		// --- E. Final Decode (Client) ---
		start = time.Now()
		val, _ := kvs.Decode(innkey, results)
		tDecode += time.Since(start)

		_, march := kv.GetValAndComp(outkey, innkey, val)
		if !march {
			fmt.Println("Keyword SIPIR Failed")
		}
		num++
	}

	fmt.Printf("Keyword PIR Benchmark (%s + %s + %s, N=2^%d, w=%d)\n", pir.Name(), batchtype, kvs.Name(), logNumsKeys, bitsPerVal)

	fmt.Printf("OFFLINE OVERHEAD:\n")
	fmt.Printf("  - Encode DB:     %v\n", encodeDuration)
	fmt.Printf("  - Generate Hint:  %v, overall Setup time:  %v, offline Communication: %.4f MB\n", genHintDuration, encodeDuration+genHintDuration, float64(db.Size()+kvs.Size())/(1024*1024))
	fmt.Printf("  - Hint Storage:  KVS: %.4f KB, PIR: %.4f KB, overall Client hint size: %.4f MB\n", float64(kvs.Size())/1024, float64(pir.GetHintSize())/1024, (float64(kvs.Size())+pir.GetHintSize())/1024/1024)

	fmt.Printf("\nONLINE LATENCY (Average over %.0f runs):\n", num)
	fmt.Printf("  1. Mapping (KVS Index):   %8.4f us\n", float64(tMapping.Nanoseconds())/num/1000)
	fmt.Printf("  2. PIR Query Gen:         %8.4f us (Up: %7.4f KBytes)\n", float64(tQuery.Nanoseconds())/num/1000, upMsgSize/num/1024)
	fmt.Printf("  3. Server Answer:         %8.4f us (Down: %7.4f KBytes)\n", float64(tServer.Nanoseconds())/num/1000, downMsgSize/num/1024)
	fmt.Printf("  5. ReconAndRefresh State: %8.4f us\n", float64(tRewindRecon.Nanoseconds())/num/1000)
	fmt.Printf("  6. KVS Decode:            %8.4f us, all client reconstruct and refresh and decode time: %8.4f us\n", float64(tDecode.Nanoseconds())/num/1000, float64(tRewindRecon.Nanoseconds()+tDecode.Nanoseconds())/num/1000)
}

// go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirSkip -benchmem -run=none -v -args -logN 20 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "piano" -type "skip"
// go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirSkip -benchmem -v
func BenchmarkKeywordSipirSkip(b *testing.B) {

	sipir := &sipir.Piano{}

	var kvs KVS.KVS

	switch kvsID {
	case "pthashkvs":
		kvs = KVS.NewPTHashKVS()
	case "bbhashkvs":
		kvs = KVS.NewBBHashKVS()
	case "bbhash2kvs":
		kvs = &KVS.BBHash2KVS{}
	case "bffkvs":
		kvs = &KVS.BFFKVS{}
	case "consensusrecsplitkvs":
		kvs = KVS.NewConsensusRecSplitKVS()
	default:
		kvs = KVS.NewPTHashKVS()
	}

	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()
	kv.Sort()

	start := time.Now()
	db := kvs.Encode(kv)
	encodeDuration := time.Since(start)

	start = time.Now()
	sipir.InitParams(db.NumEntries, db.BitsPerEntry, batchtype)
	sipir.GenerateHint(&db)
	genHintDuration := time.Since(start)

	batchSize := kvs.GetBatchSize()

	var tMapping, tQuery, tServer, tReconAndRefresh, tDecode time.Duration
	var upMsgSize, downMsgSize float64
	num := 0.0

	maxQueries := sipir.Params.Q / batchSize

	fmt.Println("********* Keyword SIPIR (Skip) Benchmark Start **********")

	for q := uint64(0); q < maxQueries; q++ {
		outkey, innkey := kv.GenRandomKey()

		// A. Keyword Mapping (Client)
		t1 := time.Now()
		targetIndexes := kvs.Lookup(outkey, innkey)
		tMapping += time.Since(t1)

		// B. Query Gen (Client)
		t2 := time.Now()
		req, state := sipir.Query(targetIndexes)
		tQuery += time.Since(t2)
		upMsgSize += float64(req.Size())

		// C. Server Answer (Server)
		t3 := time.Now()
		resp := sipir.Answer(&db, req)
		tServer += time.Since(t3)
		downMsgSize += float64(resp.Size())

		// D. Reconstruct (Client)
		t4 := time.Now()
		results := sipir.Reconstruct(resp, state)
		sipir.Refresh(targetIndexes, results, state)
		tReconAndRefresh += time.Since(t4)

		// E. Final Decode (Client)
		t6 := time.Now()
		val, _ := kvs.Decode(innkey, results)
		tDecode += time.Since(t6)

		_, march := kv.GetValAndComp(outkey, innkey, val)
		if !march {
			fmt.Println("Keyword SIPIR Failed")
		}
		num++
	}

	fmt.Printf("Keyword PIR Benchmark (%s + %s + %s, N=2^%d, w=%d)\n", sipir.Name(), batchtype, kvs.Name(), logNumsKeys, bitsPerVal)

	fmt.Printf("OFFLINE OVERHEAD:\n")
	fmt.Printf("  - Encode DB:     %v\n", encodeDuration)
	fmt.Printf("  - Generate Hint:  %v, overall Setup time:  %v, offline Communication: %.4f MB\n", genHintDuration, encodeDuration+genHintDuration, float64(db.Size()+kvs.Size())/(1024*1024))
	fmt.Printf("  - Hint Storage:  KVS: %.4f KB, PIR: %.4f KB, overall Client hint size: %.4f MB\n", float64(kvs.Size())/1024, float64(sipir.Hint.Size())/1024, (float64(kvs.Size())+sipir.Hint.Size())/1024/1024)

	fmt.Printf("\nONLINE LATENCY (Average over %.0f runs):\n", num)
	fmt.Printf("  1. Mapping (KVS Index):   %8.4f us\n", float64(tMapping.Nanoseconds())/num/1000)
	fmt.Printf("  2. PIR Query Gen:         %8.4f us (Up: %7.4f KBytes)\n", float64(tQuery.Nanoseconds())/num/1000, upMsgSize/num/1024)
	fmt.Printf("  3. Server Answer:         %8.4f us (Down: %7.4f KBytes)\n", float64(tServer.Nanoseconds())/num/1000, downMsgSize/num/1024)
	fmt.Printf("  5. ReconAndRefresh State: %8.4f us\n", float64(tReconAndRefresh.Nanoseconds())/num/1000)
	fmt.Printf("  6. KVS Decode:            %8.4f us, all client reconstruct and refresh and decode time: %8.4f us\n", float64(tDecode.Nanoseconds())/num/1000, float64(tReconAndRefresh.Nanoseconds()+tDecode.Nanoseconds())/num/1000)
}

// go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -args -logN 20 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "simplepir"

// go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -v
func BenchmarkKeywordHepir(b *testing.B) {

	var hepir simplepir.HEPIR

	switch pirID {
	case "doublepir":
		hepir = &simplepir.DoublePIR{}
	case "simplepir":
		hepir = &simplepir.SimplePIR{}
	default:
		b.Fatalf("error Scheme: piano, singleserver, singlepass")
	}

	var kvs KVS.KVS

	switch kvsID {
	case "pthashkvs":
		kvs = KVS.NewPTHashKVS()
	case "bbhashkvs":
		kvs = KVS.NewBBHashKVS()
	case "bbhash2kvs":
		kvs = &KVS.BBHash2KVS{}
	case "bffkvs":
		kvs = &KVS.BFFKVS{}
	case "consensusrecsplitkvs":
		kvs = KVS.NewConsensusRecSplitKVS()
	default:
		kvs = KVS.NewPTHashKVS()
	}
	defer kvs.Free()

	kv := &utils.KV{}
	kv.Setup(1<<logNumsKeys, bitsPerVal)
	kv.Random()
	kv.Sort()

	start := time.Now()
	db := kvs.Encode(kv)
	encodeDuration := time.Since(start)

	start = time.Now()
	hepir.InitParams(db.NumEntries, db.BitsPerEntry)
	internalDB := hepir.MakeInternalDB(&db)
	serverHint, clientHint := hepir.Setup(internalDB)
	setupDuration := time.Since(start)

	var tMapping, tOffQuery, tOnQuery, tServer, tRecon, tDecode time.Duration
	var upMsgSize, downMsgSize float64
	num := 0.0

	testRuns := uint64(100)

	fmt.Printf("********* Keyword HEPIR (%s) Benchmark Start **********\n", hepir.Name())

	for q := uint64(0); q < testRuns; q++ {
		outkey, innkey := kv.GenRandomKey()

		// A. Keyword Mapping (Client)
		t1 := time.Now()
		targetIndexes := kvs.Lookup(outkey, innkey)
		tMapping += time.Since(t1)

		// B. Query Gen (Client)
		//t2 := time.Now()
		//req, clientState := hepir.Query(targetIndexes)
		//tQuery += time.Since(t2)

		t21 := time.Now()
		EncZero, clientState := hepir.QueryOffline(uint64(len(targetIndexes)))
		tOffQuery += time.Since(t21)
		t22 := time.Now()
		req := hepir.QueryOnline(targetIndexes, EncZero)
		tOnQuery += time.Since(t22)

		upMsgSize += float64(req.Size())

		// C. Server Answer (Server)
		t3 := time.Now()
		resp := hepir.Answer(internalDB, req, serverHint)
		tServer += time.Since(t3)
		downMsgSize += float64(resp.Size())

		// D. Reconstruct (Client)
		t4 := time.Now()
		results := hepir.Reconstruct(targetIndexes, clientHint, req, resp, clientState)
		tRecon += time.Since(t4)

		// E. Final Decode (Client)
		t5 := time.Now()
		val, _ := kvs.Decode(innkey, results)
		tDecode += time.Since(t5)

		_, march := kv.GetValAndComp(outkey, innkey, val)
		if !march {
			fmt.Printf("FAILED: HEPIR mismatch at q=%d\n", q)
		}

		num++
	}

	fmt.Printf("Keyword PIR Benchmark (%s + %s + %s, N=2^%d, w=%d)\n", hepir.Name(), "plain", kvs.Name(), logNumsKeys, bitsPerVal)

	fmt.Printf("OFFLINE OVERHEAD:\n")
	fmt.Printf("  - Encode DB:     %v\n", encodeDuration)
	fmt.Printf("  - Generate Hint:  %v, overall Setup time:  %v, offline Communication: %.4f MB\n", setupDuration, encodeDuration+setupDuration, float64(clientHint.Size()+kvs.Size())/(1024*1024))
	fmt.Printf("  - Hint Storage:   KVS: %.4f KB, Client: %.4f MB (all %.4f MB), Server: %.2f MB\n",
		float64(kvs.Size())/1024, float64(clientHint.Size())/(1024*1024), float64(kvs.Size()+clientHint.Size())/(1024*1024), float64(serverHint.Size())/(1024*1024))

	fmt.Printf("\nONLINE LATENCY (Average over %.0f runs):\n", num)
	fmt.Printf("  1. Mapping (KVS Index):   %8.4f us\n", float64(tMapping.Nanoseconds())/num/1000)
	fmt.Printf("  2. PIR Query Gen:         %8.4f + %8.4f = %8.4f us (Up: %.4f KB)\n", float64(tOffQuery.Nanoseconds())/num/1000, float64(tOnQuery.Nanoseconds())/num/1000, float64((tOffQuery+tOnQuery).Nanoseconds())/num/1000, upMsgSize/num/1024)
	fmt.Printf("  3. Server Answer:         %8.4f us (Down: %.4f KB)\n", float64(tServer.Nanoseconds())/num/1000, downMsgSize/num/1024)
	fmt.Printf("  4. Reconstruct:           %8.4f us\n", float64(tRecon.Nanoseconds())/num/1000)
	fmt.Printf("  5. KVS Decode:            %8.4f us, all reconstruct and decode %8.4f\n", float64(tDecode.Nanoseconds())/num/1000, float64(tRecon.Nanoseconds()+tDecode.Nanoseconds())/num/1000)

}
