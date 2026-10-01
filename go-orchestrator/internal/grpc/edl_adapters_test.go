package grpc

import (
	"context"
	"reflect"
	"testing"
	"time"

	commonpb "github.com/ayushozha/AdobePremiereProMCP/gen/go/premierpro/common/v1"
	intelpb "github.com/ayushozha/AdobePremiereProMCP/gen/go/premierpro/intelligence/v1"
	premierepb "github.com/ayushozha/AdobePremiereProMCP/gen/go/premierpro/premiere/v1"
	orch "github.com/ayushozha/AdobePremiereProMCP/go-orchestrator/internal/orchestrator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type generatedEDLClient struct {
	intelpb.IntelligenceServiceClient
	request  *intelpb.GenerateEDLRequest
	response *intelpb.GenerateEDLResponse
}

func (c *generatedEDLClient) GenerateEDL(_ context.Context, request *intelpb.GenerateEDLRequest, _ ...grpc.CallOption) (*intelpb.GenerateEDLResponse, error) {
	// Cross real protobuf serialization, so accidental loss of nested fields is
	// observed independently of the adapter's in-memory struct representation.
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	c.request = &intelpb.GenerateEDLRequest{}
	if err := proto.Unmarshal(encoded, c.request); err != nil {
		return nil, err
	}
	encoded, err = proto.Marshal(c.response)
	if err != nil {
		return nil, err
	}
	response := &intelpb.GenerateEDLResponse{}
	return response, proto.Unmarshal(encoded, response)
}

func testEDLRange(seconds, frames uint32, frameRate float64) *orch.TimeRange {
	return &orch.TimeRange{
		InPoint:  orch.Timecode{Hours: 1, Minutes: 2, Seconds: seconds, Frames: frames, FrameRate: frameRate},
		OutPoint: orch.Timecode{Hours: 1, Minutes: 2, Seconds: seconds + 4, Frames: frames + 1, FrameRate: frameRate},
	}
}

func TestIntelAdapterCarriesSelectedMatchesAndGeneratedRangesAcrossProtobuf(t *testing.T) {
	sourceRange := testEDLRange(3, 7, 23.976)
	timelineRange := testEDLRange(30, 11, 29.97)
	suggestedRange := testEDLRange(5, 2, 23.976)
	generated := &orch.EDL{
		ID: "edl-generated", Name: "Exact source",
		SequenceResolution: orch.Resolution{Width: 320, Height: 180}, SequenceFrameRate: 24,
		Entries: []*orch.EDLEntry{{
			Index: 9, SourceAssetID: "opaque-scanner-id", SourceRange: sourceRange,
			TimelineRange: timelineRange, Track: &orch.TrackTarget{Type: orch.TrackTypeAudio, TrackIndex: 2},
			Transition: &orch.TransitionInfo{Type: "cross dissolve", DurationSeconds: 0.5, Alignment: "center"},
			Effects:    []*orch.EffectInfo{{Name: "Volume", Parameters: map[string]string{"level": "-3"}}}, Notes: "retain entry metadata",
		}},
	}
	rpc := &generatedEDLClient{response: &intelpb.GenerateEDLResponse{
		Edl: nativeEDLToProto(convertOrchestratorEDLToGRPC(generated)),
	}}
	adapter := &IntelAdapter{C: &IntelligenceClient{client: rpc, callTimeout: time.Second, logger: zap.NewNop()}}
	matches := []*orch.AssetMatch{{SegmentIndex: 7, AssetID: "opaque-scanner-id", Confidence: 0.97,
		Reasoning: "selected embedding candidate", SuggestedRange: suggestedRange}}
	actual, err := adapter.GenerateEDL(context.Background(), []*orch.ScriptSegment{{Index: 7, Content: "asset.mp4"}},
		[]*orch.AssetInfo{{ID: "opaque-scanner-id", FilePath: "/tmp/asset.mp4"}}, matches,
		&orch.EDLSettings{Resolution: orch.Resolution{Width: 320, Height: 180}, FrameRate: 24})
	if err != nil {
		t.Fatal(err)
	}
	wantMatch := &intelpb.AssetMatch{SegmentIndex: 7, AssetId: "opaque-scanner-id", Confidence: 0.97,
		Reasoning: "selected embedding candidate", SuggestedRange: nativeTimeRangeToProto(ptrRange(suggestedRange))}
	if len(rpc.request.GetMatches()) != 1 || !proto.Equal(rpc.request.Matches[0], wantMatch) {
		t.Fatalf("GenerateEDL lost or changed selected matches: %v", rpc.request.GetMatches())
	}
	if !reflect.DeepEqual(actual, generated) {
		t.Fatalf("generated EDL did not retain both ranges and metadata: got %#v; want %#v", actual.Entries[0], generated.Entries[0])
	}
}

func ptrRange(value *orch.TimeRange) *TimeRange {
	converted := convertOrchestratorTimeRangeToGRPC(value)
	return &converted
}

type executedEDLClient struct {
	premierepb.PremiereBridgeServiceClient
	request *premierepb.ExecuteEDLRequest
}

func (c *executedEDLClient) ExecuteEDL(_ context.Context, request *premierepb.ExecuteEDLRequest, _ ...grpc.CallOption) (*premierepb.ExecuteEDLResponse, error) {
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	c.request = &premierepb.ExecuteEDLRequest{}
	if err := proto.Unmarshal(encoded, c.request); err != nil {
		return nil, err
	}
	return &premierepb.ExecuteEDLResponse{SequenceId: "created", Status: commonpb.OperationStatus_OPERATION_STATUS_COMPLETED, ClipsPlaced: 1}, nil
}

func TestPremiereAdapterPreservesDistinctSourceAndTimelineRanges(t *testing.T) {
	sourceRange := testEDLRange(1, 3, 23.976)
	timelineRange := testEDLRange(20, 9, 29.97)
	rpc := &executedEDLClient{}
	adapter := &PremiereAdapter{C: &PremiereBridgeClient{client: rpc, callTimeout: time.Second, logger: zap.NewNop()}}
	result, err := adapter.ExecuteEDL(context.Background(), &orch.EDL{
		ID: "edl-execution", Name: "Execute exact ranges", SequenceFrameRate: 24,
		SequenceResolution: orch.Resolution{Width: 320, Height: 180},
		Entries: []*orch.EDLEntry{{Index: 0, SourceAssetID: "/tmp/source with spaces.mov",
			SourceRange: sourceRange, TimelineRange: timelineRange, Track: &orch.TrackTarget{Type: orch.TrackTypeVideo, TrackIndex: 1}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rpc.request.GetEdl().GetEntries()) != 1 || !rpc.request.AutoImport || !rpc.request.AutoCreateSequence {
		t.Fatalf("wrong EDL request wrapper: %v", rpc.request)
	}
	entry := rpc.request.Edl.Entries[0]
	if !proto.Equal(entry.SourceRange, nativeTimeRangeToProto(ptrRange(sourceRange))) ||
		!proto.Equal(entry.TimelineRange, nativeTimeRangeToProto(ptrRange(timelineRange))) {
		t.Fatalf("range conversion lost timecode components: %v", entry)
	}
	if entry.SourceAssetId != "/tmp/source with spaces.mov" || entry.Track.Type != commonpb.TrackType_TRACK_TYPE_VIDEO || entry.Track.TrackIndex != 1 {
		t.Fatalf("source/track identity changed: %v", entry)
	}
	if result.Status != "completed" || result.ClipsPlaced != 1 || result.SequenceID != "created" {
		t.Fatalf("execution result changed: %#v", result)
	}
}
