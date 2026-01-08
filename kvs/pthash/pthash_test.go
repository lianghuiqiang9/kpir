package pthash

import (
	"fmt"
	"math/rand"
	"testing"
)

// go test -run TestPTHash
func TestPTHash(t *testing.T) {
	const N = 1 << 20 // 先用小规模测试
	keys := make([]uint64, N)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < N; i++ {
		keys[i] = rng.Uint64()
	}

	phf := New(keys)
	if phf == nil {
		t.Fatal("Failed to initialize PTHash")
	}
	defer phf.Free()

	// 验证逻辑
	occupied := make([]bool, N)
	for _, k := range keys {
		idx := phf.Lookup(k)
		if idx >= N {
			fmt.Printf("Index out of bounds: %d", idx)
			continue
		}
		if occupied[idx] {
			fmt.Printf("Collision at index %d", idx)
		}
		occupied[idx] = true
	}
	fmt.Printf("Bits per key: %.4f\n", float64(phf.Bits())/float64(N))
}
