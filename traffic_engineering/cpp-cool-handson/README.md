# cpp-cool-handson

Hands-on C++ systems programming, concurrency, networking, and more. Built with **C++20** and **CMake**.

## Project Structure

```
cpp-cool-handson/
├── CMakeLists.txt
├── concurrency/        # Producer-Consumer, condition variables
├── multithreading/     # Thread pool, mutex patterns, deadlock avoidance
├── networking/         # TCP echo server/client (cross-platform)
├── systems/            # Signal handling, file watcher
├── ipc/                # Shared memory inter-process communication
├── memory/             # Smart pointers, custom memory pool allocator
└── io/                 # Async file I/O with futures
```

## Build

```bash
cmake -B build -DCMAKE_BUILD_TYPE=Debug
cmake --build build
```

## Targets

| Target              | Description                              |
|---------------------|------------------------------------------|
| `producer_consumer` | Bounded buffer with condition variables  |
| `thread_pool`       | Work-stealing thread pool with futures   |
| `mutex_basics`      | Mutex, shared_mutex, scoped_lock demos   |
| `tcp_echo_server`   | Multi-threaded TCP echo server           |
| `tcp_echo_client`   | TCP client for the echo server           |
| `signal_handling`   | Graceful shutdown via signal handlers    |
| `file_watcher`      | Filesystem change polling watcher        |
| `shared_memory_demo`| IPC via shared memory (Win32/POSIX)      |
| `smart_pointers`    | unique_ptr, shared_ptr, weak_ptr demos   |
| `memory_pool`       | Fixed-size pool allocator + benchmark    |
| `async_file_io`     | Async reads/writes with std::async       |

## Requirements

- C++20 compiler (MSVC, GCC 10+, Clang 12+)
- CMake 3.20+
