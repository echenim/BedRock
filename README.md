# square

```plaintext

storagerouter/
├── cmd/
│   ├── router/                # Main entry point for the routing service
│   │   ├── main.go
│   ├── retrieval/             # Main entry point for the retrieval service
│   │   ├── main.go
│   ├── metadata/              # Entry point for metadata management service
│   │   ├── main.go
├── internal/
│   ├── config/                # Configuration loading and management
│   │   ├── config.go
│   ├── router/                # Routing service logic
│   │   ├── router.go
│   │   ├── rules.go           # Routing rules engine
│   ├── retrieval/             # File retrieval logic
│   │   ├── retrieval.go
│   │   ├── cache.go           # Caching layer for faster retrieval
│   ├── metadata/              # Metadata management logic
│   │   ├── metadata.go
│   │   ├── store.go           # Interface for metadata storage backends
│   ├── storage/               # Storage backends abstraction
│   │   ├── s3.go              # Amazon S3 integration
│   │   ├── hdfs.go            # HDFS integration
│   │   ├── local.go           # Local storage integration
│   │   ├── storage.go         # Unified storage interface
│   ├── logger/                # Logging utilities
│   │   ├── logger.go
│   ├── utils/                 # Utility functions
│   │   ├── mime.go            # MIME type detection
│   │   ├── checksum.go        # Data integrity utilities
├── pkg/                       # Shared libraries for external use
│   ├── api/                   # API handlers and routing
│   │   ├── http.go            # HTTP API handlers
│   │   ├── grpc.go            # gRPC API handlers
├── test/                      # Integration and unit tests
│   ├── router_test.go
│   ├── retrieval_test.go
├── configs/                   # Configuration files
│   ├── app.yaml
│   ├── storage.yaml
├── docs/                      # Project documentation
├── go.mod                     # Go modules file
├── go.sum                     # Go modules dependencies


```



