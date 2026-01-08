package utils

import (
	"fmt"
	"math/rand"
	"testing"
)

// 全局变量防止编译器优化掉循环体
var globalSlice []uint64

// go test -bench=BenchmarkBucket -benchmem -run=none -v
func BenchmarkBucket(b *testing.B) {
	logNumsKeys := 25
	W := 32
	totalKeys := uint64(1 << logNumsKeys)

	// Data preparation
	bucket := &Bucket{}
	bucket.Setup(totalKeys, uint64(W))
	bucket.Random()

	// --- 1. Benchmark: Map Construction ---
	b.Run("MakeMap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = MakeMap(bucket.Keys, bucket.Values)
		}
	})

	// Pre-build a map for lookup performance evaluation
	Map := MakeMap(bucket.Keys, bucket.Values)

	fmt.Println("bucketSize: ", GetSerializedSize(bucket)/1024/1024, " MB,")
	fmt.Println(" MapSzie: ", GetSerializedSize(Map)/1024/1024, " MB.")

	// --- 2. Benchmark: Map Random Lookup ---
	b.Run("MapLookupRandom", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulating random access patterns
			idx := i % int(totalKeys)
			globalSlice, _ = Map[bucket.Keys[idx]]
		}
	})

	// --- 3. Benchmark: Sorting Process ---
	b.Run("Sort", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			bucket.Random() // Reset to random state to avoid pdqsort best-case optimization
			b.StartTimer()
			bucket.Sort()
		}
	})

	bucket.Random() // Reset to random state for subsequent non-sorted lookups
	keys := bucket.Keys
	fmt.Println("bucket.IsSort: ", bucket.IsSort)

	// --- 4. Benchmark: Linear/Binary Lookup (Unsorted, sequential query input) ---
	b.Run("LookupSequential", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % int(totalKeys)
			// Sequential query of keys on unsorted data
			globalSlice, _ = bucket.GetVal(keys[idx])
		}
	})

	// --- 5. Benchmark: Linear/Binary Lookup (Unsorted, random query input) ---
	b.Run("LookupRandom", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := rand.Uint64() % (totalKeys)
			// Random query of keys on unsorted data
			globalSlice, _ = bucket.GetVal(keys[idx])
		}
	})

	// Pre-sort the bucket for interpolation/binary search evaluation
	bucket.Sort()
	fmt.Println("bucket.IsSort: ", bucket.IsSort)
	sortedKeys := bucket.Keys

	// --- 6. Benchmark: Interpolation/Binary Lookup (Sorted, random query input) ---
	b.Run("InterpLookupRandom", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % int(totalKeys)
			// Querying random keys against a sorted index (leads to cache misses)
			globalSlice, _ = bucket.GetVal(keys[idx])
		}
	})

	// --- 7. Benchmark: Interpolation/Binary Lookup (Sorted, sequential query input) ---
	b.Run("InterpLookupSequential", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % int(totalKeys)
			// Query order aligns with physical memory layout (highly cache-friendly)
			globalSlice, _ = bucket.GetVal(sortedKeys[idx])
		}
	})

}
