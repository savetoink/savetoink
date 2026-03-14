package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamValue_String(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		p := StringParam("test value")
		s, err := p.String()
		require.NoError(t, err)
		assert.Equal(t, "test value", s)
	})

	t.Run("nil param", func(t *testing.T) {
		var p ParamValue
		_, err := p.String()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parameter is nil")
	})
}

func TestRequireString(t *testing.T) {
	t.Run("existing string param", func(t *testing.T) {
		params := map[string]ParamValue{
			"key1": StringParam("value1"),
		}
		s, err := RequireString(params, "key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", s)
	})

	t.Run("missing required param", func(t *testing.T) {
		params := map[string]ParamValue{
			"key1": StringParam("value1"),
		}
		_, err := RequireString(params, "key2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required parameter 'key2' not found")
	})
}

func TestOptionalString(t *testing.T) {
	t.Run("existing string param", func(t *testing.T) {
		params := map[string]ParamValue{
			"key1": StringParam("value1"),
		}
		s, err := OptionalString(params, "key1", "default")
		require.NoError(t, err)
		assert.Equal(t, "value1", s)
	})

	t.Run("missing optional param returns default", func(t *testing.T) {
		params := map[string]ParamValue{}
		s, err := OptionalString(params, "key1", "default")
		require.NoError(t, err)
		assert.Equal(t, "default", s)
	})
}
