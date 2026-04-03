cmake_minimum_required(VERSION 3.20)
project(cpp_cool_handson LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Platform-specific settings
if(WIN32)
    # Link Winsock for networking on Windows
    set(PLATFORM_LIBS ws2_32 wsock32)
else()
    find_package(Threads REQUIRED)
    set(PLATFORM_LIBS Threads::Threads)
endif()

# --- Concurrency ---
add_executable(producer_consumer concurrency/ProducerConsumer.cpp)
target_link_libraries(producer_consumer PRIVATE ${PLATFORM_LIBS})

# --- Multithreading ---
add_executable(thread_pool multithreading/ThreadPool.cpp)
target_link_libraries(thread_pool PRIVATE ${PLATFORM_LIBS})

add_executable(mutex_basics multithreading/MutexBasics.cpp)
target_link_libraries(mutex_basics PRIVATE ${PLATFORM_LIBS})

# --- Networking ---
add_executable(tcp_echo_server networking/TcpEchoServer.cpp)
target_link_libraries(tcp_echo_server PRIVATE ${PLATFORM_LIBS})

add_executable(tcp_echo_client networking/TcpEchoClient.cpp)
target_link_libraries(tcp_echo_client PRIVATE ${PLATFORM_LIBS})

# --- Systems Programming ---
add_executable(signal_handling systems/SignalHandling.cpp)
target_link_libraries(signal_handling PRIVATE ${PLATFORM_LIBS})

add_executable(file_watcher systems/FileWatcher.cpp)
target_link_libraries(file_watcher PRIVATE ${PLATFORM_LIBS})

# --- IPC ---
add_executable(shared_memory_demo ipc/SharedMemoryDemo.cpp)
target_link_libraries(shared_memory_demo PRIVATE ${PLATFORM_LIBS})

# --- Memory ---
add_executable(smart_pointers memory/SmartPointers.cpp)
target_link_libraries(smart_pointers PRIVATE ${PLATFORM_LIBS})

add_executable(memory_pool memory/MemoryPool.cpp)
target_link_libraries(memory_pool PRIVATE ${PLATFORM_LIBS})

# --- I/O ---
add_executable(async_file_io io/AsyncFileIO.cpp)
target_link_libraries(async_file_io PRIVATE ${PLATFORM_LIBS})
