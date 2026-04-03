// Signal handling demo
#include <iostream>
#include <csignal>
#include <thread>
#include <chrono>
#include <atomic>

std::atomic<bool> running{true};

void signal_handler(int signum) {
    std::cout << "\n[Signal] Caught signal " << signum << "\n";
    running = false;
}

int main() {
    // Register signal handlers
    std::signal(SIGINT, signal_handler);   // Ctrl+C
    std::signal(SIGTERM, signal_handler);  // Termination request

    std::cout << "Signal handling demo - press Ctrl+C to stop\n";
    std::cout << "PID: " <<
#ifdef _WIN32
        GetCurrentProcessId()
#else
        getpid()
#endif
    << "\n";

    int counter = 0;
    while (running) {
        std::cout << "Working... tick " << ++counter << "\r" << std::flush;
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }

    std::cout << "Graceful shutdown after " << counter << " ticks\n";
    return 0;
}
