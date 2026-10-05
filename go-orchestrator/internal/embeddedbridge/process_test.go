package embeddedbridge

import (
	"context"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Removing early-exit observation would make a dead child look like a timeout.
func TestChildExitBeforeReadiness(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestBridgeHelper")
	cmd.Env = append(os.Environ(), "BRIDGE_HELPER=exit")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := supervise(ctx, cmd, freeAddress(t))
	if err == nil {
		t.Fatal("expected child exit error")
	}
	if !strings.Contains(err.Error(), "exited before readiness") {
		t.Fatal(err)
	}
	if ctx.Err() != nil {
		t.Fatal("dead child waited until startup timeout")
	}
}

// Removing cleanup/reaping would leave the listener alive after stop.
func TestReadyChildStopsAndReaps(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	cmd := exec.Command(os.Args[0], "-test.run=TestBridgeHelper")
	cmd.Env = append(os.Environ(), "BRIDGE_HELPER=listen", "BRIDGE_HELPER_ADDR="+addr)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stop, err := supervise(ctx, cmd, addr)
	if err != nil {
		t.Fatal(err)
	}
	stop()
	stop()
	if cmd.ProcessState == nil {
		t.Fatal("child not reaped")
	}
	c, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
	if err == nil {
		c.Close()
		t.Fatal("child still listening")
	}
}

func TestOccupiedAddressDoesNotStartChild(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	cmd := exec.Command(os.Args[0], "-test.run=TestBridgeHelper")
	if _, err := supervise(context.Background(), cmd, listener.Addr().String()); err == nil {
		t.Fatal("accepted occupied port")
	}
	if cmd.Process != nil {
		t.Fatal("started duplicate bridge")
	}
}

func TestStartupTimeoutReapsChild(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestBridgeHelper")
	cmd.Env = append(os.Environ(), "BRIDGE_HELPER=idle")
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if _, err := supervise(ctx, cmd, freeAddress(t)); err == nil {
		t.Fatal("accepted unready bridge")
	}
	if cmd.ProcessState == nil {
		t.Fatal("timed out child not reaped")
	}
}

func TestBridgeHelper(t *testing.T) {
	switch os.Getenv("BRIDGE_HELPER") {
	case "exit":
		os.Exit(7)
	case "listen":
		listener, err := net.Listen("tcp", os.Getenv("BRIDGE_HELPER_ADDR"))
		if err != nil {
			os.Exit(8)
		}
		defer listener.Close()
		for {
			c, err := listener.Accept()
			if err != nil {
				os.Exit(9)
			}
			c.Close()
		}
	case "idle":
		time.Sleep(time.Minute)
		os.Exit(0)
	}
}

func freeAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	return addr
}

// Loading the entry before containment would let descendants escape the job.
func TestNodeEntryWaitsForGate(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node unavailable")
	}
	directory := t.TempDir()
	marker := directory + "/started"
	entry := directory + "/bridge.mjs"
	if err := os.WriteFile(entry, []byte("import fs from 'node:fs'; fs.writeFileSync("+strconv.Quote(marker)+", 'started');"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd, gate, err := bridgeCommand(node, entry)
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("bridge started before gate release")
	}
	if _, err := gate.Write([]byte("start\n")); err != nil {
		t.Fatal(err)
	}
	gate.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("ES module bridge did not start:", err)
	}
}
