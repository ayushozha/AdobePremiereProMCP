package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
)

// ---------------------------------------------------------------------------
// Smart Editing
// ---------------------------------------------------------------------------

// SmartCut analyzes audio for silence regions and auto-cuts dead air.
func (e *Engine) SmartCut(ctx context.Context, params *SmartCutParams) (*SmartCutResult, error) {
	if params == nil {
		return nil, fmt.Errorf("smart_cut: params must not be nil")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "smartCut", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SmartCut: %w", err)
	}
	var out SmartCutResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("SmartCut: unmarshal: %w", err)
	}
	return &out, nil
}

// SmartTrim trims clips to remove silence/dead air at start and end.
func (e *Engine) SmartTrim(ctx context.Context, params *SmartTrimParams) (*SmartTrimResult, error) {
	if params == nil {
		return nil, fmt.Errorf("smart_trim: params must not be nil")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "smartTrim", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SmartTrim: %w", err)
	}
	var out SmartTrimResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("SmartTrim: unmarshal: %w", err)
	}
	return &out, nil
}

// AutoColorMatch matches color between two clips using Lumetri analysis.
func (e *Engine) AutoColorMatch(ctx context.Context, params *AutoColorMatchParams) (*AutoColorMatchResult, error) {
	if params == nil {
		return nil, fmt.Errorf("auto_color_match: params must not be nil")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "autoColorMatch", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AutoColorMatch: %w", err)
	}
	var out AutoColorMatchResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AutoColorMatch: unmarshal: %w", err)
	}
	return &out, nil
}

// AutoAudioLevels analyzes and normalizes all audio to broadcast standards.
func (e *Engine) AutoAudioLevels(ctx context.Context, params *AutoAudioLevelsParams) (*AutoAudioLevelsResult, error) {
	if params == nil {
		return nil, fmt.Errorf("auto_audio_levels: params must not be nil")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "autoAudioLevels", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AutoAudioLevels: %w", err)
	}
	var out AutoAudioLevelsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AutoAudioLevels: unmarshal: %w", err)
	}
	return &out, nil
}

// SuggestTransitions uses AI to suggest transition types based on content analysis.
func (e *Engine) SuggestTransitions(ctx context.Context, sequenceID string) (*SuggestTransitionsResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("suggest_transitions: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "suggestTransitions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SuggestTransitions: %w", err)
	}
	var out SuggestTransitionsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("SuggestTransitions: unmarshal: %w", err)
	}
	return &out, nil
}

// SuggestMusic uses AI to suggest music timing/cuts based on scene pacing.
func (e *Engine) SuggestMusic(ctx context.Context, sequenceID string) (*SuggestMusicResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("suggest_music: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "suggestMusic", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SuggestMusic: %w", err)
	}
	var out SuggestMusicResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("SuggestMusic: unmarshal: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Content Analysis
// ---------------------------------------------------------------------------

// AnalyzeClip performs full analysis of a clip.
func (e *Engine) AnalyzeClip(ctx context.Context, filePath string) (*ClipAnalysis, error) {
	if filePath == "" {
		return nil, fmt.Errorf("analyze_clip: file_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "analyzeClip", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AnalyzeClip: %w", err)
	}
	var out ClipAnalysis
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AnalyzeClip: unmarshal: %w", err)
	}
	return &out, nil
}

// AnalyzeSequence performs full sequence analysis.
func (e *Engine) AnalyzeSequence(ctx context.Context, sequenceID string) (*SequenceAnalysis, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("analyze_sequence: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "analyzeSequence", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AnalyzeSequence: %w", err)
	}
	var out SequenceAnalysis
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AnalyzeSequence: unmarshal: %w", err)
	}
	return &out, nil
}

// GetSequenceStatistics returns summary statistics for a sequence.
func (e *Engine) GetSequenceStatistics(ctx context.Context, sequenceID string) (*SequenceStatistics, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("get_sequence_statistics: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "getSequenceStatistics", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GetSequenceStatistics: %w", err)
	}
	var out SequenceStatistics
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GetSequenceStatistics: unmarshal: %w", err)
	}
	return &out, nil
}

