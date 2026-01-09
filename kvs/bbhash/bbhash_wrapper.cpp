#include "BBHash/BooPHF.h"
#include <stdint.h>
#include <vector>

typedef boomphf::SingleHashFunctor<uint64_t> hasher_t;
typedef boomphf::mphf<uint64_t, hasher_t> boophf_t;

extern "C" {
    boophf_t* bbhash_build(const uint64_t* keys, size_t n, int threads, double gamma) {
        auto it_range = boomphf::range(keys, keys + n);
    
        return new boophf_t(n, it_range, threads, gamma, false, false); 
    }

    uint64_t bbhash_lookup(boophf_t* handle, uint64_t key) {
        if (!handle) return 0;
        return handle->lookup(key);
    }

    uint64_t bbhash_total_bits(boophf_t* handle) {
        if (!handle) return 0;
        return handle->totalBitSize();
    }

    void bbhash_free(boophf_t* handle) {
        if (handle) delete handle;
    }
}