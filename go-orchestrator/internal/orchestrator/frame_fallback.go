package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type frameSourceInfo struct {
	MediaPath       string  `json:"mediaPath"`
	Seconds         float64 `json:"seconds"`
	TimelineSeconds float64 `json:"timelineSeconds"`
}

const sourceFrameWarning = "Raw source-file frame only; excludes Premiere effects, color, overlays, transitions, reframing and timeline composition. Not a timeline preview."

// extractSourcePNG never writes to a caller-supplied path. A private directory
// prevents concurrent requests and pre-existing temporary files from colliding.
func extractSourcePNG(ctx context.Context, src frameSourceInfo) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(src.MediaPath) || math.IsNaN(src.Seconds) || math.IsInf(src.Seconds, 0) || src.Seconds < 0 {
		return nil, fmt.Errorf("invalid local source path or source time")
	}
	info, err := os.Stat(src.MediaPath)
	if err != nil {
		return nil, fmt.Errorf("source file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source must be a regular local file")
	}
	dir, err := os.MkdirTemp("", "premiere-source-frame-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "frame.png")
	cmd := exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-v", "error", "-protocol_whitelist", "file,pipe", "-i", src.MediaPath, "-ss", strconv.FormatFloat(src.Seconds, 'f', 9, 64), "-frames:v", "1", "-update", "1", "-f", "image2", path)
	// Bound error output so a malformed source cannot fill process memory.
	var stderr boundedFrameLog
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("ffmpeg source frame extraction: %w: %s", err, stderr.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ffmpeg did not produce a source PNG: %w", err)
	}
	if _, err := png.DecodeConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("ffmpeg produced invalid PNG: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

type boundedFrameLog struct{ bytes.Buffer }

func (b *boundedFrameLog) Write(p []byte) (int, error) {
	n := len(p)
	remaining := 8192 - b.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = b.Buffer.Write(p)
	}
	return n, nil
}

// Exclusive creation preserves existing files, including symbolic links.
func writeSourceFrame(path string, data []byte) error {
	if !filepath.IsAbs(path) || filepath.Ext(path) != ".png" {
		return fmt.Errorf("source frame output must be an absolute .png path")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create source frame output (existing files are never overwritten): %w", err)
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	if err = f.Close(); err != nil {
		os.Remove(path)
		return err
	}
	return nil
}

func (e *Engine) exportSourceFrameFallback(ctx context.Context, output string) (*GenericExportResult, error) {
	raw, err := e.premiere.EvalCommand(ctx, "getFrameSourceAtTime", "{}")
	if err != nil {
		return nil, fmt.Errorf("resolve source frame: %w", err)
	}
	var envelope struct {
		Success *bool           `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   string          `json:"error"`
	}
	if err = json.Unmarshal([]byte(raw), &envelope); err != nil {
		return nil, fmt.Errorf("parse source frame: %w", err)
	}
	if envelope.Success != nil {
		if !*envelope.Success {
			return nil, fmt.Errorf("resolve source frame: %s", envelope.Error)
		}
		raw = string(envelope.Data)
	}
	var mapping struct {
		Seconds         *float64 `json:"seconds"`
		TimelineSeconds *float64 `json:"timelineSeconds"`
	}
	if err = json.Unmarshal([]byte(raw), &mapping); err != nil {
		return nil, err
	}
	if mapping.Seconds == nil || mapping.TimelineSeconds == nil {
		return nil, fmt.Errorf("missing source time or timeline time; cannot safely map frame")
	}
	var src frameSourceInfo
	if err = json.Unmarshal([]byte(raw), &src); err != nil {
		return nil, err
	}
	if math.IsNaN(src.TimelineSeconds) || math.IsInf(src.TimelineSeconds, 0) || src.TimelineSeconds < 0 {
		return nil, fmt.Errorf("invalid timeline time")
	}
	data, err := extractSourcePNG(ctx, src)
	if err != nil {
		return nil, err
	}
	if err = writeSourceFrame(output, data); err != nil {
		return nil, err
	}
	return &GenericExportResult{Status: "source_file_frame", OutputPath: output, Warning: sourceFrameWarning, SourcePath: src.MediaPath, SourceSeconds: src.Seconds, TimelineSeconds: src.TimelineSeconds}, nil
}
