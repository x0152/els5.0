package ports

import "context"

type MediaStream struct {
	Index int
	Lang  string
	Title string
}

type MediaProbe struct {
	DurationMs int
	VideoCodec string
	Audio      []MediaStream
	Subs       []MediaStream
}

type Transcoder interface {
	Probe(ctx context.Context, srcPath string) (MediaProbe, error)
	ExtractSubtitleSRT(ctx context.Context, srcPath string, streamIndex int) ([]byte, error)
	ExtractThumbnail(ctx context.Context, srcPath, outPath string, atSeconds float64) error
	BuildAudioVariant(ctx context.Context, srcPath string, audioStreamIndex int, copyVideo bool, outPath string) error
}
