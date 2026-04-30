//go:build windows

package embeddedbridge

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"go.uber.org/zap"
	"golang.org/x/sys/windows"
)

// Start spawns the TypeScript bridge (node ts-bridge/dist/index.js) in a Win32 Job Object
// with JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE. When server.exe terminates for any reason, the OS
// closes handles for the dead process → last job handle closes → node is forcibly terminated.
func Start(logger *zap.Logger) (stop func(), err error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("embeddedbridge: Executable: %w", err)
	}
	binDir := filepath.Dir(exePath)
	goOrchRoot := filepath.Clean(filepath.Join(binDir, ".."))
	repoRoot := filepath.Clean(filepath.Join(goOrchRoot, ".."))
	bridgeWD := filepath.Join(repoRoot, "ts-bridge")
	bridgeJS := filepath.Join(bridgeWD, "dist", "index.js")

	if _, err := os.Stat(bridgeJS); err != nil {
		return nil, fmt.Errorf("embeddedbridge: missing bridge script %s: %w", bridgeJS, err)
	}
	nodeExe, err := exec.LookPath("node.exe")
	if err != nil {
		nodeExe, err = exec.LookPath("node")
	}
	if err != nil {
		return nil, fmt.Errorf("embeddedbridge: node not on PATH: %w", err)
	}

	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("embeddedbridge: CreateJobObject: %w", err)
	}

	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		_ = windows.CloseHandle(job)
		return nil, fmt.Errorf("embeddedbridge: SetInformationJobObject(KILL_ON_JOB_CLOSE): %w", err)
	}

	cmd := exec.Command(nodeExe, bridgeJS)
	cmd.Dir = bridgeWD
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_BREAKAWAY_FROM_JOB,
	}

	if err := cmd.Start(); err != nil {
		_ = windows.CloseHandle(job)
		return nil, fmt.Errorf("embeddedbridge: starting node: %w", err)
	}

	procH, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = cmd.Process.Kill()
		_ = windows.CloseHandle(job)
		return nil, fmt.Errorf("embeddedbridge: OpenProcess(pid=%d): %w", cmd.Process.Pid, err)
	}
	if err := windows.AssignProcessToJobObject(job, procH); err != nil {
		_ = windows.CloseHandle(procH)
		_ = cmd.Process.Kill()
		_ = windows.CloseHandle(job)
		return nil, fmt.Errorf("embeddedbridge: AssignProcessToJobObject: %w", err)
	}
	_ = windows.CloseHandle(procH)

	logger.Info("embedded ts-bridge started under kill-on-close job",
		zap.Int("pid", cmd.Process.Pid),
		zap.String("script", bridgeJS),
	)

	return func() {
		if err := windows.CloseHandle(job); err != nil {
			logger.Warn("embeddedbridge: CloseHandle(job)", zap.Error(err))
		}
	}, nil
}
