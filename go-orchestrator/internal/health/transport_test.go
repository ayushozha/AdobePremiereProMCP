package health

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestGRPCTransportProbeRejectsDisconnectedDependency(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	probe := GRPCTransportProbe(listener.Addr().String())
	if err := probe(context.Background()); err != nil {
		t.Fatalf("running gRPC server: %v", err)
	}
	server.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := probe(ctx); err == nil {
		t.Fatal("disconnected gRPC dependency reported healthy")
	}
}

func TestPremiereReadinessRequiresRunningApplicationAndSanitizesErrors(t *testing.T) {
	for _, tc := range []struct {
		running bool
		err     error
		want    string
	}{
		{true, nil, ""},
		{false, nil, "Premiere is not connected"},
		{false, errors.New("secret-token"), "bridge ping failed"},
	} {
		probe := PremiereReadinessProbe(func(ctx context.Context) (bool, error) {
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("probe has no deadline")
			}
			return tc.running, tc.err
		})
		err := probe(context.Background())
		if tc.want == "" && err != nil {
			t.Fatal(err)
		}
		if tc.want != "" && (err == nil || err.Error() != tc.want) {
			t.Fatalf("got %v, want %q", err, tc.want)
		}
	}
}

func TestReadinessRejectsStaleSuccessfulProbe(t *testing.T) {
	checker := NewChecker(nil)
	for _, health := range checker.services {
		health.Status = StatusHealthy
		health.LastCheck = time.Now()
	}
	if !checker.IsReady() {
		t.Fatal("recent successful probes should be ready")
	}
	checker.services["media-engine"].LastCheck = time.Now().Add(-2 * time.Minute)
	if checker.IsReady() {
		t.Fatal("stale success was reported ready")
	}
}
