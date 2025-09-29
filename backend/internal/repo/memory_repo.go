package repo

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/google/uuid"

	"chorequest/backend/graph/model"
)

type MemoryRepo struct {
	mu sync.RWMutex

	children map[string]*model.Child
	quests   map[string]*model.Quest
	rewards  map[string]*model.Reward

	assignments map[string]*model.Assignment

	childrenByParent   map[string][]string
	questsByParent     map[string][]string
	rewardsByParent    map[string][]string
	assignmentsByChild map[string][]string
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		children:           make(map[string]*model.Child),
		quests:             make(map[string]*model.Quest),
		rewards:            make(map[string]*model.Reward),
		assignments:        make(map[string]*model.Assignment),
		childrenByParent:   make(map[string][]string),
		questsByParent:     make(map[string][]string),
		rewardsByParent:    make(map[string][]string),
		assignmentsByChild: make(map[string][]string),
	}
}

func (r *MemoryRepo) CreateChild(_ context.Context, in model.NewChild) (*model.Child, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	child := &model.Child{ID: id, ParentID: in.ParentID, Name: in.Name, Xp: 0, Gold: 0}
	r.children[id] = child
	r.childrenByParent[in.ParentID] = append(r.childrenByParent[in.ParentID], id)
	return cloneChild(child), nil
}

func (r *MemoryRepo) ListChildren(_ context.Context, parentID string) ([]*model.Child, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.childrenByParent[parentID]
	res := make([]*model.Child, 0, len(ids))
	for _, id := range ids {
		if child, ok := r.children[id]; ok {
			res = append(res, cloneChild(child))
		}
	}
	return res, nil
}

func (r *MemoryRepo) UpdateChild(_ context.Context, parentID, childID string, in model.UpdateChildInput) (*model.Child, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	child, ok := r.children[childID]
	if !ok || child.ParentID != parentID {
		return nil, errors.New("child not found")
	}
	child.Name = in.Name
	return cloneChild(child), nil
}

func (r *MemoryRepo) DeleteChild(_ context.Context, parentID, childID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	child, ok := r.children[childID]
	if !ok || child.ParentID != parentID {
		return errors.New("child not found")
	}
	delete(r.children, childID)
	r.childrenByParent[parentID] = removeString(r.childrenByParent[parentID], childID)

	if ids, ok := r.assignmentsByChild[childID]; ok {
		for _, aid := range ids {
			delete(r.assignments, aid)
		}
		delete(r.assignmentsByChild, childID)
	}
	return nil
}

func (r *MemoryRepo) CreateQuest(_ context.Context, in model.NewQuest) (*model.Quest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	quest := &model.Quest{ID: id, ParentID: in.ParentID, Title: in.Title, Description: in.Description, Xp: in.Xp, Gold: in.Gold}
	r.quests[id] = quest
	r.questsByParent[in.ParentID] = append(r.questsByParent[in.ParentID], id)
	return cloneQuest(quest), nil
}

func (r *MemoryRepo) ListQuests(_ context.Context, parentID string) ([]*model.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.questsByParent[parentID]
	res := make([]*model.Quest, 0, len(ids))
	for _, id := range ids {
		if quest, ok := r.quests[id]; ok {
			res = append(res, cloneQuest(quest))
		}
	}
	return res, nil
}

func (r *MemoryRepo) GetQuestByID(_ context.Context, questID string) (*model.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quest, ok := r.quests[questID]
	if !ok {
		return nil, errors.New("quest not found")
	}
	return cloneQuest(quest), nil
}

func (r *MemoryRepo) UpdateQuest(_ context.Context, parentID, questID string, in model.UpdateQuestInput) (*model.Quest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	quest, ok := r.quests[questID]
	if !ok || quest.ParentID != parentID {
		return nil, errors.New("quest not found")
	}
	quest.Title = in.Title
	quest.Description = in.Description
	quest.Xp = in.Xp
	quest.Gold = in.Gold

	// Update cached quest on assignments
	for _, assignment := range r.assignments {
		if assignment.Quest != nil && assignment.Quest.ID == questID {
			assignment.Quest = cloneQuest(quest)
		}
	}

	return cloneQuest(quest), nil
}

func (r *MemoryRepo) DeleteQuest(_ context.Context, parentID, questID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	quest, ok := r.quests[questID]
	if !ok || quest.ParentID != parentID {
		return errors.New("quest not found")
	}

	delete(r.quests, questID)
	r.questsByParent[parentID] = removeString(r.questsByParent[parentID], questID)

	// Remove any assignments referencing this quest
	for childID, ids := range r.assignmentsByChild {
		filtered := ids[:0]
		for _, aid := range ids {
			a := r.assignments[aid]
			if a != nil && a.Quest != nil && a.Quest.ID == questID {
				delete(r.assignments, aid)
				continue
			}
			filtered = append(filtered, aid)
		}
		if len(filtered) == 0 {
			delete(r.assignmentsByChild, childID)
		} else {
			r.assignmentsByChild[childID] = filtered
		}
	}
	return nil
}

func (r *MemoryRepo) AssignQuest(_ context.Context, questID, childID string) (*model.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	quest, ok := r.quests[questID]
	if !ok {
		return nil, errors.New("quest not found")
	}
	if _, ok := r.children[childID]; !ok {
		return nil, errors.New("child not found")
	}

	id := uuid.NewString()
	assignment := &model.Assignment{
		ID:        id,
		Quest:     cloneQuest(quest),
		ChildID:   childID,
		Status:    "ASSIGNED",
		CreatedAt: NowRFC3339(),
	}
	r.assignments[id] = assignment
	r.assignmentsByChild[childID] = append(r.assignmentsByChild[childID], id)
	return cloneAssignment(assignment), nil
}