// DetectJumpCuts detects potential jump cuts (similar adjacent frames).
func (e *Engine) DetectJumpCuts(ctx context.Context, sequenceID string, threshold float64) (*JumpCutResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("detect_jump_cuts: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
		"threshold":  threshold,
	})
	result, err := e.premiere.EvalCommand(ctx, "detectJumpCuts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DetectJumpCuts: %w", err)
	}
	var out JumpCutResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("DetectJumpCuts: unmarshal: %w", err)
	}
	return &out, nil
}

// DetectAudioIssues detects audio clipping, silence, and level mismatches.
func (e *Engine) DetectAudioIssues(ctx context.Context, sequenceID string) (*AudioIssuesResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("detect_audio_issues: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "detectAudioIssues", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("DetectAudioIssues: %w", err)
	}
	var out AudioIssuesResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("DetectAudioIssues: unmarshal: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// AI Script-to-Edit Pipeline (enhanced)
// ---------------------------------------------------------------------------

// GenerateRoughCut generates a rough cut from script and assets.
func (e *Engine) GenerateRoughCut(ctx context.Context, params *RoughCutParams) (*RoughCutResult, error) {
	if params == nil {
		return nil, fmt.Errorf("generate_rough_cut: params must not be nil")
	}
	if params.AssetsDirectory == "" {
		return nil, fmt.Errorf("generate_rough_cut: assets_directory must not be empty")
	}
	if params.ScriptText == "" && params.ScriptPath == "" {
		return nil, fmt.Errorf("generate_rough_cut: either script_text or script_path must be provided")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateRoughCut", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateRoughCut: %w", err)
	}
	var out RoughCutResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GenerateRoughCut: unmarshal: %w", err)
	}
	return &out, nil
}

// RefineEdit performs an AI refinement pass on an existing edit.
func (e *Engine) RefineEdit(ctx context.Context, params *RefineEditParams) (*RefineEditResult, error) {
	if params == nil {
		return nil, fmt.Errorf("refine_edit: params must not be nil")
	}
	if params.SequenceID == "" {
		return nil, fmt.Errorf("refine_edit: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "refineEdit", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("RefineEdit: %w", err)
	}
	var out RefineEditResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("RefineEdit: unmarshal: %w", err)
	}
	return &out, nil
}

// AddBRollSuggestions suggests B-roll placement based on dialogue analysis.
func (e *Engine) AddBRollSuggestions(ctx context.Context, sequenceID string) (*BRollSuggestionsResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("add_broll_suggestions: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "addBRollSuggestions", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AddBRollSuggestions: %w", err)
	}
	var out BRollSuggestionsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AddBRollSuggestions: unmarshal: %w", err)
	}
	return &out, nil
}

// GenerateTrailer generates a trailer or highlight reel from a sequence.
func (e *Engine) GenerateTrailer(ctx context.Context, params *GenerateTrailerParams) (*GenerateTrailerResult, error) {
	if params == nil {
		return nil, fmt.Errorf("generate_trailer: params must not be nil")
	}
	if params.SequenceID == "" {
		return nil, fmt.Errorf("generate_trailer: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateTrailer", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateTrailer: %w", err)
	}
	var out GenerateTrailerResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GenerateTrailer: unmarshal: %w", err)
	}
	return &out, nil
}

// CreateSocialCuts creates social media cuts from a sequence.
func (e *Engine) CreateSocialCuts(ctx context.Context, params *SocialCutParams) (*SocialCutResult, error) {
	if params == nil {
		return nil, fmt.Errorf("create_social_cuts: params must not be nil")
	}
	if params.SequenceID == "" {
		return nil, fmt.Errorf("create_social_cuts: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "createSocialCuts", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateSocialCuts: %w", err)
	}
	var out SocialCutResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("CreateSocialCuts: unmarshal: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Smart Organisation
// ---------------------------------------------------------------------------

// AutoOrganizeProject organises the project bins by content type.
func (e *Engine) AutoOrganizeProject(ctx context.Context, params *AutoOrganizeParams) (*AutoOrganizeResult, error) {
	if params == nil {
		return nil, fmt.Errorf("auto_organize_project: params must not be nil")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"params": params,
	})
	result, err := e.premiere.EvalCommand(ctx, "autoOrganizeProject", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AutoOrganizeProject: %w", err)
	}
	var out AutoOrganizeResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AutoOrganizeProject: unmarshal: %w", err)
	}
	return &out, nil
}

