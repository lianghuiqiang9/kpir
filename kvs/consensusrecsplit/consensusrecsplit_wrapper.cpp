#include "consensusrecsplit_wrapper.hpp"
#include <memory>
#include <vector>

struct MPHWrapper {
    std::unique_ptr<consensus::ConsensusRecSplit<8192, 0.01>> mph;
};

extern "C" {

MPHWrapper* create_mph(const uint64_t* keys, size_t num_keys) {
    auto wrapper = new MPHWrapper();
    std::vector<uint64_t> v(keys, keys + num_keys);
    wrapper->mph = std::make_unique<consensus::ConsensusRecSplit<8192, 0.01>>(v);
    return wrapper;
}

uint64_t query_mph(MPHWrapper* wrapper, uint64_t key) {
    return wrapper->mph->operator()(key);
}

size_t mph_bits(MPHWrapper* wrapper) {
    return wrapper->mph->getBits();
}

void free_mph(MPHWrapper* wrapper) {
    delete wrapper;
}

}
