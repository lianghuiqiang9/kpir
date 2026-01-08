package bbhash

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// go test -run TestBBHash
func TestBBHash(t *testing.T) {
	NumKeys := uint64(1 << 20)
	keys := make([]uint64, NumKeys)
	for i := uint64(0); i < NumKeys; i++ {
		keys[i] = rand.Uint64()
	}

	bb := New(keys)

	defer bb.Free()

	for _, k := range keys {
		idx := bb.Lookup(k)
		//fmt.Println("idx: ", idx)
		idx++
	}
	bbbits := bb.Bits()
	fmt.Printf("Total Bits: %d\n", bbbits)
	fmt.Printf("BBHash size = %.4f bits,  radio = %.4f bits/key\n", float64(bbbits), float64(bbbits)/float64(NumKeys))
}
