package orchestrator

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestAutoEditPassesMatchedAssetsAndSourceRangesToEDLGeneration(t *testing.T) {
	segments := []*ScriptSegment{
		{Index: 0, Type: SegmentTypeBRoll, Content: "Use second.mp4"},
		{Index: 1, Type: SegmentTypeBRoll, Content: "Then first.mp4"},
	}
	assets := []*AssetInfo{
		{ID: "scanner-uuid-1", FilePath: "/fixtures/first.mp4", FileName: "first.mp4"},
		{ID: "scanner-uuid-2", FilePath: "/fixtures/second.mp4", FileName: "second.mp4"},
	}
	matches := []*AssetMatch{
		{SegmentIndex: 0, AssetID: "scanner-uuid-2", Confidence: 0.95, Reasoning: "Explicit source",
			SuggestedRange: &TimeRange{InPoint: Timecode{Seconds: 3, FrameRate: 24}, OutPoint: Timecode{Seconds: 5, FrameRate: 24}}},
		{SegmentIndex: 1, AssetID: "scanner-uuid-1", Confidence: 0.85, Reasoning: "Second script segment"},
	}
	settings := &EDLSettings{Resolution: Resolution{Width: 320, Height: 180}, FrameRate: 24}
	intel := &mockIntelClient{
		parseResult: &ParsedScript{Segments: segments, Metadata: &ScriptMetadata{SegmentCount: 2}},
		matchResult: &MatchResult{Matches: matches},
		edlResult:   &EDL{Entries: []*EDLEntry{{SourceAssetID: "scanner-uuid-2", SourceRange: matches[0].SuggestedRange}, {SourceAssetID: "scanner-uuid-1"}}},
	}
	premiere := &mockPremiereClient{edlExecResult: &EDLExecutionResult{SequenceID: "sequence", Status: "completed", ClipsPlaced: 2}}
	engine := newTestEngine(
		&mockMediaClient{scanResult: &ScanResult{Assets: assets, MediaFilesFound: 2}},
		intel,
		premiere,
	)
	result, err := engine.AutoEdit(context.Background(), &AutoEditParams{
		ScriptText: "Use second.mp4; then first.mp4", AssetsDirectory: "/fixtures", EDLSettings: settings,
	})
	if err != nil {
		t.Fatalf("AutoEdit failed: %v", err)
	}
	if !reflect.DeepEqual(intel.generatedMatches, matches) {
		t.Fatalf("EDL generator received matches %+v, want %+v including source range", intel.generatedMatches, matches)
	}
	if !reflect.DeepEqual(intel.generatedSegments, segments) || !reflect.DeepEqual(intel.generatedAssets, assets) {
		t.Fatal("propagating matches changed the original script segments or scanned assets")
	}
	if intel.generatedSettings != settings {
		t.Fatal("EDL generator did not receive the supplied settings")
	}
	if result.EDL != intel.edlResult || result.EDL.Entries[0].SourceAssetID != "scanner-uuid-2" || result.EDL.Entries[1].SourceAssetID != "scanner-uuid-1" {
		t.Fatal("execution rewrote the canonical generated EDL asset IDs")
	}
	if premiere.executedEDL == result.EDL || premiere.executedEDL.Entries[0] == result.EDL.Entries[0] {
		t.Fatal("Premiere execution must receive a separate EDL and entry copy")
	}
	if premiere.executedEDL.Entries[0].SourceAssetID != "/fixtures/second.mp4" || premiere.executedEDL.Entries[1].SourceAssetID != "/fixtures/first.mp4" {
		t.Fatalf("Premiere received opaque source IDs instead of exact matched file paths: %+v", premiere.executedEDL.Entries)
	}
	if !reflect.DeepEqual(premiere.executedEDL.Entries[0].SourceRange, matches[0].SuggestedRange) {
		t.Fatal("execution path resolution lost the matched source range")
	}
}

func autoEditBoundaryFixture() (*Engine, *mockMediaClient, *mockIntelClient, *mockPremiereClient) {
	media := &mockMediaClient{scanResult: &ScanResult{MediaFilesFound: 1, Assets: []*AssetInfo{{ID: "opaque-id", FilePath: "/fixtures/source.mp4"}}}}
	intel := &mockIntelClient{
		parseResult: &ParsedScript{Segments: []*ScriptSegment{{Index: 0}}, Metadata: &ScriptMetadata{SegmentCount: 1}},
		matchResult: &MatchResult{Matches: []*AssetMatch{{SegmentIndex: 0, AssetID: "opaque-id"}}},
		edlResult:   &EDL{Entries: []*EDLEntry{{SourceAssetID: "opaque-id"}}},
	}
	premiere := &mockPremiereClient{edlExecResult: &EDLExecutionResult{SequenceID: "sequence", Status: "completed", ClipsPlaced: 1}, exportResult: &ExportResult{}}
	return newTestEngine(media, intel, premiere), media, intel, premiere
}

