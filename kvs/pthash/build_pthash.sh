
g++ -O3 -fPIC -shared -march=native -std=c++17 pthash_wrapper.cpp \
    -I ./pthash/include \
    -I ./pthash/external/bits/external/essentials/include \
    -I ./pthash/external/mm_file/include \
    -I ./pthash/external/xxHash \
    -I ./pthash/external/bits/include \
    -lpthread -o libpthash_wrapper.so

go test -run TestPTHash