
cd ConsensusRecSplit
mkdir build
cd build
cmake ..
make

cd ..
cd ..

sed -i '2i#include <iostream>' ConsensusRecSplit/include/consensus/UnalignedBitVector.h

g++ -std=c++20 -O3 -fPIC -shared \
  -I./ConsensusRecSplit/include \
  -I./ConsensusRecSplit/extlib/ips2ra/include \
  -I./ConsensusRecSplit/extlib/util/include \
  -I./ConsensusRecSplit/extlib/util/extern/pasta-bit-vector/include \
  -I./ConsensusRecSplit/extlib/tlx \
  -I./ConsensusRecSplit/build/_deps/pasta_utils-src/include \
  -I./ConsensusRecSplit/extlib/fips/include \
  -o libconsensusrecsplit_wrapper.so \
  consensusrecsplit_wrapper.cpp

go test -run TestConsensusRecSplit