func assertAutoEditFailedBeforeExport(t *testing.T, result *AutoEditResult, premiere *mockPremiereClient, step string) {
	t.Helper()
	last := result.Steps[len(result.Steps)-1]
	if last.Name != step || last.Status != "failed" || last.Error == "" {
		t.Fatalf("expected failed %s step, got %+v", step, last)
	}
	if premiere.exportCalls != 0 {
		t.Fatal("failed EDL must not start an export")
	}
}

func TestAutoEditRejectsInvalidEDLBeforePremiereMutation(t *testing.T) {
	cases := []struct {
		name      string
		step      string
		message   string
		configure func(*mockMediaClient, *mockIntelClient)
	}{
		{"nil EDL", "generate_edl", "no entries", func(_ *mockMediaClient, intel *mockIntelClient) { intel.edlResult = nil }},
		{"empty EDL", "generate_edl", "no entries", func(_ *mockMediaClient, intel *mockIntelClient) { intel.edlResult.Entries = nil }},
		{"nil entry", "execute_edl", "no source asset ID", func(_ *mockMediaClient, intel *mockIntelClient) { intel.edlResult.Entries[0] = nil }},
		{"empty source ID", "execute_edl", "no source asset ID", func(_ *mockMediaClient, intel *mockIntelClient) { intel.edlResult.Entries[0].SourceAssetID = "" }},
		{"missing source", "execute_edl", "absent from the scan", func(_ *mockMediaClient, intel *mockIntelClient) {
			intel.edlResult.Entries[0].SourceAssetID = "unscanned-id"
		}},
		{"duplicate source IDs", "execute_edl", "ambiguous", func(media *mockMediaClient, _ *mockIntelClient) {
			media.scanResult.Assets = append(media.scanResult.Assets, media.scanResult.Assets[0])
		}},
		{"missing path", "execute_edl", "no file path", func(media *mockMediaClient, _ *mockIntelClient) { media.scanResult.Assets[0].FilePath = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, media, intel, premiere := autoEditBoundaryFixture()
			tc.configure(media, intel)
			result, err := engine.AutoEdit(context.Background(), &AutoEditParams{ScriptText: "Use source.mp4", AssetsDirectory: "/fixtures", OutputName: "/tmp/export.mp4"})
			if err == nil || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("expected %q error, got %v", tc.message, err)
			}
			if premiere.executeEDLCalls != 0 {
				t.Fatal("invalid EDL must fail before native execution")
			}
			assertAutoEditFailedBeforeExport(t, result, premiere, tc.step)
		})
	}
}

func TestAutoEditRejectsIncompleteExecutionBeforeExport(t *testing.T) {
	cases := []struct {
		name      string
		execution *EDLExecutionResult
	}{
		{"nil result", nil},
		{"missing sequence identity", &EDLExecutionResult{Status: "completed", ClipsPlaced: 1}},
		{"host failure", &EDLExecutionResult{SequenceID: "sequence", Status: "failed", ClipsPlaced: 1}},
		{"pending result", &EDLExecutionResult{SequenceID: "sequence", Status: "running", ClipsPlaced: 1}},
		{"partial placement", &EDLExecutionResult{SequenceID: "sequence", Status: "completed", ClipsPlaced: 0}},
		{"extra placement", &EDLExecutionResult{SequenceID: "sequence", Status: "completed", ClipsPlaced: 2}},
		{"host errors", &EDLExecutionResult{SequenceID: "sequence", Status: "completed", ClipsPlaced: 1, Errors: []string{"transition failed"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, _, _, premiere := autoEditBoundaryFixture()
			premiere.edlExecResult = tc.execution
			result, err := engine.AutoEdit(context.Background(), &AutoEditParams{ScriptText: "Use source.mp4", AssetsDirectory: "/fixtures", OutputName: "/tmp/export.mp4"})
			if err == nil {
				t.Fatal("incomplete execution must return an error")
			}
			if result.ExecutionResult != tc.execution {
				t.Fatal("partial execution evidence was discarded")
			}
			if premiere.executeEDLCalls != 1 {
				t.Fatal("expected exactly one native execution attempt")
			}
			assertAutoEditFailedBeforeExport(t, result, premiere, "execute_edl")
		})
	}
}
