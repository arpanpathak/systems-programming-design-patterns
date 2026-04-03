// Producer-Consumer using condition variables and mutex
#include <iostream>
#include <thread>
#include <mutex>
#include <condition_variable>
#include <queue>
#include <chrono>
#include <atomic>

class BoundedBuffer {
    std::queue<int> buffer_;
    size_t capacity_;
    std::mutex mtx_;
    std::condition_variable not_full_;
    std::condition_variable not_empty_;

public:
    explicit BoundedBuffer(size_t capacity) : capacity_(capacity) {}

    void produce(int item) {
        std::unique_lock<std::mutex> lock(mtx_);
        not_full_.wait(lock, [this] { return buffer_.size() < capacity_; });
        buffer_.push(item);
        std::cout << "[Producer] Produced: " << item
                  << " | Buffer size: " << buffer_.size() << "\n";
        not_empty_.notify_one();
    }

    int consume() {
        std::unique_lock<std::mutex> lock(mtx_);
        not_empty_.wait(lock, [this] { return !buffer_.empty(); });
        int item = buffer_.front();
        buffer_.pop();
        std::cout << "[Consumer] Consumed: " << item
                  << " | Buffer size: " << buffer_.size() << "\n";
        not_full_.notify_one();
        return item;
    }
};

int main() {
    BoundedBuffer buffer(5);
    std::atomic<bool> done{false};

    // Producer thread
    std::thread producer([&] {
        for (int i = 1; i <= 20; ++i) {
            buffer.produce(i);
            std::this_thread::sleep_for(std::chrono::milliseconds(50));
        }
        done = true;
    });

    // Consumer threads
    std::thread consumer1([&] {
        while (!done || true) {
            buffer.consume();
            std::this_thread::sleep_for(std::chrono::milliseconds(120));
            if (done) break;
        }
    });

    std::thread consumer2([&] {
        while (!done || true) {
            buffer.consume();
            std::this_thread::sleep_for(std::chrono::milliseconds(150));
            if (done) break;
        }
    });

    producer.join();
    // Give consumers time to drain
    std::this_thread::sleep_for(std::chrono::milliseconds(500));
    done = true;
    consumer1.detach();
    consumer2.detach();

    std::cout << "\n=== Producer-Consumer demo complete ===\n";
    return 0;
}
