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

	families map[string]*model.Family

	children map[string]*model.Child
	quests   map[string]*model.Quest
	rewards  map[string]*model.Reward

	assignments map[string]*model.Assignment

	childrenByFamily   map[string][]string
	questsByFamily     map[string][]string
	rewardsByFamily    map[string][]string
	assignmentsByChild map[string][]string
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		families:           make(map[string]*model.Family),
		children:           make(map[string]*model.Child),
		quests:             make(map[string]*model.Quest),
		rewards:            make(map[string]*model.Reward),
		assignments:        make(map[string]*model.Assignment),
		childrenByFamily:   make(map[string][]string),
		questsByFamily:     make(map[string][]string),
		rewardsByFamily:    make(map[string][]string),
		assignmentsByChild: make(map[string][]string),
	}
}

// Family operations
func (r *MemoryRepo) CreateFamily(_ context.Context, in model.NewFamily) (*model.Family, error) {
	if in.OwnerID == "" {
		return nil, errors.New("ownerId required")
	}
	if in.Name == "" {
		return nil, errors.New("name required")
	}
	id := uuid.NewString()
	fam := &model.Family{ID: id, Name: in.Name, OwnerID: in.OwnerID, CreatedAt: NowRFC3339()}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.families[id] = fam
	return cloneFamily(fam), nil
}

func (r *MemoryRepo) GetFamily(_ context.Context, familyID string) (*model.Family, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fam, ok := r.families[familyID]
	if !ok {
		return nil, errors.New("family not found")
	}
	return cloneFamily(fam), nil
}

func (r *MemoryRepo) ListFamiliesForOwner(_ context.Context, ownerID string) ([]*model.Family, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := []*model.Family{}
	for _, fam := range r.families {
		if fam.OwnerID == ownerID {
			res = append(res, cloneFamily(fam))
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].CreatedAt > res[j].CreatedAt })
	return res, nil
}

