---
name: goxf Framework Core Development
description: Guidelines, architecture rules, and functional module overviews for developing the goxf microservice framework.
---

# goxf Framework Development Guide

`goxf` is a high-performance, enterprise-grade Go microservice framework. It abstracts and coordinates core microservice components—such as configuration, logging, databases, and background jobs—within a unified lifecycle. 

When working on `goxf`, AI agents and developers must strictly adhere to the following architecture rules and module definitions.

## 核心开发哲学 (Core Design Principles)
1. **配置驱动 (Configuration-Driven)**: The framework heavily relies on centralized configuration (`conf` package). Components should initialize their states based on uniform yaml/toml configurations, supporting multi-environment deployments and daemon-based hot-reloading naturally.
2. **生命周期托管 (Lifecycle Controlled)**: Do not use isolated `init()` functions for critical connections. All external IO, background jobs, and HTTP/RPC servers must be registered to the global `Service` container. This guarantees proper start-up sequencing and safe, graceful shutdowns during system termination.
3. **监控前置与开箱即用 (Observability First)**: Third-party drivers (like internal Gin, Redis, or Cron) must be wrapped securely. The framework should automatically inject standard logging (`zap`), tracing (`jaeger`), and metrics (`prometheus`), shielding business developers from repetitive interceptor logic.
4. **强类型与防错 (Type Safety & Fallbacks)**: Leverage modern Go features (like Go 1.18+ Generics) to provide safe handles to developers. When component configuration is missing, the framework should fall back to sensible defaults with logging warnings, rather than halting execution unexpectedly.
---

## 核心功能模块全景 (Functional Modules)

### 1. 运行容器底层与 CLI 生命周期 (`application`, `goxf.go`, `flag`)
- **功能**: The entry point `Service` context and CLI subsystem. Manages initialization of configs, logger, traces, metrics, subcommands, and OS exit signals.
- **生命周期分层**:
  - `InitBase()`: Core essentials (flag parsing, config load via `-c`, `application.Runtime` setup, `log.Logger` init). No network listener, no banner, no ETCD registration.
  - `Bootstrap()`: Full microservice container setup (`InitBase()` + banner + Jaeger Tracer + ETCD Registry).
  - `Run(servers...)`: Serves protocols and blocks on OS signals.
- **Subcommand Interception**:
  - Use `flag.AddCommand(cmd)` for offline/tool subcommands (no config overhead).
  - Use `flag.AddCommandWithBase(cmd)` (or `flag.AddBaseCommand`) for subcommands requiring configs/database/logger (e.g. `migrate`, `seed`, `worker`).
  - When a subcommand matches, `goxf.NewService()` automatically intercepts execution, runs the subcommand, and exits cleanly (`os.Exit(0)`) without starting network servers.
- **原则**: Use `hooks.Register` (Stages: `BeforeLoadConfig`, `BeforeRun`, `BeforeStop`, `AfterStop`) to manage cross-module shutdown/startup sequences cleanly.

### 2. 配置中心 (`conf`)
- **多实例与 Functional Options (`LoadFromSource` / `NewConfFromSource`)**:
  - `conf.LoadFromSource(addr, opts...)`: Loads and deep-merges into the global singleton.
  - `conf.NewConfFromSource(addr, opts...)`: Creates an isolated, standalone `*conf.Conf` instance with separate delimiters, password, and state.
  - Functional Options: `conf.WithWatch`, `conf.WithUnmarshal`, `conf.WithKeyDelimiter`, `conf.WithPassword`, `conf.WithIgnoreGlobalPassword`.
- **多源合并 (`.local` Override)**: Supports unified parsing of YAML/JSON/TOML formats. It features an intelligent `.local` fallback mechanism: whenever a main config (e.g., `config.yaml`) is loaded, it automatically deep-merges any sibling `.local` file (e.g., `config.local.yaml`). This prevents git pollution from developer-specific environments.
- **热更新与加密安全 (`.enc`)**: 
  - Dynamic ciphertext mapping prevents multi-config overwriting (`config.yaml` $\rightarrow$ `system.enc`, `custom.yaml` $\rightarrow$ `custom.enc`).
  - Supports transparent decryption of any `*.enc` file.
  - Explicit path prompts during password validation (`Enter config password for [...]`).

### 3. 可观测性 (`log`, `metric`, `tracer`)
- **log**: A high-performance asynchronous logger built over Uber `zap` supporting rotation and context-aware injection.
- **metric**: Prometheus bindings providing instant dashboards. **All new modules must embed monitoring**. (e.g., `metric.ServerHandleHistogram`, `metric.ClientHandleCounter`).
- **tracer**: OpenTracing / Jaeger standardization for distributed request propagation.

### 4. 服务治理与通信 (`server`, `client`, `auth`)
- **server**: Standardizes protocols API boundaries (HTTP via `gin`, gRPC, WebSockets), incorporating default CORS and Logger interceptors.
- **client**: Built-in native wrappers for external nodes (`credis` for Redis, `cetcd` for ETCD, `carangodb` for ArangoDB).
- **auth**: JWT certificate and Token lifecycle validation for enterprise security borders.

### 5. 后台任务引擎 (`job/cron`, `job/queue`)
- **`job/cron`**: Cron engine (based on `robfig/cron`) for standard Unix-cron syntax scheduling with unified panic-recovery and metric recording.
- **`job/queue` (Async & Delayed Jobs)**: 
    - Exposes a strongly-typed API (`queue.NewWorker[T]`).
    - **Brokers**: Pluggable storage via the `Broker` interface. 
        - `MemoryBroker`: Optimized lock-free Min-Heap O(1) scheduling for single-node extreme performance.
        - `credis.QueueBroker`: High-availability distributed layout using Redis `ZSET` + `Lua Script Sweeper` for zero-loss delayed missions (e.g., *Order timeouts*).
    - **Usage**: Always call `worker.Enqueue(ctx, data)` or `worker.EnqueueAfter(ctx, data, delay)` directly from the Worker handle to trace metrics accurately.

### 6. 本地缓存与统一响应 (`cache/golanglru`, `api`)
- **`cache/golanglru`**: High-performance in-memory cache wrapper over `github.com/hashicorp/golang-lru/v2`.
    - Strongly-typed generic interface (`Cache[K comparable, V any]`) and convenient `golanglru.New(...)`.
    - Configurations: `name`, `size`, `expiration` (TTL, supports human-friendly duration strings like `"30m"`), `algorithm` (`lru` default or `2q`), `enableMetric` (Prometheus monitoring via `metric.ClientHandleCounter` / `metric.ClientHandleHistogram`).
    - Lifecycle managed via `hooks.Stage_AfterStop`.
- **`api`**: Standardized response wrapper (`api.BaseRes`) and semantic HTTP status code mapping (`CodeToHTTPStatus`) eliminating magic numbers, fully mapped against `ecode/common_ecode.go` with safe map isolation against concurrent mutations.

### 7. 标准脚手架支撑 (`i18n`, `ecode`, `util`)
- **i18n**: International language mapper.
- **ecode**: Global business error codes registry carrying underlying stack traces.
- **util**: Everyday formatters (`xfmt`), security helpers, and strings utilities.

---

## 开发测试规约
When adding or altering public signatures in `goxf`:
1. Always adapt the existing cases in the `examples/` directory to match the new behavior.
2. Manually test updates by running example programs: `go run examples/<module>/main.go`.

