// Simple fixed-size memory pool allocator
#include <iostream>
#include <vector>
#include <cstddef>
#include <cassert>
#include <chrono>

class MemoryPool {
    struct Block {
        Block* next;
    };

    std::vector<char> pool_;
    Block* free_list_ = nullptr;
    size_t block_size_;
    size_t num_blocks_;
    size_t allocated_ = 0;

public:
    MemoryPool(size_t block_size, size_t num_blocks)
        : block_size_(std::max(block_size, sizeof(Block)))
        , num_blocks_(num_blocks)
        , pool_(block_size_ * num_blocks)
    {
        // Build free list
        for (size_t i = 0; i < num_blocks_; ++i) {
            auto* block = reinterpret_cast<Block*>(pool_.data() + i * block_size_);
            block->next = free_list_;
            free_list_ = block;
        }
    }

    void* allocate() {
        if (!free_list_) {
            throw std::bad_alloc();
        }
        Block* block = free_list_;
        free_list_ = free_list_->next;
        ++allocated_;
        return block;
    }

    void deallocate(void* ptr) {
        auto* block = static_cast<Block*>(ptr);
        block->next = free_list_;
        free_list_ = block;
        --allocated_;
    }

    size_t allocated_count() const { return allocated_; }
    size_t capacity() const { return num_blocks_; }
};

struct Particle {
    float x, y, z;
    float vx, vy, vz;
    float life;
};

int main() {
    constexpr size_t N = 100'000;

    // Benchmark: pool allocator vs new/delete
    std::cout << "=== Memory Pool Benchmark (" << N << " allocations) ===\n\n";

    // Pool allocator
    {
        MemoryPool pool(sizeof(Particle), N);
        std::vector<void*> ptrs;
        ptrs.reserve(N);

        auto start = std::chrono::high_resolution_clock::now();
        for (size_t i = 0; i < N; ++i) {
            ptrs.push_back(pool.allocate());
        }
        for (auto* p : ptrs) {
            pool.deallocate(p);
        }
        auto end = std::chrono::high_resolution_clock::now();
        auto us = std::chrono::duration_cast<std::chrono::microseconds>(end - start).count();
        std::cout << "Pool allocator:    " << us << " us\n";
        std::cout << "  Allocated: " << pool.allocated_count()
                  << " / " << pool.capacity() << "\n";
    }

    // Standard new/delete
    {
        std::vector<Particle*> ptrs;
        ptrs.reserve(N);

        auto start = std::chrono::high_resolution_clock::now();
        for (size_t i = 0; i < N; ++i) {
            ptrs.push_back(new Particle{});
        }
        for (auto* p : ptrs) {
            delete p;
        }
        auto end = std::chrono::high_resolution_clock::now();
        auto us = std::chrono::duration_cast<std::chrono::microseconds>(end - start).count();
        std::cout << "new/delete:        " << us << " us\n";
    }

    return 0;
}
