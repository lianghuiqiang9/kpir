package utils

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

const MaxNumsPerBucket uint64 = 1 << 25

type EncodedDB struct {
	Data           []uint64
	NumEntries     uint64
	Uint64PerEntry uint64
	BitsPerEntry   uint64
}

func (db *EncodedDB) Size() int {
	if db == nil {
		return 0
	}
	size := int(unsafe.Sizeof(*db))
	size += cap(db.Data) * 8

	return size
}

func (db *EncodedDB) SizeGB() float64 {
	return float64(db.Size()) / (1024 * 1024 * 1024)
}

func (db *EncodedDB) InitParams(numEntries uint64, bitsPerEntry uint64) {
	if bitsPerEntry%32 != 0 {
		fmt.Println("bitsPerVal should be 32 * k")
		os.Exit(1)
	}
	uint64PerEntry := (bitsPerEntry + 63) / 64

	db.NumEntries = NextPerfectSquare(numEntries)
	db.BitsPerEntry = bitsPerEntry
	db.Uint64PerEntry = uint64PerEntry

	totalWords := numEntries * uint64PerEntry

	fmt.Printf("Database Initialized: %d entries, %d uint64 per entry (Total: %d uint64s)\n",
		numEntries, uint64PerEntry, totalWords)
}

func (db *EncodedDB) Random() {
	db.Data = make([]uint64, db.NumEntries*db.Uint64PerEntry)
	var lastWordMask uint64 = math.MaxUint64
	if db.BitsPerEntry%64 != 0 {
		lastWordMask = 0xFFFFFFFF
	}

	for i := uint64(0); i < db.NumEntries; i++ {
		for j := uint64(0); j < db.Uint64PerEntry; j++ {
			idx := i*db.Uint64PerEntry + j
			val := rand.Uint64()
			if j == db.Uint64PerEntry-1 {
				val &= lastWordMask
			}
			db.Data[idx] = val
		}
	}
}

func (db *EncodedDB) LoadDB(id string, filepath string) int64 {

	dbFile := filepath + id + "db.bin"
	//totalWords := numEntries * uint64PerEntry

	//db.NumEntries = numEntries
	//db.Uint64PerEntry = uint64PerEntry

	if FileExists(dbFile) {
		fmt.Printf("Loading Database [%s] from disk...\n", id)

		db.Data = make([]uint64, db.NumEntries*db.Uint64PerEntry)

		if err := LoadUint64Slice(db.Data, dbFile); err != nil {
			log.Fatalf("Load DB failed: %v", err)
		}

	} else {
		fmt.Printf("Database [%s] not found. Generating new random data...\n", id)

		db.Random()

		if err := SaveUint64Slice(db.Data, dbFile); err != nil {
			fmt.Printf("Warning: Save DB failed: %v\n", err)
		}
	}

	var dbSize int64
	if s, err := os.Stat(dbFile); err == nil {
		dbSize = s.Size()
	}

	return dbSize
}

func (db *EncodedDB) GetEntry(entryIdx uint64) []uint64 {
	//out := make([]uint64, db.Uint64PerEntry)
	start := entryIdx * uint64(db.Uint64PerEntry)
	end := start + uint64(db.Uint64PerEntry)
	return db.Data[start:end]
}

func (db *EncodedDB) GetBatchEntry(idx []uint64) [][]uint64 {
	out := make([][]uint64, len(idx))
	for i, x := range idx {
		out[i] = db.GetEntry(x)
	}

	return out
}

func (db *EncodedDB) EntryEqualsData(entryIdx uint64, data []uint64) bool {
	entry := db.GetEntry(entryIdx)
	//fmt.Println(entry, data)
	if len(entry) != len(data) {
		return false
	}
	for i := range entry {
		if entry[i] != data[i] {
			return false
		}
	}
	return true
}

func (db *EncodedDB) BatchEntryEqualsData(indexes []uint64, batchData [][]uint64) bool {

	if len(indexes) != len(batchData) {
		return false
	}

	for i := range indexes {
		if !db.EntryEqualsData(indexes[i], batchData[i]) {
			return false
		}
	}

	return true
}

