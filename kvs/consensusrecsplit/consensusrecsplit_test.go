package consensusrecsplit

import (
	"fmt"
	"math/rand"
	"testing"
)

// go test -run TestConsensusRecSplit
func TestConsensusRecSplit(t *testing.T) {
	const N = 1 << 20 // 先用小规模测试
	keys := make([]uint64, N)
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < N; i++ {
		keys[i] = rng.Uint64()
	}

	// 直接调用 main.go 中定义的导出函数
	mph := New(keys)

	defer mph.Free()

	// 验证逻辑
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
