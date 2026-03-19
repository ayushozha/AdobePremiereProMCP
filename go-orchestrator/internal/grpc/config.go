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

	// DialTimeout is the maximum duration to wait when establishing a gRPC connection.
	DialTimeout time.Duration

	// CallTimeout is the default per-RPC deadline applied to individual calls.
	CallTimeout time.Duration
}

// DefaultConfig returns a ClientsConfig with sensible defaults for local development.
func DefaultConfig() *ClientsConfig {
	return &ClientsConfig{
		MediaAddr:    "localhost:50052",
		IntelAddr:    "localhost:50053",
		PremiereAddr: "localhost:50054",
		DialTimeout:  10 * time.Second,
		CallTimeout:  30 * time.Second,
	}
}
