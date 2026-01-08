
# go 调用c++库
1. 编译
cd ConsensusRecSplit
mkdir build
cd build
cmake ..
make

2. Add #include<iostream> to ConsensusRecSplit/include/consensus/UnalignedBitVector.h

3. 
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

2. 链接到路径
    cd ..
    cd ..
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:.
3. 运行
    go test -run TestConsensusRecSplit