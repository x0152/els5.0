package ports

import "context"

type ImageAspect string

const (
	ImageAspectSquare    ImageAspect = "square"
	ImageAspectLandscape ImageAspect = "landscape"
	ImageAspectPortrait  ImageAspect = "portrait"
	ImageAspectWide      ImageAspect = "wide"
)

type ImageReference struct {
	Data        []byte
	ContentType string
}

type ImageOptions struct {
	Aspect     ImageAspect
	Size       string
	References []ImageReference
}

type ImageGenerator interface {
	IsAvailable() bool
	GenerateImageBytes(ctx context.Context, prompt string, opts *ImageOptions) ([]byte, error)
}
