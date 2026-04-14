
# SIPIR

# Piano
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=1 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=26 -bitsPerEntry=32 -batch=1 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=27 -bitsPerEntry=32 -batch=1 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=1 -type="skip" -pirID="piano"

go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=64 -batch=1 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=128 -batch=1 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=192 -batch=1 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=256 -batch=1 -type="skip" -pirID="piano"

# SinglePass
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=1 -type="rewind" -pirID="singlepass"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=1 -type="rewind" -pirID="singlepass"

# SingleServer
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=1 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=26 -bitsPerEntry=32 -batch=1 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=27 -bitsPerEntry=32 -batch=1 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=1 -type="rewind" -pirID="singleserver"

go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=64 -batch=1 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=128 -batch=1 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=192 -batch=1 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=256 -batch=1 -type="rewind" -pirID="singleserver"

# Batch SIPIR

# Piano with Skip
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=3 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=26 -bitsPerEntry=32 -batch=3 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=27 -bitsPerEntry=32 -batch=3 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=3 -type="skip" -pirID="piano"

go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=64 -batch=3 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=128 -batch=3 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=192 -batch=3 -type="skip" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=256 -batch=3 -type="skip" -pirID="piano"

# Piano with Rewind
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="piano"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=26 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="piano"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=27 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="piano"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="piano"

go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=64 -batch=3 -type="rewind" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=128 -batch=3 -type="rewind" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=192 -batch=3 -type="rewind" -pirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=256 -batch=3 -type="rewind" -pirID="piano"

# SinglePass with Rewind
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="singlepass"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="singlepass"

# SingleServer with Rewind
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=26 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=27 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=28 -bitsPerEntry=32 -batch=3 -type="rewind" -pirID="singleserver"

go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=64 -batch=3 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=128 -batch=3 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=192 -batch=3 -type="rewind" -pirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerEntry=256 -batch=3 -type="rewind" -pirID="singleserver"
