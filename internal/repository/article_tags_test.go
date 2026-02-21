// Package repository provides article tag index persistence tests.
package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockArticleTags struct {
	addTagIndexCalls        [][]string
	removeTagIndexCalls     [][]string
	getArticleIDsByTagCalls [][]string
	getArticlesByTagsCalls  [][][]string
}

func (m *MockArticleTags) AddTagIndex(_ context.Context, account, articleID, tag, createdAt string) error {
	for _, call := range m.addTagIndexCalls {
		if call[0] == account && call[1] == articleID && call[2] == tag {
			return errors.New("duplicate tag")
		}
	}
	m.addTagIndexCalls = append(m.addTagIndexCalls, []string{account, articleID, tag, createdAt})
	return nil
}

func (m *MockArticleTags) RemoveTagIndex(_ context.Context, account, articleID, tag, createdAt string) error {
	m.removeTagIndexCalls = append(m.removeTagIndexCalls, []string{account, articleID, tag, createdAt})
	return nil
}

func (m *MockArticleTags) GetArticleIDsByTag(_ context.Context, account, tag string, page, pageSize int) ([]string, map[string]interface{}, int, error) {
	m.getArticleIDsByTagCalls = append(m.getArticleIDsByTagCalls, []string{account, tag})
	ids := []string{"article-1", "article-2", "article-3"}
	emptyMap := make(map[string]interface{})
	return ids, emptyMap, 3, nil
}

func (m *MockArticleTags) GetArticlesByTags(ctx context.Context, account string, tags []string, page, pageSize int) ([]string, int, error) {
	tagsCopy := make([]string, len(tags))
	copy(tagsCopy, tags)
	m.getArticlesByTagsCalls = append(m.getArticlesByTagsCalls, [][]string{tagsCopy})
	return []string{"article-1"}, 1, nil
}

func NewMockArticleTags() *MockArticleTags {
	return &MockArticleTags{}
}

func TestArticleTags_AddTagIndex(t *testing.T) {
	mock := NewMockArticleTags()
	account := "test-account"
	articleID := "test-article-id"
	tag := "golang"
	createdAt := "2024-01-15T10:00:00Z"

	err := mock.AddTagIndex(context.Background(), account, articleID, tag, createdAt)
	require.NoError(t, err)
	require.Len(t, mock.addTagIndexCalls, 1)
	require.Equal(t, account, mock.addTagIndexCalls[0][0])
	require.Equal(t, articleID, mock.addTagIndexCalls[0][1])
	require.Equal(t, tag, mock.addTagIndexCalls[0][2])
	require.Equal(t, createdAt, mock.addTagIndexCalls[0][3])
}

func TestArticleTags_AddTagIndex_Duplicate(t *testing.T) {
	mock := NewMockArticleTags()
	account := "test-account"
	articleID := "test-article-id"
	tag := "golang"
	createdAt := "2024-01-15T10:00:00Z"

	_ = mock.AddTagIndex(context.Background(), account, articleID, tag, createdAt)
	err := mock.AddTagIndex(context.Background(), account, articleID, tag, createdAt)
	require.Error(t, err)
}

func TestArticleTags_RemoveTagIndex(t *testing.T) {
	mock := NewMockArticleTags()
	account := "test-account"
	articleID := "test-article-id"
	tag := "golang"
	createdAt := "2024-01-15T10:00:00Z"

	_ = mock.AddTagIndex(context.Background(), account, articleID, tag, createdAt)

	err := mock.RemoveTagIndex(context.Background(), account, articleID, tag, createdAt)
	require.NoError(t, err)

	require.Len(t, mock.removeTagIndexCalls, 1)
	require.Equal(t, account, mock.removeTagIndexCalls[0][0])
	require.Equal(t, articleID, mock.removeTagIndexCalls[0][1])
	require.Equal(t, tag, mock.removeTagIndexCalls[0][2])
	require.Equal(t, createdAt, mock.removeTagIndexCalls[0][3])
}

func TestArticleTags_GetArticleIDsByTag(t *testing.T) {
	mock := NewMockArticleTags()
	account := "test-account"
	tag := "golang"

	ids, _, total, err := mock.GetArticleIDsByTag(context.Background(), account, tag, 1, 2)
	require.NoError(t, err)
	require.Equal(t, 3, len(ids))
	require.Equal(t, 3, total)
}

func TestArticleTags_GetArticlesByTags_SingleTag(t *testing.T) {
	mock := NewMockArticleTags()
	account := "test-account"
	tag := "golang"

	ids, total, err := mock.GetArticlesByTags(context.Background(), account, []string{tag}, 1, 10)
	require.NoError(t, err)
	require.Equal(t, 1, len(ids))
	require.Equal(t, 1, total)
}