// XORInplace a = a ^ b
func (db *EncodedDB) XORInplace(a []uint64, b []uint64) {

	_ = a[db.Uint64PerEntry-1]
	_ = b[db.Uint64PerEntry-1]

	for k := uint64(0); k < db.Uint64PerEntry; k++ {
		a[k] ^= b[k]
	}
}
func GenRandomIndexes(batchSize, numEntries uint64) []uint64 {
	targetIndexes := make([]uint64, batchSize)
	for i := uint64(0); i < batchSize; i++ {
		targetIndexes[i] = rand.Uint64() % numEntries
	}
	return targetIndexes
}

type Bucket struct {
	TotalKeys uint64
	Keys      []uint64
	Values    []uint64

	Uint64PerVal uint64
	BitsPerVal   uint64
	IsSort       bool
}

func (b *Bucket) Size() int {
	if b == nil {
		return 0
	}

	size := int(unsafe.Sizeof(*b))

	size += cap(b.Keys) * 8

	size += cap(b.Values) * 8

	return size
}
func (b *Bucket) Setup(totalKeys uint64, bitsPerVal uint64) {
	b.TotalKeys = totalKeys
	if (bitsPerVal+32)%64 != 0 {
		fmt.Println("bitsPerVal should be 32 + 64 * k")
		os.Exit(1)
	}
	b.Uint64PerVal = (bitsPerVal / 64) + 1

	b.BitsPerVal = bitsPerVal
	b.IsSort = false
}

func (b *Bucket) Random() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b.Keys = make([]uint64, b.TotalKeys)

	w := b.Uint64PerVal

	valLen := b.TotalKeys * w
	b.Values = make([]uint64, valLen)
	uniqueKeys := GenerateUniqueUint64s(int(b.TotalKeys))
	copy(b.Keys, uniqueKeys)
	var idx uint64 = 0
	for i := uint64(0); i < b.TotalKeys; i++ {

		for j := uint64(0); j < w-1; j++ {
			b.Values[idx] = r.Uint64()
			idx++
		}

		b.Values[idx] = uint64(r.Uint32())
		idx++
	}
	b.IsSort = false
}

func (b *Bucket) Print(limit int) {
	size := b.TotalKeys
	w := b.Uint64PerVal
	isSort := b.IsSort
	fmt.Printf("  [Bucket Statistics] Size: %d | Uint64PerEntry: %d | IsSort: %t\n", size, w, isSort)

	if size == 0 {
		fmt.Println("  (Empty Bucket)")
		return
	}

	// 确定打印条数
	displayCount := int(size)
	if limit > 0 && displayCount > limit {
		displayCount = limit
	}

	for j := 0; j < displayCount; j++ {

		key := b.Keys[j]

		start := uint64(j) * w
		end := start + w
		valSlice := b.Values[start:end]

		fmt.Printf("  ├── [Item: %-4d] InnKey: %-15d => Value: %v\n", j, key, valSlice)
	}

	if limit > 0 && uint64(limit) < size {
		fmt.Printf("  └── ... and %d more items in this bucket\n", size-uint64(limit))
	}
}

func (b *Bucket) GetVal(key uint64) ([]uint64, bool) {
	if b.IsSort {
		return b.GetValInterpolation(key)
	}

	for i, v := range b.Keys {
		if v == key {
			start := uint64(i) * b.Uint64PerVal
			end := start + b.Uint64PerVal
			return b.Values[start:end], true
		}
	}
	return nil, false
}

func (b *Bucket) GetValAndComp(innKey uint64, targetValue []uint64) (found bool, match bool) {

	val, exists := b.GetVal(innKey)
	if !exists {
		return false, false
	}

	if len(val) != len(targetValue) {
		return true, false
	}

	for i := 0; i < len(val); i++ {
		if val[i] != targetValue[i] {
			return true, false
		}
	}

	return true, true
}

func (b *Bucket) Sort() {

	b.Keys, b.Values = Sort(b.Keys, b.Values)

	b.IsSort = true
}

func (b *Bucket) GetValInterpolation(key uint64) ([]uint64, bool) {

	return GetValInterpolation(b.Keys, b.Values, int(b.Uint64PerVal), key)

}

