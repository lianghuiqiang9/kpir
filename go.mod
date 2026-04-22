module main

go 1.24.11

replace (
	github.com/FastFilter/xorfilter => ./kvs/xorfilter
	github.com/local/consensusrecsplit => ./kvs/consensusrecsplit
	github.com/local/hepir/simplepir => ./hepir/simplepir
	github.com/local/kvs => ./kvs
	github.com/local/cppir => ./cppir
	github.com/local/utils => ./utils
	github.com/local/bbhash => ./kvs/bbhash
	github.com/local/pthash => ./kvs/pthash
)

require (
	github.com/local/hepir/simplepir v0.0.0-00010101000000-000000000000
	github.com/local/kvs v0.0.0-00010101000000-000000000000
	github.com/local/cppir v0.0.0-00010101000000-000000000000
	github.com/local/utils v0.0.0-00010101000000-000000000000
	github.com/FastFilter/xorfilter v0.0.0-00010101000000-000000000000 // indirect
	github.com/local/bbhash v0.0.0-00010101000000-000000000000 // indirect
	github.com/local/pthash v0.0.0-00010101000000-000000000000 // indirect
	github.com/local/consensusrecsplit v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/sync v0.12.0 // indirect
)

require github.com/relab/bbhash v0.0.0-20250331135148-7358f69256fb // indirect
