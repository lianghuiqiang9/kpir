#pragma once
#include <cstdint>
#include <cstddef>
#include "ConsensusRecSplit.h"
#ifdef __cplusplus
extern "C" {
#endif

typedef struct MPHWrapper MPHWrapper;


MPHWrapper* create_mph(const uint64_t* keys, size_t num_keys);


uint64_t query_mph(MPHWrapper* wrapper, uint64_t key);


size_t mph_bits(MPHWrapper* wrapper);


void free_mph(MPHWrapper* wrapper);

#ifdef __cplusplus
}
#endif
