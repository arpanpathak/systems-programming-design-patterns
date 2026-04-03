// Shared memory IPC demo using memory-mapped files (cross-platform concept)
// On Windows, uses named shared memory via Win32 API
// On Linux/macOS, uses POSIX shared memory
#include <iostream>
#include <cstring>
#include <thread>
#include <chrono>

#ifdef _WIN32
    #include <windows.h>
#else
    #include <sys/mman.h>
    #include <sys/stat.h>
    #include <fcntl.h>
    #include <unistd.h>
#endif

constexpr const char* SHM_NAME = "cpp_cool_shm";
constexpr size_t SHM_SIZE = 256;

struct SharedData {
    int counter;
    char message[240];
};

void writer_process() {
#ifdef _WIN32
    HANDLE hMapFile = CreateFileMappingA(
        INVALID_HANDLE_VALUE, NULL, PAGE_READWRITE, 0, SHM_SIZE, SHM_NAME);
    if (!hMapFile) { std::cerr << "CreateFileMapping failed\n"; return; }
    auto* data = static_cast<SharedData*>(
        MapViewOfFile(hMapFile, FILE_MAP_ALL_ACCESS, 0, 0, SHM_SIZE));
#else
    int fd = shm_open(SHM_NAME, O_CREAT | O_RDWR, 0666);
    ftruncate(fd, SHM_SIZE);
    auto* data = static_cast<SharedData*>(
        mmap(nullptr, SHM_SIZE, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0));
#endif

    if (!data) { std::cerr << "Mapping failed\n"; return; }

    for (int i = 1; i <= 10; ++i) {
        data->counter = i;
        snprintf(data->message, sizeof(data->message), "Hello from writer, tick %d", i);
        std::cout << "[Writer] Wrote: counter=" << i << "\n";
        std::this_thread::sleep_for(std::chrono::milliseconds(200));
    }

#ifdef _WIN32
    UnmapViewOfFile(data);
    CloseHandle(hMapFile);
#else
    munmap(data, SHM_SIZE);
    close(fd);
#endif
}

void reader_process() {
    std::this_thread::sleep_for(std::chrono::milliseconds(100)); // let writer init

#ifdef _WIN32
    HANDLE hMapFile = OpenFileMappingA(FILE_MAP_ALL_ACCESS, FALSE, SHM_NAME);
    if (!hMapFile) { std::cerr << "OpenFileMapping failed\n"; return; }
    auto* data = static_cast<SharedData*>(
        MapViewOfFile(hMapFile, FILE_MAP_ALL_ACCESS, 0, 0, SHM_SIZE));
#else
    int fd = shm_open(SHM_NAME, O_RDONLY, 0666);
    auto* data = static_cast<SharedData*>(
        mmap(nullptr, SHM_SIZE, PROT_READ, MAP_SHARED, fd, 0));
#endif

    if (!data) { std::cerr << "Mapping failed\n"; return; }

    int last_seen = 0;
    while (last_seen < 10) {
        if (data->counter != last_seen) {
            last_seen = data->counter;
            std::cout << "[Reader] Read: counter=" << last_seen
                      << " msg=\"" << data->message << "\"\n";
        }
        std::this_thread::sleep_for(std::chrono::milliseconds(50));
    }

#ifdef _WIN32
    UnmapViewOfFile(data);
    CloseHandle(hMapFile);
#else
    munmap(data, SHM_SIZE);
    close(fd);
    shm_unlink(SHM_NAME);
#endif
}

int main() {
    std::cout << "=== Shared Memory IPC Demo ===\n\n";
    std::thread writer(writer_process);
    std::thread reader(reader_process);
    writer.join();
    reader.join();
    std::cout << "\n=== Done ===\n";
    return 0;
}
