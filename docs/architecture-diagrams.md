# Architecture Diagrams

Visual diagrams of the Hybrid-Grid distributed build system architecture.

## System Overview

```mermaid
graph TB
    subgraph Client["Client Machine"]
        CLI[hgbuild CLI]
        LC[Local Cache]
    end

    subgraph Coordinator["Coordinator Node"]
        COORD[hg-coord]
        SCHED[Scheduler<br/>P2C Algorithm]
        REG[Worker Registry]
        DASH[Dashboard<br/>:8080]
        METRICS[Prometheus<br/>Metrics]
    end

    subgraph Workers["Worker Pool"]
        W1[hg-worker 1<br/>macOS ARM64]
        W2[hg-worker 2<br/>Linux x86_64]
        W3[hg-worker 3<br/>Raspberry Pi]
        W4[hg-worker N<br/>...]
    end

    CLI -->|"1. Submit Task<br/>gRPC"| COORD
    COORD --> SCHED
    COORD --> REG
    COORD --> DASH
    COORD --> METRICS

    SCHED -->|"2. Dispatch"| W1
    SCHED -->|"2. Dispatch"| W2
    SCHED -->|"2. Dispatch"| W3

    W1 -->|"3. Result"| COORD
    W2 -->|"3. Result"| COORD
    W3 -->|"3. Result"| COORD

    COORD -->|"4. Return"| CLI
    CLI -.->|"Cache"| LC

    W1 & W2 & W3 -->|Heartbeat| REG
```

## Compilation Flow

```mermaid
sequenceDiagram
    participant C as hgbuild
    participant LC as Local Cache
    participant CO as Coordinator
    participant S as Scheduler
    participant W as Worker
    participant RC as Remote Cache

    C->>C: 1. Parse compiler args
    C->>LC: 2. Check cache (source hash)

    alt Cache Hit
        LC-->>C: Return cached object
        C->>CO: Report cache hit (async)
    else Cache Miss
        C->>CO: 3. SubmitTask(source, args)
        CO->>S: 4. Schedule task
        S->>S: Select best worker (P2C)
        S->>W: 5. Forward task
        W->>W: 6. Compile with gcc/clang
        W-->>CO: 7. Return object file
        CO-->>C: 8. Return result
        C->>LC: 9. Store in cache
        C->>C: 10. Write output file
    end
```

## Worker Selection (P2C Algorithm)

```mermaid
flowchart TD
    subgraph Selection["Power of Two Choices"]
        START[New Task] --> FILTER[Filter by<br/>Capabilities]
        FILTER --> RANDOM[Pick 2 Random<br/>Candidates]
        RANDOM --> SCORE1[Score Worker 1]
        RANDOM --> SCORE2[Score Worker 2]

        SCORE1 --> COMPARE{Compare<br/>Scores}
        SCORE2 --> COMPARE

        COMPARE -->|Best Score| SELECT[Select Winner]
        SELECT --> DISPATCH[Dispatch Task]
    end

    subgraph Scoring["Score Factors"]
        CPU[CPU Cores]
        MEM[Available Memory]
        LOAD[Current Load]
        RTT[Network Latency]
        ARCH[Architecture Match]
    end

    SCORE1 -.-> Scoring
    SCORE2 -.-> Scoring

    style Selection fill:#e1f5fe
    style Scoring fill:#fff3e0
```

## Fault Tolerance

```mermaid
stateDiagram-v2
    [*] --> Closed: Initial State

    Closed --> Open: Failures >= Threshold
    Closed --> Closed: Success

    Open --> HalfOpen: Timeout Expires
    Open --> Open: Requests Fail Fast

    HalfOpen --> Closed: Test Request Succeeds
    HalfOpen --> Open: Test Request Fails

    note right of Closed
        Normal operation
        All requests pass through
    end note

    note right of Open
        Circuit tripped
        Requests fail immediately
        Worker marked unhealthy
    end note

    note right of HalfOpen
        Testing recovery
        Single request allowed
    end note
```

