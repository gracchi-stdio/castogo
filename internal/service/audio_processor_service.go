package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type ProcessingOptions struct {
	TargetLUFS    float64 // loudness target (-16 for podcasts)
	TargetBitrate string  // "192k"
	TargetSample  int     // 44100
	Threads       int     // 4
}

func DefaultProcessingOptions() ProcessingOptions {
	return ProcessingOptions{
		TargetLUFS:    -16,
		TargetBitrate: "192k",
		TargetSample:  44100,
		Threads:       4,
	}
}

type ProcessingResult struct {
	OutputPath string
	Duration   float64
	FileSize   int64
	Bitrate    int
	SampleRate int
	Channels   int
}

type AudioProcessor interface {
	Process(ctx context.Context, inputPath string, opts ProcessingOptions) (*ProcessingResult, error)
}

type FFmpegProcessor struct{}

func NewFFmpegProcessor() *FFmpegProcessor {
	return &FFmpegProcessor{}
}

func (p *FFmpegProcessor) Process(ctx context.Context, inputPath string, opts ProcessingOptions) (*ProcessingResult, error) {
	outputPath, err := createTempFile("processed-*.mp3")
	if err != nil {
		return nil, fmt.Errorf("create temp output file: %w", err)
	}

	args := []string{
		"-i", inputPath,
		"-af", fmt.Sprintf("loudnorm=I=%.0f:TP=-1:LRA=11", opts.TargetLUFS),
		"-c:a", "libmp3lame",
		"-b:a", opts.TargetBitrate,
		"-ar", strconv.Itoa(opts.TargetSample),
		"-threads", strconv.Itoa(opts.Threads),
		"-y", outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outputPath)
		return nil, fmt.Errorf("ffmpeg failed: %w\n%s", err, string(output))
	}

	result, err := parseFFmpegStats(output)
	if err == nil {
		result.OutputPath = outputPath
	}

	// get file size regardless of stats parsing
	info, _ := os.Stat(outputPath)
	if info != nil {
		result.FileSize = info.Size()
	}

	return result, nil
}

func createTempFile(pattern string) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

func parseFFmpegStats(output []byte) (*ProcessingResult, error) {
	result := &ProcessingResult{}
	lines := string(output)

	// FFmpeg prints stats like: "size=    1234kB time=00:01:23.45 bitrate= 192.0kbits/s"
	for _, line := range strings.Split(lines, "\n") {
		if strings.Contains(line, "size=") && strings.Contains(line, "time=") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "time=") {
					t := strings.TrimPrefix(part, "time=")
					result.Duration = parseDuration(t)
				}
				if strings.HasPrefix(part, "bitrate=") {
					b := strings.TrimPrefix(part, "bitrate=")
					b = strings.TrimSuffix(b, "kbits/s")
					if v, err := strconv.Atoi(strings.TrimSpace(b)); err == nil {
						result.Bitrate = v
					}
				}
			}
		}
	}

	return result, nil
}

// parseDuration converts FFmpeg time format "HH:MM:SS.ms" to seconds
func parseDuration(s string) float64 {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)
	return hours*3600 + minutes*60 + seconds
}

// TempFile manages a temporary file with automatic cleanup
type TempFile struct {
	Path string
}

func NewTempFile(input io.Reader, ext string) (*TempFile, error) {
	f, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(f, input); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	f.Close()

	return &TempFile{Path: f.Name()}, nil
}

func (t *TempFile) Cleanup() {
	os.Remove(t.Path)
}
