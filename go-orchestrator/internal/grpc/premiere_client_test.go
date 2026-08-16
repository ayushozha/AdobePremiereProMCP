package grpc

import (
	"strings"
	"testing"
)

func TestUnwrapEvalEnvelopeReturnsDataOnSuccess(t *testing.T) {
	// Every ExtendScript function in core.jsx/premiere.jsx returns
	// JSON.stringify({success, data}) via its _ok() helper.
	got, err := unwrapEvalEnvelope("getSequenceList", `{"success":true,"data":{"count":1,"sequences":[]}}`)
	if err != nil {
		t.Fatalf("unwrapEvalEnvelope() error = %v", err)
	}
	want := `{"count":1,"sequences":[]}`
	if got != want {
		t.Fatalf("unwrapEvalEnvelope() = %q, want %q", got, want)
	}
}

func TestUnwrapEvalEnvelopeReturnsErrorOnFailure(t *testing.T) {
	_, err := unwrapEvalEnvelope("getSequenceList", `{"success":false,"error":"No project is open"}`)
	if err == nil {
		t.Fatal("unwrapEvalEnvelope() expected an error, got nil")
	}
	if got, want := err.Error(), "No project is open"; !strings.Contains(got, want) {
		t.Fatalf("unwrapEvalEnvelope() error = %q, want it to contain %q", got, want)
	}
}

func TestUnwrapEvalEnvelopeUsesFallbackMessageWhenErrorFieldEmpty(t *testing.T) {
	_, err := unwrapEvalEnvelope("getSequenceList", `{"success":false}`)
	if err == nil {
		t.Fatal("unwrapEvalEnvelope() expected an error, got nil")
	}
}

func TestUnwrapEvalEnvelopePassesThroughNonEnvelopedResults(t *testing.T) {
	// Not every caller goes through _ok()/_err() (e.g. bare scalars); those
	// must pass through unchanged rather than being rejected.
	got, err := unwrapEvalEnvelope("someLegacyCommand", `{"count":1,"sequences":[]}`)
	if err != nil {
		t.Fatalf("unwrapEvalEnvelope() error = %v", err)
	}
	want := `{"count":1,"sequences":[]}`
	if got != want {
		t.Fatalf("unwrapEvalEnvelope() = %q, want %q", got, want)
	}
}
