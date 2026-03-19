package health

import (
	"context"
	"time"
)

// probeTimeout is the maximum duration a single health probe is allowed to
// take before it is cancelled.
const probeTimeout = 5 * time.Second

// Probe is a function that checks a backend service's health. It returns nil
// if the service is reachable and operating normally, or an error describing
// the failure.
type Probe func(ctx context.Context) error

// MediaClient is the minimal interface needed to probe the Rust media engine.
type MediaClient interface {
	// Ping sends a lightweight health-check RPC to the media engine and
	// returns an error if the service is unreachable.
	Ping(ctx context.Context) error
}

// IntelClient is the minimal interface needed to probe the Python intelligence
// service.
type IntelClient interface {
	// Ping sends a lightweight health-check RPC to the intelligence service
	// and returns an error if the service is unreachable.
	Ping(ctx context.Context) error
}

// PremiereClient is the minimal interface needed to probe the TypeScript
// Premiere bridge.
type PremiereClient interface {
	// Ping sends a lightweight health-check RPC to the Premiere bridge and
	// returns an error if the service is unreachable.
	Ping(ctx context.Context) error
}

// MediaProbe returns a Probe that checks the Rust media engine by issuing a
// lightweight Ping RPC. The probe enforces a 5-second timeout.
func MediaProbe(client MediaClient) Probe {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		return client.Ping(ctx)
	}
}

// IntelligenceProbe returns a Probe that checks the Python intelligence
// service by issuing a lightweight Ping RPC. The probe enforces a 5-second
// timeout.
func IntelligenceProbe(client IntelClient) Probe {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		return client.Ping(ctx)
	}
}

// PremiereProbe returns a Probe that checks the TypeScript Premiere bridge by
// issuing a lightweight Ping RPC. The probe enforces a 5-second timeout.
func PremiereProbe(client PremiereClient) Probe {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		return client.Ping(ctx)
	}
}
