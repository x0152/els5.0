package quest

import "context"

type MissionCatalogItem struct {
	Mission *CustomMission
	Started bool
}

type MissionRepository interface {
	Insert(ctx context.Context, userID string, mission *CustomMission) error
	Fork(ctx context.Context, userID, authorID string, mission *CustomMission) error
	Save(ctx context.Context, userID string, mission *CustomMission) error
	Update(ctx context.Context, userID, id string, mutate func(*CustomMission) error) error
	GetByID(ctx context.Context, userID, id string) (*CustomMission, error)
	GetOrigin(ctx context.Context, id string) (*CustomMission, string, error)
	GetAllByID(ctx context.Context, id string) ([]*CustomMission, error)
	List(ctx context.Context, userID string) ([]MissionCatalogItem, error)
	Delete(ctx context.Context, id string) error
}

type ProfileRepository interface {
	Get(ctx context.Context, userID string) (PlayerProfile, error)
	Save(ctx context.Context, userID string, profile PlayerProfile) error
}
