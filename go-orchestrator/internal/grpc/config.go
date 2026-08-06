package grpc

import "time"

// ClientsConfig holds connection parameters for the three backend gRPC services.
type ClientsConfig struct {
	// MediaAddr is the address of the Rust media engine service (e.g., "localhost:50052").
	MediaAddr string

	// IntelAddr is the address of the Python intelligence service (e.g., "localhost:50053").
	IntelAddr string

	// PremiereAddr is the address of the TypeScript Premiere bridge service (e.g., "localhost:50054").
	PremiereAddr string

	// Per-service RPC deadlines. grpc.NewClient is intentionally lazy, so
	// backend availability is established by bounded calls rather than startup.
	MediaCallTimeout    time.Duration
	IntelCallTimeout    time.Duration
	PremiereCallTimeout time.Duration
}

// DefaultConfig returns a ClientsConfig with sensible defaults for local development.
func DefaultConfig() *ClientsConfig {
	return &ClientsConfig{
		MediaAddr:           "localhost:50052",
		IntelAddr:           "localhost:50053",
		PremiereAddr:        "localhost:50054",
		MediaCallTimeout:    30 * time.Second,
		IntelCallTimeout:    60 * time.Second,
		PremiereCallTimeout: 30 * time.Second,
	}
}