func (r *MemoryRepo) ListAssignmentsForChild(_ context.Context, childID string) ([]*model.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := append([]string(nil), r.assignmentsByChild[childID]...)
	res := make([]*model.Assignment, 0, len(ids))
	for _, id := range ids {
		if assignment, ok := r.assignments[id]; ok {
			res = append(res, cloneAssignment(assignment))
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt > res[j].CreatedAt
	})
	return res, nil
}

func (r *MemoryRepo) CompleteAssignment(_ context.Context, assignmentID string) (*model.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	assignment, ok := r.assignments[assignmentID]
	if !ok {
		return nil, errors.New("assignment not found")
	}
	if assignment.Quest == nil {
		return nil, errors.New("quest not loaded")
	}
	child, ok := r.children[assignment.ChildID]
	if !ok {
		return nil, errors.New("child not found")
	}

	if assignment.Status != "COMPLETED" {
		done := NowRFC3339()
		assignment.Status = "COMPLETED"
		assignment.CompletedAt = &done
		child.Xp += assignment.Quest.Xp
		child.Gold += assignment.Quest.Gold
	}

	return cloneAssignment(assignment), nil
}

func (r *MemoryRepo) CreateReward(_ context.Context, in model.NewReward) (*model.Reward, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.NewString()
	reward := &model.Reward{ID: id, ParentID: in.ParentID, Name: in.Name, XpThreshold: in.XpThreshold}
	r.rewards[id] = reward
	r.rewardsByParent[in.ParentID] = append(r.rewardsByParent[in.ParentID], id)
	return cloneReward(reward), nil
}

func (r *MemoryRepo) ListRewards(_ context.Context, parentID string) ([]*model.Reward, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.rewardsByParent[parentID]
	res := make([]*model.Reward, 0, len(ids))
	for _, id := range ids {
		if reward, ok := r.rewards[id]; ok {
			res = append(res, cloneReward(reward))
		}
	}
	return res, nil
}

func (r *MemoryRepo) UpdateReward(_ context.Context, parentID, rewardID string, in model.UpdateRewardInput) (*model.Reward, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reward, ok := r.rewards[rewardID]
	if !ok || reward.ParentID != parentID {
		return nil, errors.New("reward not found")
	}
	reward.Name = in.Name
	reward.XpThreshold = in.XpThreshold
	return cloneReward(reward), nil
}

func (r *MemoryRepo) DeleteReward(_ context.Context, parentID, rewardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reward, ok := r.rewards[rewardID]
	if !ok || reward.ParentID != parentID {
		return errors.New("reward not found")
	}

	delete(r.rewards, rewardID)
	r.rewardsByParent[parentID] = removeString(r.rewardsByParent[parentID], rewardID)
	return nil
}

func (r *MemoryRepo) PurchaseItem(_ context.Context, childID, _ string, priceGold int) (*model.Child, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	child, ok := r.children[childID]
	if !ok {
		return nil, errors.New("child not found")
	}
	if child.Gold < priceGold {
		return nil, errors.New("insufficient gold")
	}
	child.Gold -= priceGold
	return cloneChild(child), nil
}

func (r *MemoryRepo) SeedDemoData() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.children) > 0 || len(r.quests) > 0 || len(r.rewards) > 0 {
		return
	}

	parentID := "parent-1"
	child := &model.Child{ID: "child-1", ParentID: parentID, Name: "Alex", Xp: 120, Gold: 45}
	quest1 := &model.Quest{ID: "quest-1", ParentID: parentID, Title: "Clean Room", Description: strPtr("Tidy up and vacuum"), Xp: 50, Gold: 10}
	quest2 := &model.Quest{ID: "quest-2", ParentID: parentID, Title: "Do Dishes", Description: strPtr("Load and run dishwasher"), Xp: 30, Gold: 8}
	reward := &model.Reward{ID: "reward-1", ParentID: parentID, Name: "Movie Night", XpThreshold: 200}

	r.children[child.ID] = child
	r.childrenByParent[parentID] = []string{child.ID}

	r.quests[quest1.ID] = quest1
	r.quests[quest2.ID] = quest2
	r.questsByParent[parentID] = []string{quest1.ID, quest2.ID}

	r.rewards[reward.ID] = reward
	r.rewardsByParent[parentID] = []string{reward.ID}

	assignment := &model.Assignment{
		ID:        "assignment-1",
		Quest:     cloneQuest(quest1),
		ChildID:   child.ID,
		Status:    "ASSIGNED",
		CreatedAt: NowRFC3339(),
	}
	r.assignments[assignment.ID] = assignment
	r.assignmentsByChild[child.ID] = []string{assignment.ID}
}

func cloneChild(in *model.Child) *model.Child {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneQuest(in *model.Quest) *model.Quest {
	if in == nil {
		return nil
	}
	out := *in
	if in.Description != nil {
		v := *in.Description
		out.Description = &v
	}
	return &out
}

func cloneReward(in *model.Reward) *model.Reward {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneAssignment(in *model.Assignment) *model.Assignment {
	if in == nil {
		return nil
	}
	out := *in
	out.Quest = cloneQuest(in.Quest)
	if in.CompletedAt != nil {
		v := *in.CompletedAt
		out.CompletedAt = &v
	}
	return &out
}

func removeString(list []string, target string) []string {
	for i, v := range list {
		if v == target {
			return append(list[:i], list[i+1:]...)
		}
	}
	return list
}

func strPtr(v string) *string {
	return &v
}

var _ Repo = (*MemoryRepo)(nil)
