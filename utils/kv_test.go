package utils

import (
	"fmt"
	"testing"
	"time"
)

// go test -run TestKV
func TestKV(t *testing.T) {
	bucket := &Bucket{}
	bucket.Setup(1<<25, 32) // MaxNumsInnKey
	bucket.Random()

	bucket.Print(3)
	val, flag := bucket.GetVal(14310880194869718394)
	fmt.Println("", val, flag)

	bucket2 := &Bucket{}
	bucket2.LoadBuckets("", "", 1<<12, 96, 100)
	bucket2.Print(4)

	val2, flag2 := bucket2.GetVal(7631819921086983138)
	fmt.Println("", val2, flag2)

	kv := &KV{}
	kv.Setup(1<<12, 96)
	kv.Random()

	kv.Print(3)

	kv2 := &KV{}
	kv2.LoadKV("", "", 1<<12, 160)
	kv2.Print(4)

	val3, flag3 := kv2.GetVal(3, 259169879257162931)
	fmt.Println("", val3, flag3)

	kv2.Sort()
	kv2.Print(20)
	val4, flag4 := kv2.GetVal(3, 259169879257162931)
	fmt.Println("", val4, flag4)
}

// go test -run TestGetVal
func TestGetVal(t *testing.T) {
	logNumsKeys := 25
	bucket := &Bucket{}
	bucket.Setup(1<<logNumsKeys, 32) // MaxNumsInnKey
	bucket.Random()

	KeysRand := bucket.Keys
	ValsRand := bucket.Values

	startBuild := time.Now()
	Map := MakeMap(KeysRand, ValsRand)
	buildTime := time.Since(startBuild)
	fmt.Printf("Map 2^%d = %d keys: %s, MapSize: %d MB\n", logNumsKeys, len(bucket.Keys), buildTime, GetSerializedSize(Map)/1024/1024)

	startBuild = time.Now()
	for j := 0; j < len(KeysRand); j++ {
		v, _ := Map[KeysRand[j]]
		v[0] = v[0] + 1
	}

	buildTime = time.Since(startBuild)
	fmt.Printf("Find 2^%d = %d keys: %s\n", logNumsKeys, len(bucket.Keys), buildTime)

	startBuild = time.Now()
	bucket.Sort()
	buildTime = time.Since(startBuild)
	fmt.Printf("Sort 2^%d = %d keys: %s\n", logNumsKeys, len(bucket.Keys), buildTime)

	Keys := bucket.Keys
	Vals := bucket.Values
	W := int(bucket.Uint64PerVal)

	startBuild = time.Now()
	for j := 0; j < len(Keys); j++ {
		v, _ := GetValInterpolation(Keys, Vals, W, KeysRand[j])

		v[0] = v[0] + 1
	}

	buildTime = time.Since(startBuild)
	fmt.Printf("Find 2^%d = %d keys: %s\n", logNumsKeys, len(bucket.Keys), buildTime)

	startBuild = time.Now()
	for j := 0; j < len(Keys); j++ {
		v, _ := GetValInterpolation(Keys, Vals, W, Keys[j])
		v[0] = v[0] + 1
	}
	buildTime = time.Since(startBuild)
	fmt.Printf("Find sort 2^%d = %d keys: %s\n", logNumsKeys, len(bucket.Keys), buildTime)

}
