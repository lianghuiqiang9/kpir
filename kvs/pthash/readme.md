
1. 
cd pthash
git submodule update --init --recursive

2. 

g++ -O3 -fPIC -shared -march=native -std=c++17 pthash_wrapper.cpp \
    -I ./pthash/include \
    -I ./pthash/external/bits/external/essentials/include \
    -I ./pthash/external/mm_file/include \
    -I ./pthash/external/xxHash \
    -I ./pthash/external/bits/include \
    -lpthread -o libpthash_wrapper.so

3. 

