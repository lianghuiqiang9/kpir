

go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=bffkvs

go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=pthashkvs

go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=consensusrecsplitkvs