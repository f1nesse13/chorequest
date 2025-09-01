package db

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// EnsureSingleTable creates a generic single-table model suitable for a wide range of entities.
// Keys: PK, SK (both strings). GSIs: GSI1(PK/SK), GSI2(PK/SK)
// Table name is env DYNAMO_TABLE_NAME or provided name (fallback: chorequest)
func EnsureSingleTable(ctx context.Context, c *Client, name string) error {
    if name == "" {
        name = os.Getenv("DYNAMO_TABLE_NAME")
    }
    if name == "" {
        name = "chorequest"
    }

    // Check if table exists
    _, err := c.Dynamo.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(name)})
    if err == nil {
        return nil
    }

    // Create
    _, err = c.Dynamo.CreateTable(ctx, &dynamodb.CreateTableInput{
        TableName: aws.String(name),
        AttributeDefinitions: []types.AttributeDefinition{
            {AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI2PK"), AttributeType: types.ScalarAttributeTypeS},
            {AttributeName: aws.String("GSI2SK"), AttributeType: types.ScalarAttributeTypeS},
        },
        KeySchema: []types.KeySchemaElement{
            {AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
            {AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
        },
        BillingMode: types.BillingModePayPerRequest,
        GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
            {
                IndexName: aws.String("GSI1"),
                KeySchema: []types.KeySchemaElement{
                    {AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
                    {AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
                },
                Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
            },
            {
                IndexName: aws.String("GSI2"),
                KeySchema: []types.KeySchemaElement{
                    {AttributeName: aws.String("GSI2PK"), KeyType: types.KeyTypeHash},
                    {AttributeName: aws.String("GSI2SK"), KeyType: types.KeyTypeRange},
                },
                Projection: &types.Projection{ProjectionType: types.ProjectionTypeAll},
            },
        },
    })
    if err != nil {
        return fmt.Errorf("create table: %w", err)
    }

    // Wait until active
    for i := 0; i < 30; i++ {
        out, err := c.Dynamo.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(name)})
        if err == nil && out.Table != nil && out.Table.TableStatus == types.TableStatusActive {
            return nil
        }
        time.Sleep(2 * time.Second)
    }
    return fmt.Errorf("table %s not active in time", name)
}

