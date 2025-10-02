package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"chorequest/backend/graph/model"
)

type SQLiteRepo struct {
	DB *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{DB: db}
}

func (r *SQLiteRepo) CreateChild(ctx context.Context, in model.NewChild) (*model.Child, error) {
	if err := r.ensureParent(ctx, in.ParentID); err != nil {
		return nil, err
	}
	id := uuid.NewString()
	if _, err := r.DB.ExecContext(ctx, `INSERT INTO children (id, parent_id, name, xp, gold, created_at) VALUES (?, ?, ?, 0, 0, ?)`, id, in.ParentID, in.Name, NowRFC3339()); err != nil {
		return nil, fmt.Errorf("insert child: %w", err)
	}
	return &model.Child{ID: id, ParentID: in.ParentID, Name: in.Name, Xp: 0, Gold: 0}, nil
}

func (r *SQLiteRepo) ListChildren(ctx context.Context, parentID string) ([]*model.Child, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, xp, gold FROM children WHERE parent_id = ? ORDER BY created_at DESC, name ASC`, parentID)
	if err != nil {
		return nil, fmt.Errorf("query children: %w", err)
	}
	defer rows.Close()
	var items []*model.Child
	for rows.Next() {
		var c model.Child
		c.ParentID = parentID
		if err := rows.Scan(&c.ID, &c.Name, &c.Xp, &c.Gold); err != nil {
			return nil, err
		}
		items = append(items, &c)
	}
	return items, rows.Err()
}

func (r *SQLiteRepo) UpdateChild(ctx context.Context, parentID, childID string, in model.UpdateChildInput) (*model.Child, error) {
	res, err := r.DB.ExecContext(ctx, `UPDATE children SET name = ? WHERE id = ? AND parent_id = ?`, in.Name, childID, parentID)
	if err != nil {
		return nil, fmt.Errorf("update child: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("child not found")
	}
	var xp, gold int
	if err := r.DB.QueryRowContext(ctx, `SELECT xp, gold FROM children WHERE id = ?`, childID).Scan(&xp, &gold); err != nil {
		return nil, fmt.Errorf("select child: %w", err)
	}
	return &model.Child{ID: childID, ParentID: parentID, Name: in.Name, Xp: xp, Gold: gold}, nil
}

func (r *SQLiteRepo) DeleteChild(ctx context.Context, parentID, childID string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM assignments WHERE child_id = ?`, childID); err != nil {
		return fmt.Errorf("delete child assignments: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM children WHERE id = ? AND parent_id = ?`, childID, parentID)
	if err != nil {
		return fmt.Errorf("delete child: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("child not found")
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *SQLiteRepo) CreateQuest(ctx context.Context, in model.NewQuest) (*model.Quest, error) {
	if err := r.ensureParent(ctx, in.ParentID); err != nil {
		return nil, err
	}
	id := uuid.NewString()
	if _, err := r.DB.ExecContext(ctx, `INSERT INTO quests (id, parent_id, title, description, xp, gold, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, in.ParentID, in.Title, in.Description, in.Xp, in.Gold, NowRFC3339()); err != nil {
		return nil, fmt.Errorf("insert quest: %w", err)
	}
	return &model.Quest{ID: id, ParentID: in.ParentID, Title: in.Title, Description: in.Description, Xp: in.Xp, Gold: in.Gold}, nil
}

func (r *SQLiteRepo) ListQuests(ctx context.Context, parentID string) ([]*model.Quest, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, title, description, xp, gold FROM quests WHERE parent_id = ? ORDER BY created_at DESC`, parentID)
	if err != nil {
		return nil, fmt.Errorf("query quests: %w", err)
	}
	defer rows.Close()
	var items []*model.Quest
	for rows.Next() {
		var q model.Quest
		q.ParentID = parentID
		if err := rows.Scan(&q.ID, &q.Title, &q.Description, &q.Xp, &q.Gold); err != nil {
			return nil, err
		}
		items = append(items, &q)
	}
	return items, rows.Err()
}

func (r *SQLiteRepo) GetQuestByID(ctx context.Context, questID string) (*model.Quest, error) {
	var q model.Quest
	if err := r.DB.QueryRowContext(ctx, `SELECT id, parent_id, title, description, xp, gold FROM quests WHERE id = ?`, questID).Scan(&q.ID, &q.ParentID, &q.Title, &q.Description, &q.Xp, &q.Gold); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("quest not found")
		}
		return nil, fmt.Errorf("select quest: %w", err)
	}
	return &q, nil
}

func (r *SQLiteRepo) UpdateQuest(ctx context.Context, parentID, questID string, in model.UpdateQuestInput) (*model.Quest, error) {
	res, err := r.DB.ExecContext(ctx, `UPDATE quests SET title = ?, description = ?, xp = ?, gold = ? WHERE id = ? AND parent_id = ?`, in.Title, in.Description, in.Xp, in.Gold, questID, parentID)
	if err != nil {
		return nil, fmt.Errorf("update quest: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("quest not found")
	}
	return &model.Quest{ID: questID, ParentID: parentID, Title: in.Title, Description: in.Description, Xp: in.Xp, Gold: in.Gold}, nil
}

func (r *SQLiteRepo) DeleteQuest(ctx context.Context, parentID, questID string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM assignments WHERE quest_id = ?`, questID); err != nil {
		return fmt.Errorf("delete quest assignments: %w", err)
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM quests WHERE id = ? AND parent_id = ?`, questID, parentID)
	if err != nil {
		return fmt.Errorf("delete quest: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("quest not found")
	}
	return tx.Commit()
}

func (r *SQLiteRepo) AssignQuest(ctx context.Context, questID, childID string) (*model.Assignment, error) {
	quest, err := r.GetQuestByID(ctx, questID)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	created := NowRFC3339()
	if _, err := r.DB.ExecContext(ctx, `INSERT INTO assignments (id, quest_id, child_id, status, created_at) VALUES (?, ?, ?, 'ASSIGNED', ?)`, id, questID, childID, created); err != nil {
		return nil, fmt.Errorf("insert assignment: %w", err)
	}
	return &model.Assignment{ID: id, Quest: quest, ChildID: childID, Status: "ASSIGNED", CreatedAt: created}, nil
}

func (r *SQLiteRepo) ListAssignmentsForChild(ctx context.Context, childID string) ([]*model.Assignment, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT a.id, a.status, a.created_at, a.completed_at, q.id, q.parent_id, q.title, q.description, q.xp, q.gold FROM assignments a JOIN quests q ON q.id = a.quest_id WHERE a.child_id = ? ORDER BY a.created_at DESC`, childID)
	if err != nil {
		return nil, fmt.Errorf("query assignments: %w", err)
	}
	defer rows.Close()
	var res []*model.Assignment
	for rows.Next() {
		var assignment model.Assignment
		var quest model.Quest
		var completed sql.NullString
		if err := rows.Scan(&assignment.ID, &assignment.Status, &assignment.CreatedAt, &completed, &quest.ID, &quest.ParentID, &quest.Title, &quest.Description, &quest.Xp, &quest.Gold); err != nil {
			return nil, err
		}
		if completed.Valid {
			assignment.CompletedAt = &completed.String
		}
		assignment.ChildID = childID
		assignment.Quest = &quest
		res = append(res, &assignment)
	}
	return res, rows.Err()
}

func (r *SQLiteRepo) CompleteAssignment(ctx context.Context, assignmentID string) (*model.Assignment, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status string
	var createdAt string
	var completed sql.NullString
	var questID string
	var childID string
	if err := tx.QueryRowContext(ctx, `SELECT status, created_at, completed_at, quest_id, child_id FROM assignments WHERE id = ?`, assignmentID).Scan(&status, &createdAt, &completed, &questID, &childID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("assignment not found")
		}
		return nil, err
	}
	if status == "COMPLETED" {
		quest, _ := r.GetQuestByID(ctx, questID)
		return &model.Assignment{ID: assignmentID, Status: status, CreatedAt: createdAt, CompletedAt: completedStringPtr(completed), Quest: quest, ChildID: childID}, nil
	}

	quest, err := r.GetQuestByID(ctx, questID)
	if err != nil {
		return nil, err
	}
	done := NowRFC3339()
	if _, err := tx.ExecContext(ctx, `UPDATE assignments SET status = 'COMPLETED', completed_at = ? WHERE id = ?`, done, assignmentID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE children SET xp = xp + ?, gold = gold + ? WHERE id = ?`, quest.Xp, quest.Gold, childID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &model.Assignment{ID: assignmentID, Quest: quest, ChildID: childID, Status: "COMPLETED", CreatedAt: createdAt, CompletedAt: &done}, nil
}

func (r *SQLiteRepo) DeleteAssignment(ctx context.Context, assignmentID string) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM assignments WHERE id = ?`, assignmentID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("assignment not found")
	}
	return nil
}

func (r *SQLiteRepo) CreateReward(ctx context.Context, in model.NewReward) (*model.Reward, error) {
	if err := r.ensureParent(ctx, in.ParentID); err != nil {
		return nil, err
	}
	id := uuid.NewString()
	if _, err := r.DB.ExecContext(ctx, `INSERT INTO rewards (id, parent_id, name, xp_threshold, created_at) VALUES (?, ?, ?, ?, ?)`, id, in.ParentID, in.Name, in.XpThreshold, NowRFC3339()); err != nil {
		return nil, fmt.Errorf("insert reward: %w", err)
	}
	return &model.Reward{ID: id, ParentID: in.ParentID, Name: in.Name, XpThreshold: in.XpThreshold}, nil
}

func (r *SQLiteRepo) ListRewards(ctx context.Context, parentID string) ([]*model.Reward, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, name, xp_threshold FROM rewards WHERE parent_id = ? ORDER BY created_at DESC`, parentID)
	if err != nil {
		return nil, fmt.Errorf("query rewards: %w", err)
	}
	defer rows.Close()
	var items []*model.Reward
	for rows.Next() {
		var rw model.Reward
		rw.ParentID = parentID
		if err := rows.Scan(&rw.ID, &rw.Name, &rw.XpThreshold); err != nil {
			return nil, err
		}
		items = append(items, &rw)
	}
	return items, rows.Err()
}

