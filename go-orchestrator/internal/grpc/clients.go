// clients.go is the top-level client manager that owns connections to all three
// PremierPro MCP backend services.
//
// PROTO STUB NOTE: Once proto stubs are generated into gen/go/, the individual
// client wrappers (media_client.go, intelligence_client.go, premiere_client.go)
// will import the generated stubs and convert between proto types and the
// Go-native types defined in types.go. No changes are needed in this file.
package grpc

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// Clients holds gRPC client wrappers for every backend service in the
// PremierPro MCP system.
type Clients struct {
	// Media is the Rust media engine client (scan, probe, thumbnails, waveform, scenes).
	Media *MediaEngineClient

	// Intel is the Python intelligence client (parse, EDL, match, pacing).
	Intel *IntelligenceClient

	// Premiere is the TypeScript Premiere bridge client (project, timeline, clips, export).
	Premiere *PremiereBridgeClient

	logger *zap.Logger
}

// NewClients creates gRPC connections to all three backend services and returns
// a Clients manager. If any connection fails, already-opened connections are
// closed before returning the error.
func NewClients(cfg *ClientsConfig, logger *zap.Logger) (*Clients, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	logger.Info("initialising gRPC clients",
		zap.String("media_addr", cfg.MediaAddr),
		zap.String("intel_addr", cfg.IntelAddr),
		zap.String("premiere_addr", cfg.PremiereAddr),
	)

	// Connect to each service sequentially so we get clear error attribution.
	media, err := newMediaEngineClient(cfg.MediaAddr, cfg.DialTimeout, cfg.CallTimeout, logger)
	if err != nil {
		return nil, fmt.Errorf("create media engine client: %w", err)
	}

	intel, err := newIntelligenceClient(cfg.IntelAddr, cfg.DialTimeout, cfg.CallTimeout, logger)
	if err != nil {
		_ = media.close()
		return nil, fmt.Errorf("create intelligence client: %w", err)
	}

	premiere, err := newPremiereBridgeClient(cfg.PremiereAddr, cfg.DialTimeout, cfg.CallTimeout, logger)
	if err != nil {
		_ = media.close()
		_ = intel.close()
		return nil, fmt.Errorf("create premiere bridge client: %w", err)
	}

	logger.Info("all gRPC clients connected")

	return &Clients{
		Media:    media,
		Intel:    intel,
		Premiere: premiere,
		logger:   logger,
	}, nil
}

// Close shuts down all gRPC connections. It attempts to close every connection
// even if one fails, and returns a combined error.
func (c *Clients) Close() error {
	c.logger.Info("closing all gRPC clients")

	var errs []error

	if c.Media != nil {
		if err := c.Media.close(); err != nil {
			errs = append(errs, fmt.Errorf("close media client: %w", err))
		}
	}
	if c.Intel != nil {
		if err := c.Intel.close(); err != nil {
			errs = append(errs, fmt.Errorf("close intelligence client: %w", err))
		}
	}
	if c.Premiere != nil {
		if err := c.Premiere.close(); err != nil {
			errs = append(errs, fmt.Errorf("close premiere client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("closing gRPC clients: %w", errors.Join(errs...))
	}

	c.logger.Info("all gRPC clients closed")
	return nil
}
