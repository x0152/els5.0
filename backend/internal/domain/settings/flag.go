package settings

import "context"

const FlagEventProcessing = "event_processing"

type FlagRepository interface {
	GetFlag(ctx context.Context, key string) (bool, error)
	SetFlag(ctx context.Context, key string, value bool) error
}
