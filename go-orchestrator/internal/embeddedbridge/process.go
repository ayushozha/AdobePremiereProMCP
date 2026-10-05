package embeddedbridge

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"
)

func available(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("bridge address %s unavailable: %w", addr, err)
	}
	return listener.Close()
}

// supervise owns only the process it starts. Readiness means a TCP listener,
// not that Premiere is connected. A failed start always kills and reaps it.
func supervise(ctx context.Context, cmd *exec.Cmd, addr string, release ...func() error) (func(), error) {
	if err := available(addr); err != nil {
		return nil, err
	}
	closeJob, err := startContained(cmd)
	if err != nil {
		return nil, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var once sync.Once
	stop := func() {
		once.Do(func() {
			// Kill the entire contained family before Wait: descendants may hold
			// os/exec's stdout copy pipe open even after the direct child exits.
			closeJob()
			_ = cmd.Process.Kill()
			<-done
		})
	}
	// Preserve Wait's result for readiness while leaving cleanup a reusable token.
	exited := func(err error) (func(), error) {
		done <- err
		stop()
		return nil, fmt.Errorf("bridge exited before readiness: %v", err)
	}
	for _, unlock := range release {
		if err := unlock(); err != nil {
			stop()
			return nil, fmt.Errorf("releasing bridge startup: %w", err)
		}
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			return exited(err)
		case <-ctx.Done():
			stop()
			return nil, fmt.Errorf("waiting for bridge listener: %w", ctx.Err())
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
			if err == nil {
				conn.Close()
				select {
				case err := <-done:
					return exited(err)
				default:
				}
				return stop, nil
			}
		}
	}
}

// Node starts no application code until its owner releases this private pipe.
func bridgeCommand(node, entry string) (*exec.Cmd, io.WriteCloser, error) {
	cmd := exec.Command(node, "-e", "process.stdin.once('data',()=>{process.stdin.pause();import(require('node:url').pathToFileURL(process.argv[1]).href).catch(e=>{console.error(e);process.exit(1);});});", entry)
	gate, err := cmd.StdinPipe()
	return cmd, gate, err
}
