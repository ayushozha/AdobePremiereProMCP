package health

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestCheckerInitialState(t *testing.T) {
	c := NewChecker(zap.NewNop())

	// All three default services should start as unhealthy (unknown until
	// first check).
	for _, name := range []string{"media-engine", "intelligence", "premiere-bridge"} {
		sh := c.GetStatus(name)
		if sh == nil {
			t.Fatalf("expected ServiceHealth entry for %q", name)
		}
		if sh.Status != StatusUnhealthy {
			t.Fatalf("%s: initial status = %v, want StatusUnhealthy", name, sh.Status)
		}
	}
}

func TestCheckerIsReady(t *testing.T) {
	c := NewChecker(zap.NewNop())

	// Register a probe that succeeds for premiere-bridge.
	c.RegisterProbe("premiere-bridge", func(_ context.Context) error {
		return nil
	})

	// Before any check, premiere-bridge is unhealthy -- not ready.
	if c.IsReady() {
		t.Fatal("should not be ready before first check")
	}

	// Run the check.
	c.Check(context.Background(), "premiere-bridge")

	if !c.IsReady() {
		t.Fatal("should be ready after successful premiere-bridge check")
	}
}

func TestCheckerNotReady(t *testing.T) {
	c := NewChecker(zap.NewNop())

	// Register a probe that always fails.
	c.RegisterProbe("premiere-bridge", func(_ context.Context) error {
		return errors.New("unreachable")
	})

	// Run enough checks to exceed the unhealthy threshold (3).
	for i := 0; i < unhealthyThreshold; i++ {
		c.Check(context.Background(), "premiere-bridge")
	}

	if c.IsReady() {
		t.Fatal("should not be ready when premiere-bridge is unhealthy")
	}

	sh := c.GetStatus("premiere-bridge")
	if sh == nil {
		t.Fatal("expected ServiceHealth for premiere-bridge")
	}
	if sh.Status != StatusUnhealthy {
		t.Fatalf("premiere-bridge status: got %v, want StatusUnhealthy", sh.Status)
	}
}

func TestCheckerDegradedAfterSingleFailure(t *testing.T) {
	c := NewChecker(zap.NewNop())

	c.RegisterProbe("premiere-bridge", func(_ context.Context) error {
		return errors.New("blip")
	})

	// One failure should be degraded, not unhealthy (threshold is 3).
	c.Check(context.Background(), "premiere-bridge")

	sh := c.GetStatus("premiere-bridge")
	if sh == nil {
		t.Fatal("expected ServiceHealth for premiere-bridge")
	}
	if sh.Status != StatusDegraded {
		t.Fatalf("status after 1 failure: got %v, want StatusDegraded", sh.Status)
	}
}

func TestCheckerGetAllStatuses(t *testing.T) {
	c := NewChecker(zap.NewNop())

	all := c.GetAllStatuses()
	if len(all) != 3 {
		t.Fatalf("expected 3 services, got %d", len(all))
	}

	for _, name := range []string{"media-engine", "intelligence", "premiere-bridge"} {
		if _, ok := all[name]; !ok {
			t.Fatalf("missing service %q in GetAllStatuses", name)
		}
	}
}

func TestCheckerGetStatusUnknown(t *testing.T) {
	c := NewChecker(zap.NewNop())

	sh := c.GetStatus("nonexistent-service")
	if sh != nil {
		t.Fatalf("expected nil for unknown service, got %+v", sh)
	}
}

func TestCheckerRecovery(t *testing.T) {
	c := NewChecker(zap.NewNop())

	callCount := 0
	c.RegisterProbe("premiere-bridge", func(_ context.Context) error {
		callCount++
		if callCount <= unhealthyThreshold {
			return errors.New("down")
		}
		return nil // recovered
	})

	// Drive it to unhealthy.
	for i := 0; i < unhealthyThreshold; i++ {
		c.Check(context.Background(), "premiere-bridge")
	}
	if c.IsReady() {
		t.Fatal("should not be ready after threshold failures")
	}

	// Now the probe succeeds -- should recover to healthy.
	c.Check(context.Background(), "premiere-bridge")

	if !c.IsReady() {
		t.Fatal("should be ready after recovery")
	}

	sh := c.GetStatus("premiere-bridge")
	if sh.Consecutive != 0 {
		t.Fatalf("consecutive failures after recovery: got %d, want 0", sh.Consecutive)
	}
}

func TestCheckerRegisterProbeCreatesEntry(t *testing.T) {
	c := NewChecker(zap.NewNop())

	c.RegisterProbe("custom-service", func(_ context.Context) error {
		return nil
	})

	sh := c.GetStatus("custom-service")
	if sh == nil {
		t.Fatal("RegisterProbe should create a ServiceHealth entry for unknown service names")
	}
	if sh.Status != StatusUnhealthy {
		t.Fatalf("initial status for new service: got %v, want StatusUnhealthy", sh.Status)
	}

	// After checking, should be healthy.
	c.Check(context.Background(), "custom-service")

	sh = c.GetStatus("custom-service")
	if sh.Status != StatusHealthy {
		t.Fatalf("status after successful check: got %v, want StatusHealthy", sh.Status)
	}
}

func TestStatusString(t *testing.T) {
	cases := []struct {
		status Status
		want   string
	}{
		{StatusHealthy, "healthy"},
		{StatusDegraded, "degraded"},
		{StatusUnhealthy, "unhealthy"},
		{Status(99), "unknown"},
	}
	for _, tc := range cases {
		got := tc.status.String()
		if got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}
