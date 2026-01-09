#include "pthash.hpp"
#include <vector>
#include <stdint.h>

extern "C" {
    typedef void* PTHashPtr;

    PTHashPtr pthash_init(uint64_t* keys, size_t n) {
        using namespace pthash;

        typedef dense_partitioned_phf<xxhash_128, opt_bucketer, R_int, true> pthash_type;
        
        pthash_type* f = new pthash_type();
        build_configuration config;
        config.alpha = 0.98;
        config.lambda = 5;
        config.num_threads = 1;

        config.dense_partitioning = (n > 100000); 
        config.verbose = false;

        std::vector<uint64_t> keys_vec(keys, keys + n);
        f->build_in_internal_memory(keys_vec.begin(), keys_vec.size(), config);
        return (void*)f;
    }

    uint64_t pthash_lookup(PTHashPtr ptr, uint64_t key) {
        using namespace pthash;
        typedef dense_partitioned_phf<xxhash_128, opt_bucketer, R_int, true> pthash_type;
        return (*(pthash_type*)ptr)(key);
    }
    
    uint64_t pthash_get_bits(PTHashPtr ptr) {
        using namespace pthash;
        typedef dense_partitioned_phf<xxhash_128, opt_bucketer, R_int, true> pthash_type;
        return static_cast<pthash_type*>(ptr)->num_bits();
    }

    void pthash_free(PTHashPtr ptr) {
        using namespace pthash;
        typedef dense_partitioned_phf<xxhash_128, opt_bucketer, R_int, true> pthash_type;
        delete (pthash_type*)ptr;
    }
}