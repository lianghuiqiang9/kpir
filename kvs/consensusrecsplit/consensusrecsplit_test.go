package consensusrecsplit

import (
	"fmt"
	"math/rand"
	"testing"
)

// go test -run TestConsensusRecSplit
func TestConsensusRecSplit(t *testing.T) {
	const N = 1 << 20
	keys := make([]uint64, N)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < N; i++ {
		keys[i] = rng.Uint64()
	}
	mph := New(keys)

	defer mph.Free()

	occupied := make([]bool, N)
	for _, k := range keys {
		idx := mph.Lookup(k)
		if idx >= N {
			fmt.Printf("Index out of bounds: %d", idx)
			continue
		}
		if occupied[idx] {
			fmt.Printf("Collision at index %d", idx)
		}
		occupied[idx] = true
	}

	fmt.Printf("Bits per key: %.4f\n", float64(mph.Bits())/float64(N))
}
