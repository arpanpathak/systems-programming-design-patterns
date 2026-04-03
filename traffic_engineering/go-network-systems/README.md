# Go Network Systems - Apple Network Traffic Team Prep

Comprehensive Go project covering **concurrency**, **OS/systems programming**, **networking**, and **proxy implementations** — targeting Apple Network Traffic team interview preparation.

## Project Structure

Each directory is a standalone runnable program with its own `main()`.

```
go-network-systems/
├── concurrency/
│   ├── goroutines/           # Basic goroutines + WaitGroup
│   ├── channels/             # Unbuffered, buffered, direction
│   ├── select_patterns/      # Select, non-blocking, done channel, timeout
│   ├── fan_out_fan_in/       # Fan-out / fan-in pattern
│   ├── pipeline/             # Pipeline stages + or-channel
│   ├── mutex/                # Mutex + RWMutex
│   ├── sync_primitives/      # sync.Once, Cond, Map, Pool
│   ├── atomic_ops/           # Atomic operations, CAS, barrier
│   ├── semaphore_context/    # Semaphore + context cancellation/timeout
│   ├── worker_pool/          # Worker pool with job queue
│   ├── rate_limiter/         # Token bucket rate limiter
│   ├── circuit_breaker/      # Circuit breaker (closed/open/half-open)
│   ├── pubsub/               # Pub/sub with topics
│   ├── map_reduce/           # Map-reduce + bounded parallelism
│   ├── graceful_shutdown/    # Graceful shutdown + errgroup
│   ├── reactor_pattern/      # Reactor event loop
│   ├── lockfree_stack/       # Treiber stack with CAS
│   ├── concurrent_queue/     # Bounded queue with sync.Cond
│   ├── ring_buffer/          # SPSC lock-free ring buffer
│   ├── sharded_map/          # FNV-hash sharded concurrent map
│   ├── lru_cache/            # Concurrent LRU cache
│   ├── work_stealing/        # Work-stealing deque
│   └── rcu/                  # Read-Copy-Update pattern
├── systems/
│   ├── signals/              # Signal handling for graceful shutdown
│   ├── process/              # Process execution + environment variables
│   ├── file_io/              # File I/O, directory ops, atomic writes
│   ├── pipes/                # io.Pipe for inter-goroutine streaming
│   ├── sysinfo/              # CPU, memory, runtime stats
│   ├── file_watcher/         # Polling-based file watcher
│   ├── stack_vs_heap/        # Stack vs heap, escape analysis
│   ├── garbage_collector/    # sync.Pool, GC, finalizers
│   └── unsafe_pointers/      # Unsafe ops, struct alignment, cache lines
├── networking/
│   ├── tcp_echo/             # TCP echo server/client with keep-alive
│   ├── connection_pool/      # TCP connection pooling
│   ├── kv_protocol/          # Line-based key-value protocol server
│   ├── udp_echo/             # UDP echo server/client + MTU reference
│   ├── reliable_udp/         # Stop-and-Wait ARQ over UDP
│   ├── http_server/          # HTTP server + middleware chain + SSE
│   ├── http_client/          # HTTP client with retry + backoff
│   ├── dns/                  # DNS resolution + custom resolver
│   ├── tls_inspection/       # TLS connection + certificate inspection
│   ├── protocol_concepts/    # TCP/TLS/HTTP2/QUIC/NAT reference
│   ├── event_driven_server/  # Reactor pattern server + netpoller concepts
│   ├── bandwidth_throttler/  # Token bucket bandwidth throttler
│   ├── packet_parsing/       # IP/TCP header parsing from raw bytes
│   └── flow_tracking/        # Flow tracker, traffic shaping, network stats
└── proxy/
    ├── tcp_proxy/            # TCP L4 bidirectional proxy
    ├── reverse_proxy/        # HTTP reverse proxy + round-robin LB
    ├── connect_tunnel/       # HTTP CONNECT tunnel for HTTPS
    ├── consistent_hash/      # Consistent hashing for routing
    └── socks5/               # Full SOCKS5 server (RFC 1928)
```

## Running

Each directory is a standalone `main` package. Run any module directly:

```bash
cd go-network-systems

# Examples
go run ./concurrency/goroutines/
go run ./concurrency/lockfree_stack/
go run ./networking/tcp_echo/
go run ./proxy/socks5/

# Build all
go build ./...
```

## Topics Covered

