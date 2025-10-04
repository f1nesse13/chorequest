package repo

import (
	"context"
	"time"

	"chorequest/backend/graph/model"
)

// Repo defines persistence operations for the domain.
type Repo interface {
	CreateChild(ctx context.Context, in model.NewChild) (*model.Child, error)
	ListChildren(ctx context.Context, familyID string) ([]*model.Child, error)
	GetChild(ctx context.Context, childID string) (*model.Child, error)
	UpdateChild(ctx context.Context, familyID, childID string, in model.UpdateChildInput) (*model.Child, error)
	DeleteChild(ctx context.Context, familyID, childID string) error

	CreateQuest(ctx context.Context, in model.NewQuest) (*model.Quest, error)
	ListQuests(ctx context.Context, familyID string) ([]*model.Quest, error)
	GetQuestByID(ctx context.Context, questID string) (*model.Quest, error)
	UpdateQuest(ctx context.Context, familyID, questID string, in model.UpdateQuestInput) (*model.Quest, error)
	DeleteQuest(ctx context.Context, familyID, questID string) error

	AssignQuest(ctx context.Context, familyID, questID, childID string) (*model.Assignment, error)
	ListAssignmentsForChild(ctx context.Context, childID string) ([]*model.Assignment, error)
	CompleteAssignment(ctx context.Context, assignmentID string) (*model.Assignment, error)
	DeleteAssignment(ctx context.Context, assignmentID string) error

	CreateReward(ctx context.Context, in model.NewReward) (*model.Reward, error)
	ListRewards(ctx context.Context, familyID string) ([]*model.Reward, error)
	UpdateReward(ctx context.Context, familyID, rewardID string, in model.UpdateRewardInput) (*model.Reward, error)
	DeleteReward(ctx context.Context, familyID, rewardID string) error

	CreateFamily(ctx context.Context, in model.NewFamily) (*model.Family, error)
	GetFamily(ctx context.Context, familyID string) (*model.Family, error)
	ListFamiliesForOwner(ctx context.Context, ownerID string) ([]*model.Family, error)
	RenameFamily(ctx context.Context, familyID, name string) (*model.Family, error)
	DeleteFamily(ctx context.Context, familyID string) error

	PurchaseItem(ctx context.Context, childID, itemName string, priceGold int) (*model.Child, error)
}

// NowRFC3339 returns a UTC RFC3339 timestamp.
func NowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }
