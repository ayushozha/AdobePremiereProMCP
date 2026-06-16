package orchestrator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type frameSourceInfo struct {
	MediaPath string  `json:"mediaPath"`
	Seconds   float64 `json:"seconds"`
}

// exportFrameViaFFmpeg extracts a single frame from the VOD under the playhead.
// Used when Premiere 24.x QE exportFramePNG throws despite the method existing.
func (e *Engine) exportFrameViaFFmpeg(ctx context.Context, outputPath string, seconds float64) error {
	src, err := e.getFrameSourceAtPlayhead(ctx, seconds)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	args := []string{
		"-y",
		"-ss", strconv.FormatFloat(src.Seconds, 'f', 3, 64),
		"-i", src.MediaPath,
		"-frames:v", "1",
		"-update", "1",
		outputPath,
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		return fmt.Errorf("ffmpeg frame extract: %w (%s)", runErr, string(out))
	}
	if _, statErr := os.Stat(outputPath); statErr != nil {
		return fmt.Errorf("ffmpeg did not create output file: %w", statErr)
	}
	return nil
}

func (e *Engine) captureFrameBase64ViaFFmpeg(ctx context.Context, seconds float64) (*FrameCaptureResult, error) {
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("mcp_frame_%d.png", os.Getpid()))
	defer os.Remove(tempFile)

	if err := e.exportFrameViaFFmpeg(ctx, tempFile, seconds); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(tempFile)
	if err != nil {
		return nil, fmt.Errorf("read ffmpeg frame: %w", err)
	}
	return &FrameCaptureResult{
		ImageBase64: base64.StdEncoding.EncodeToString(data),
		Format:      "png",
		Timecode:    seconds,
	}, nil
}

func (e *Engine) getFrameSourceAtPlayhead(ctx context.Context, seconds float64) (*frameSourceInfo, error) {
	script := fmt.Sprintf(`(function(){
		var seq=app.project.activeSequence;
		if(!seq) return JSON.stringify({error:"no active sequence"});
		var pos=seq.getPlayerPosition();
		var t=%f;
		if(t>0){var tm=new Time();tm.seconds=t;seq.setPlayerPosition(tm.ticks);pos=seq.getPlayerPosition();}
		var sec=pos.seconds;
		for(var ti=0;ti<seq.videoTracks.numTracks;ti++){
			var tr=seq.videoTracks[ti];
			for(var ci=0;ci<tr.clips.numItems;ci++){
				var c=tr.clips[ci];
				if(c.start.seconds<=sec&&c.end.seconds>=sec){
					var p=c.projectItem?c.projectItem.getMediaPath():"";
					if(p){return JSON.stringify({mediaPath:p,seconds:sec});}
				}
			}
		}
		return JSON.stringify({error:"no media clip at playhead"});
	})()`, seconds)

	argsJSON, _ := json.Marshal(map[string]any{
		"script":   script,
		"validate": false,
	})
	raw, err := e.premiere.EvalCommand(ctx, "executeSecureScript", string(argsJSON))
	if err != nil {
		return nil, fmt.Errorf("get frame source: %w", err)
	}

	inner := unwrapEvalResult(raw)

	var info frameSourceInfo
	if err := json.Unmarshal([]byte(inner), &info); err != nil {
		return nil, fmt.Errorf("parse frame source %q: %w", inner, err)
	}
	if info.MediaPath == "" {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal([]byte(inner), &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("empty media path at playhead")
	}
	if info.Seconds <= 0 && seconds > 0 {
		info.Seconds = seconds
	}
	return &info, nil
}

// unwrapEvalResult peels MCP/CEP JSON wrappers until we reach the inner payload string.
func unwrapEvalResult(raw string) string {
	inner := raw
	for depth := 0; depth < 4; depth++ {
		var wrapper struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
			Result  string          `json:"result"`
			Error   string          `json:"error"`
		}
		if err := json.Unmarshal([]byte(inner), &wrapper); err != nil {
			break
		}
		if wrapper.Error != "" && !wrapper.Success {
			return inner
		}
		if wrapper.Result != "" {
			inner = wrapper.Result
			continue
		}
		if len(wrapper.Data) > 0 {
			var dataObj struct {
				Result string `json:"result"`
			}
			if json.Unmarshal(wrapper.Data, &dataObj) == nil && dataObj.Result != "" {
				inner = dataObj.Result
				continue
			}
			inner = string(wrapper.Data)
			continue
		}
		break
	}
	return inner
}
