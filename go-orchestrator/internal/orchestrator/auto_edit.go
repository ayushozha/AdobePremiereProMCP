package orchestrator

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// AutoEdit performs the full automated editing pipeline:
//
//  1. Fan-out (parallel):  Scan assets (Rust) + Parse script (Python)
//  2. Sequential:          Match assets to script segments (Python)
//  3. Sequential:          Generate EDL from matches (Python)
//  4. Execute:             Assemble timeline in Premiere (TypeScript)
//  5. Optional:            Export the finished sequence
//
// Each step is logged with timing and produces a StepStatus entry in the
// result so callers can report granular progress.
func (e *Engine) AutoEdit(ctx context.Context, params *AutoEditParams) (*AutoEditResult, error) {
	if params == nil {
		return nil, fmt.Errorf("auto_edit: params must not be nil")
	}
	if params.AssetsDirectory == "" {
		return nil, fmt.Errorf("auto_edit: assets_directory must not be empty")
	}
	if params.ScriptText == "" && params.ScriptPath == "" {
		return nil, fmt.Errorf("auto_edit: either script_text or script_path must be provided")
	}

	pipelineStart := time.Now()
	result := &AutoEditResult{}

	e.logger.Info("auto_edit: starting pipeline",
		zap.String("assets_dir", params.AssetsDirectory),
		zap.Bool("has_script_text", params.ScriptText != ""),
		zap.String("script_path", params.ScriptPath),
		zap.String("script_format", params.ScriptFormat),
		zap.String("match_strategy", params.MatchStrategy),
		zap.String("output_name", params.OutputName),
	)

	// -----------------------------------------------------------------------
	// Step 1: Fan-out — scan assets + parse script in parallel
	// -----------------------------------------------------------------------
	var (
		scanResult   *ScanResult
		parsedScript *ParsedScript
	)

	step1Start := time.Now()
	g, gctx := errgroup.WithContext(ctx)

	// 1a. Rust engine: scan assets directory
	g.Go(func() error {
		e.logger.Info("auto_edit: step 1a — scanning assets",
			zap.String("dir", params.AssetsDirectory),
			zap.Bool("recursive", params.Recursive),
		)
		scanStart := time.Now()
		res, err := e.media.ScanAssets(gctx, params.AssetsDirectory, params.Recursive, params.Extensions)
		if err != nil {
			return fmt.Errorf("scan assets: %w", err)
		}
		e.logger.Info("auto_edit: step 1a — scan complete",
			zap.Uint32("media_files", res.MediaFilesFound),
			zap.Uint32("total_scanned", res.TotalFilesScanned),
			zap.Duration("elapsed", time.Since(scanStart)),
		)
		scanResult = res
		return nil
	})

	// 1b. Python: parse script
	g.Go(func() error {
		e.logger.Info("auto_edit: step 1b — parsing script",
			zap.String("format", params.ScriptFormat),
		)
		parseStart := time.Now()
		res, err := e.intel.ParseScript(gctx, params.ScriptText, params.ScriptPath, params.ScriptFormat)
		if err != nil {
			return fmt.Errorf("parse script: %w", err)
		}
		e.logger.Info("auto_edit: step 1b — parse complete",
			zap.Uint32("segments", res.Metadata.SegmentCount),
			zap.String("title", res.Metadata.Title),
			zap.Duration("elapsed", time.Since(parseStart)),
		)
		parsedScript = res
		return nil
	})

	if err := g.Wait(); err != nil {
		step1Duration := time.Since(step1Start)
		result.Steps = append(result.Steps, &StepStatus{
			Name:     "fan_out_scan_and_parse",
			Status:   "failed",
			Duration: step1Duration,
			Error:    err.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Error("auto_edit: step 1 failed", zap.Error(err))
		return result, fmt.Errorf("auto_edit step 1 (fan-out): %w", err)
	}

	result.ScanResult = scanResult
	result.ParsedScript = parsedScript
	result.Steps = append(result.Steps, &StepStatus{
		Name:     "fan_out_scan_and_parse",
		Status:   "completed",
		Duration: time.Since(step1Start),
		Detail: fmt.Sprintf("scanned %d media files, parsed %d script segments",
			scanResult.MediaFilesFound, parsedScript.Metadata.SegmentCount),
	})

	// Early exit: no media files found
	if scanResult.MediaFilesFound == 0 {
		noAssetsErr := fmt.Errorf("auto_edit: asset scan found 0 media files in %q — nothing to edit", params.AssetsDirectory)
		result.Steps = append(result.Steps, &StepStatus{
			Name:   "match_assets",
			Status: "skipped",
			Error:  noAssetsErr.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Warn("auto_edit: aborting — no media assets found",
			zap.String("dir", params.AssetsDirectory),
		)
		return result, noAssetsErr
	}

	// Early exit: no script segments parsed
	if len(parsedScript.Segments) == 0 {
		noSegmentsErr := fmt.Errorf("auto_edit: script parsing produced 0 segments — nothing to edit")
		result.Steps = append(result.Steps, &StepStatus{
			Name:   "match_assets",
			Status: "skipped",
			Error:  noSegmentsErr.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Warn("auto_edit: aborting — no script segments parsed")
		return result, noSegmentsErr
	}

	// -----------------------------------------------------------------------
	// Step 2: Match assets to script segments (Python)
	// -----------------------------------------------------------------------
	step2Start := time.Now()
	e.logger.Info("auto_edit: step 2 — matching assets to segments",
		zap.Int("segments", len(parsedScript.Segments)),
		zap.Int("assets", len(scanResult.Assets)),
		zap.String("strategy", params.MatchStrategy),
	)

	matchResult, err := e.intel.MatchAssets(ctx, parsedScript.Segments, scanResult.Assets, params.MatchStrategy)
	if err != nil {
		step2Duration := time.Since(step2Start)
		result.Steps = append(result.Steps, &StepStatus{
			Name:     "match_assets",
			Status:   "failed",
			Duration: step2Duration,
			Error:    err.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Error("auto_edit: step 2 failed", zap.Error(err))
		return result, fmt.Errorf("auto_edit step 2 (match assets): %w", err)
	}

	result.MatchResult = matchResult
	result.Steps = append(result.Steps, &StepStatus{
		Name:     "match_assets",
		Status:   "completed",
		Duration: time.Since(step2Start),
		Detail: fmt.Sprintf("matched %d segments, %d unmatched",
			len(matchResult.Matches), len(matchResult.Unmatched)),
	})

	e.logger.Info("auto_edit: step 2 complete",
		zap.Int("matches", len(matchResult.Matches)),
		zap.Int("unmatched", len(matchResult.Unmatched)),
		zap.Duration("elapsed", time.Since(step2Start)),
	)

	if len(matchResult.Matches) == 0 {
		noMatchErr := fmt.Errorf("auto_edit: asset matching produced 0 matches — cannot build EDL")
		result.Steps = append(result.Steps, &StepStatus{
			Name:   "generate_edl",
			Status: "skipped",
			Error:  noMatchErr.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Warn("auto_edit: aborting — no matches found")
		return result, noMatchErr
	}

	// -----------------------------------------------------------------------
	// Step 3: Generate EDL from matches (Python)
	// -----------------------------------------------------------------------
	step3Start := time.Now()
	e.logger.Info("auto_edit: step 3 — generating EDL",
		zap.Int("matches", len(matchResult.Matches)),
	)

	edlSettings := params.EDLSettings
	if edlSettings == nil {
		edlSettings = &EDLSettings{
			Resolution:                Resolution{Width: 1920, Height: 1080},
			FrameRate:                 29.97,
			DefaultTransition:         "cross_dissolve",
			DefaultTransitionDuration: 0.5,
			Pacing:                    PacingPresetModerate,
		}
		e.logger.Debug("auto_edit: using default EDL settings")
	}

	edl, err := e.intel.GenerateEDL(ctx, parsedScript.Segments, scanResult.Assets, edlSettings)
	if err != nil {
		step3Duration := time.Since(step3Start)
		result.Steps = append(result.Steps, &StepStatus{
			Name:     "generate_edl",
			Status:   "failed",
			Duration: step3Duration,
			Error:    err.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Error("auto_edit: step 3 failed", zap.Error(err))
		return result, fmt.Errorf("auto_edit step 3 (generate EDL): %w", err)
	}

	result.EDL = edl
	result.Steps = append(result.Steps, &StepStatus{
		Name:     "generate_edl",
		Status:   "completed",
		Duration: time.Since(step3Start),
		Detail:   fmt.Sprintf("EDL %q with %d entries", edl.Name, len(edl.Entries)),
	})

	e.logger.Info("auto_edit: step 3 complete",
		zap.String("edl_id", edl.ID),
		zap.String("edl_name", edl.Name),
		zap.Int("entries", len(edl.Entries)),
		zap.Duration("elapsed", time.Since(step3Start)),
	)

	// -----------------------------------------------------------------------
	// Step 4: Execute EDL in Premiere Pro (TypeScript bridge)
	// -----------------------------------------------------------------------
	step4Start := time.Now()
	e.logger.Info("auto_edit: step 4 — executing EDL in Premiere Pro",
		zap.Int("entries", len(edl.Entries)),
	)

	execResult, err := e.premiere.ExecuteEDL(ctx, edl)
	if err != nil {
		step4Duration := time.Since(step4Start)
		result.Steps = append(result.Steps, &StepStatus{
			Name:     "execute_edl",
			Status:   "failed",
			Duration: step4Duration,
			Error:    err.Error(),
		})
		result.TotalDuration = time.Since(pipelineStart)
		e.logger.Error("auto_edit: step 4 failed", zap.Error(err))
		return result, fmt.Errorf("auto_edit step 4 (execute EDL): %w", err)
	}

	result.ExecutionResult = execResult
	result.Steps = append(result.Steps, &StepStatus{
		Name:     "execute_edl",
		Status:   "completed",
		Duration: time.Since(step4Start),
		Detail: fmt.Sprintf("placed %d clips, added %d transitions, %d warnings",
			execResult.ClipsPlaced, execResult.TransitionsAdded, len(execResult.Warnings)),
	})

	e.logger.Info("auto_edit: step 4 complete",
		zap.String("sequence_id", execResult.SequenceID),
		zap.Uint32("clips_placed", execResult.ClipsPlaced),
		zap.Uint32("transitions_added", execResult.TransitionsAdded),
		zap.Int("errors", len(execResult.Errors)),
		zap.Int("warnings", len(execResult.Warnings)),
		zap.Duration("elapsed", time.Since(step4Start)),
	)

	// Log any non-fatal errors or warnings from EDL execution.
	if len(execResult.Errors) > 0 {
		e.logger.Warn("auto_edit: EDL execution reported errors",
			zap.Strings("errors", execResult.Errors),
		)
	}
	if len(execResult.Warnings) > 0 {
		e.logger.Warn("auto_edit: EDL execution reported warnings",
			zap.Strings("warnings", execResult.Warnings),
		)
	}

	// -----------------------------------------------------------------------
	// Step 5 (Optional): Export if output_name was provided
	// -----------------------------------------------------------------------
	if params.OutputName != "" {
		step5Start := time.Now()
		e.logger.Info("auto_edit: step 5 — exporting sequence",
			zap.String("output_name", params.OutputName),
			zap.Int("preset", int(params.ExportPreset)),
		)

		exportPreset := params.ExportPreset
		if exportPreset == ExportPresetUnspecified {
			exportPreset = ExportPresetH264_1080P
		}

		exportParams := &ExportParams{
			SequenceID: execResult.SequenceID,
			OutputPath: params.OutputName,
			Preset:     exportPreset,
		}

		exportRes, err := e.premiere.ExportSequence(ctx, exportParams)
		if err != nil {
			step5Duration := time.Since(step5Start)
			result.Steps = append(result.Steps, &StepStatus{
				Name:     "export",
				Status:   "failed",
				Duration: step5Duration,
				Error:    err.Error(),
			})
			// Export failure is not fatal to the overall pipeline — the
			// timeline was assembled successfully, so we return the result
			// with a partial success indication.
			result.TotalDuration = time.Since(pipelineStart)
			e.logger.Error("auto_edit: step 5 (export) failed — timeline was assembled successfully",
				zap.Error(err),
			)
			return result, fmt.Errorf("auto_edit step 5 (export): %w", err)
		}

		result.ExportResult = exportRes
		result.Steps = append(result.Steps, &StepStatus{
			Name:     "export",
			Status:   "completed",
			Duration: time.Since(step5Start),
			Detail:   fmt.Sprintf("export job %q started, output: %s", exportRes.JobID, exportRes.OutputPath),
		})

		e.logger.Info("auto_edit: step 5 complete",
			zap.String("job_id", exportRes.JobID),
			zap.String("status", exportRes.Status),
			zap.String("output", exportRes.OutputPath),
			zap.Duration("elapsed", time.Since(step5Start)),
		)
	} else {
		result.Steps = append(result.Steps, &StepStatus{
			Name:   "export",
			Status: "skipped",
			Detail: "no output_name provided",
		})
		e.logger.Debug("auto_edit: step 5 — export skipped (no output_name)")
	}

	// -----------------------------------------------------------------------
	// Done
	// -----------------------------------------------------------------------
	result.TotalDuration = time.Since(pipelineStart)

	e.logger.Info("auto_edit: pipeline complete",
		zap.Duration("total_duration", result.TotalDuration),
		zap.Int("steps_completed", countStepsByStatus(result.Steps, "completed")),
		zap.Int("steps_skipped", countStepsByStatus(result.Steps, "skipped")),
		zap.Int("steps_failed", countStepsByStatus(result.Steps, "failed")),
	)

	return result, nil
}

// countStepsByStatus counts how many steps have a given status string.
func countStepsByStatus(steps []*StepStatus, status string) int {
	n := 0
	for _, s := range steps {
		if s.Status == status {
			n++
		}
	}
	return n
}
