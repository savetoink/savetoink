package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shaftoe/savetoink/internal/model"
)

// GetUserProfile implements UserProfileRepository.GetUserProfile.
func (d *DynamoDB) GetUserProfile(ctx context.Context, account string) (*model.UserProfile, error) {
	// remove all these checks, validate in the factory call
if d.profileTableName == "" {
		return nil, errors.New("user profile table not configured")
	}

	resp, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.profileTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if resp.Item == nil {
		return nil, nil
	}

	var profile model.UserProfile
	if unmarshalErr := attributevalue.UnmarshalMap(resp.Item, &profile); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal user profile: %w", unmarshalErr)
	}

	return &profile, nil
}

// PutUserProfile implements UserProfileRepository.PutUserProfile.
func (d *DynamoDB) PutUserProfile(ctx context.Context, profile *model.UserProfile) error {
	if d.profileTableName == "" {
		return errors.New("user profile table not configured")
	}

	if profile.Account == "" {
		return errors.New("account field is required")
	}

	item, err := attributevalue.MarshalMap(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %w", err)
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.profileTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to store user profile: %w", err)
	}

	return nil
}

// DeleteUserProfile implements UserProfileRepository.DeleteUserProfile.
func (d *DynamoDB) DeleteUserProfile(ctx context.Context, account string) error {
	if d.profileTableName == "" {
		return errors.New("user profile table not configured")
	}

	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.profileTableName),
		Key: map[string]types.AttributeValue{
			attributeNameAccount: &types.AttributeValueMemberS{Value: account},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user profile: %w", err)
	}

	return nil
}
