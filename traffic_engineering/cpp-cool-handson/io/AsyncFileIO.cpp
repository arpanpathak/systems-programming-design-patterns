// Async file I/O using std::async and futures
#include <iostream>
#include <fstream>
#include <string>
#include <future>
#include <vector>
#include <chrono>
#include <filesystem>

namespace fs = std::filesystem;

// Async file write
std::future<bool> async_write(const std::string& path, const std::string& content) {
    return std::async(std::launch::async, [path, content] {
        std::ofstream file(path);
        if (!file) return false;
        file << content;
        std::cout << "[Write] Wrote " << content.size() << " bytes to " << path << "\n";
        return true;
    });
}

// Async file read
std::future<std::string> async_read(const std::string& path) {
    return std::async(std::launch::async, [path] {
        std::ifstream file(path);
        if (!file) return std::string("ERROR: could not open " + path);
        std::string content((std::istreambuf_iterator<char>(file)),
                             std::istreambuf_iterator<char>());
        std::cout << "[Read]  Read " << content.size() << " bytes from " << path << "\n";
        return content;
    });
}

// Async file copy
std::future<bool> async_copy(const std::string& src, const std::string& dst) {
    return std::async(std::launch::async, [src, dst] {
        try {
            fs::copy_file(src, dst, fs::copy_options::overwrite_existing);
            std::cout << "[Copy]  " << src << " -> " << dst << "\n";
            return true;
        } catch (const fs::filesystem_error& e) {
            std::cerr << "[Copy Error] " << e.what() << "\n";
            return false;
        }
    });
}

int main() {
    std::cout << "=== Async File I/O Demo ===\n\n";
    const std::string dir = "async_io_test";
    fs::create_directories(dir);

    // Fire off multiple async writes
    std::vector<std::future<bool>> write_futures;
    for (int i = 0; i < 5; ++i) {
        std::string path = dir + "/file_" + std::to_string(i) + ".txt";
        std::string content = "Content of file " + std::to_string(i) +
                              " -- written asynchronously!\n";
        write_futures.push_back(async_write(path, content));
    }

    // Wait for all writes
    for (auto& f : write_futures) f.get();
    std::cout << "\nAll writes complete.\n\n";

    // Fire off multiple async reads
    std::vector<std::future<std::string>> read_futures;
    for (int i = 0; i < 5; ++i) {
        std::string path = dir + "/file_" + std::to_string(i) + ".txt";
        read_futures.push_back(async_read(path));
    }

    // Collect reads
    std::cout << "\n--- File Contents ---\n";
    for (int i = 0; i < 5; ++i) {
        std::cout << "file_" << i << ": " << read_futures[i].get();
    }

    // Cleanup
    fs::remove_all(dir);
    std::cout << "\nCleaned up test directory.\n";
    return 0;
}
