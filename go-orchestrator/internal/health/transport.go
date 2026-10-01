package health

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCTransportProbe checks the HTTP/2 gRPC transport for dependencies without
// a health RPC. It does not claim the dependency's application/model is ready.
func GRPCTransportProbe(address string) Probe {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return errors.New("invalid gRPC endpoint")
		}
		defer connection.Close()
		connection.Connect()
		for {
			state := connection.GetState()
			if state == connectivity.Ready {
				return nil
			}
			if state == connectivity.Shutdown || !connection.WaitForStateChange(ctx, state) {
				return errors.New("gRPC transport unavailable")
			}
		}
	}
}

// PremiereReadinessProbe checks both the bridge RPC and its application
// connection. It bounds probe time and keeps raw RPC errors out of HTTP output.
func PremiereReadinessProbe(ping func(context.Context) (bool, error)) Probe {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, probeTimeout)
		defer cancel()
		running, err := ping(ctx)
		if err != nil {
			return errors.New("bridge ping failed")
		}
		if !running {
			return errors.New("Premiere is not connected")
		}
		return nil
	}
}