// AITagClips generates AI tags and metadata for a clip.
func (e *Engine) AITagClips(ctx context.Context, filePath string) (*TagClipsResult, error) {
	if filePath == "" {
		return nil, fmt.Errorf("tag_clips: file_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath": filePath,
	})
	result, err := e.premiere.EvalCommand(ctx, "aITagClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("AITagClips: %w", err)
	}
	var out TagClipsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("AITagClips: unmarshal: %w", err)
	}
	return &out, nil
}

// FindSimilarClips finds visually or audibly similar clips.
func (e *Engine) FindSimilarClips(ctx context.Context, filePath string, maxResults int) (*FindSimilarResult, error) {
	if filePath == "" {
		return nil, fmt.Errorf("find_similar_clips: file_path must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"filePath":   filePath,
		"maxResults": maxResults,
	})
	result, err := e.premiere.EvalCommand(ctx, "findSimilarClips", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("FindSimilarClips: %w", err)
	}
	var out FindSimilarResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("FindSimilarClips: unmarshal: %w", err)
	}
	return &out, nil
}

// SuggestReplacements suggests better clips for timeline positions.
func (e *Engine) SuggestReplacements(ctx context.Context, sequenceID string) (*SuggestReplacementsResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("suggest_replacements: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "suggestReplacements", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("SuggestReplacements: %w", err)
	}
	var out SuggestReplacementsResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("SuggestReplacements: unmarshal: %w", err)
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Workflow Automation
// ---------------------------------------------------------------------------

// CreateReviewMarkers adds review markers at key decision points.
func (e *Engine) CreateReviewMarkers(ctx context.Context, sequenceID string) (*ReviewMarkersResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("create_review_markers: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "createReviewMarkers", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateReviewMarkers: %w", err)
	}
	var out ReviewMarkersResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("CreateReviewMarkers: unmarshal: %w", err)
	}
	return &out, nil
}

// GenerateEditSummary generates a text summary of the current edit.
func (e *Engine) GenerateEditSummary(ctx context.Context, sequenceID string) (*EditSummaryResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("generate_edit_summary: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "generateEditSummary", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("GenerateEditSummary: %w", err)
	}
	var out EditSummaryResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("GenerateEditSummary: unmarshal: %w", err)
	}
	return &out, nil
}

// EstimateRenderTime estimates the render time for the current sequence.
func (e *Engine) EstimateRenderTime(ctx context.Context, sequenceID string) (*RenderTimeEstimate, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("estimate_render_time: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
	})
	result, err := e.premiere.EvalCommand(ctx, "estimateRenderTime", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("EstimateRenderTime: %w", err)
	}
	var out RenderTimeEstimate
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("EstimateRenderTime: unmarshal: %w", err)
	}
	return &out, nil
}

// CheckDeliverySpecs checks if a sequence meets delivery specifications.
func (e *Engine) CheckDeliverySpecs(ctx context.Context, sequenceID string, standard string) (*DeliverySpecResult, error) {
	if sequenceID == "" {
		return nil, fmt.Errorf("check_delivery_specs: sequence_id must not be empty")
	}
	argsJSON, _ := json.Marshal(map[string]any{
		"sequenceID": sequenceID,
		"standard":   standard,
	})
	result, err := e.premiere.EvalCommand(ctx, "checkDeliverySpecs", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CheckDeliverySpecs: %w", err)
	}
	var out DeliverySpecResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("CheckDeliverySpecs: unmarshal: %w", err)
	}
	return &out, nil
}

// CreateProjectReport generates a comprehensive project report.
func (e *Engine) CreateProjectReport(ctx context.Context) (*ProjectReportResult, error) {
	argsJSON, _ := json.Marshal(map[string]any{})
	result, err := e.premiere.EvalCommand(ctx, "createProjectReport", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("CreateProjectReport: %w", err)
	}
	var out ProjectReportResult
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		return nil, fmt.Errorf("CreateProjectReport: unmarshal: %w", err)
	}
	return &out, nil
}
