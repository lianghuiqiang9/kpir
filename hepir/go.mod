module hepir

go 1.24.11

replace (
github.com/local/utils => ../utils
github.com/local/simplepir => ./simplepir
)

require (
	github.com/local/utils v0.0.0-00010101000000-000000000000
	github.com/local/simplepir v0.0.0-00010101000000-000000000000
)