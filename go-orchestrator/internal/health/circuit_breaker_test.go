package health

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker("test-closed", 3, 100*time.Millisecond, zap.NewNop())

	callCount := 0
	err := cb.Execute(context.Background(), func(_ context.Context) error {
		callCount++
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("call count: got %d, want 1", callCount)
	}
	if cb.GetState() != StateClosed {
		t.Fatalf("state: got %v, want Closed", cb.GetState())
	}
}

func TestCircuitBreakerOpens(t *testing.T) {
	maxFailures := 3
	cb := NewCircuitBreaker("test-opens", maxFailures, 100*time.Millisecond, zap.NewNop())

	testErr := errors.New("service down")

	// Cause maxFailures consecutive failures to trip the breaker.
	for i := 0; i < maxFailures; i++ {
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return testErr
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("state after %d failures: got %v, want Open", maxFailures, cb.GetState())
	}
}

func TestCircuitBreakerRejects(t *testing.T) {
	maxFailures := 2
	cb := NewCircuitBreaker("test-rejects", maxFailures, 5*time.Second, zap.NewNop())

	// Open the breaker.
	for i := 0; i < maxFailures; i++ {
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return errors.New("fail")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("expected Open state, got %v", cb.GetState())
	}

	// Next call should be rejected immediately without running fn.
	called := false
	err := cb.Execute(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})

	if called {
		t.Fatal("fn should not have been called while circuit is open")
	}
	if err == nil {
		t.Fatal("expected ErrCircuitOpen error")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got: %v", err)
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	maxFailures := 2
	resetTimeout := 50 * time.Millisecond
	cb := NewCircuitBreaker("test-half-open", maxFailures, resetTimeout, zap.NewNop())

	// Open the breaker.
	for i := 0; i < maxFailures; i++ {
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return errors.New("fail")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("expected Open state, got %v", cb.GetState())
	}

	// Wait for the reset timeout to elapse.
	time.Sleep(resetTimeout + 10*time.Millisecond)

	// The next call should be allowed through (half-open probe).
	probed := false
	err := cb.Execute(context.Background(), func(_ context.Context) error {
		probed = true
		return nil // success resets to Closed
	})

	if err != nil {
		t.Fatalf("unexpected error on half-open probe: %v", err)
	}
	if !probed {
		t.Fatal("expected probe function to be called in half-open state")
	}
	if cb.GetState() != StateClosed {
		t.Fatalf("expected Closed state after successful probe, got %v", cb.GetState())
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	maxFailures := 2
	resetTimeout := 50 * time.Millisecond
	cb := NewCircuitBreaker("test-half-open-fail", maxFailures, resetTimeout, zap.NewNop())

	// Open the breaker.
	for i := 0; i < maxFailures; i++ {
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return errors.New("fail")
		})
	}

	// Wait for the reset timeout.
	time.Sleep(resetTimeout + 10*time.Millisecond)

	// Probe fails -- should re-open.
	_ = cb.Execute(context.Background(), func(_ context.Context) error {
		return errors.New("still broken")
	})

	if cb.GetState() != StateOpen {
		t.Fatalf("expected Open state after failed probe, got %v", cb.GetState())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	maxFailures := 2
	cb := NewCircuitBreaker("test-reset", maxFailures, 5*time.Second, zap.NewNop())

	// Open the breaker.
	for i := 0; i < maxFailures; i++ {
		_ = cb.Execute(context.Background(), func(_ context.Context) error {
			return errors.New("fail")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("expected Open state, got %v", cb.GetState())
	}

	// Manual reset.
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Fatalf("expected Closed state after Reset, got %v", cb.GetState())
	}

	// Calls should flow through again.
	called := false
	err := cb.Execute(context.Background(), func(_ context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if !called {
		t.Fatal("fn should have been called after reset")
	}
}

func TestCircuitBreakerDefaultValues(t *testing.T) {
	// Pass 0 for both maxFailures and resetTimeout -- should use defaults.
	cb := NewCircuitBreaker("defaults", 0, 0, zap.NewNop())
	if cb.maxFailures != defaultMaxFailures {
		t.Fatalf("maxFailures: got %d, want %d", cb.maxFailures, defaultMaxFailures)
	}
	if cb.resetTimeout != defaultResetTimeout {
		t.Fatalf("resetTimeout: got %v, want %v", cb.resetTimeout, defaultResetTimeout)
	}
}

func TestStateString(t *testing.T) {
	cases := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}
	for _, tc := range cases {
		got := tc.state.String()
		if got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}
