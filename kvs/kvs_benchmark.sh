
# bff-kvs
go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=bffkvs

# bbhashkvs
go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=bbhashkvs

# pthashkvs
go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=pthashkvs

# consensusrecsplitkvs
go test -bench=BenchmarkKVStore -benchmem -run=none -v -args -logN=25 -bitsPerVal=32 -kvsID=consensusrecsplitkvs