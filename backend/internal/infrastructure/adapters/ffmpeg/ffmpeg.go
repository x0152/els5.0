package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/els/backend/internal/domain/shared/ports"
)

type Transcoder struct{}

func New() *Transcoder { return &Transcoder{} }

var textSubCodecs = map[string]bool{"subrip": true, "ass": true, "ssa": true, "mov_text": true, "webvtt": true}

type probeStream struct {
	Index     int    `json:"index"`
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Tags      struct {
		Language string `json:"language"`
		Title    string `json:"title"`
	} `json:"tags"`
}

type probeOutput struct {
	Streams []probeStream `json:"streams"`
	Format  struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func (t *Transcoder) Probe(ctx context.Context, srcPath string) (ports.MediaProbe, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", // #nosec G204 -- fixed binary with argv, no shell expansion.
		"-show_entries", "stream=index,codec_type,codec_name:stream_tags=language,title:format=duration",
		"-of", "json", srcPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return ports.MediaProbe{}, fmt.Errorf("ffprobe: %s", detail(stderr, err))
	}

	var out probeOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return ports.MediaProbe{}, fmt.Errorf("parse ffprobe output: %w", err)
	}

	res := ports.MediaProbe{}
	if secs, err := strconv.ParseFloat(strings.TrimSpace(out.Format.Duration), 64); err == nil {
		res.DurationMs = int(secs * 1000)
	}
	for _, s := range out.Streams {
		info := ports.MediaStream{Index: s.Index, Lang: s.Tags.Language, Title: s.Tags.Title}
		switch s.CodecType {
		case "video":
			if res.VideoCodec == "" {
				res.VideoCodec = strings.ToLower(s.CodecName)
			}
		case "audio":
			res.Audio = append(res.Audio, info)
		case "subtitle":
			if textSubCodecs[strings.ToLower(s.CodecName)] {
				res.Subs = append(res.Subs, info)
			}
		}
	}
	return res, nil
}

func (t *Transcoder) ExtractSubtitleSRT(ctx context.Context, srcPath string, streamIndex int) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-i", srcPath, // #nosec G204 -- fixed binary with argv, no shell expansion.
		"-map", fmt.Sprintf("0:%d", streamIndex), "-f", "srt", "pipe:1")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg subtitle extract: %s", detail(stderr, err))
	}
	return stdout.Bytes(), nil
}

func (t *Transcoder) ExtractThumbnail(ctx context.Context, srcPath, outPath string, atSeconds float64) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-v", "error", // #nosec G204 -- fixed binary with argv, no shell expansion.
		"-ss", strconv.FormatFloat(atSeconds, 'f', 3, 64), "-i", srcPath,
		"-vf", "scale=640:-1", "-frames:v", "1", outPath)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg thumbnail: %s", detail(stderr, err))
	}
	return nil
}

func (t *Transcoder) ExtractFrameJPEG(ctx context.Context, src string, atSeconds float64) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", "-v", "error", // #nosec G204 -- fixed binary with argv, no shell expansion.
		"-ss", strconv.FormatFloat(atSeconds, 'f', 3, 64), "-i", src,
		"-frames:v", "1", "-vf", "scale=768:-1", "-f", "image2", "-c:v", "mjpeg", "pipe:1")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg frame extract: %s", detail(stderr, err))
	}
	return stdout.Bytes(), nil
}

func (t *Transcoder) BuildAudioVariant(ctx context.Context, srcPath string, audioStreamIndex int, copyVideo bool, outPath string) error {
	args := []string{"-y", "-v", "error", "-i", srcPath, "-map", "0:v:0", "-map", fmt.Sprintf("0:%d", audioStreamIndex)}
	if copyVideo {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", "22", "-pix_fmt", "yuv420p")
	}
	args = append(args, "-c:a", "aac", "-b:a", "192k", "-ac", "2", "-movflags", "+faststart", "-f", "mp4", outPath)

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", args...) // #nosec G204 -- fixed binary with argv, no shell expansion.
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg build variant: %s", detail(stderr, err))
	}
	return nil
}

func detail(stderr bytes.Buffer, err error) string {
	if d := strings.TrimSpace(stderr.String()); d != "" {
		return d
	}
	return err.Error()
}
