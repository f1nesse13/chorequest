package repo

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "github.com/google/uuid"

    "chorequest/backend/graph/model"
)

type DynamoRepo struct {
    DB    *dynamodb.Client
    Table string
}

func NewDynamoRepo(db *dynamodb.Client, table string) *DynamoRepo {
    if table == "" { table = "chorequest" }
    return &DynamoRepo{DB: db, Table: table}
}

type item struct {
    PK       string `dynamodbav:"PK"`
    SK       string `dynamodbav:"SK"`
    Type     string `dynamodbav:"Type"`
    GSI1PK   string `dynamodbav:"GSI1PK,omitempty"`
    GSI1SK   string `dynamodbav:"GSI1SK,omitempty"`
    GSI2PK   string `dynamodbav:"GSI2PK,omitempty"`
    GSI2SK   string `dynamodbav:"GSI2SK,omitempty"`

    // Common
    ParentID string  `dynamodbav:"ParentID,omitempty"`
    ChildID  string  `dynamodbav:"ChildID,omitempty"`
    QuestID  string  `dynamodbav:"QuestID,omitempty"`
    Name     string  `dynamodbav:"Name,omitempty"`
    Title    string  `dynamodbav:"Title,omitempty"`
    Desc     *string `dynamodbav:"Description,omitempty"`
    XP       int     `dynamodbav:"XP,omitempty"`
    Gold     int     `dynamodbav:"Gold,omitempty"`
    XPThresh int     `dynamodbav:"XPThreshold,omitempty"`
    Status   string  `dynamodbav:"Status,omitempty"`
    Created  string  `dynamodbav:"CreatedAt,omitempty"`
    DoneAt   *string `dynamodbav:"CompletedAt,omitempty"`
}

// Key builders
func pkParent(parentID string) string { return "PARENT#" + parentID }
func skChild(childID string) string  { return "CHILD#" + childID }
func skQuest(questID string) string  { return "QUEST#" + questID }
func skReward(rewardID string) string { return "REWARD#" + rewardID }
func pkChild(childID string) string  { return "CHILD#" + childID }
func skAssign(assignID string) string { return "ASSIGN#" + assignID }
func gsi2Key(tag, id string) (string, string) { return tag + "#" + id, "META" }

// Children
func (r *DynamoRepo) CreateChild(ctx context.Context, in model.NewChild) (*model.Child, error) {
    cid := uuid.NewString()
    it := item{
        PK: pkParent(in.ParentID), SK: skChild(cid), Type: "Child",
        ParentID: in.ParentID, Name: in.Name, XP: 0, Gold: 0,
    }
    g2pk, g2sk := gsi2Key("CHILD", cid)
    it.GSI2PK, it.GSI2SK = g2pk, g2sk
    av, _ := attributevalue.MarshalMap(it)
    if _, err := r.DB.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.Table), Item: av, ConditionExpression: aws.String("attribute_not_exists(PK)")}); err != nil {
        return nil, err
    }
    return &model.Child{ID: cid, ParentID: in.ParentID, Name: in.Name, Xp: 0, Gold: 0}, nil
}

func (r *DynamoRepo) ListChildren(ctx context.Context, parentID string) ([]*model.Child, error) {
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: pkParent(parentID)},
            ":sk": &types.AttributeValueMemberS{Value: "CHILD#"},
        },
    })
    if err != nil { return nil, err }
    res := make([]*model.Child, 0, len(out.Items))
    for _, m := range out.Items {
        var it item
        if err := attributevalue.UnmarshalMap(m, &it); err != nil { return nil, err }
        id := strings.TrimPrefix(it.SK, "CHILD#")
        res = append(res, &model.Child{ID: id, ParentID: it.ParentID, Name: it.Name, Xp: it.XP, Gold: it.Gold})
    }
    return res, nil
}

// Quests
func (r *DynamoRepo) CreateQuest(ctx context.Context, in model.NewQuest) (*model.Quest, error) {
    qid := uuid.NewString()
    it := item{PK: pkParent(in.ParentID), SK: skQuest(qid), Type: "Quest", ParentID: in.ParentID, Title: in.Title, Desc: in.Description, XP: in.Xp, Gold: in.Gold}
    g2pk, g2sk := gsi2Key("QUEST", qid)
    it.GSI2PK, it.GSI2SK = g2pk, g2sk
    av, _ := attributevalue.MarshalMap(it)
    if _, err := r.DB.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.Table), Item: av, ConditionExpression: aws.String("attribute_not_exists(PK)")}); err != nil {
        return nil, err
    }
    return &model.Quest{ID: qid, ParentID: in.ParentID, Title: in.Title, Description: in.Description, Xp: in.Xp, Gold: in.Gold}, nil
}

