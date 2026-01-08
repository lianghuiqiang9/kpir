module kvs

go 1.24.11

replace (
	github.com/FastFilter/xorfilter => ./xorfilter
	github.com/local/utils => ../utils
	github.com/local/consensusrecsplit => ./consensusrecsplit
	github.com/local/bbhash => ./bbhash
	github.com/local/pthash => ./pthash
)

require (
	github.com/FastFilter/xorfilter v0.0.0-00010101000000-000000000000
	github.com/local/utils v0.0.0-00010101000000-000000000000
	github.com/local/consensusrecsplit v0.0.0-00010101000000-000000000000
	github.com/local/bbhash v0.0.0-00010101000000-000000000000
	github.com/local/pthash v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.9.0
)

require golang.org/x/sync v0.12.0 // indirect

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/relab/bbhash v0.0.0-20250331135148-7358f69256fb
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
