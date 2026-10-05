//go:build windows

package embeddedbridge

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"go.uber.org/zap"
	"golang.org/x/sys/windows"
)

// Start embeds only the TypeScript bridge. Rust and Python remain separately managed.
func Start(ctx context.Context, addr string, logger *zap.Logger) (func(), error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("embedded bridge address: %w", err)
	}
	ip := net.ParseIP(host)
	if host == "localhost" {
		host = "127.0.0.1"
	} else if ip == nil || !ip.IsLoopback() {
		return nil, fmt.Errorf("embedded bridge requires a loopback address")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("embedded bridge requires a port from 1 to 65535")
	}
	addr = net.JoinHostPort(host, port)
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	bridgeDir := filepath.Clean(filepath.Join(filepath.Dir(executable), "..", "..", "ts-bridge"))
	entry := filepath.Join(bridgeDir, "dist", "index.js")
	if _, err := os.Stat(entry); err != nil {
		return nil, fmt.Errorf("build ts-bridge first (%s): %w", entry, err)
	}
	node, err := exec.LookPath("node.exe")
	if err != nil {
		return nil, fmt.Errorf("node.exe not on PATH: %w", err)
	}
	// Keep JavaScript execution behind a pipe until Job Object assignment succeeds.
	cmd, gate, err := bridgeCommand(node, entry)
	if err != nil {
		return nil, err
	}
	defer gate.Close()
	cmd.Dir, cmd.Stdout, cmd.Stderr = bridgeDir, io.Discard, os.Stderr
	for _, entry := range os.Environ() {
		key := strings.SplitN(entry, "=", 2)[0]
		if !strings.EqualFold(key, "BRIDGE_GRPC_HOST") && !strings.EqualFold(key, "BRIDGE_GRPC_PORT") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "BRIDGE_GRPC_HOST="+host, "BRIDGE_GRPC_PORT="+port)

	startupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	stop, err := supervise(startupCtx, cmd, addr, func() error { _, err := io.WriteString(gate, "start\n"); return err })
	if err != nil {
		return nil, err
	}
	logger.Info("embedded TypeScript bridge listener started", zap.Int("pid", cmd.Process.Pid), zap.String("addr", addr))
	return stop, nil
}

func startContained(cmd *exec.Cmd) (func(), error) {
	job, err := windows.CreateJobObject(nil, nil) // no inheritable security attributes
	if err != nil {
		return nil, fmt.Errorf("CreateJobObject: %w", err)
	}
	closeJob := func() { _ = windows.CloseHandle(job) }
	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		closeJob()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		closeJob()
		return nil, err
	}
	handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err == nil {
		err = windows.AssignProcessToJobObject(job, handle)
		windows.CloseHandle(handle)
	}
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		closeJob()
		return nil, fmt.Errorf("containing bridge in Windows Job Object: %w", err)
	}
	return closeJob, nil
}
