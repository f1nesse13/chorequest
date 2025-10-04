package main

import (
	"context"
	"log"
	"os"

	"chorequest/backend/graph/model"
	repopkg "chorequest/backend/internal/repo"
	sqlitedb "chorequest/backend/internal/sqlite"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "./data/dev.sqlite"
	}
	db, err := sqlitedb.Open(ctx, sqlitePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repopkg.NewSQLiteRepo(db)

	ownerID := os.Getenv("SEED_OWNER_ID")
	if ownerID == "" {
		ownerID = "parent-1"
	}
	familyName := os.Getenv("SEED_FAMILY_NAME")
	if familyName == "" {
		familyName = "Demo Family"
	}
	family, err := repo.CreateFamily(ctx, model.NewFamily{OwnerID: ownerID, Name: familyName})
	if err != nil {
		log.Fatal(err)
	}

	child, err := repo.CreateChild(ctx, model.NewChild{FamilyID: family.ID, Name: "Alex"})
	if err != nil {
		log.Fatal(err)
	}

	q1, err := repo.CreateQuest(ctx, model.NewQuest{FamilyID: family.ID, Title: "Clean Room", Description: ptr("Tidy up and vacuum"), Xp: 50, Gold: 10})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := repo.CreateQuest(ctx, model.NewQuest{FamilyID: family.ID, Title: "Do Dishes", Description: ptr("Load and run dishwasher"), Xp: 30, Gold: 8}); err != nil {
		log.Fatal(err)
	}

	if _, err := repo.CreateReward(ctx, model.NewReward{FamilyID: family.ID, Name: "Movie Night", XpThreshold: 200}); err != nil {
		log.Fatal(err)
	}

	if _, err := repo.AssignQuest(ctx, family.ID, q1.ID, child.ID); err != nil {
		log.Fatal(err)
	}

	log.Printf("Seeded family=%s child=%s", family.ID, child.ID)
}

func ptr[T any](v T) *T { return &v }
