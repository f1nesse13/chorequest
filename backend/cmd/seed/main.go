package main

import (
    "context"
    "log"
    "os"

    "chorequest/backend/internal/db"
    repopkg "chorequest/backend/internal/repo"
    "chorequest/backend/graph/model"
)

func main() {
    ctx := context.Background()
    client, err := db.New(ctx)
    if err != nil { log.Fatal(err) }
    if err := db.EnsureSingleTable(ctx, client, os.Getenv("DYNAMO_TABLE_NAME")); err != nil { log.Fatal(err) }

    repo := repopkg.NewDynamoRepo(client.Dynamo, os.Getenv("DYNAMO_TABLE_NAME"))

    parentID := os.Getenv("SEED_PARENT_ID")
    if parentID == "" { parentID = "parent-1" }

    child, err := repo.CreateChild(ctx, model.NewChild{ParentID: parentID, Name: "Alex"})
    if err != nil { log.Fatal(err) }

    q1, err := repo.CreateQuest(ctx, model.NewQuest{ParentID: parentID, Title: "Clean Room", Description: ptr("Tidy up and vacuum"), Xp: 50, Gold: 10})
    if err != nil { log.Fatal(err) }
    if _, err := repo.CreateQuest(ctx, model.NewQuest{ParentID: parentID, Title: "Do Dishes", Description: ptr("Load and run dishwasher"), Xp: 30, Gold: 8}); err != nil { log.Fatal(err) }

    if _, err := repo.CreateReward(ctx, model.NewReward{ParentID: parentID, Name: "Movie Night", XpThreshold: 200}); err != nil { log.Fatal(err) }

    if _, err := repo.AssignQuest(ctx, q1.ID, child.ID); err != nil { log.Fatal(err) }

    log.Printf("Seeded parent=%s child=%s", parentID, child.ID)
}

func ptr[T any](v T) *T { return &v }