### Concurrency (23 modules)
| Module | Concepts |
|--------|----------|
| `goroutines` | Goroutine lifecycle, WaitGroup |
| `channels` | Unbuffered, buffered, direction |
| `select_patterns` | Select, non-blocking, done channel, timeout |
| `fan_out_fan_in` | Fan-out / fan-in pattern |
| `pipeline` | Pipeline stages, or-channel |
| `mutex` | Mutex, RWMutex |
| `sync_primitives` | sync.Once, sync.Cond, sync.Map, sync.Pool |
| `atomic_ops` | Atomic operations, CAS, barrier |
| `semaphore_context` | Semaphore, context cancellation/timeout/value |
| `worker_pool` | Worker pool with job queue |
| `rate_limiter` | Token bucket rate limiter |
| `circuit_breaker` | Closed/open/half-open states |
| `pubsub` | Pub/sub with topics |
| `map_reduce` | Map-reduce + bounded parallelism |
| `graceful_shutdown` | Graceful shutdown + errgroup |
| `reactor_pattern` | Reactor event loop |
| `lockfree_stack` | Treiber stack with CAS |
| `concurrent_queue` | Bounded queue with sync.Cond |
| `ring_buffer` | SPSC lock-free ring buffer |
| `sharded_map` | FNV-hash sharded concurrent map |
| `lru_cache` | Concurrent LRU cache |
| `work_stealing` | Work-stealing deque |
| `rcu` | Read-Copy-Update pattern |

### Systems Programming (9 modules)
| Module | Concepts |
|--------|----------|
| `signals` | Signal handling for graceful shutdown |
| `process` | Process execution, environment variables |
| `file_io` | File I/O, directory ops, atomic writes |
| `pipes` | io.Pipe for inter-goroutine streaming |
| `sysinfo` | CPU, memory, runtime stats |
| `file_watcher` | Polling-based file watcher |
| `stack_vs_heap` | Stack vs heap, escape analysis |
| `garbage_collector` | sync.Pool, GC internals, finalizers |
| `unsafe_pointers` | Unsafe ops, struct alignment, cache lines |

### Networking (14 modules)
| Module | Concepts |
|--------|----------|
| `tcp_echo` | TCP server/client, keep-alive, TCP_NODELAY |
| `connection_pool` | TCP connection pooling |
| `kv_protocol` | Line-based key-value protocol server |
| `udp_echo` | UDP echo server/client, MTU reference |
| `reliable_udp` | Stop-and-Wait ARQ over UDP |
| `http_server` | HTTP server + middleware chain + SSE |
| `http_client` | HTTP client, retry with exponential backoff |
| `dns` | DNS resolution (A, MX, NS, TXT, CNAME, SRV), custom resolver |
| `tls_inspection` | TLS connection + certificate inspection |
| `protocol_concepts` | TCP/TLS 1.3/HTTP/2/QUIC/NAT/LB reference |
| `event_driven_server` | Reactor pattern server, epoll/kqueue/IOCP concepts |
| `bandwidth_throttler` | Token bucket bandwidth throttler |
| `packet_parsing` | IP/TCP header parsing from raw bytes |
| `flow_tracking` | Flow tracking, traffic shaping, DPI/QoS concepts |

### Proxy (5 modules)
| Module | Concepts |
|--------|----------|
| `tcp_proxy` | TCP L4 bidirectional proxy |
| `reverse_proxy` | HTTP reverse proxy + round-robin load balancing |
| `connect_tunnel` | HTTP CONNECT tunnel for HTTPS |
| `consistent_hash` | Consistent hashing for routing |
| `socks5` | Full SOCKS5 server (RFC 1928), auth, address types |

## Key Interview Topics

### Apple Network Traffic Specifics
- **Network.framework**: Modern transport API (replaces BSD sockets)
- **NSURLSession**: Foundation-level HTTP/HTTPS
- **NEFilterProvider**: Content filtering
- **NEDNSProxyProvider**: DNS proxy provider
- **NEPacketTunnelProvider**: VPN/tunneling
- **App Transport Security (ATS)**: HTTPS enforcement
- **Private Relay (iCloud)**: MASQUE-based two-hop proxy
- **Encrypted Client Hello (ECH)**: Privacy for TLS SNI
- **Oblivious DNS over HTTPS (ODoH)**: Privacy-preserving DNS

### Performance Concepts
- **Zero-copy I/O**: sendfile(), splice(), io_uring
- **Kernel bypass**: DPDK, XDP, AF_XDP
- **Lock-free data structures**: CAS-based algorithms
- **Memory-mapped I/O**: mmap for efficient file access
- **Connection pooling**: Reuse TCP connections
- **SO_REUSEPORT**: Multiple listeners on same port

### Go Runtime Internals
- **M:N scheduling**: Goroutines on OS threads
- **Netpoller**: Async I/O via epoll/kqueue/IOCP
- **Work stealing**: Scheduler balances goroutines across Ps
- **GC**: Concurrent tri-color mark-and-sweep with write barriers
- **Stack growth**: Contiguous stacks, copy on growth