func (r *DynamoRepo) ListQuests(ctx context.Context, parentID string) ([]*model.Quest, error) {
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: pkParent(parentID)},
            ":sk": &types.AttributeValueMemberS{Value: "QUEST#"},
        },
    })
    if err != nil { return nil, err }
    res := make([]*model.Quest, 0, len(out.Items))
    for _, m := range out.Items {
        var it item
        if err := attributevalue.UnmarshalMap(m, &it); err != nil { return nil, err }
        id := strings.TrimPrefix(it.SK, "QUEST#")
        res = append(res, &model.Quest{ID: id, ParentID: it.ParentID, Title: it.Title, Description: it.Desc, Xp: it.XP, Gold: it.Gold})
    }
    return res, nil
}

func (r *DynamoRepo) GetQuestByID(ctx context.Context, questID string) (*model.Quest, error) {
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        IndexName:              aws.String("GSI2"),
        KeyConditionExpression: aws.String("GSI2PK = :pk AND GSI2SK = :sk"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: "QUEST#" + questID},
            ":sk": &types.AttributeValueMemberS{Value: "META"},
        },
        Limit: aws.Int32(1),
    })
    if err != nil { return nil, err }
    if len(out.Items) == 0 { return nil, errors.New("quest not found") }
    var it item
    if err := attributevalue.UnmarshalMap(out.Items[0], &it); err != nil { return nil, err }
    id := strings.TrimPrefix(it.SK, "QUEST#")
    return &model.Quest{ID: id, ParentID: it.ParentID, Title: it.Title, Description: it.Desc, Xp: it.XP, Gold: it.Gold}, nil
}

// Rewards
func (r *DynamoRepo) CreateReward(ctx context.Context, in model.NewReward) (*model.Reward, error) {
    rid := uuid.NewString()
    it := item{PK: pkParent(in.ParentID), SK: skReward(rid), Type: "Reward", ParentID: in.ParentID, Name: in.Name, XPThresh: in.XpThreshold}
    av, _ := attributevalue.MarshalMap(it)
    if _, err := r.DB.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.Table), Item: av, ConditionExpression: aws.String("attribute_not_exists(PK)")}); err != nil {
        return nil, err
    }
    return &model.Reward{ID: rid, ParentID: in.ParentID, Name: in.Name, XpThreshold: in.XpThreshold}, nil
}

func (r *DynamoRepo) ListRewards(ctx context.Context, parentID string) ([]*model.Reward, error) {
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: pkParent(parentID)},
            ":sk": &types.AttributeValueMemberS{Value: "REWARD#"},
        },
    })
    if err != nil { return nil, err }
    res := make([]*model.Reward, 0, len(out.Items))
    for _, m := range out.Items {
        var it item
        if err := attributevalue.UnmarshalMap(m, &it); err != nil { return nil, err }
        id := strings.TrimPrefix(it.SK, "REWARD#")
        res = append(res, &model.Reward{ID: id, ParentID: it.ParentID, Name: it.Name, XpThreshold: it.XPThresh})
    }
    return res, nil
}

// Assignments
func (r *DynamoRepo) AssignQuest(ctx context.Context, questID, childID string) (*model.Assignment, error) {
    q, err := r.GetQuestByID(ctx, questID)
    if err != nil { return nil, err }
    aid := uuid.NewString()
    it := item{
        PK: pkChild(childID), SK: skAssign(aid), Type: "Assignment",
        ChildID: childID, QuestID: questID, Status: "ASSIGNED", Created: NowRFC3339(),
        GSI1PK: "QUEST#" + questID, GSI1SK: "ASSIGN#" + aid,
    }
    av, _ := attributevalue.MarshalMap(it)
    if _, err := r.DB.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.Table), Item: av, ConditionExpression: aws.String("attribute_not_exists(PK)")}); err != nil {
        return nil, err
    }
    return &model.Assignment{ID: aid, Quest: q, ChildID: childID, Status: it.Status, CreatedAt: it.Created}, nil
}

func (r *DynamoRepo) ListAssignmentsForChild(ctx context.Context, childID string) ([]*model.Assignment, error) {
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName: aws.String(r.Table),
        KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: pkChild(childID)},
            ":sk": &types.AttributeValueMemberS{Value: "ASSIGN#"},
        },
    })
    if err != nil { return nil, err }
    res := make([]*model.Assignment, 0, len(out.Items))
    for _, m := range out.Items {
        var it item
        if err := attributevalue.UnmarshalMap(m, &it); err != nil { return nil, err }
        q, _ := r.GetQuestByID(ctx, it.QuestID)
        id := strings.TrimPrefix(it.SK, "ASSIGN#")
        res = append(res, &model.Assignment{ID: id, Quest: q, ChildID: it.ChildID, Status: it.Status, CreatedAt: it.Created, CompletedAt: it.DoneAt})
    }
    return res, nil
}

