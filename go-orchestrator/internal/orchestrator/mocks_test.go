package orchestrator

import (
	"context"
)

// ---------------------------------------------------------------------------
// mockMediaClient
// ---------------------------------------------------------------------------

type mockMediaClient struct {
	scanResult    *ScanResult
	scanErr       error
	probeResult   *AssetInfo
	probeErr      error
	waveformResult *WaveformResult
	waveformErr   error
	sceneResult   *SceneResult
	sceneErr      error
}

func (m *mockMediaClient) ScanAssets(_ context.Context, _ string, _ bool, _ []string) (*ScanResult, error) {
	return m.scanResult, m.scanErr
}

func (m *mockMediaClient) ProbeMedia(_ context.Context, _ string) (*AssetInfo, error) {
	return m.probeResult, m.probeErr
}

func (m *mockMediaClient) AnalyzeWaveform(_ context.Context, _ string, _ *WaveformOptions) (*WaveformResult, error) {
	return m.waveformResult, m.waveformErr
}

func (m *mockMediaClient) DetectScenes(_ context.Context, _ string, _ float64) (*SceneResult, error) {
	return m.sceneResult, m.sceneErr
}

// ---------------------------------------------------------------------------
// mockIntelClient
// ---------------------------------------------------------------------------

type mockIntelClient struct {
	parseResult  *ParsedScript
	parseErr     error
	matchResult  *MatchResult
	matchErr     error
	edlResult    *EDL
	edlErr       error
	pacingResult *PacingResult
	pacingErr    error
}

func (m *mockIntelClient) ParseScript(_ context.Context, _ string, _ string, _ string) (*ParsedScript, error) {
	return m.parseResult, m.parseErr
}

func (m *mockIntelClient) GenerateEDL(_ context.Context, _ []*ScriptSegment, _ []*AssetInfo, _ *EDLSettings) (*EDL, error) {
	return m.edlResult, m.edlErr
}

func (m *mockIntelClient) MatchAssets(_ context.Context, _ []*ScriptSegment, _ []*AssetInfo, _ string) (*MatchResult, error) {
	return m.matchResult, m.matchErr
}

func (m *mockIntelClient) AnalyzePacing(_ context.Context, _ *EDL, _ string) (*PacingResult, error) {
	return m.pacingResult, m.pacingErr
}

// ---------------------------------------------------------------------------
// mockPremiereClient
// ---------------------------------------------------------------------------

type mockPremiereClient struct {
	pingResult       *PingResult
	pingErr          error
	projectState     *ProjectState
	projectStateErr  error
	seqResult        *SequenceResult
	seqErr           error
	importResult     *ImportResult
	importErr        error
	placeResult      *ClipResult
	placeErr         error
	removeErr        error
	transitionErr    error
	textResult       *ClipResult
	textErr          error
	audioErr         error
	timelineState    *TimelineState
	timelineErr      error
	exportResult     *ExportResult
	exportErr        error
	edlExecResult    *EDLExecutionResult
	edlExecErr       error
	evalResult       string
	evalErr          error
	evalAudioResult  map[string]any
	evalAudioErr     error
	evalImmResult    map[string]any
	evalImmErr       error
}

func (m *mockPremiereClient) Ping(_ context.Context) (*PingResult, error) {
	return m.pingResult, m.pingErr
}

func (m *mockPremiereClient) GetProjectState(_ context.Context) (*ProjectState, error) {
	return m.projectState, m.projectStateErr
}

func (m *mockPremiereClient) CreateSequence(_ context.Context, _ *CreateSequenceParams) (*SequenceResult, error) {
	return m.seqResult, m.seqErr
}

func (m *mockPremiereClient) ImportMedia(_ context.Context, _ string, _ string) (*ImportResult, error) {
	return m.importResult, m.importErr
}

func (m *mockPremiereClient) PlaceClip(_ context.Context, _ *PlaceClipParams) (*ClipResult, error) {
	return m.placeResult, m.placeErr
}

func (m *mockPremiereClient) RemoveClip(_ context.Context, _ string, _ string) error {
	return m.removeErr
}

func (m *mockPremiereClient) AddTransition(_ context.Context, _ *TransitionParams) error {
	return m.transitionErr
}

func (m *mockPremiereClient) AddText(_ context.Context, _ *TextParams) (*ClipResult, error) {
	return m.textResult, m.textErr
}

func (m *mockPremiereClient) SetAudioLevel(_ context.Context, _ string, _ string, _ float64) error {
	return m.audioErr
}

func (m *mockPremiereClient) GetTimelineState(_ context.Context, _ string) (*TimelineState, error) {
	return m.timelineState, m.timelineErr
}

func (m *mockPremiereClient) ExportSequence(_ context.Context, _ *ExportParams) (*ExportResult, error) {
	return m.exportResult, m.exportErr
}

func (m *mockPremiereClient) ExecuteEDL(_ context.Context, _ *EDL) (*EDLExecutionResult, error) {
	return m.edlExecResult, m.edlExecErr
}

func (m *mockPremiereClient) EvalCommand(_ context.Context, _ string, _ string) (string, error) {
	return m.evalResult, m.evalErr
}

func (m *mockPremiereClient) EvalAudioCommand(_ context.Context, _ string, _ map[string]any) (map[string]any, error) {
	return m.evalAudioResult, m.evalAudioErr
}

func (m *mockPremiereClient) EvalImmersiveCommand(_ context.Context, _ string, _ map[string]any) (map[string]any, error) {
	return m.evalImmResult, m.evalImmErr
}
