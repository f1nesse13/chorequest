package db

import (
    "context"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Client wraps AWS DynamoDB client for app usage.
type Client struct {
    Dynamo *dynamodb.Client
}

// New creates a DynamoDB client. If DYNAMODB_ENDPOINT is provided, it uses it (e.g., for local dev).
func New(ctx context.Context) (*Client, error) {
    // Load default config
    cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
        if ep := os.Getenv("DYNAMODB_ENDPOINT"); ep != "" {
            o.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
                return aws.Endpoint{URL: ep, HostnameImmutable: true}, nil
            })
        }
        if r := os.Getenv("AWS_REGION"); r != "" {
            o.Region = r
        } else {
            o.Region = "us-east-1"
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return &Client{Dynamo: dynamodb.NewFromConfig(cfg)}, nil
}

// Helper exports to avoid unused imports while scaffolding.
var _ = attributevalue.Marshal

