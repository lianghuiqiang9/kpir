
# compare with chalametPIR and KPIR

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 256 -kvsID "bffkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 512 -kvsID "bffkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 1024 -kvsID "bffkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 2048 -kvsID "bffkvs" -pirID "simplepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 256 -kvsID "bffkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 512 -kvsID "bffkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 1024 -kvsID "bffkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 2048 -kvsID "bffkvs" -pirID "simplepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 256 -kvsID "pthashkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 512 -kvsID "pthashkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 1024 -kvsID "pthashkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 20 -bitsPerVal 2048 -kvsID "pthashkvs" -pirID "simplepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 256 -kvsID "pthashkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 512 -kvsID "pthashkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 1024 -kvsID "pthashkvs" -pirID "simplepir"
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 22 -bitsPerVal 2048 -kvsID "pthashkvs" -pirID "simplepir"

# piano with skip with bffkvs

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirSkip -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "bffkvs" -pirID "piano" -type "skip"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirSkip -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "bffkvs" -pirID "piano" -type "skip"

# piano with pthashkvs

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirSkip -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "piano" -type "skip"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirSkip -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "piano" -type "skip"

# singlepass with rewind with bffkvs

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "bffkvs" -pirID "singlepass" -type "rewind"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "bffkvs" -pirID "singlepass" -type "rewind"

# singlepass with pthashkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "singlepass" -type "rewind"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "singlepass" -type "rewind"


# singleserver with rewind with bffkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "bffkvs" -pirID "singleserver" -type "rewind"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "bffkvs" -pirID "singleserver" -type "rewind"

# singleserver with pthashkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "singleserver" -type "rewind"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordSipirRewind -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "singleserver" -type "rewind"

# simplepir with bffkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "bffkvs" -pirID "simplepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "bffkvs" -pirID "simplepir"

# simplepir with pthashkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "simplepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "simplepir"

# doublepir with bffkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "bffkvs" -pirID "doublepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "bffkvs" -pirID "doublepir"

# doublepir with pthashkvs
go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 25 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "doublepir"

go test kpir_benchmark_test.go -bench=BenchmarkKeywordHepir -benchmem -run=none -v -timeout 0 -args -logN 28 -bitsPerVal 32 -kvsID "pthashkvs" -pirID "doublepir"


