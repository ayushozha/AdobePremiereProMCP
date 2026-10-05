package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSourcePNGRejectsInvalidSources(t *testing.T) {
	for _, src := range []frameSourceInfo{{}, {MediaPath: "https://example.com/video", Seconds: 1}, {MediaPath: "/missing", Seconds: math.NaN()}, {MediaPath: "/missing", Seconds: -1}} {
		if _, err := extractSourcePNG(context.Background(), src); err == nil {
			t.Fatalf("accepted invalid source: %+v", src)
		}
	}
}
func TestExtractSourcePNGUsesSourceOffset(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	path := filepath.Join(t.TempDir(), "two colors.mkv")
	cmd := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=red:s=16x16:r=10:d=1", "-f", "lavfi", "-i", "color=blue:s=16x16:r=10:d=1", "-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0", "-c:v", "ffv1", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, out)
	}
	data, err := extractSourcePNG(context.Background(), frameSourceInfo{MediaPath: path, Seconds: 1.4})
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	r, _, b, _ := img.At(8, 8).RGBA()
	if b < 50000 || r > 5000 {
		t.Fatalf("wrong source frame: red=%d blue=%d", r, b)
	}
	if _, err := extractSourcePNG(context.Background(), frameSourceInfo{MediaPath: path, Seconds: 4}); err == nil {
		t.Fatal("accepted empty frame after EOF")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := extractSourcePNG(ctx, frameSourceInfo{MediaPath: path, Seconds: 0}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
func TestWriteSourceFramePreservesExistingFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "frame.png")
	if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeSourceFrame(path, []byte("replacement")); err == nil {
		t.Fatal("overwrote existing file")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "keep" {
		t.Fatal("existing output changed")
	}
	link := filepath.Join(dir, "link.png")
	if err := os.Symlink(path, link); err == nil {
		if err := writeSourceFrame(link, []byte("replacement")); err == nil {
			t.Fatal("followed symlink")
		}
	}
	if err := writeSourceFrame(filepath.Join(dir, "new.png"), []byte("png")); err != nil {
		t.Fatal(err)
	}
}

type failingFramePremiere struct{ mockPremiereClient }

func (p *failingFramePremiere) EvalCommand(_ context.Context, command, args string) (string, error) {
	if command == "exportFrame" {
		return "", errors.New("native unavailable")
	}
	return "", errors.New("source refused")
}
func TestExportFrameFallbackRequiresOptIn(t *testing.T) {
	e := &Engine{premiere: &failingFramePremiere{}}
	_, err := e.ExportFrame(context.Background(), &ExportFrameParams{OutputPath: filepath.Join(t.TempDir(), "frame.png")})
	if err == nil || !strings.Contains(err.Error(), "native unavailable") {
		t.Fatalf("native failure hidden: %v", err)
	}
	_, err = e.ExportFrame(context.Background(), &ExportFrameParams{OutputPath: filepath.Join(t.TempDir(), "frame.png"), AllowSourceFallback: true})
	if err == nil || !strings.Contains(err.Error(), "source refused") {
		t.Fatalf("source refusal not surfaced: %v", err)
	}
}

func TestExtractSourcePNGMissingFFmpeg(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.mkv")
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := extractSourcePNG(context.Background(), frameSourceInfo{MediaPath: path}); err == nil || !strings.Contains(err.Error(), "executable file not found") {
		t.Fatalf("missing ffmpeg: %v", err)
	}
}
func TestExtractSourcePNGReportsFFmpegFailure(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	path := filepath.Join(t.TempDir(), "broken.mkv")
	os.WriteFile(path, []byte("not a video"), 0600)
	if _, err := extractSourcePNG(context.Background(), frameSourceInfo{MediaPath: path}); err == nil || !strings.Contains(err.Error(), "ffmpeg source frame extraction") {
		t.Fatalf("decode failure hidden: %v", err)
	}
}
func TestSourceFallbackRejectsIncompleteMetadata(t *testing.T) {
	e := &Engine{premiere: &mockPremiereClient{evalResult: `{"mediaPath":"/missing"}`}}
	_, err := e.exportSourceFrameFallback(context.Background(), filepath.Join(t.TempDir(), "frame.png"))
	if err == nil || !strings.Contains(err.Error(), "missing source time") {
		t.Fatalf("incomplete mapping accepted: %v", err)
	}
}

type sourceFramePremiere struct {
	mockPremiereClient
	source string
}

func (p *sourceFramePremiere) EvalCommand(_ context.Context, command, args string) (string, error) {
	if command == "exportFrame" {
		return "", errors.New("native unavailable")
	}
	payload, _ := json.Marshal(frameSourceInfo{MediaPath: p.source, Seconds: 0, TimelineSeconds: 3})
	return string(payload), nil
}
func TestExportFrameSourceFallbackReportsProvenance(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "source.png")
	f, err := os.Create(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	f.Close()
	e := &Engine{premiere: &sourceFramePremiere{source: source}}
	result, err := e.ExportFrame(context.Background(), &ExportFrameParams{OutputPath: filepath.Join(dir, "output.png"), Format: "PNG", AllowSourceFallback: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "source_file_frame" || result.SourcePath != source || result.TimelineSeconds != 3 || !strings.Contains(result.Warning, "Not a timeline preview") {
		t.Fatalf("misleading provenance: %+v", result)
	}
	if _, err := os.Stat(result.OutputPath); err != nil {
		t.Fatal(err)
	}
}
