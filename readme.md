
# build

git submodule update --init --recursive

## build bbhash

cd ./kvs/bbhash/
bash build_bbhash.sh

## build pthash

cd ..
cd ./pthash/
bash build_pthash.sh

## build consensusrecsplit

cd ..
cd ./consensusrecsplit/
bash build_consensusrecsplit.sh

# run

cd ..
cd ..
go run main.go

cd kvs
bash kvs_benchmark.sh

cd ..
cd hepir
bash hepir_benchmark.sh

cd ..
cd sipir
bash sipir_benchmark.sh

cd ..
bash kpir_benchmark.sh