func (r *MemoryRepo) RenameFamily(_ context.Context, familyID, name string) (*model.Family, error) {
	if name == "" {
		return nil, errors.New("name required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	fam, ok := r.families[familyID]
	if !ok {
		return nil, errors.New("family not found")
	}
	fam.Name = name
	return cloneFamily(fam), nil
}

func (r *MemoryRepo) DeleteFamily(_ context.Context, familyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.families[familyID]; !ok {
		return errors.New("family not found")
	}
	delete(r.families, familyID)

	if ids := r.childrenByFamily[familyID]; len(ids) > 0 {
		for _, id := range ids {
			delete(r.children, id)
			if assigns := r.assignmentsByChild[id]; len(assigns) > 0 {
				for _, aid := range assigns {
					delete(r.assignments, aid)
				}
				delete(r.assignmentsByChild, id)
			}
		}
		delete(r.childrenByFamily, familyID)
	}

	if ids := r.questsByFamily[familyID]; len(ids) > 0 {
		for _, id := range ids {
			delete(r.quests, id)
		}
		delete(r.questsByFamily, familyID)
	}

	if ids := r.rewardsByFamily[familyID]; len(ids) > 0 {
		for _, id := range ids {
			delete(r.rewards, id)
		}
		delete(r.rewardsByFamily, familyID)
	}

	return nil
}

// Helper ensures family exists (creates placeholder if needed).
func (r *MemoryRepo) ensureFamilyUnlocked(familyID string) {
	if familyID == "" {
		return
	}
	if _, ok := r.families[familyID]; !ok {
		r.families[familyID] = &model.Family{ID: familyID, Name: "Family " + familyID, OwnerID: familyID, CreatedAt: NowRFC3339()}
	}
}

// Child operations
func (r *MemoryRepo) CreateChild(_ context.Context, in model.NewChild) (*model.Child, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if in.FamilyID == "" {
		return nil, errors.New("familyId required")
	}
	r.ensureFamilyUnlocked(in.FamilyID)

	id := uuid.NewString()
	child := &model.Child{ID: id, FamilyID: in.FamilyID, Name: in.Name, Xp: 0, Gold: 0}
	r.children[id] = child
	r.childrenByFamily[in.FamilyID] = append(r.childrenByFamily[in.FamilyID], id)
	return cloneChild(child), nil
}

func (r *MemoryRepo) ListChildren(_ context.Context, familyID string) ([]*model.Child, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.childrenByFamily[familyID]
	res := make([]*model.Child, 0, len(ids))
	for _, id := range ids {
		if child, ok := r.children[id]; ok {
			res = append(res, cloneChild(child))
		}
	}
	return res, nil
}

func (r *MemoryRepo) GetChild(_ context.Context, childID string) (*model.Child, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	child, ok := r.children[childID]
	if !ok {
		return nil, errors.New("child not found")
	}
	return cloneChild(child), nil
}

func (r *MemoryRepo) UpdateChild(_ context.Context, familyID, childID string, in model.UpdateChildInput) (*model.Child, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	child, ok := r.children[childID]
	if !ok || child.FamilyID != familyID {
		return nil, errors.New("child not found")
	}
	child.Name = in.Name
	return cloneChild(child), nil
}

func (r *MemoryRepo) DeleteChild(_ context.Context, familyID, childID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	child, ok := r.children[childID]
	if !ok || child.FamilyID != familyID {
		return errors.New("child not found")
	}
	delete(r.children, childID)
	r.childrenByFamily[familyID] = removeString(r.childrenByFamily[familyID], childID)

	if ids, ok := r.assignmentsByChild[childID]; ok {
		for _, aid := range ids {
			delete(r.assignments, aid)
		}
		delete(r.assignmentsByChild, childID)
	}
	return nil
}

// Quest operations
func (r *MemoryRepo) CreateQuest(_ context.Context, in model.NewQuest) (*model.Quest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if in.FamilyID == "" {
		return nil, errors.New("familyId required")
	}
	r.ensureFamilyUnlocked(in.FamilyID)

	id := uuid.NewString()
	quest := &model.Quest{ID: id, FamilyID: in.FamilyID, Title: in.Title, Description: in.Description, Xp: in.Xp, Gold: in.Gold}
	r.quests[id] = quest
	r.questsByFamily[in.FamilyID] = append(r.questsByFamily[in.FamilyID], id)
	return cloneQuest(quest), nil
}

func (r *MemoryRepo) ListQuests(_ context.Context, familyID string) ([]*model.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.questsByFamily[familyID]
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

func (r *MemoryRepo) UpdateQuest(_ context.Context, familyID, questID string, in model.UpdateQuestInput) (*model.Quest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	quest, ok := r.quests[questID]
	if !ok || quest.FamilyID != familyID {
		return nil, errors.New("quest not found")
	}
	quest.Title = in.Title
	quest.Description = in.Description
	quest.Xp = in.Xp
	quest.Gold = in.Gold

	for _, assignment := range r.assignments {
		if assignment.Quest != nil && assignment.Quest.ID == questID {
			assignment.Quest = cloneQuest(quest)
		}
	}

	return cloneQuest(quest), nil
}

func (r *MemoryRepo) DeleteQuest(_ context.Context, familyID, questID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	quest, ok := r.quests[questID]
	if !ok || quest.FamilyID != familyID {
		return errors.New("quest not found")
	}

	delete(r.quests, questID)
	r.questsByFamily[familyID] = removeString(r.questsByFamily[familyID], questID)

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

// Assignment operations
func (r *MemoryRepo) AssignQuest(_ context.Context, familyID, questID, childID string) (*model.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	quest, ok := r.quests[questID]
	if !ok || quest.FamilyID != familyID {
		return nil, errors.New("quest not found")
	}
	child, ok := r.children[childID]
	if !ok || child.FamilyID != familyID {
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
	sort.Slice(res, func(i, j int) bool { return res[i].CreatedAt > res[j].CreatedAt })
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

func (r *MemoryRepo) DeleteAssignment(_ context.Context, assignmentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	assignment, ok := r.assignments[assignmentID]
	if !ok {
		return errors.New("assignment not found")
	}
	delete(r.assignments, assignmentID)

	if list, ok := r.assignmentsByChild[assignment.ChildID]; ok {
		filtered := list[:0]
		for _, id := range list {
			if id != assignmentID {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) == 0 {
			delete(r.assignmentsByChild, assignment.ChildID)
		} else {
			r.assignmentsByChild[assignment.ChildID] = filtered
		}
	}
	return nil
}

// Reward operations
func (r *MemoryRepo) CreateReward(_ context.Context, in model.NewReward) (*model.Reward, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if in.FamilyID == "" {
		return nil, errors.New("familyId required")
	}
	r.ensureFamilyUnlocked(in.FamilyID)

	id := uuid.NewString()
	reward := &model.Reward{ID: id, FamilyID: in.FamilyID, Name: in.Name, XpThreshold: in.XpThreshold}
	r.rewards[id] = reward
	r.rewardsByFamily[in.FamilyID] = append(r.rewardsByFamily[in.FamilyID], id)
	return cloneReward(reward), nil
}

func (r *MemoryRepo) ListRewards(_ context.Context, familyID string) ([]*model.Reward, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.rewardsByFamily[familyID]
	res := make([]*model.Reward, 0, len(ids))
	for _, id := range ids {
		if reward, ok := r.rewards[id]; ok {
			res = append(res, cloneReward(reward))
		}
	}
	return res, nil
}

func (r *MemoryRepo) UpdateReward(_ context.Context, familyID, rewardID string, in model.UpdateRewardInput) (*model.Reward, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reward, ok := r.rewards[rewardID]
	if !ok || reward.FamilyID != familyID {
		return nil, errors.New("reward not found")
	}
	reward.Name = in.Name
	reward.XpThreshold = in.XpThreshold
	return cloneReward(reward), nil
}

func (r *MemoryRepo) DeleteReward(_ context.Context, familyID, rewardID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reward, ok := r.rewards[rewardID]
	if !ok || reward.FamilyID != familyID {
		return errors.New("reward not found")
	}

	delete(r.rewards, rewardID)
	r.rewardsByFamily[familyID] = removeString(r.rewardsByFamily[familyID], rewardID)
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

// Demo seed data for dev usage.
func (r *MemoryRepo) SeedDemoData() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.families) > 0 {
		return
	}

	familyID := "family-1"
	ownerID := "parent-1"
	r.families[familyID] = &model.Family{ID: familyID, Name: "Demo Family", OwnerID: ownerID, CreatedAt: NowRFC3339()}

	child := &model.Child{ID: "child-1", FamilyID: familyID, Name: "Alex", Xp: 120, Gold: 45}
	quest1 := &model.Quest{ID: "quest-1", FamilyID: familyID, Title: "Clean Room", Description: strPtr("Tidy up and vacuum"), Xp: 50, Gold: 10}
	quest2 := &model.Quest{ID: "quest-2", FamilyID: familyID, Title: "Do Dishes", Description: strPtr("Load and run dishwasher"), Xp: 30, Gold: 8}
	reward := &model.Reward{ID: "reward-1", FamilyID: familyID, Name: "Movie Night", XpThreshold: 200}

	r.children[child.ID] = child
	r.childrenByFamily[familyID] = []string{child.ID}

	r.quests[quest1.ID] = quest1
	r.quests[quest2.ID] = quest2
	r.questsByFamily[familyID] = []string{quest1.ID, quest2.ID}

	r.rewards[reward.ID] = reward
	r.rewardsByFamily[familyID] = []string{reward.ID}

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

func cloneFamily(in *model.Family) *model.Family {
	if in == nil {
		return nil
	}
	out := *in
	return &out
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
