//go:build windows

package embeddedbridge

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// This is compiled on other platforms but must execute on Windows to prove
// forced owner termination closes the job and kills both Node generations.
func TestJobKillsDescendantsWhenOwnerIsTerminated(t *testing.T) {
	node, err := exec.LookPath("node.exe")
	if err != nil {
		t.Skip("node.exe unavailable")
	}
	dir := t.TempDir()
	entry := filepath.Join(dir, "bridge.cjs")
	marker := filepath.Join(dir, "pids.json")
	script := `const {spawn}=require('node:child_process');const fs=require('node:fs');const child=spawn(process.execPath,['-e','setInterval(()=>{},1000)'],{stdio:'ignore'});fs.writeFileSync(process.env.JOB_MARKER,JSON.stringify([process.pid,child.pid]));setInterval(()=>{},1000);`
	if err := os.WriteFile(entry, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	owner := exec.Command(os.Args[0], "-test.run=TestJobOwnerHelper")
	owner.Env = append(os.Environ(), "JOB_OWNER=1", "JOB_NODE="+node, "JOB_ENTRY="+entry, "JOB_MARKER="+marker)
	owner.Stderr = os.Stderr
	if err := owner.Start(); err != nil {
		t.Fatal(err)
	}
	defer owner.Process.Kill()
	var pids []uint32
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(marker)
		if err == nil && json.Unmarshal(data, &pids) == nil && len(pids) == 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if len(pids) != 2 {
		t.Fatal("contained Node descendants did not start")
	}
	var handles []windows.Handle
	for _, pid := range pids {
		handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, pid)
		if err != nil {
			t.Fatal(err)
		}
		defer windows.CloseHandle(handle)
		handles = append(handles, handle)
	}
	if err := owner.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	owner.Wait()
	for index, handle := range handles {
		status, err := windows.WaitForSingleObject(handle, 5000)
		if err != nil || status != windows.WAIT_OBJECT_0 {
			t.Fatalf("Node PID %d survived owner death: status=%d err=%v", pids[index], status, err)
		}
	}
}

func TestJobOwnerHelper(t *testing.T) {
	if os.Getenv("JOB_OWNER") != "1" {
		return
	}
	cmd, gate, err := bridgeCommand(os.Getenv("JOB_NODE"), os.Getenv("JOB_ENTRY"))
	if err != nil {
		os.Exit(2)
	}
	cmd.Stderr = os.Stderr
	closeJob, err := startContained(cmd)
	if err != nil {
		os.Exit(3)
	}
	defer closeJob()
	if _, err := gate.Write([]byte("start\n")); err != nil {
		os.Exit(4)
	}
	gate.Close()
	time.Sleep(time.Minute)
}

// cmd.Wait waits for os/exec's stdout copy pipe. A descendant inheriting that
// pipe keeps it open after the root dies, so the whole job must close first.
func TestStopClosesJobBeforeWaitingForInheritedOutput(t *testing.T) {
	node, err := exec.LookPath("node.exe")
	if err != nil {
		t.Skip("node.exe unavailable")
	}
	entry := filepath.Join(t.TempDir(), "bridge.cjs")
	addr := freeAddress(t)
	host, port, _ := net.SplitHostPort(addr)
	script := `const {spawn}=require('node:child_process');spawn(process.execPath,['-e','setInterval(()=>{},1000)'],{stdio:'inherit'});require('node:net').createServer(c=>c.end()).listen(` + port + `,` + strconv.Quote(host) + `);`
	if err := os.WriteFile(entry, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	cmd, gate, err := bridgeCommand(node, entry)
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()
	cmd.Stdout = io.Discard
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stop, err := supervise(ctx, cmd, addr, func() error { _, err := gate.Write([]byte("start\n")); return err })
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stop blocked waiting for inherited stdout")
	}
}
