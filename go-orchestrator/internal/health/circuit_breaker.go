package health

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrCircuitOpen is returned by Execute when the circuit breaker is in the
// Open state and the reset timeout has not yet elapsed.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State represents the operational state of a CircuitBreaker.
type State int

const (
	// StateClosed is normal operation: requests flow through and failures are
	// counted.
	StateClosed State = iota
	// StateOpen means too many consecutive failures have occurred. All calls
	// are rejected immediately until the reset timeout expires.
	StateOpen
	// StateHalfOpen allows a single probe request through to test whether the
	// downstream service has recovered.
	StateHalfOpen
)

// String returns a human-readable label for the circuit breaker state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// defaultMaxFailures is the number of consecutive failures before the circuit
// opens.
const defaultMaxFailures = 5

// defaultResetTimeout is how long the circuit stays open before transitioning
// to half-open.
const defaultResetTimeout = 30 * time.Second

// CircuitBreaker implements a simple three-state circuit breaker around calls
// to an external service.
type CircuitBreaker struct {
	name         string
	maxFailures  int
	resetTimeout time.Duration
	state        State
	failures     int
	lastFailure  time.Time
	mu           sync.Mutex
	logger       *zap.Logger
}

// NewCircuitBreaker creates a CircuitBreaker with the given parameters. Pass
// zero for maxFailures or resetTimeout to use defaults (5 failures, 30 s).
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = defaultMaxFailures
	}
	if resetTimeout <= 0 {
		resetTimeout = defaultResetTimeout
	}
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
		logger:       logger,
	}
}

// Execute runs fn within the circuit breaker. The behaviour depends on the
// current state:
//
//   - Closed: fn is called normally. On error the failure count increments; if
//     the threshold is reached the circuit opens.
//   - Open: fn is NOT called and ErrCircuitOpen is returned immediately. If
//     the reset timeout has elapsed the state transitions to HalfOpen and the
//     call is allowed through.
//   - HalfOpen: a single call is allowed through. On success the circuit
//     closes; on failure it reopens.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	cb.mu.Lock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			// Transition to half-open: allow a probe request through.
			cb.state = StateHalfOpen
			cb.logger.Info("circuit breaker transitioning to half-open",
				zap.String("breaker", cb.name),
			)
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("%s: %w", cb.name, ErrCircuitOpen)
		}
	case StateHalfOpen:
		// Already half-open -- allow the call through.
	case StateClosed:
		// Normal operation.
	}

	// Release the lock while the potentially slow call executes.
	cb.mu.Unlock()

	err := fn(ctx)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		switch cb.state {
		case StateHalfOpen:
			// Probe failed; reopen the circuit.
			cb.state = StateOpen
			cb.logger.Warn("circuit breaker reopened after half-open failure",
				zap.String("breaker", cb.name),
				zap.Error(err),
			)
		case StateClosed:
			if cb.failures >= cb.maxFailures {
				cb.state = StateOpen
				cb.logger.Warn("circuit breaker opened",
					zap.String("breaker", cb.name),
					zap.Int("failures", cb.failures),
					zap.Error(err),
				)
			}
		}
		return err
	}

	// Success: reset the circuit.
	if cb.state == StateHalfOpen {
		cb.logger.Info("circuit breaker closed after successful probe",
			zap.String("breaker", cb.name),
		)
	}
	cb.failures = 0
	cb.state = StateClosed
	return nil
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset manually returns the circuit breaker to the Closed state and zeroes
// the failure counter.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.logger.Info("circuit breaker manually reset", zap.String("breaker", cb.name))
}
