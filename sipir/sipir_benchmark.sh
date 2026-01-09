
# SIPIR

# Piano
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=1 -type="skip" -sipirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=1 -type="skip" -sipirID="piano"

# SinglePass
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=1 -type="rewind" -sipirID="singlepass"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=1 -type="rewind" -sipirID="singlepass"

# SingleServer
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=1 -type="rewind" -sipirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=1 -type="rewind" -sipirID="singleserver"


# Batch SIPIR

# Piano with Skip
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=3 -type="skip" -sipirID="piano"
go test -bench=BenchmarkSkip -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=3 -type="skip" -sipirID="piano"

# Piano with Rewind
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=3 -type="rewind" -sipirID="piano"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=3 -type="rewind" -sipirID="piano"

# SinglePass with Rewind
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=3 -type="rewind" -sipirID="singlepass"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=3 -type="rewind" -sipirID="singlepass"

# SingleServer with Rewind
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=20 -bitsPerVal=32 -batch=3 -type="rewind" -sipirID="singleserver"
go test -bench=BenchmarkRewind -benchmem -run=none -v -timeout 0 -args -logN=25 -bitsPerVal=32 -batch=3 -type="rewind" -sipirID="singleserver"
