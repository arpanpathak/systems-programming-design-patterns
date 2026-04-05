# Advanced Systems Programming

This repository is a comprehensive, modular study guide for mastering **Advanced Go Concurrency, High-Performance Proxy Data Structures, OS/Systems Programming, and Network Resilience**. The examples mirror the internal architectures of industry-standard ingress controllers and Service Mesh proxies (like Envoy, NGINX, and Kubernetes networking).

All files are heavily commented to explain *why* these patterns are used in production Edge environments, making this the perfect curriculum to ace systems programming interview.

---

## 📚 Table of Contents & File Index

### 1. High-Performance Data Structures
In massive proxy architectures, global mutex locks destroy performance. These structures unlock high concurrency.
- **[Sharded LRU Cache](./concurrent-lru-cache/concurrent-lru-cache.go)**: Eliminates global lock contention by hashing keys across an array of 256 individual `sync.RWMutex` shards. 
- **[Lock-Free Ring Buffer](./thread-safe-queue/ring_buffer.go)**: A zero-allocation, lock-free circular queue utilizing `sync/atomic` to process packets without triggering the Go Garbage Collector.
- **[Blocking Queue](./thread-safe-queue/blocking-queue.go)**: FIFO queue utilizing `sync.Cond` for efficient producer-consumer synchronization.
- **[Hierarchical Timing Wheel](./data_structures/timing_wheel.go)**: Scales millions of active socket timeouts efficiently with O(1) tick execution, replacing the resource-heavy `time.AfterFunc()`.
- **[Lock-Free Fundamentals](./traffic_engineering/go-network-systems/concurrency/lock_free/main.go)**: Implementing atomic-based lock-free data structures.

### 2. OS Kernel & Systems Programming
Understanding what the Go Standard Library hides under the abstractions.
- **[Robust File I/O](./os_io/file_operations.go)**: User-space batching (`bufio`), Kernel page caches, atomic appends (`O_APPEND`), and zero-copy DMA utilizing the `sendfile` syscall.
- **[WAL Durability (`fsync`)](./os_io/durability_fsync.go)**: Guaranteeing hardware-level disk persistence using the Unix `fsync` syscall.
- **[Basic File Operations](./os_io/file_main.go)**: Standard file I/O, appending, and hex binary encoding for systems data.
- **[Systems: File IO Deep Dive](./traffic_engineering/go-network-systems/systems/file_io/main.go)**: Low-level file descriptors and buffer management.
- **[Systems: Pipes & IPC](./traffic_engineering/go-network-systems/systems/pipes/main.go)**: Inter-process communication using Unix pipes.
- **[Systems: Signals Handling](./traffic_engineering/go-network-systems/systems/signals/main.go)**: Managing OS signals for graceful process control.
- **[Systems: Unsafe Pointers](./traffic_engineering/go-network-systems/systems/unsafe_pointers/main.go)**: Bypassing Go's type safety for direct memory access.
- **[Memory Mapping (`mmap`)](./systems/unix_mmap.go)**: Mapping absolute physical hard drive space into Virtual Memory natively for zero-copy file serving (used by BoltDB/caches).
- **[Darwin/macOS Kqueue Event Loop](./systems/unix_kqueue_darwin.go)**: Mimics Envoy's core routing engine. Shows how to handle 10,000+ sockets on a single thread using the `kqueue` syscall (equivalent to Linux `epoll`).
- **[Raw Socket Provisioning](./systems/unix_sockets_darwin.go)**: Bypassing Go's `net` package entirely to craft non-blocking Unix TCP sockets via C-level calls (`unix.Socket`, `unix.Bind`, `unix.Listen`).
- **[Systems: GC & Allocations](./traffic_engineering/go-network-systems/systems/garbage_collector/main.go)**: Understanding stack vs heap and Go's garbage collector.
- **[Systems: Process Management](./traffic_engineering/go-network-systems/systems/process/main.go)**: Command execution, timeouts, and environment variables.
- **[Systems: File Watcher](./traffic_engineering/go-network-systems/systems/file_watcher/main.go)**: Monitoring filesystem events for dynamic configuration reloads.

### 3. Concurrency Mastery
Controlling the Go Runtime flawlessly under massive request load.
- **[Channels Deep Dive](./concurrency_core/channels.go)**: Buffered vs. Unbuffered behavior, multiplexing via `select`, and directional locking.
- **[Concurrency: Select Patterns](./traffic_engineering/go-network-systems/concurrency/select_patterns/main.go)**: Advanced multiplexing, default cases, and non-blocking operations.
- **[ErrGroup Request Fan-Out](./concurrency_core/errgroup_example.go)**: Multi-backend fetching where one failure cancels the entire aggregate group.
- **[Bounded Worker Pool](./workerpool/pool.go)**: Preventing OOM crashes by strictly pacing Goroutine spawning using a Job Dispatcher pipeline.
- **[Context Deep Dive](./traffic_engineering/go-network-systems/concurrency/context/main.go)**: Advanced context propagation, cancellation trees, and value-passing.
- **[WaitGroup Synchronization](./waitgroup/waitgroup_example.go)**: Coordinating parallel task execution with `sync.WaitGroup`.
- **[Classic Ping-Pong](./classic-ping-pong/main.go)**: Demonstrating synchronized communication between two goroutines using unbuffered channels and `sync.WaitGroup`.

