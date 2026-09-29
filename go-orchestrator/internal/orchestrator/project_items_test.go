package orchestrator

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"testing"
)

func TestGetProjectItemsPreservesHostMediaPaths(t *testing.T) {
	// Exact field names emitted by CEP getProjectItems, after its success envelope.
	client := &mockPremiereClient{evalResult: `{"binPath":"/","itemCount":2,"items":[{"index":0,"name":"e2e_test_pattern.mp4","type":"clip","mediaPath":"/fixtures/e2e_test_pattern.mp4"},{"index":1,"name":"Footage","type":"bin","childCount":3}]}`}
	engine := &Engine{premiere: client, logger: zap.NewNop()}
	result, err := engine.GetProjectItems(context.Background(), "/")
	if err != nil {
		t.Fatal(err)
	}
	if result.BinPath != "/" || result.ItemCount != 2 || len(result.Items) != 2 {
		t.Fatalf("lost host inventory fields: %+v", result)
	}
	if result.Items[0].MediaPath != "/fixtures/e2e_test_pattern.mp4" {
		t.Fatalf("lost imported media path: %+v", result.Items[0])
	}
	if result.Items[1].ChildCount != 3 {
		t.Fatalf("lost bin child count: %+v", result.Items[1])
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var public map[string]any
	if err := json.Unmarshal(encoded, &public); err != nil {
		t.Fatal(err)
	}
	if public["item_count"] != float64(2) || public["bin_path"] != "/" {
		t.Fatalf("public schema changed: %s", encoded)
	}
	item := public["items"].([]any)[0].(map[string]any)
	if item["media_path"] != "/fixtures/e2e_test_pattern.mp4" || item["mediaPath"] != nil {
		t.Fatalf("public media_path contract changed: %s", encoded)
	}
}

func TestGetProjectItemsRetainsSnakeCaseAndRejectsInvalidPayload(t *testing.T) {
	client := &mockPremiereClient{evalResult: `{"bin_path":"/Footage","item_count":1,"items":[{"name":"audio.wav","type":"clip","media_path":"/fixtures/audio.wav","child_count":0}]}`}
	engine := &Engine{premiere: client, logger: zap.NewNop()}
	result, err := engine.GetProjectItems(context.Background(), "/Footage")
	if err != nil {
		t.Fatal(err)
	}
	if result.ItemCount != 1 || result.BinPath != "/Footage" || result.Items[0].MediaPath != "/fixtures/audio.wav" {
		t.Fatalf("snake_case regression: %+v", result)
	}
	client.evalResult = `{"items":"wrong type"}`
	if _, err := engine.GetProjectItems(context.Background(), "/"); err == nil {
		t.Fatal("invalid native payload was accepted")
	}
}
