package posts

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)


const PostTableName string = "PostTable"
const PostIDGSI string = "GSI_PostID"

// DynamoDBPostTable is a repository for DynamoDB operations on posts
type DynamoDBPostTable struct {
	dynamoClient *dynamodb.Client
}

// CreatePostTableIfNotExists creates the DynamoDB table with all GSIs and LSIs if it doesn't exist
func CreatePostTableIfNotExists(ctx context.Context, dynamoClient *dynamodb.Client) error {
	// Check if table already exists
	_, err := dynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(PostTableName),
	})
	if err == nil {
		// Table exists, nothing to do
		return nil
	}

	// Table doesn't exist, create it
	_, err = dynamoClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(PostTableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("UserID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("CreatedAt"),
				AttributeType: types.ScalarAttributeTypeN,
			},
			{
				AttributeName: aws.String("PostID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("UserID"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("CreatedAt"),
				KeyType:       types.KeyTypeRange,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String(PostIDGSI),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("PostID"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return fmt.Errorf("failed to create DynamoDB table %s: %w", PostTableName, err)
	}
	return nil
}

// NewDynamoDBPostTable creates a new posts table repository
// It ensures the table exists (creates it if needed) and tests the connection
func NewDynamoDBPostTable(ctx context.Context, dynamoClient *dynamodb.Client) (*DynamoDBPostTable, error) {
	// Ensure table exists (create if it doesn't)
	if err := CreatePostTableIfNotExists(ctx, dynamoClient); err != nil {
		return nil, fmt.Errorf("failed to ensure DynamoDB table %s exists: %w", PostTableName, err)
	}

	// Test connection by describing the table - fail fast if connection fails
	_, err := dynamoClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(PostTableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DynamoDB table %s: %w", PostTableName, err)
	}

	return &DynamoDBPostTable{
		dynamoClient: dynamoClient,
	}, nil
}

func (t *DynamoDBPostTable) PutPost(ctx context.Context, post *Post) error {
	storage := DynamoDBPostToStorage(post)
	valueMap, err := attributevalue.MarshalMap(storage)
	if err != nil {
		return fmt.Errorf("error during PUT to %s: %w", PostTableName, err)
	}
	
	_, err = t.dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		Item:      valueMap,
		TableName: aws.String(PostTableName),
	})
	if err != nil {
		return fmt.Errorf("failed to put post: %w", err)
	}
	return nil
}

// ListPostsByUserID returns all posts authored by the user with id userID
func (t *DynamoDBPostTable) ListPostsByUserID(ctx context.Context, userID uuid.UUID) ([]Post, error) {
	params := &dynamodb.QueryInput{
		TableName: aws.String(PostTableName),
		KeyConditionExpression: aws.String("UserID = :userID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID.String()},
		},
		ScanIndexForward: aws.Bool(false), // Sort by CreatedAt descending
	}

	result, err := t.dynamoClient.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %w", err)
	}

	var storageModels []DynamoDBPostStorageModel
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &storageModels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal posts: %w", err)
	}

	posts := make([]Post, 0, len(storageModels))
	for _, storage := range storageModels {
		post, err := DynamoDBStorageToPost(&storage)
		if err != nil {
			return nil, fmt.Errorf("failed to convert storage to post: %w", err)
		}
		posts = append(posts, *post)
	}

	return posts, nil
}

// GetPostByID retrieves a post by its ID using the GSI_PostID index
func (t *DynamoDBPostTable) GetPostByID(ctx context.Context, postID uuid.UUID) (*Post, error) {
	params := &dynamodb.QueryInput{
		TableName:              aws.String(PostTableName),
		IndexName:              aws.String(PostIDGSI),
		KeyConditionExpression: aws.String("PostID = :postID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":postID": &types.AttributeValueMemberS{Value: postID.String()},
		},
		ConsistentRead: aws.Bool(false),
	}

	result, err := t.dynamoClient.Query(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query post by ID %s: %w", postID, err)
	}

	if len(result.Items) == 0 {
		return nil, ErrPostNotFound
	}

	if len(result.Items) > 1 {
		return nil, fmt.Errorf("multiple posts found with ID %s", postID)
	}

	var storage DynamoDBPostStorageModel
	if err := attributevalue.UnmarshalMap(result.Items[0], &storage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post: %w", err)
	}

	post, err := DynamoDBStorageToPost(&storage)
	if err != nil {
		return nil, fmt.Errorf("failed to convert storage to post: %w", err)
	}

	return post, nil
}

// DeletePost removes a post by post ID
func (t *DynamoDBPostTable) DeletePost(ctx context.Context, postID uuid.UUID) error {
	// First get the post to find its primary key
	post, err := t.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to find post with ID %s for deletion: %w", postID.String(), err)
	}

	// Delete from table using primary key (UserID, CreatedAt)
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(PostTableName),
		Key: map[string]types.AttributeValue{
			"UserID":    &types.AttributeValueMemberS{Value: post.UserID.String()},
			"CreatedAt": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", post.CreatedAt.UnixMilli())},
		},
	}

	_, err = t.dynamoClient.DeleteItem(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}


