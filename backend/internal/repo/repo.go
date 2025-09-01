package repo

import (
    "context"
    "time"

    "chorequest/backend/graph/model"
)

// Repo defines operations for the domain backed by DynamoDB.
type Repo interface {
    CreateChild(ctx context.Context, in model.NewChild) (*model.Child, error)
    ListChildren(ctx context.Context, parentID string) ([]*model.Child, error)

    CreateQuest(ctx context.Context, in model.NewQuest) (*model.Quest, error)
    ListQuests(ctx context.Context, parentID string) ([]*model.Quest, error)
    GetQuestByID(ctx context.Context, questID string) (*model.Quest, error)

    AssignQuest(ctx context.Context, questID, childID string) (*model.Assignment, error)
    ListAssignmentsForChild(ctx context.Context, childID string) ([]*model.Assignment, error)
    CompleteAssignment(ctx context.Context, assignmentID string) (*model.Assignment, error)

    CreateReward(ctx context.Context, in model.NewReward) (*model.Reward, error)
    ListRewards(ctx context.Context, parentID string) ([]*model.Reward, error)

    PurchaseItem(ctx context.Context, childID, itemName string, priceGold int) (*model.Child, error)
}

// NowRFC3339 returns a UTC RFC3339 timestamp.
func NowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