func (r *SQLiteRepo) UpdateReward(ctx context.Context, parentID, rewardID string, in model.UpdateRewardInput) (*model.Reward, error) {
	res, err := r.DB.ExecContext(ctx, `UPDATE rewards SET name = ?, xp_threshold = ? WHERE id = ? AND parent_id = ?`, in.Name, in.XpThreshold, rewardID, parentID)
	if err != nil {
		return nil, fmt.Errorf("update reward: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("reward not found")
	}
	return &model.Reward{ID: rewardID, ParentID: parentID, Name: in.Name, XpThreshold: in.XpThreshold}, nil
}

func (r *SQLiteRepo) DeleteReward(ctx context.Context, parentID, rewardID string) error {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM rewards WHERE id = ? AND parent_id = ?`, rewardID, parentID)
	if err != nil {
		return fmt.Errorf("delete reward: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("reward not found")
	}
	return nil
}

func (r *SQLiteRepo) PurchaseItem(ctx context.Context, childID, _ string, priceGold int) (*model.Child, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var parentID string
	var name string
	var xp int
	var gold int
	if err := tx.QueryRowContext(ctx, `SELECT parent_id, name, xp, gold FROM children WHERE id = ?`, childID).Scan(&parentID, &name, &xp, &gold); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("child not found")
		}
		return nil, err
	}
	if gold < priceGold {
		return nil, errors.New("insufficient gold")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE children SET gold = gold - ? WHERE id = ?`, priceGold, childID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &model.Child{ID: childID, ParentID: parentID, Name: name, Xp: xp, Gold: gold - priceGold}, nil
}

func (r *SQLiteRepo) ensureParent(ctx context.Context, parentID string) error {
	if parentID == "" {
		return errors.New("parentID required")
	}
	_, err := r.DB.ExecContext(ctx, `INSERT INTO parents (id) VALUES (?) ON CONFLICT(id) DO NOTHING`, parentID)
	return err
}

func completedStringPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

var _ Repo = (*SQLiteRepo)(nil)
