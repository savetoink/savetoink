package content

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAppJSON                 = "application/json"
	testURLHTTPS                = "https://example.com"
	testURLHTTP                 = "http://example.com"
	testURLFTP                  = "ftp://example.com"
	testURLFile                 = "file://example.com"
	testURLEmptyHost            = "http://"
	testURLNoHost               = "http:"
	errMsgURLMustUseHTTPOrHTTPS = "url must use http or https scheme"
	errMsgURLMustHaveHost       = "url must have host"
	errMsgContentNotValidHTML   = "content does not appear to be valid HTML"
	errMsgContentErrorPage      = "content appears to be an error page"
)

type mockRoundTripper struct {
	firstResp  *http.Response
	firstErr   error
	secondResp *http.Response
	secondErr  error
	callCount  int
}

type errorReader struct{}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, assert.AnError
}

func (e *errorReader) Close() error {
	return nil
}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	m.callCount++
	if m.callCount == 1 {
		return m.firstResp, m.firstErr
	}
	return m.secondResp, m.secondErr
}

func TestFetcherType_String(t *testing.T) {
	tests := []struct {
		name     string
		fetcher  FetcherType
		expected string
	}{
		{fetcherTypeGoStr, FetcherTypeGo, fetcherTypeGoStr},
		{fetcherTypeBrowserlessStr, FetcherTypeBrowserless, fetcherTypeBrowserlessStr},
		{fetcherTypeUnknownStr, FetcherType(999), fetcherTypeUnknownStr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.fetcher.String())
		})
	}
}

func TestValidateParsedURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantErr   bool
		errString string
	}{
		{"valid https", testURLHTTPS, false, ""},
		{"valid http", testURLHTTP, false, ""},
		{"invalid scheme ftp", testURLFTP, true, errMsgURLMustUseHTTPOrHTTPS},
		{"invalid scheme file", testURLFile, true, errMsgURLMustUseHTTPOrHTTPS},
		{"empty host", testURLEmptyHost, true, errMsgURLMustHaveHost},
		{"no host", testURLNoHost, true, errMsgURLMustHaveHost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			require.NoError(t, err)

			err = validateParsedURL(u)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Equal(t, tt.errString, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHTMLContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantErr   bool
		errString string
	}{
		{
			name:      "valid html with doctype",
			content:   "<!DOCTYPE html><html><body>test</body></html>",
			wantErr:   false,
			errString: "",
		},
		{
			name:      "valid html lowercase",
			content:   "<html><body>test</body></html>",
			wantErr:   false,
			errString: "",
		},
		{
			name:      "valid html uppercase",
			content:   "<HTML><BODY>test</BODY></HTML>",
			wantErr:   false,
			errString: "",
		},
		{
			name:      "not html",
			content:   "just some plain text",
			wantErr:   true,
			errString: errMsgContentNotValidHTML,
		},
		{
			name: "error page pattern",
			content: "<html><body>This website is using a security service to protect " +
				"itself from online attacks. The action you just performed triggered " +
				"the security solution.</body></html>",
			wantErr:   true,
			errString: errMsgContentErrorPage,
		},
		{
			name: "error page pattern case insensitive",
			content: "<html><body>THIS WEBSITE IS USING A SECURITY SERVICE TO " +
				"PROTECT ITSELF FROM ONLINE ATTACKS. THE ACTION YOU JUST PERFORMED " +
				"TRIGGERED THE SECURITY SOLUTION.</body></html>",
			wantErr:   true,
			errString: errMsgContentErrorPage,
		},
		{
			name:      "empty content",
			content:   "",
			wantErr:   true,
			errString: errMsgContentNotValidHTML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHTMLContent([]byte(tt.content))
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetcher_Fetch(t *testing.T) {
	validURL, _ := url.Parse("https://example.com/article")
	validHTML := "<!DOCTYPE html><html><body>Test Content</body></html>"

	t.Run("invalid URL scheme", func(t *testing.T) {
		fetcher := NewFetcher("")
		u, _ := url.Parse("ftp://example.com")

		_, err := fetcher.Fetch(context.Background(), u)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid URL")
	})

	t.Run("invalid URL empty host", func(t *testing.T) {
		fetcher := NewFetcher("")
		u, _ := url.Parse("http://")

		_, err := fetcher.Fetch(context.Background(), u)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid URL")
	})

	t.Run("HTTP client error without browserless", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstErr: assert.AnError,
				},
			},
			browserlessKey: "",
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch URL")
	})

	t.Run("non-200 status without browserless", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
			},
			browserlessKey: "",
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code: 404")
	})

	t.Run("non-HTML content type without browserless", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
				},
			},
			browserlessKey: "",
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected HTML content")
	})

	t.Run("successful fetch with Go client", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(validHTML)),
					},
				},
			},
			browserlessKey: "",
		}

		content, err := fetcher.Fetch(context.Background(), validURL)
		require.NoError(t, err)
		assert.Equal(t, FetcherTypeGo, content.Type)
		assert.Equal(t, validURL, content.URL)

		html, err := io.ReadAll(content.HTML)
		require.NoError(t, err)
		assert.Equal(t, validHTML, string(html))
	})

	t.Run("HTTP client error with browserless fallback", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstErr: assert.AnError,
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(validHTML)),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		content, err := fetcher.Fetch(context.Background(), validURL)
		require.NoError(t, err)
		assert.Equal(t, FetcherTypeBrowserless, content.Type)
	})

	t.Run("non-200 status with browserless fallback", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(strings.NewReader("")),
					},
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(validHTML)),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		content, err := fetcher.Fetch(context.Background(), validURL)
		require.NoError(t, err)
		assert.Equal(t, FetcherTypeBrowserless, content.Type)
	})

	t.Run("non-HTML content type with browserless fallback", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(validHTML)),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		content, err := fetcher.Fetch(context.Background(), validURL)
		require.NoError(t, err)
		assert.Equal(t, FetcherTypeBrowserless, content.Type)
	})

	t.Run("browserless fetch HTTP error", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondErr: assert.AnError,
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(validHTML)),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "browserless")
	})

	t.Run("browserless returns non-200 status", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondResp: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "browserless returned status code: 500")
	})

	t.Run("browserless returns invalid HTML", func(t *testing.T) {
		invalidHTML := notHTML

		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(invalidHTML)),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "browserless returned invalid content")
	})

	t.Run("successful browserless fallback", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       io.NopCloser(strings.NewReader(validHTML)),
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		content, err := fetcher.Fetch(context.Background(), validURL)
		require.NoError(t, err)
		assert.Equal(t, FetcherTypeBrowserless, content.Type)
		assert.Equal(t, validURL, content.URL)

		html, err := io.ReadAll(content.HTML)
		require.NoError(t, err)
		assert.Equal(t, validHTML, string(html))
	})

	t.Run("context canceled during fetch", func(t *testing.T) {
		fetcher := NewFetcher("")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := fetcher.Fetch(ctx, validURL)
		assert.Error(t, err)
	})

	t.Run("context canceled during browserless fetch", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondErr: context.Canceled,
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "browserless")
	})

	t.Run("browserless returns error reading body", func(t *testing.T) {
		fetcher := &Fetcher{
			client: &http.Client{
				Transport: &mockRoundTripper{
					firstResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testAppJSON}},
						Body:       io.NopCloser(strings.NewReader("{}")),
					},
					secondResp: &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{testContentType: []string{testTextHTML}},
						Body:       &errorReader{},
					},
				},
			},
			browserlessKey: testBrowserlessKey,
		}

		_, err := fetcher.Fetch(context.Background(), validURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read browserless response")
	})
}

func TestFetcher_NewFetcher(t *testing.T) {
	t.Run("creates fetcher without browserless key", func(t *testing.T) {
		fetcher := NewFetcher("")
		assert.NotNil(t, fetcher)
		assert.NotNil(t, fetcher.client)
		assert.Empty(t, fetcher.browserlessKey)
	})

	t.Run("creates fetcher with browserless key", func(t *testing.T) {
		fetcher := NewFetcher(testBrowserlessKey)
		assert.NotNil(t, fetcher)
		assert.NotNil(t, fetcher.client)
		assert.Equal(t, testBrowserlessKey, fetcher.browserlessKey)
	})
}
