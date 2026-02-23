package consts

import (
	"strings"
	"testing"
)

func TestGetRandomUserAgent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"single call"},
		{"multiple calls"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := GetRandomUserAgent()

			if ua == "" {
				t.Error("GetRandomUserAgent() returned empty string")
			}

			if !strings.HasPrefix(ua, "Mozilla/5.0") {
				t.Errorf("GetRandomUserAgent() returned invalid user agent: %s", ua)
			}
		})
	}
}

func TestGetRandomUserAgentVariations(t *testing.T) {
	agents := make(map[string]bool)

	iterations := 100
	for range iterations {
		ua := GetRandomUserAgent()
		agents[ua] = true
	}

	if len(agents) == 1 {
		t.Error("GetRandomUserAgent() appears to always return the same agent")
	}

	for ua := range agents {
		if ua == "" {
			t.Error("Found empty user agent in results")
		}
		if !strings.HasPrefix(ua, "Mozilla/5.0") {
			t.Errorf("Found invalid user agent in results: %s", ua)
		}
	}
}

func TestUserAgents(t *testing.T) {
	for _, ua := range userAgents {
		if ua == "" {
			t.Error("user agent is empty")
		}
		if !strings.HasPrefix(ua, "Mozilla/5.0") {
			t.Errorf("user agent does not start with Mozilla/5.0: %s", ua)
		}
	}
}
