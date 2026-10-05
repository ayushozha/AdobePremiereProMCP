//go:build !windows

package embeddedbridge

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
)

func Start(_ context.Context, _ string, _ *zap.Logger) (func(), error) {
	return nil, fmt.Errorf("--embed-ts-bridge is only supported on Windows")
}

func startContained(cmd *exec.Cmd) (func(), error) {
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return func() {}, nil
}
