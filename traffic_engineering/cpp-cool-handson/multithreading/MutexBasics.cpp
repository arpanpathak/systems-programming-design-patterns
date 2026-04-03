// Mutex, lock_guard, unique_lock, shared_mutex, and deadlock avoidance
#include <iostream>
#include <thread>
#include <mutex>
#include <shared_mutex>
#include <vector>
#include <chrono>

// --- 1. Basic mutex with lock_guard ---
class Counter {
    int count_ = 0;
    std::mutex mtx_;
public:
    void increment() {
        std::lock_guard<std::mutex> lock(mtx_);
        ++count_;
    }
    int get() const { return count_; }
};

// --- 2. Reader-Writer lock with shared_mutex ---
class SharedConfig {
    std::string value_ = "default";
    mutable std::shared_mutex mtx_;
public:
    std::string read() const {
        std::shared_lock lock(mtx_);  // multiple readers OK
        return value_;
    }
    void write(const std::string& val) {
        std::unique_lock lock(mtx_);  // exclusive writer
        value_ = val;
    }
};

// --- 3. Deadlock avoidance with std::scoped_lock ---
class BankAccount {
    double balance_;
    std::mutex mtx_;
    std::string name_;
public:
    BankAccount(std::string name, double balance)
        : name_(std::move(name)), balance_(balance) {}

    friend void transfer(BankAccount& from, BankAccount& to, double amount) {
        // scoped_lock acquires both locks without deadlock
        std::scoped_lock lock(from.mtx_, to.mtx_);
        if (from.balance_ >= amount) {
            from.balance_ -= amount;
            to.balance_ += amount;
            std::cout << "Transferred " << amount
                      << " from " << from.name_ << " to " << to.name_ << "\n";
        }
    }

    void print() const {
        std::cout << name_ << ": $" << balance_ << "\n";
    }
};

int main() {
    // 1. Counter demo
    std::cout << "=== Mutex Counter ===\n";
    Counter counter;
    std::vector<std::thread> threads;
    for (int i = 0; i < 10; ++i) {
        threads.emplace_back([&counter] {
            for (int j = 0; j < 1000; ++j) counter.increment();
        });
    }
    for (auto& t : threads) t.join();
    std::cout << "Final count: " << counter.get() << " (expected 10000)\n\n";

    // 2. Reader-Writer demo
    std::cout << "=== Shared Mutex ===\n";
    SharedConfig config;
    std::thread writer([&config] {
        for (int i = 0; i < 5; ++i) {
            config.write("value_" + std::to_string(i));
            std::this_thread::sleep_for(std::chrono::milliseconds(50));
        }
    });
    std::thread reader([&config] {
        for (int i = 0; i < 10; ++i) {
            std::cout << "Read: " << config.read() << "\n";
            std::this_thread::sleep_for(std::chrono::milliseconds(25));
        }
    });
    writer.join();
    reader.join();

    // 3. Deadlock avoidance demo
    std::cout << "\n=== Scoped Lock (Deadlock Avoidance) ===\n";
    BankAccount alice("Alice", 1000);
    BankAccount bob("Bob", 1000);
    std::thread t1([&] { for (int i = 0; i < 5; ++i) transfer(alice, bob, 100); });
    std::thread t2([&] { for (int i = 0; i < 5; ++i) transfer(bob, alice, 50); });
    t1.join();
    t2.join();
    alice.print();
    bob.print();

    return 0;
}
