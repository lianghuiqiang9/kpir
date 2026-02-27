
# simplepir

go test -bench=BenchmarkHepir -benchmem -run=none -v -args -logN=25 -bitsPerEntry=32 -batch=1 -pirID="simplepir"

go test -bench=BenchmarkHepir -benchmem -run=none -v -args -logN=28 -bitsPerEntry=32 -batch=1 -pirID="simplepir"

# doublepir

go test -bench=BenchmarkHepir -benchmem -run=none -v -args -logN=25 -bitsPerEntry=32 -batch=1 -pirID="doublepir"

go test -bench=BenchmarkHepir -benchmem -run=none -v -args -logN=28 -bitsPerEntry=32 -batch=1 -pirID="doublepir"
