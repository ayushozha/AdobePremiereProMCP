//go:build !windows

package embeddedbridge

import (
	"fmt"
	"go.uber.org/zap"
)

// Start is not implemented on this platform (-embed-ts-bridge is Windows-only).
func Start(_ *zap.Logger) (func(), error) {
	return nil, fmt.Errorf("embeddedbridge: -embed-ts-bridge is only supported on Windows")
}
