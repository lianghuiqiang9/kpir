
g++ -O3 -std=c++11 -fPIC -shared -o libbbhash_wrapper.so bbhash_wrapper.cpp

go test -run TestBBHash