## Cache Architecture

```mermaid
flowchart LR
    subgraph Client["Client Side"]
        SRC[Source File] --> HASH[xxHash64]
        HASH --> KEY[Cache Key]

        KEY --> CHECK{Cache<br/>Lookup}
        CHECK -->|Hit| OBJ1[Cached Object]
        CHECK -->|Miss| REMOTE[Remote Compile]
        REMOTE --> STORE[Store Result]
        STORE --> OBJ2[New Object]
    end

    subgraph CacheStore["Cache Storage"]
        direction TB
        DISK[(~/.hybridgrid/cache)]
        META[Metadata Index]
        LRU[LRU Eviction]

        DISK --- META
        META --- LRU
    end

    KEY -.-> CacheStore
    OBJ1 -.-> CacheStore
    STORE -.-> CacheStore

    style Client fill:#e8f5e9
    style CacheStore fill:#fce4ec
```

## mDNS Auto-Discovery

```mermaid
sequenceDiagram
    participant CO as Coordinator
    participant mDNS as mDNS Network
    participant W as Worker
    participant C as Client

    CO->>mDNS: 1. Announce _hybridgrid-coord._tcp
    W->>mDNS: 2. Browse for coordinators
    mDNS-->>W: 3. Found coordinator at 192.168.1.100:9000
    W->>CO: 4. Connect & Register

    Note over W,CO: Worker now connected

    C->>mDNS: 5. Browse for coordinators
    mDNS-->>C: 6. Found coordinator at 192.168.1.100:9000
    C->>CO: 7. Submit compilation task

    Note over C,CO: Zero-config networking
```

## Cross-Compilation Flow

```mermaid
flowchart TB
    subgraph Client["Client (macOS ARM64)"]
        REQ[Build Request<br/>Target: Linux x86_64]
    end

    subgraph Coordinator
        MATCH{Match<br/>Capabilities}
    end

    subgraph Workers
        W1[macOS ARM64<br/>Native]
        W2[Linux x86_64<br/>Native ✓]
        W3[Raspberry Pi<br/>ARM64]
        DOCKER[Docker Worker<br/>dockcross]
    end

    REQ --> MATCH
    MATCH -->|Native Match| W2
    MATCH -->|No Native| DOCKER

    W2 -->|Compile| RESULT1[Object File<br/>ELF x86_64]
    DOCKER -->|Cross-compile| RESULT2[Object File<br/>ELF x86_64]

    style W2 fill:#c8e6c9
    style DOCKER fill:#fff9c4
```

## Request Processing Pipeline

```mermaid
flowchart LR
    subgraph Ingress
        GRPC[gRPC Server]
        AUTH[Auth Check]
        VALID[Validation]
    end

    subgraph Processing
        QUEUE[Task Queue]
        SCHED[Scheduler]
        DISPATCH[Dispatcher]
    end

    subgraph Execution
        WORKER[Worker Pool]
        COMPILE[Compilation]
        RESULT[Result Handler]
    end

    subgraph Egress
        CACHE[Cache Store]
        METRICS[Metrics Update]
        RESPONSE[Response]
    end

    GRPC --> AUTH --> VALID --> QUEUE
    QUEUE --> SCHED --> DISPATCH
    DISPATCH --> WORKER --> COMPILE --> RESULT
    RESULT --> CACHE --> METRICS --> RESPONSE

    style Ingress fill:#e3f2fd
    style Processing fill:#f3e5f5
    style Execution fill:#e8f5e9
    style Egress fill:#fff3e0
```

## Dashboard Architecture

