package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultMaxDuration  = 180 * time.Second
	defaultMaxBytes     = 60 * 1024 * 1024 // 60 MB
	defaultFrames       = 8
	defaultFrameMaxSide = 768
	defaultAudioMaxSec  = 90
	maxAudioBytes       = 1024 * 1024 // 1 MB
	maxTitleRunes       = 300
	maxDescRunes        = 2000
	maxSubtitleRunes    = 4000
)

// NewFromPath searches PATH for yt-dlp and ffmpeg and populates standard defaults.
func NewFromPath() *Fetcher {
	ytDlp, _ := exec.LookPath("yt-dlp")
	ffmpeg, _ := exec.LookPath("ffmpeg")
	return &Fetcher{
		YtDlp:        ytDlp,
		FFmpeg:       ffmpeg,
		MaxDuration:  defaultMaxDuration,
		MaxBytes:     defaultMaxBytes,
		Frames:       defaultFrames,
		FrameMaxSide: defaultFrameMaxSide,
		AudioMaxSec:  defaultAudioMaxSec,
	}
}

// Enabled returns true if both yt-dlp and ffmpeg paths are configured.
func (f *Fetcher) Enabled() bool {
	return f != nil && f.YtDlp != "" && f.FFmpeg != ""
}

// ytdlpInfoJSON matches standard yt-dlp metadata fields.
type ytdlpInfoJSON struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Uploader    string  `json:"uploader"`
	UploadDate  string  `json:"upload_date"`
	Duration    float64 `json:"duration"`
	ViewCount   int64   `json:"view_count"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`

	// Platform AI-generated flags
	AILabel       bool `json:"is_ai_generated"`
	AIGenerated   bool `json:"ai_generated"`
	TikTokAILabel bool `json:"is_synthetic"`
	TikTokAITag   bool `json:"aigc_label_type"`
}

