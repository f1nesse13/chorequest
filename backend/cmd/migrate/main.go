package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"

	"chorequest/backend/graph/model"
	"chorequest/backend/internal/db"
	repopkg "chorequest/backend/internal/repo"
	sqlitedb "chorequest/backend/internal/sqlite"
)

func main() {
	ctx := context.Background()
	_ = godotenv.Load()

	var sqlitePath string
	flag.StringVar(&sqlitePath, "sqlite", os.Getenv("SQLITE_PATH"), "Path to SQLite database")
	var tableName string
	flag.StringVar(&tableName, "table", os.Getenv("DYNAMO_TABLE_NAME"), "DynamoDB table name")
	flag.Parse()

	if sqlitePath == "" {
		log.Fatal("sqlite path is required (set SQLITE_PATH or pass --sqlite)")
	}
	if tableName == "" {
		tableName = "chorequest"
	}

	dynamoClient, err := db.New(ctx)
	if err != nil {
		log.Fatalf("create dynamo client: %v", err)
	}
	if dynamoClient == nil || dynamoClient.Dynamo == nil {
		log.Fatal("dynamo client not configured; set AWS credentials / DYNAMODB_ENDPOINT")
	}

	sqliteDB, err := sqlitedb.Open(ctx, sqlitePath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer sqliteDB.Close()

	srcRepo := repopkg.NewDynamoRepo(dynamoClient.Dynamo, tableName)

	parentIDs, err := listParents(ctx, dynamoClient.Dynamo, tableName)
	if err != nil {
		log.Fatalf("list parents: %v", err)
	}
	if len(parentIDs) == 0 {
		log.Println("no parents found; nothing to migrate")
		return
	}

	log.Printf("found %d parents to migrate", len(parentIDs))

	if err := migrateFamilies(ctx, sqliteDB, srcRepo, parentIDs); err != nil {
		log.Fatalf("migrate families: %v", err)
	}

	log.Println("migration complete")
}

type parentProjection struct {
	ParentID string `dynamodbav:"ParentID"`
	Type     string `dynamodbav:"Type"`
}

func listParents(ctx context.Context, client *dynamodb.Client, table string) ([]string, error) {
	seen := map[string]struct{}{}
	input := &dynamodb.ScanInput{
		TableName:            aws.String(table),
		ProjectionExpression: aws.String("ParentID, Type"),
	}
	paginator := dynamodb.NewScanPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scan page: %w", err)
		}
		for _, item := range page.Items {
			var proj parentProjection
			if err := attributevalue.UnmarshalMap(item, &proj); err != nil {
				return nil, fmt.Errorf("unmarshal parent projection: %w", err)
			}
			if proj.ParentID != "" {
				seen[proj.ParentID] = struct{}{}
			}
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

func migrateFamilies(ctx context.Context, db *sql.DB, repo repopkg.Repo, parentIDs []string) error {
	return withTx(ctx, db, func(tx *sql.Tx) error {
		for _, parentID := range parentIDs {
			if err := migrateFamily(ctx, tx, repo, parentID); err != nil {
				return fmt.Errorf("parent %s: %w", parentID, err)
			}
		}
		return nil
	})
}

func migrateFamily(ctx context.Context, tx *sql.Tx, repo repopkg.Repo, parentID string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO parents (id) VALUES (?) ON CONFLICT(id) DO NOTHING`, parentID); err != nil {
		return fmt.Errorf("upsert parent: %w", err)
	}

	children, err := repo.ListChildren(ctx, parentID)
	if err != nil {
		return fmt.Errorf("list children: %w", err)
	}
	for _, child := range children {
		if err := upsertChild(ctx, tx, child); err != nil {
			return err
		}
	}

	quests, err := repo.ListQuests(ctx, parentID)
	if err != nil {
		return fmt.Errorf("list quests: %w", err)
	}
	for _, quest := range quests {
		if err := upsertQuest(ctx, tx, quest); err != nil {
			return err
		}
	}

	rewards, err := repo.ListRewards(ctx, parentID)
	if err != nil {
		return fmt.Errorf("list rewards: %w", err)
	}
	for _, reward := range rewards {
		if err := upsertReward(ctx, tx, reward); err != nil {
			return err
		}
	}

	for _, child := range children {
		assignments, err := repo.ListAssignmentsForChild(ctx, child.ID)
		if err != nil {
			return fmt.Errorf("list assignments for %s: %w", child.ID, err)
		}
		for _, assignment := range assignments {
			if err := upsertAssignment(ctx, tx, assignment); err != nil {
				return err
			}
		}
	}

	return nil
}

func upsertChild(ctx context.Context, tx *sql.Tx, child *model.Child) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO children (id, parent_id, name, xp, gold, created_at)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET parent_id = excluded.parent_id, name = excluded.name, xp = excluded.xp, gold = excluded.gold`,
		child.ID, child.ParentID, child.Name, child.Xp, child.Gold, repopkg.NowRFC3339()); err != nil {
		return fmt.Errorf("upsert child %s: %w", child.ID, err)
	}
	return nil
}

func upsertQuest(ctx context.Context, tx *sql.Tx, quest *model.Quest) error {
	var desc any
	if quest.Description != nil {
		desc = *quest.Description
	} else {
		desc = nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO quests (id, parent_id, title, description, xp, gold, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET parent_id = excluded.parent_id, title = excluded.title, description = excluded.description, xp = excluded.xp, gold = excluded.gold`,
		quest.ID, quest.ParentID, quest.Title, desc, quest.Xp, quest.Gold, repopkg.NowRFC3339()); err != nil {
		return fmt.Errorf("upsert quest %s: %w", quest.ID, err)
	}
	return nil
}

func upsertReward(ctx context.Context, tx *sql.Tx, reward *model.Reward) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO rewards (id, parent_id, name, xp_threshold, created_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET parent_id = excluded.parent_id, name = excluded.name, xp_threshold = excluded.xp_threshold`,
		reward.ID, reward.ParentID, reward.Name, reward.XpThreshold, repopkg.NowRFC3339()); err != nil {
		return fmt.Errorf("upsert reward %s: %w", reward.ID, err)
	}
	return nil
}

func upsertAssignment(ctx context.Context, tx *sql.Tx, assignment *model.Assignment) error {
	if assignment.Quest == nil {
		return errors.New("assignment missing quest")
	}
	var completed any
	if assignment.CompletedAt != nil {
		completed = *assignment.CompletedAt
	} else {
		completed = nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO assignments (id, quest_id, child_id, status, created_at, completed_at)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET quest_id = excluded.quest_id, child_id = excluded.child_id, status = excluded.status, created_at = excluded.created_at, completed_at = excluded.completed_at`,
		assignment.ID, assignment.Quest.ID, assignment.ChildID, assignment.Status, assignment.CreatedAt, completed); err != nil {
		return fmt.Errorf("upsert assignment %s: %w", assignment.ID, err)
	}
	return nil
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