- **[Concurrency: Work Stealing](./traffic_engineering/go-network-systems/concurrency/work_stealing/main.go)**: Custom scheduler patterns and task distribution.
- **[Lock-Free Hot Swapping (xDS)](./advanced_concurrency/lockfree_config.go)**: How Proxies update massive routing tables dynamically while traffic flows by swinging pointers with `sync/atomic.Value`.
- **[Atomic Benchmarks](./traffic_engineering/go-network-systems/concurrency/mutex/main.go)**: Comparing performance between `sync.Mutex` and `sync/atomic`.

### 4. Edge Proxying, Traffic & Routing
The lifeblood of the Data Plane.
- **[L7 Reverse Proxy & Header Injection](./proxy/reverse_proxy.go)**: Appending `X-Forwarded-For` and trace IDs dynamically using Go's `httputil.ReverseProxy`.
- **[Networking: Bandwidth Throttler](./traffic_engineering/go-network-systems/networking/bandwidth_throttler/main.go)**: Token-bucket traffic shaping for `io` streams.
- **[L4 TLS SNI Inspector](./proxy/tls_sni/sni_inspector.go)**: Peeking inside raw TLS *ClientHello* frames to extract the Server Name Indication (SNI) for routing decisions *without* decrypting the SSL payload!
- **[Networking: TCP/HTTP/UDP Servers](./traffic_engineering/go-network-systems/networking/tcp_server/main.go)**: Building custom protocol handlers from the ground up.
- **[TCP Keep-Alive Connection Pool](./proxy/connection_pool/pool.go)**: Reusing upstream proxy sockets and gracefully blocking queued Goroutines using `sync.Cond`.
- **[Networking: Port Scanner](./traffic_engineering/go-network-systems/networking/port_scanner/main.go)**: Efficient reconnaissance using parallelized socket connections.
- **[Passive Outlier Detection](./traffic_engineering/passive_outlier.go)**: Envoy's pattern for dynamically monitoring and ejecting a bad upstream node mathematically without active health checks.
- **[Graceful Shutdown Listener](./graceful/shutdown.go)**: Trapping `SIGINT/SIGTERM` OS Signals to drain active proxy traffic connections gracefully during deployments safely.
- **[Networking Basics](./computer-networks/networking-basics.go)**: Foundational TCP/IP server implementation.

### 5. Load Balancing Algorithms
Distributing throughput flawlessly across cluster farms.
- **[Thread-Safe Round Robin](./loadbalancer/round_robin.go)**: Purely atomic, lock-free routing iteration (`atomic.AddUint32`).
- **[Consistent Hashing Ring](./loadbalancer/consistent_hash.go)**: Implementing a Virtual-Node Hash map with `crc32` hashing and `sort.Search` for deterministic sticky-session routing without mass-resharding.

### 6. Mesh Resiliency & Rate Limiting
Protecting your cluster from cascading downtime and Thundering Herds.
- **[Circuit Breaker State Machine](./resilience/circuit_breaker.go)**: Mathematical `Closed` -> `Open` -> `HalfOpen` states used to fast-fail traffic and protect a dying upstream database.
- **[Sliding Window Log Rate Limiter](./rate-limiter/sliding_window.go)**: Perfect-precision time-pruning API rate limiting vs Token Buckets.
- **[Token Bucket Rate Limiter (API)](./rate-limiter/token-bucket.go)**: Robust API-level rate limiting implementation.
- **[Bulkhead Resource Isolation](./resiliency/bulkhead_isolation.go)**: Using Weighted Semaphores (`golang.org/x/sync/semaphore`) to ensure an external API outage never consumes all proxy resources.
- **[Exponential Jitter Backoff](./resiliency/jitter_backoff.go)**: Randomized retry algorithms designed to stop 50,000 proxies from crashing a database upon reboot (Thundering Herd shield).
- **[Raw Network Server Protocols](./tcp/server.go) & [UDP Host (`sync.Pool`)](./udp/server.go)**: Foundational packet framing, TCP Read Deadlines, and UDP zero-allocation Datagram parsing.
- **[Simple Rate Limiter](./main.go)**: Lightweight token-refill implementation.
