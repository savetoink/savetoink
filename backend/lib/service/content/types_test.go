package content

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessArticleEvent_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		event    ProcessArticleEvent
		wantJSON string
	}{
		{
			name: "full event with all fields",
			event: ProcessArticleEvent{
				RequestID:      "req-123",
				URL:            "https://example.com/article",
				ArticleID:      "article-456",
				AccountID:      "account-789",
				InheritedAttrs: []map[string]any{{"key1": "value1"}, {"key2": "value2"}},
				SendOnComplete: true,
			},
			wantJSON: `{"request_id":"req-123","url":"https://example.com/article",` +
				`"article_id":"article-456","account_id":"account-789",` +
				`"inherited_attrs":[{"key1":"value1"},{"key2":"value2"}],"send_on_complete":true}`,
		},
		{
			name: "minimal event",
			event: ProcessArticleEvent{
				RequestID: "req-123",
				URL:       "https://example.com/article",
				ArticleID: "article-456",
				AccountID: "account-789",
			},
			wantJSON: `{"request_id":"req-123","url":"https://example.com/article",` +
				`"article_id":"article-456","account_id":"account-789",` +
				`"inherited_attrs":null,"send_on_complete":false}`,
		},
		{
			name: "event with empty inherited attrs",
			event: ProcessArticleEvent{
				RequestID:      "req-123",
				URL:            "https://example.com/article",
				ArticleID:      "article-456",
				AccountID:      "account-789",
				InheritedAttrs: []map[string]any{},
				SendOnComplete: false,
			},
			wantJSON: `{"request_id":"req-123","url":"https://example.com/article",` +
				`"article_id":"article-456","account_id":"account-789",` +
				`"inherited_attrs":[],"send_on_complete":false}`,
		},
		{
			name: "event with complex inherited attrs",
			event: ProcessArticleEvent{
				RequestID: "req-123",
				URL:       "https://example.com/article",
				ArticleID: "article-456",
				AccountID: "account-789",
				InheritedAttrs: []map[string]any{
					{
						"string_key": "string_value",
						"number_key": 42,
						"bool_key":   true,
						"nested_key": map[string]any{"nested": "value"},
					},
				},
				SendOnComplete: true,
			},
			wantJSON: `{"request_id":"req-123","url":"https://example.com/article",` +
				`"article_id":"article-456","account_id":"account-789",` +
				`"inherited_attrs":[{"string_key":"string_value","number_key":42,` +
				`"bool_key":true,"nested_key":{"nested":"value"}}],"send_on_complete":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.event)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(data))

			var unmarshaled ProcessArticleEvent
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			assert.Equal(t, tt.event.RequestID, unmarshaled.RequestID)
			assert.Equal(t, tt.event.URL, unmarshaled.URL)
			assert.Equal(t, tt.event.ArticleID, unmarshaled.ArticleID)
			assert.Equal(t, tt.event.AccountID, unmarshaled.AccountID)
			assert.Equal(t, tt.event.SendOnComplete, unmarshaled.SendOnComplete)
			assert.Equal(t, len(tt.event.InheritedAttrs), len(unmarshaled.InheritedAttrs))
		})
	}
}

func TestProcessArticleEvent_NilInheritedAttrs(t *testing.T) {
	event := ProcessArticleEvent{
		RequestID:      "req-123",
		URL:            "https://example.com/article",
		ArticleID:      "article-456",
		AccountID:      "account-789",
		InheritedAttrs: nil,
		SendOnComplete: false,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled ProcessArticleEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.RequestID, unmarshaled.RequestID)
	assert.Equal(t, event.URL, unmarshaled.URL)
	assert.Equal(t, event.ArticleID, unmarshaled.ArticleID)
	assert.Equal(t, event.AccountID, unmarshaled.AccountID)
	assert.Equal(t, event.SendOnComplete, unmarshaled.SendOnComplete)
	assert.Nil(t, unmarshaled.InheritedAttrs)
}

func TestProcessArticleEvent_EmptyInheritedAttrs(t *testing.T) {
	event := ProcessArticleEvent{
		RequestID:      "req-123",
		URL:            "https://example.com/article",
		ArticleID:      "article-456",
		AccountID:      "account-789",
		InheritedAttrs: []map[string]any{},
		SendOnComplete: false,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled ProcessArticleEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.RequestID, unmarshaled.RequestID)
	assert.Equal(t, event.URL, unmarshaled.URL)
	assert.Equal(t, event.ArticleID, unmarshaled.ArticleID)
	assert.Equal(t, event.AccountID, unmarshaled.AccountID)
	assert.Equal(t, event.SendOnComplete, unmarshaled.SendOnComplete)
	assert.NotNil(t, unmarshaled.InheritedAttrs)
	assert.Empty(t, unmarshaled.InheritedAttrs)
}

func TestProcessArticleEvent_InheritedAttrsComplex(t *testing.T) {
	event := ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-456",
		AccountID: "account-789",
		InheritedAttrs: []map[string]any{
			{
				"trace_id":    "trace-123",
				"user_agent":  "Mozilla/5.0",
				"request_ts":  "2024-03-15T10:30:00Z",
				"client_ip":   "192.168.1.1",
				"request_id":  "req-123",
				"correlation": "corr-456",
			},
		},
		SendOnComplete: true,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled ProcessArticleEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.RequestID, unmarshaled.RequestID)
	assert.Equal(t, event.URL, unmarshaled.URL)
	assert.Equal(t, event.ArticleID, unmarshaled.ArticleID)
	assert.Equal(t, event.AccountID, unmarshaled.AccountID)
	assert.Equal(t, event.SendOnComplete, unmarshaled.SendOnComplete)

	require.Len(t, unmarshaled.InheritedAttrs, 1)
	assert.Equal(t, "trace-123", unmarshaled.InheritedAttrs[0]["trace_id"])
	assert.Equal(t, "Mozilla/5.0", unmarshaled.InheritedAttrs[0]["user_agent"])
	assert.Equal(t, "192.168.1.1", unmarshaled.InheritedAttrs[0]["client_ip"])
	assert.Equal(t, "req-123", unmarshaled.InheritedAttrs[0]["request_id"])
}

func TestProcessArticleEvent_EmptyFields(t *testing.T) {
	event := ProcessArticleEvent{
		RequestID: "req-123",
		URL:       "https://example.com/article",
		ArticleID: "article-456",
		AccountID: "account-789",
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled ProcessArticleEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, event.RequestID, unmarshaled.RequestID)
	assert.Equal(t, event.URL, unmarshaled.URL)
	assert.Equal(t, event.ArticleID, unmarshaled.ArticleID)
	assert.Equal(t, event.AccountID, unmarshaled.AccountID)
	assert.Nil(t, unmarshaled.InheritedAttrs)
	assert.False(t, unmarshaled.SendOnComplete)
}
