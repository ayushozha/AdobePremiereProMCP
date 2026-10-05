package orchestrator

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"strings"
	"testing"
)

func TestGetSequenceListNormalizesHostMetadata(t *testing.T) {
	canonical := `{"count":1,"active_sequence_id":"seq-a","sequences":[{"index":0,"name":"Cut","sequence_id":"seq-a","frame_size_horizontal":1920,"frame_size_vertical":1080,"timebase":"10584000000","video_track_count":3,"audio_track_count":2,"is_active":true}]}`
	legacy := `{"count":1,"activeSequenceID":"seq-a","sequences":[{"index":0,"name":"Cut","sequenceID":"seq-a","frameSizeHorizontal":1920,"frameSizeVertical":1080,"timebase":"10584000000","videoTrackCount":3,"audioTrackCount":2,"isActive":true}]}`
	for name, raw := range map[string]string{"canonical": canonical, "camelCase": legacy, "wrapped": `{"success":true,"data":` + legacy + `}`} {
		t.Run(name, func(t *testing.T) {
			engine := &Engine{premiere: &mockPremiereClient{evalResult: raw}, logger: zap.NewNop()}
			got, err := engine.GetSequenceList(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if got.Count != 1 || got.ActiveSequenceID != "seq-a" || len(got.Sequences) != 1 {
				t.Fatalf("lost list metadata: %+v", got)
			}
			item := got.Sequences[0]
			if item.SequenceID != "seq-a" || item.FrameSizeHorizontal != 1920 || item.FrameSizeVertical != 1080 || item.VideoTrackCount != 3 || item.AudioTrackCount != 2 || !item.IsActive {
				t.Fatalf("lost sequence fields: %+v", item)
			}
			public, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(public, &fields)
			if fields["active_sequence_id"] == nil || fields["activeSequenceID"] != nil {
				t.Fatalf("public schema changed: %s", public)
			}
		})
	}
}

func TestGetSequenceListRejectsFalseOrUnverifiableResults(t *testing.T) {
	for _, raw := range []string{
		`{}`, `null`, `{"success":false,"error":"Host cannot read sequences"}`,
		`{"success":true,"data":null}`, `{"count":1,"sequences":[]}`,
		`{"count":1,"sequences":[{"name":"Cut"}]}`,
		`{"count":2,"sequences":[{"sequence_id":"same"},{"sequence_id":"same"}]}`,
		`{"count":1,"sequences":[{"sequence_id":"a","sequenceID":"b"}]}`,
		`{"count":1,"active_sequence_id":"absent","sequences":[{"sequence_id":"a"}]}`,
		`{"count":1,"sequences":[{"index":0,"name":"Cut","sequence_id":"a"}]}`,
		`{"count":1,"active_sequence_id":"a","sequences":[{"index":0,"name":"Cut","sequence_id":"a","frame_size_horizontal":1920,"frame_size_vertical":1080,"timebase":"1","video_track_count":1,"audio_track_count":1,"is_active":false}]}`,
	} {
		t.Run(raw, func(t *testing.T) {
			engine := &Engine{premiere: &mockPremiereClient{evalResult: raw}, logger: zap.NewNop()}
			got, err := engine.GetSequenceList(context.Background())
			if err == nil {
				t.Fatalf("accepted unverifiable result: %+v", got)
			}
		})
	}
}

func TestGetSequenceListAllowsEmptyProject(t *testing.T) {
	engine := &Engine{premiere: &mockPremiereClient{evalResult: `{"count":0,"active_sequence_id":"","sequences":[]}`}, logger: zap.NewNop()}
	got, err := engine.GetSequenceList(context.Background())
	if err != nil || got.Count != 0 || len(got.Sequences) != 0 {
		t.Fatalf("empty project: %+v %v", got, err)
	}
}

func TestSequenceListRejectsUnreadableTimebaseAndActiveMetadata(t *testing.T) {
	valid := `{"count":1,"active_sequence_id":"","sequences":[{"index":0,"name":"Cut","sequence_id":"a","frame_size_horizontal":1920,"frame_size_vertical":1080,"timebase":"10584000000","video_track_count":1,"audio_track_count":1,"is_active":false}]}`
	for _, raw := range []string{strings.Replace(valid, `"active_sequence_id":"",`, "", 1), strings.Replace(valid, `"active_sequence_id":""`, `"active_sequence_id":null`, 1), strings.Replace(valid, `"10584000000"`, `""`, 1), strings.Replace(valid, `"10584000000"`, `"invalid"`, 1), strings.Replace(valid, `"10584000000"`, `"0"`, 1), strings.Replace(valid, `"10584000000"`, `"-1"`, 1)} {
		if err := decodeSequenceList(raw, &SequenceListResult{}); err == nil {
			t.Fatalf("accepted incomplete metadata: %s", raw)
		}
	}
}
