#include "BBHash/BooPHF.h"
#include <stdint.h>
#include <vector>

// 实例化 uint64_t 版本的模板
typedef boomphf::SingleHashFunctor<uint64_t> hasher_t;
typedef boomphf::mphf<uint64_t, hasher_t> boophf_t;

extern "C" {
    // 1. Build: 构建并返回对象指针
    boophf_t* bbhash_build(const uint64_t* keys, size_t n, int threads, double gamma) {
        auto it_range = boomphf::range(keys, keys + n);
    
        // 关键点：将倒数第二个参数设为 false
        return new boophf_t(n, it_range, threads, gamma, false, false); 
    }

    // 2. Lookup: 正向查询
    uint64_t bbhash_lookup(boophf_t* handle, uint64_t key) {
        if (!handle) return 0;
        return handle->lookup(key);
    }

    // 3. totalBitSize: 获取占用位数
    uint64_t bbhash_total_bits(boophf_t* handle) {
        if (!handle) return 0;
        return handle->totalBitSize();
    }

    // 必须提供释放函数
    void bbhash_free(boophf_t* handle) {
        if (handle) delete handle;
    }
}