func (b *Bucket) LoadBuckets(id string, filepath string, numKeys uint64, BitsPerVal uint64, index uint64) {

	fileName := fmt.Sprintf("%s%s_bucket_%d.bin", filepath, id, index)
	b.Setup(numKeys, BitsPerVal)

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("无法处理桶文件 %d: %v", index, err)
	}
	defer f.Close()

	if len(b.Keys) == 0 {
		b.Keys = make([]uint64, b.TotalKeys)
	}
	if len(b.Values) == 0 {
		b.Values = make([]uint64, b.TotalKeys*b.Uint64PerVal)
	}

	fi, _ := f.Stat()
	expectedSize := int64(len(b.Keys)*8 + len(b.Values)*8)

	if fi.Size() >= expectedSize {
		binary.Read(f, binary.LittleEndian, b.Keys)
		binary.Read(f, binary.LittleEndian, b.Values)
	} else {
		b.Random()
		f.Seek(0, 0)
		binary.Write(f, binary.LittleEndian, b.Keys)
		binary.Write(f, binary.LittleEndian, b.Values)
		f.Sync()
	}
}

type KV struct {
	TotalKeys     uint64
	BucketCount   uint64
	BucketIndices []uint64
	BucketSizes   []uint64

	Buckets []*Bucket

	Uint64PerVal uint64
	BitsPerVal   uint64
	IsSort       bool
}

// Size returns the total memory consumption of the KV structure in bytes,
// including all underlying buckets and slices.
func (kv *KV) Size() int {
	if kv == nil {
		return 0
	}

	// 1. Size of the KV struct itself (fixed fields + slice headers)
	size := int(unsafe.Sizeof(*kv))

	// 2. Size of the administrative slices (Capacity * 8 bytes)
	size += (cap(kv.BucketIndices)) * 8
	size += (cap(kv.BucketSizes)) * 8

	// 3. Size of the Buckets slice header and pointers
	// Each pointer in the slice takes 8 bytes on 64-bit systems
	size += (cap(kv.Buckets)) * 8

	// 4. Recursive size of each Bucket object
	for _, b := range kv.Buckets {
		if b != nil {
			size += b.Size()
		}
	}

	return size
}

func (kv *KV) Setup(totalKeys uint64, BitsPerVal uint64) {

	kv.TotalKeys = totalKeys

	if (BitsPerVal+32)%64 != 0 {
		fmt.Println("BitsPerVal should be 32 + 64 * k")
		os.Exit(1)
	}

	kv.Uint64PerVal = (BitsPerVal / 64) + 1
	kv.IsSort = false

	kv.BucketCount = (totalKeys + MaxNumsPerBucket - 1) / MaxNumsPerBucket

	kv.BucketIndices = make([]uint64, kv.BucketCount)
	kv.BucketSizes = make([]uint64, kv.BucketCount)
	kv.Buckets = make([]*Bucket, kv.BucketCount)

	for i := uint64(0); i < kv.BucketCount; i++ {

		kv.BucketIndices[i] = i

		size := MaxNumsPerBucket
		if i == kv.BucketCount-1 {
			if remainder := totalKeys % MaxNumsPerBucket; remainder != 0 {
				size = remainder
			}
		}

		kv.BucketSizes[i] = size

		b := &Bucket{}
		b.Setup(size, BitsPerVal)
		kv.Buckets[i] = b
	}
}

func NextPerfectSquare(n uint64) uint64 {
	root := math.Ceil(math.Sqrt(float64(n)))
	return uint64(root * root)
}

func GenerateUniqueUint64s(n int) []uint64 {
	set := make(map[uint64]struct{})
	results := make([]uint64, 0, n)

	for len(results) < n {
		val := rand.Uint64()
		if _, exists := set[val]; !exists {
			set[val] = struct{}{}
			results = append(results, val)
		}
	}
	return results
}

func (kv *KV) GenRandomKey() (uint64, uint64) {
	outkey := rand.Uint64() % kv.BucketCount
	innkeyIndex := rand.Uint64() % kv.Buckets[outkey].TotalKeys
	innkey := kv.Buckets[outkey].Keys[innkeyIndex]
	return outkey, innkey
}

