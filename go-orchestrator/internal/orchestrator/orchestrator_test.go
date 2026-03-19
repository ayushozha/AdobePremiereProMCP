package orchestrator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// helper: build an Engine with the three mock clients.
func newTestEngine(media *mockMediaClient, intel *mockIntelClient, premiere *mockPremiereClient) *Engine {
	return New(media, intel, premiere, zap.NewNop())
}

// -----------------------------------------------------------------------
// TestNew
// -----------------------------------------------------------------------

func TestNew(t *testing.T) {
	t.Run("all clients provided", func(t *testing.T) {
		eng := newTestEngine(&mockMediaClient{}, &mockIntelClient{}, &mockPremiereClient{})
		if eng == nil {
			t.Fatal("expected non-nil Engine")
		}
	})

	t.Run("nil logger uses nop", func(t *testing.T) {
		eng := New(&mockMediaClient{}, &mockIntelClient{}, &mockPremiereClient{}, nil)
		if eng == nil {
			t.Fatal("expected non-nil Engine with nil logger")
		}
	})

	t.Run("nil media panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic for nil MediaClient")
			}
		}()
		New(nil, &mockIntelClient{}, &mockPremiereClient{}, zap.NewNop())
	})

	t.Run("nil intel panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic for nil IntelClient")
			}
		}()
		New(&mockMediaClient{}, nil, &mockPremiereClient{}, zap.NewNop())
	})

	t.Run("nil premiere panics", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic for nil PremiereClient")
			}
		}()
		New(&mockMediaClient{}, &mockIntelClient{}, nil, zap.NewNop())
	})
}

// -----------------------------------------------------------------------
// TestPing
// -----------------------------------------------------------------------