// Fetch downloads the social video using yt-dlp, extracts metadata, frames, and audio.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*Clip, error) {
	if !f.Enabled() {
		return nil, ErrNoTools
	}

	platform, ok := Platform(rawURL)
	if !ok {
		return nil, ErrUnsupported
	}

	extractor := ExtractorForPlatform(platform)
	if extractor == "" {
		return nil, ErrUnsupported
	}

	start := time.Now()

	// Create temp directory for this fetch
	tmpDir, err := os.MkdirTemp("", "amba-video-*")
	if err != nil {
		return nil, fmt.Errorf("video: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	maxDurSec := int(f.MaxDuration.Seconds())
	if maxDurSec <= 0 {
		maxDurSec = int(defaultMaxDuration.Seconds())
	}
	maxBytes := f.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}

	// Prepare output template
	outputTemplate := filepath.Join(tmpDir, "v.%(ext)s")

	// yt-dlp arguments (strict allowlist, no generic extractor, -i ignores subtitle 429s)
	args := []string{
		"--ignore-config",
		"--no-playlist",
		"--no-warnings",
		"--no-progress",
		"--restrict-filenames",
		"--no-part",
		"--no-mtime",
		"--socket-timeout", "10",
		"--retries", "1",
		"-i",
		"--use-extractors", extractor,
		"--max-filesize", fmt.Sprintf("%d", maxBytes),
		"--match-filter", fmt.Sprintf("duration<=%d", maxDurSec),
		"-f", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[height<=720][ext=mp4]/best[height<=720]/best",
		"--merge-output-format", "mp4",
		"--write-info-json",
		"--write-subs",
		"--write-auto-subs",
		"--sub-langs", "id.*,en.*",
		"--sub-format", "vtt",
		"-o", outputTemplate,
		"--",
		rawURL,
	}

	var cmd *exec.Cmd
	if f.cmdRunner != nil {
		cmd = f.cmdRunner(ctx, f.YtDlp, args...)
	} else {
		cmd = exec.CommandContext(ctx, f.YtDlp, args...)
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	cmdErr := cmd.Run()

	// Locate info.json
	infoPath := filepath.Join(tmpDir, "v.info.json")
	infoData, readInfoErr := os.ReadFile(infoPath)
	if readInfoErr != nil {
		matches, _ := filepath.Glob(filepath.Join(tmpDir, "*.info.json"))
		if len(matches) > 0 {
			infoData, readInfoErr = os.ReadFile(matches[0])
		}
	}

	// Locate video file (v.mp4 or other extension)
	videoFile := filepath.Join(tmpDir, "v.mp4")
	if _, statErr := os.Stat(videoFile); os.IsNotExist(statErr) {
		candidates, _ := filepath.Glob(filepath.Join(tmpDir, "v.*"))
		found := ""
		for _, c := range candidates {
			ext := strings.ToLower(filepath.Ext(c))
			if ext != ".json" && ext != ".vtt" && ext != ".srt" {
				found = c
				break
			}
		}
		if found != "" {
			videoFile = found
		} else {
			videoFile = ""
		}
	}

	// If cmdErr returned an error and we did not get a valid video file or info JSON, map error
	if cmdErr != nil && (readInfoErr != nil || videoFile == "") {
		mappedErr := mapYtDlpError(ctx, cmdErr, stderrBuf.String())
		slog.Warn("video fetch failed",
			"platform", platform,
			"duration_ms", time.Since(start).Milliseconds(),
			"error_category", mappedErr.Error(),
		)
		return nil, mappedErr
	}

	if readInfoErr != nil || videoFile == "" {
		mappedErr := mapYtDlpError(ctx, cmdErr, stderrBuf.String())
		return nil, mappedErr
	}

	var info ytdlpInfoJSON
	if err := json.Unmarshal(infoData, &info); err != nil {
		return nil, ErrDownload
	}

	aiLabel := detectAILabel(infoData, &info)

	// Verify duration limit from metadata
	if f.MaxDuration > 0 && time.Duration(info.Duration*float64(time.Second)) > f.MaxDuration {
		return nil, ErrTooLong
	}

	// Read subtitles if available
	subtitles := loadSubtitles(tmpDir)

	// Extract keyframes using ffmpeg
	framesCount := f.Frames
	if framesCount <= 0 {
		framesCount = defaultFrames
	}
	frameMaxSide := f.FrameMaxSide
	if frameMaxSide <= 0 {
		frameMaxSide = defaultFrameMaxSide
	}

	frames, err := f.extractFrames(ctx, videoFile, info.Duration, framesCount, frameMaxSide, tmpDir)
	if err != nil {
		slog.Warn("video keyframe extraction failed", "platform", platform, "error", err.Error())
		if len(frames) == 0 {
			return nil, ErrDownload
		}
	}

	// Extract audio (optional)
	audioMaxSec := f.AudioMaxSec
	if audioMaxSec <= 0 {
		audioMaxSec = defaultAudioMaxSec
	}
	audioBytes := f.extractAudio(ctx, videoFile, audioMaxSec, tmpDir)

	clip := &Clip{
		Platform:    platform,
		ID:          info.ID,
		Title:       truncateRunes(info.Title, maxTitleRunes),
		Description: truncateRunes(info.Description, maxDescRunes),
		Uploader:    info.Uploader,
		UploadDate:  info.UploadDate,
		DurationSec: info.Duration,
		ViewCount:   info.ViewCount,
		Width:       info.Width,
		Height:      info.Height,
		AILabel:     aiLabel,
		Subtitles:   subtitles,
		Frames:      frames,
		Audio:       audioBytes,
	}

	slog.Info("video fetched",
		"platform", platform,
		"duration_sec", info.Duration,
		"frames", len(frames),
		"audio_bytes", len(audioBytes),
		"elapsed_ms", time.Since(start).Milliseconds(),
	)

	return clip, nil
}

func detectAILabel(rawJSON []byte, info *ytdlpInfoJSON) bool {
	if info.AILabel || info.AIGenerated || info.TikTokAILabel || info.TikTokAITag {
		return true
	}
	lowerJSON := strings.ToLower(string(rawJSON))
	patterns := []string{
		`"aigc_label_type":1`,
		`"is_ai_generated":true`,
		`"is_synthetic":true`,
		`"aigc_info":`,
	}
	for _, p := range patterns {
		if strings.Contains(lowerJSON, p) {
			return true
		}
	}
	return false
}

func loadSubtitles(dir string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, "*.vtt"))
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err == nil && len(data) > 0 {
			cleaned := cleanVTT(data, maxSubtitleRunes)
			if cleaned != "" {
				return cleaned
			}
		}
	}
	return ""
}