func (kv *KV) Random() {

	if len(kv.Buckets) == 0 {
		kv.Buckets = make([]*Bucket, kv.BucketCount)
		for i := range kv.Buckets {
			if kv.Buckets[i].TotalKeys == 0 {
				kv.Buckets[i].Setup(kv.BucketSizes[i], 32)
			}
		}
	}

	for i := uint64(0); i < kv.BucketCount; i++ {
		kv.Buckets[i].Random()

		if i%10000 == 0 {
			fmt.Printf("\rGenerating Buckets: %d/%d (%.2f%%)",
				i, kv.BucketCount, float64(i)/float64(kv.BucketCount)*100)
		}
	}
	kv.IsSort = false

	fmt.Println("\nData generation completed successfully.")
}

func (kv *KV) Print(limit int) {
	fmt.Printf("--- KV Bucket-Based Storage Statistics ---\n")
	fmt.Printf("Total Keys:    %d\n", kv.TotalKeys)
	fmt.Printf("Total Buckets: %d\n", kv.BucketCount)
	fmt.Printf("------------------------------------------\n")

	bucketLimit := 10
	if int(kv.BucketCount) < bucketLimit {
		bucketLimit = int(kv.BucketCount)
	}

	for i := 0; i < bucketLimit; i++ {
		fmt.Printf("Bucket [%d]: Index %d\n", i, kv.BucketIndices[i])
		kv.Buckets[i].Print(limit)
		fmt.Println()
	}

	if int(kv.BucketCount) > bucketLimit {
		fmt.Printf("... and %d more buckets hidden.\n", kv.BucketCount-uint64(bucketLimit))
	}
}
func (kv *KV) Sort() {
	for i := uint64(0); i < kv.BucketCount; i++ {
		kv.Buckets[i].Sort()
	}
	kv.IsSort = true
}

func (kv *KV) GetVal(outKey uint64, innKey uint64) ([]uint64, bool) {
	if outKey >= kv.BucketCount {
		return nil, false
	}
	return kv.Buckets[outKey].GetVal(innKey)
}

func (kv *KV) GetValAndComp(outKey uint64, innKey uint64, targetValue []uint64) (found bool, match bool) {
	if outKey >= kv.BucketCount {
		return false, false
	}
	return kv.Buckets[outKey].GetValAndComp(innKey, targetValue)
}

// MergeBuckets consolidates individual bucket files into a single master kv.bin
func (kv *KV) MergeBuckets(id string, filepath string) {
	masterFile := filepath + id + "kv.bin"
	out, err := os.Create(masterFile)
	if err != nil {
		log.Fatalf("Failed to create merged file: %v", err)
	}
	defer out.Close()

	fmt.Printf("Starting to merge %d bucket files into %s...\n", kv.BucketCount, masterFile)

	// First pass: sequentially merge all bucket Keys
	for i := uint64(0); i < kv.BucketCount; i++ {
		binary.Write(out, binary.LittleEndian, kv.Buckets[i].Keys)
	}

	// Second pass: sequentially merge all bucket Values
	for i := uint64(0); i < kv.BucketCount; i++ {
		binary.Write(out, binary.LittleEndian, kv.Buckets[i].Values)
	}

	fmt.Println("Merge completed.")
}

func (kv *KV) LoadKV(id string, filepath string, numKeys uint64, uint64PerVal uint64) {
	// 1. Initialize metadata and bucket structures
	kv.Setup(numKeys, uint64PerVal)

	fmt.Println("Processing bucket files sequentially...")

	for i := uint64(0); i < kv.BucketCount; i++ {
		// Load data for each individual bucket
		// Keywords: Skip (if file exists and is valid) / Rewind (if data needs regeneration)
		kv.Buckets[i].LoadBuckets(id, filepath, kv.BucketSizes[i], uint64PerVal, i)

		// Progress indicator for large-scale data (e.g., 2^30 entries)
		if i%5000 == 0 || i == kv.BucketCount-1 {
			fmt.Printf("\rProgress: %d/%d (%.2f%%)",
				i+1, kv.BucketCount, float64(i+1)/float64(kv.BucketCount)*100)
		}
	}
	fmt.Println("\nAll individual buckets have been processed.")

	// 3. Optional: Consolidate buckets into a single master file
	// kv.MergeBuckets(id, filepath)
}
