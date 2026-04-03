// File watcher - polls for file changes using filesystem
#include <iostream>
#include <filesystem>
#include <unordered_map>
#include <string>
#include <chrono>
#include <thread>

namespace fs = std::filesystem;

class FileWatcher {
    std::string watch_path_;
    std::chrono::milliseconds interval_;
    std::unordered_map<std::string, fs::file_time_type> file_times_;

    void check_for_changes() {
        // Check for new or modified files
        for (auto& entry : fs::recursive_directory_iterator(watch_path_)) {
            if (!entry.is_regular_file()) continue;
            auto path = entry.path().string();
            auto last_write = fs::last_write_time(entry);

            if (file_times_.find(path) == file_times_.end()) {
                file_times_[path] = last_write;
                std::cout << "[Created]  " << path << "\n";
            } else if (file_times_[path] != last_write) {
                file_times_[path] = last_write;
                std::cout << "[Modified] " << path << "\n";
            }
        }

        // Check for deleted files
        auto it = file_times_.begin();
        while (it != file_times_.end()) {
            if (!fs::exists(it->first)) {
                std::cout << "[Deleted]  " << it->first << "\n";
                it = file_times_.erase(it);
            } else {
                ++it;
            }
        }
    }

public:
    FileWatcher(std::string path, std::chrono::milliseconds interval)
        : watch_path_(std::move(path)), interval_(interval) {}

    void start() {
        std::cout << "Watching: " << watch_path_ << " (every "
                  << interval_.count() << "ms)\n";
        std::cout << "Press Ctrl+C to stop\n\n";

        while (true) {
            check_for_changes();
            std::this_thread::sleep_for(interval_);
        }
    }
};

int main(int argc, char* argv[]) {
    std::string path = (argc > 1) ? argv[1] : ".";

    if (!fs::exists(path)) {
        std::cerr << "Path does not exist: " << path << "\n";
        return 1;
    }

    FileWatcher watcher(path, std::chrono::milliseconds(1000));
    watcher.start();
    return 0;
}