func (f *Fetcher) extractFrames(ctx context.Context, videoFile string, duration float64, count, maxSide int, tmpDir string) ([]Frame, error) {
	if count <= 0 {
		return nil, nil
	}

	if duration <= 0 {
		duration = 10.0
	}

	var timestamps []float64
	if count == 1 {
		timestamps = []float64{duration * 0.5}
	} else {
		start := duration * 0.05
		end := duration * 0.95
		step := (end - start) / float64(count-1)
		for i := 0; i < count; i++ {
			ts := start + float64(i)*step
			if ts > duration {
				ts = duration
			}
			timestamps = append(timestamps, math.Round(ts*100)/100)
		}
	}

	scaleFilter := fmt.Sprintf("scale='if(gt(iw,ih),min(%d,iw),-2)':'if(gt(iw,ih),-2,min(%d,ih))'", maxSide, maxSide)

	var frames []Frame
	for i, ts := range timestamps {
		outFramePath := filepath.Join(tmpDir, fmt.Sprintf("frame_%02d.jpg", i))
		args := []string{
			"-ss", fmt.Sprintf("%.3f", ts),
			"-i", videoFile,
			"-vf", scaleFilter,
			"-vframes", "1",
			"-q:v", "3", // ~quality 80
			"-y",
			outFramePath,
		}

		var cmd *exec.Cmd
		if f.cmdRunner != nil {
			cmd = f.cmdRunner(ctx, f.FFmpeg, args...)
		} else {
			cmd = exec.CommandContext(ctx, f.FFmpeg, args...)
		}

		if err := cmd.Run(); err != nil {
			continue
		}

		jpegData, err := os.ReadFile(outFramePath)
		if err == nil && len(jpegData) > 0 {
			frames = append(frames, Frame{
				AtSec: ts,
				JPEG:  jpegData,
			})
		}
	}

	return frames, nil
}

func (f *Fetcher) extractAudio(ctx context.Context, videoFile string, maxSec int, tmpDir string) []byte {
	outAudioPath := filepath.Join(tmpDir, "audio.mp3")
	args := []string{
		"-i", videoFile,
		"-t", fmt.Sprintf("%d", maxSec),
		"-vn",
		"-ac", "1", // mono
		"-ar", "16000", // 16 kHz
		"-b:a", "32k",
		"-y",
		outAudioPath,
	}

	var cmd *exec.Cmd
	if f.cmdRunner != nil {
		cmd = f.cmdRunner(ctx, f.FFmpeg, args...)
	} else {
		cmd = exec.CommandContext(ctx, f.FFmpeg, args...)
	}

	if err := cmd.Run(); err != nil {
		return nil
	}

	audioData, err := os.ReadFile(outAudioPath)
	if err != nil || len(audioData) == 0 {
		return nil
	}

	if len(audioData) > maxAudioBytes {
		return audioData[:maxAudioBytes]
	}
	return audioData
}

func mapYtDlpError(ctx context.Context, cmdErr error, stderr string) error {
	if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
		return ErrTimeout
	}

	lower := strings.ToLower(stderr)
	switch {
	case strings.Contains(lower, "private video") ||
		strings.Contains(lower, "login required") ||
		strings.Contains(lower, "sign in") ||
		strings.Contains(lower, "requires authentication") ||
		strings.Contains(lower, "this video is only available to registered users") ||
		strings.Contains(lower, "video unavailable"):
		return ErrPrivate

	case strings.Contains(lower, "does not pass filter") ||
		(strings.Contains(lower, "duration") && strings.Contains(lower, "filter")):
		return ErrTooLong

	case strings.Contains(lower, "larger than max-filesize") ||
		strings.Contains(lower, "file is larger than"):
		return ErrTooBig

	case strings.Contains(lower, "unsupported url") ||
		strings.Contains(lower, "no suitable extractor") ||
		strings.Contains(lower, "not found"):
		return ErrUnsupported

	case strings.Contains(lower, "timed out") || strings.Contains(lower, "timeout"):
		return ErrTimeout

	default:
		return ErrDownload
	}
}