func TestPing(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &PingResult{
			PremiereRunning: true,
			PremiereVersion: "25.1",
			ProjectOpen:     true,
			BridgeMode:      "cep",
		}
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{pingResult: expected},
		)

		got, err := eng.Ping(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != expected {
			t.Fatalf("got %+v, want %+v", got, expected)
		}
	})

	t.Run("error", func(t *testing.T) {
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{pingErr: errors.New("connection refused")},
		)

		_, err := eng.Ping(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "could not reach Premiere Pro") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

// -----------------------------------------------------------------------
// TestGetProject
// -----------------------------------------------------------------------

func TestGetProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &ProjectState{
			ProjectName: "TestProject",
			ProjectPath: "/tmp/test.prproj",
			Sequences:   []*SequenceInfo{{ID: "seq-1", Name: "Main"}},
			BinCount:    3,
			IsSaved:     true,
		}
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{projectState: expected},
		)

		got, err := eng.GetProject(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProjectName != expected.ProjectName {
			t.Fatalf("project name: got %q, want %q", got.ProjectName, expected.ProjectName)
		}
		if len(got.Sequences) != 1 {
			t.Fatalf("sequence count: got %d, want 1", len(got.Sequences))
		}
	})

	t.Run("error", func(t *testing.T) {
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{projectStateErr: errors.New("no project")},
		)

		_, err := eng.GetProject(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "could not retrieve project state") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

// -----------------------------------------------------------------------
// TestCreateSequence
// -----------------------------------------------------------------------

func TestCreateSequence(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &SequenceResult{SequenceID: "seq-42", Name: "My Seq"}
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{seqResult: expected},
		)

		params := &CreateSequenceParams{
			Name:       "My Seq",
			Resolution: Resolution{Width: 1920, Height: 1080},
			FrameRate:  24,
		}
		got, err := eng.CreateSequence(context.Background(), params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.SequenceID != "seq-42" {
			t.Fatalf("sequence id: got %q, want %q", got.SequenceID, "seq-42")
		}
	})

	t.Run("nil params", func(t *testing.T) {
		eng := newTestEngine(&mockMediaClient{}, &mockIntelClient{}, &mockPremiereClient{})

		_, err := eng.CreateSequence(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for nil params")
		}
		if !strings.Contains(err.Error(), "params must not be nil") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("delegate error", func(t *testing.T) {
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{seqErr: errors.New("fail")},
		)

		params := &CreateSequenceParams{Name: "X", Resolution: Resolution{Width: 1920, Height: 1080}, FrameRate: 30}
		_, err := eng.CreateSequence(context.Background(), params)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "failed to create sequence") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

// -----------------------------------------------------------------------
// TestImportMedia
// -----------------------------------------------------------------------

func TestImportMedia(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := &ImportResult{ProjectItemID: "item-1", Name: "clip.mp4"}
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{importResult: expected},
		)

		got, err := eng.ImportMedia(context.Background(), "/tmp/clip.mp4", "Footage")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProjectItemID != "item-1" {
			t.Fatalf("project item id: got %q, want %q", got.ProjectItemID, "item-1")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		eng := newTestEngine(&mockMediaClient{}, &mockIntelClient{}, &mockPremiereClient{})

		_, err := eng.ImportMedia(context.Background(), "", "Footage")
		if err == nil {
			t.Fatal("expected error for empty path")
		}
		if !strings.Contains(err.Error(), "file_path is required") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("delegate error", func(t *testing.T) {
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{importErr: errors.New("boom")},
		)

		_, err := eng.ImportMedia(context.Background(), "/tmp/clip.mp4", "")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "failed to import") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

// -----------------------------------------------------------------------
// TestEvalCommand
// -----------------------------------------------------------------------

func TestEvalCommand(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{evalResult: `{"ok":true}`},
		)

		got, err := eng.EvalCommand(context.Background(), "doSomething", `{"key":"val"}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != `{"ok":true}` {
			t.Fatalf("result: got %q, want %q", got, `{"ok":true}`)
		}
	})

	t.Run("empty function name", func(t *testing.T) {
		eng := newTestEngine(&mockMediaClient{}, &mockIntelClient{}, &mockPremiereClient{})

		_, err := eng.EvalCommand(context.Background(), "", "{}")
		if err == nil {
			t.Fatal("expected error for empty function name")
		}
		if !strings.Contains(err.Error(), "function_name is required") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("delegate error", func(t *testing.T) {
		eng := newTestEngine(
			&mockMediaClient{},
			&mockIntelClient{},
			&mockPremiereClient{evalErr: errors.New("timeout")},
		)

		_, err := eng.EvalCommand(context.Background(), "fn", "{}")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), `command "fn" failed`) {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

// -----------------------------------------------------------------------
// TestAutoEdit — full success path
// -----------------------------------------------------------------------

func TestAutoEdit(t *testing.T) {
	media := &mockMediaClient{
		scanResult: &ScanResult{
			Assets: []*AssetInfo{
				{ID: "a1", FilePath: "/tmp/a.mp4", FileName: "a.mp4"},
				{ID: "a2", FilePath: "/tmp/b.mp4", FileName: "b.mp4"},
			},
			TotalFilesScanned:   10,
			MediaFilesFound:     2,
			ScanDurationSeconds: 0.5,
		},
	}
	intel := &mockIntelClient{
		parseResult: &ParsedScript{
			Segments: []*ScriptSegment{
				{Index: 0, Type: SegmentTypeDialogue, Content: "Hello"},
				{Index: 1, Type: SegmentTypeBRoll, Content: "Cityscape"},
			},
			Metadata: &ScriptMetadata{
				Title:        "Test Script",
				Format:       "youtube",
				SegmentCount: 2,
			},
		},
		matchResult: &MatchResult{
			Matches: []*AssetMatch{
				{SegmentIndex: 0, AssetID: "a1", Confidence: 0.9},
				{SegmentIndex: 1, AssetID: "a2", Confidence: 0.85},
			},
		},
		edlResult: &EDL{
			ID:   "edl-1",
			Name: "auto-edit-edl",
			Entries: []*EDLEntry{
				{Index: 0, SourceAssetID: "a1"},
				{Index: 1, SourceAssetID: "a2"},
			},
		},
	}
	premiere := &mockPremiereClient{
		edlExecResult: &EDLExecutionResult{
			SequenceID:       "seq-100",
			Status:           "completed",
			ClipsPlaced:      2,
			TransitionsAdded: 1,
		},
	}

	eng := newTestEngine(media, intel, premiere)

	params := &AutoEditParams{
		ScriptText:      "Hello world\nCityscape B-Roll",
		AssetsDirectory: "/tmp/footage",
		Recursive:       true,
	}

	result, err := eng.AutoEdit(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all steps completed
	completedCount := 0
	for _, s := range result.Steps {
		if s.Status == "completed" {
			completedCount++
		}
	}
	// Steps: fan_out_scan_and_parse, match_assets, generate_edl, execute_edl, export(skipped)
	if completedCount != 4 {
		t.Fatalf("completed steps: got %d, want 4", completedCount)
	}

	if result.ScanResult == nil {
		t.Fatal("expected ScanResult to be set")
	}
	if result.ParsedScript == nil {
		t.Fatal("expected ParsedScript to be set")
	}
	if result.MatchResult == nil {
		t.Fatal("expected MatchResult to be set")
	}
	if result.EDL == nil {
		t.Fatal("expected EDL to be set")
	}
	if result.ExecutionResult == nil {
		t.Fatal("expected ExecutionResult to be set")
	}
	if result.ExecutionResult.ClipsPlaced != 2 {
		t.Fatalf("clips placed: got %d, want 2", result.ExecutionResult.ClipsPlaced)
	}
	if result.ExportResult != nil {
		t.Fatal("expected ExportResult to be nil (no output_name)")
	}
}

// -----------------------------------------------------------------------
// TestAutoEditNoAssets — scan returns 0 media files
// -----------------------------------------------------------------------

func TestAutoEditNoAssets(t *testing.T) {
	media := &mockMediaClient{
		scanResult: &ScanResult{
			Assets:              []*AssetInfo{},
			TotalFilesScanned:   5,
			MediaFilesFound:     0,
			ScanDurationSeconds: 0.1,
		},
	}
	intel := &mockIntelClient{
		parseResult: &ParsedScript{
			Segments: []*ScriptSegment{
				{Index: 0, Content: "test"},
			},
			Metadata: &ScriptMetadata{
				Title:        "Empty",
				SegmentCount: 1,
			},
		},
	}
	premiere := &mockPremiereClient{}

	eng := newTestEngine(media, intel, premiere)

	params := &AutoEditParams{
		ScriptText:      "Some script",
		AssetsDirectory: "/empty/dir",
	}

	result, err := eng.AutoEdit(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for no assets")
	}
	if !strings.Contains(err.Error(), "0 media files") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result even on error")
	}
	// The match_assets step should be skipped
	found := false
	for _, s := range result.Steps {
		if s.Name == "match_assets" && s.Status == "skipped" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected match_assets step to be skipped")
	}
}

// -----------------------------------------------------------------------
// TestAutoEditNoScript — empty script_text and script_path
// -----------------------------------------------------------------------

func TestAutoEditNoScript(t *testing.T) {
	eng := newTestEngine(&mockMediaClient{}, &mockIntelClient{}, &mockPremiereClient{})

	params := &AutoEditParams{
		AssetsDirectory: "/tmp/footage",
		// ScriptText and ScriptPath are both empty
	}

	_, err := eng.AutoEdit(context.Background(), params)
	if err == nil {
		t.Fatal("expected error for missing script")
	}
	if !strings.Contains(err.Error(), "script_text or script_path must be provided") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