```mermaid
flowchart TB
    subgraph Frontend["Web Dashboard"]
        UI[HTML/JS UI]
        WS[WebSocket Client]
        CHARTS[Real-time Charts]
    end

    subgraph Backend["Coordinator"]
        HTTP[HTTP Server :8080]
        API[REST API]
        WSS[WebSocket Server]
        DATA[Data Aggregator]
    end

    subgraph Sources["Data Sources"]
        REG[Worker Registry]
        TASK[Task Manager]
        CACHE[Cache Stats]
        PROM[Prometheus Metrics]
    end

    UI --> HTTP
    WS <-->|Real-time| WSS
    CHARTS --> WS

    HTTP --> API --> DATA
    WSS --> DATA
    DATA --> REG
    DATA --> TASK
    DATA --> CACHE
    DATA --> PROM

    style Frontend fill:#e1f5fe
    style Backend fill:#f3e5f5
    style Sources fill:#e8f5e9
```

## Component Interactions

```mermaid
graph LR
    subgraph CLI["hgbuild"]
        PARSE[Arg Parser]
        PP[Preprocessor]
        CLIENT[gRPC Client]
    end

    subgraph Coord["hg-coord"]
        SERVER[gRPC Server]
        SCHED[Scheduler]
        REG[Registry]
        DASH[Dashboard]
    end

    subgraph Worker["hg-worker"]
        EXEC[Executor]
        COMP[Compiler]
        CB[Circuit Breaker]
    end

    PARSE --> PP --> CLIENT
    CLIENT <-->|gRPC| SERVER
    SERVER --> SCHED
    SCHED <--> REG
    SERVER --> DASH

    SCHED <-->|gRPC| EXEC
    EXEC --> COMP
    EXEC <--> CB

    style CLI fill:#bbdefb
    style Coord fill:#c8e6c9
    style Worker fill:#ffecb3
```

## Metrics Collection

```mermaid
flowchart TB
    subgraph Components
        CO[Coordinator]
        W1[Worker 1]
        W2[Worker 2]
        CLI[hgbuild]
    end

    subgraph Metrics["Prometheus Metrics"]
        TASKS[hybridgrid_tasks_total]
        DUR[hybridgrid_task_duration_seconds]
        CACHE[hybridgrid_cache_hits/misses]
        WORKERS[hybridgrid_workers_total]
        CIRCUIT[hybridgrid_circuit_state]
    end

    subgraph Monitoring
        PROM[Prometheus]
        GRAF[Grafana]
        ALERT[Alertmanager]
    end

    CO --> |/metrics| PROM
    W1 --> |/metrics| PROM
    W2 --> |/metrics| PROM

    CO --> TASKS & DUR & CACHE & WORKERS & CIRCUIT
    W1 & W2 --> DUR

    PROM --> GRAF
    PROM --> ALERT

    style Metrics fill:#fff3e0
    style Monitoring fill:#e8eaf6
```

## Network Topology Options

```mermaid
graph TB
    subgraph LAN["LAN Setup (mDNS)"]
        L_CLI[hgbuild] -->|mDNS| L_COORD[Coordinator]
        L_W1[Worker 1] -->|mDNS| L_COORD
        L_W2[Worker 2] -->|mDNS| L_COORD
    end

    subgraph WAN["WAN Setup (Explicit)"]
        W_CLI[hgbuild] -->|"IP:Port"| W_COORD[Coordinator<br/>Public IP]
        W_W1[Worker 1<br/>VPN] --> W_COORD
        W_W2[Worker 2<br/>Tunnel] --> W_COORD
    end

    subgraph Docker["Docker Compose"]
        D_CLI[hgbuild] -->|"docker network"| D_COORD[hg-coord]
        D_W1[hg-worker] --> D_COORD
        D_W2[hg-worker] --> D_COORD
    end

    style LAN fill:#e8f5e9
    style WAN fill:#fff3e0
    style Docker fill:#e3f2fd
```

## Rendering Diagrams

To render these Mermaid diagrams:

1. **GitHub**: Diagrams render automatically in markdown files
2. **VS Code**: Install "Mermaid Preview" extension
3. **CLI**: Use mermaid-cli
   ```bash
   npx @mermaid-js/mermaid-cli -i architecture-diagrams.md -o docs/images/
   ```
4. **Web**: Paste into [Mermaid Live Editor](https://mermaid.live)