func (r *DynamoRepo) CompleteAssignment(ctx context.Context, assignmentID string) (*model.Assignment, error) {
    // Lookup assignment via GSI1 by ID
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        IndexName:              aws.String("GSI1"),
        KeyConditionExpression: aws.String("GSI1SK = :sk"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":sk": &types.AttributeValueMemberS{Value: "ASSIGN#" + assignmentID},
        },
        Limit: aws.Int32(1),
    })
    if err != nil { return nil, err }
    if len(out.Items) == 0 { return nil, errors.New("assignment not found") }
    var it item
    if err := attributevalue.UnmarshalMap(out.Items[0], &it); err != nil { return nil, err }

    // Get quest and child item (via GSI2)
    q, err := r.GetQuestByID(ctx, it.QuestID)
    if err != nil { return nil, err }
    chq, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        IndexName:              aws.String("GSI2"),
        KeyConditionExpression: aws.String("GSI2PK = :pk AND GSI2SK = :sk"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: "CHILD#" + strings.TrimPrefix(it.PK, "CHILD#")},
            ":sk": &types.AttributeValueMemberS{Value: "META"},
        },
        Limit: aws.Int32(1),
    })
    if err != nil { return nil, err }
    if len(chq.Items) == 0 { return nil, errors.New("child not found") }
    var ch item
    if err := attributevalue.UnmarshalMap(chq.Items[0], &ch); err != nil { return nil, err }

    done := NowRFC3339()
    // Transaction: mark assignment completed if not already, and add XP/Gold to child
    _, err = r.DB.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
        TransactItems: []types.TransactWriteItem{
            { Update: &types.Update{ TableName: aws.String(r.Table),
                Key: map[string]types.AttributeValue{
                    "PK": &types.AttributeValueMemberS{Value: it.PK},
                    "SK": &types.AttributeValueMemberS{Value: it.SK},
                },
                UpdateExpression:          aws.String("SET #S = :s, CompletedAt = :d"),
                ConditionExpression:       aws.String("attribute_not_exists(CompletedAt) AND #S <> :s"),
                ExpressionAttributeNames:  map[string]string{"#S": "Status"},
                ExpressionAttributeValues: map[string]types.AttributeValue{":s": &types.AttributeValueMemberS{Value: "COMPLETED"}, ":d": &types.AttributeValueMemberS{Value: done}},
            }},
            { Update: &types.Update{ TableName: aws.String(r.Table),
                Key: map[string]types.AttributeValue{
                    "PK": &types.AttributeValueMemberS{Value: ch.PK},
                    "SK": &types.AttributeValueMemberS{Value: ch.SK},
                },
                UpdateExpression:          aws.String("ADD XP :xp, Gold :g"),
                ExpressionAttributeValues: map[string]types.AttributeValue{":xp": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", q.Xp)}, ":g": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", q.Gold)}},
            }},
        },
    })
    if err != nil { return nil, err }

    id := strings.TrimPrefix(it.SK, "ASSIGN#")
    return &model.Assignment{ID: id, Quest: q, ChildID: it.ChildID, Status: "COMPLETED", CreatedAt: it.Created, CompletedAt: &done}, nil
}

func (r *DynamoRepo) PurchaseItem(ctx context.Context, childID, itemName string, priceGold int) (*model.Child, error) {
    // Load child via GSI2
    out, err := r.DB.Query(ctx, &dynamodb.QueryInput{
        TableName:              aws.String(r.Table),
        IndexName:              aws.String("GSI2"),
        KeyConditionExpression: aws.String("GSI2PK = :pk AND GSI2SK = :sk"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":pk": &types.AttributeValueMemberS{Value: "CHILD#" + childID},
            ":sk": &types.AttributeValueMemberS{Value: "META"},
        },
        Limit: aws.Int32(1),
    })
    if err != nil { return nil, err }
    if len(out.Items) == 0 { return nil, errors.New("child not found") }
    var it item
    if err := attributevalue.UnmarshalMap(out.Items[0], &it); err != nil { return nil, err }

    // Attempt spend
    // We don't enforce non-negative here; add ConditionExpression if desired.
    _, err = r.DB.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(r.Table),
        Key:       map[string]types.AttributeValue{"PK": &types.AttributeValueMemberS{Value: it.PK}, "SK": &types.AttributeValueMemberS{Value: it.SK}},
        UpdateExpression: aws.String("ADD Gold :delta"),
        ConditionExpression: aws.String("Gold >= :cost"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":delta": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", -priceGold)},
            ":cost":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", priceGold)},
        },
        ReturnValues: types.ReturnValueAllNew,
    })
    if err != nil { return nil, err }

    // Return updated projection (best effort minimal fields)
    return &model.Child{ID: childID, ParentID: it.ParentID, Name: it.Name, Xp: it.XP, Gold: it.Gold - priceGold}, nil
}
