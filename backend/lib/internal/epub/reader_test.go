package epub

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewReader(t *testing.T) {
	reader := NewReader()
	require.NotNil(t, reader, "NewReader should return a non-nil Reader")
	require.NotNil(t, reader.client, "Reader client should be initialized")
}

func TestReadFromURL_FileScheme(t *testing.T) {
	reader := NewReader()
	ctx := context.Background()

	testDataPath := filepath.Join("..", "..", "..", "cli", "savetoink", "e2e", "testdata", "article.orig.epub")
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("test EPUB file not found")
	}

	u, err := url.Parse("file://" + testDataPath)
	require.NoError(t, err, "URL parse should succeed")

	epubReader, title, err := reader.ReadFromURL(ctx, u)
	require.NoError(t, err, "ReadFromURL should succeed")
	require.NotEmpty(t, title, "title should not be empty")
	require.NotNil(t, epubReader, "reader should not be nil")

	// Verify that we can read from the reader
	data, err := io.ReadAll(epubReader)
	require.NoError(t, err, "reading epub data should succeed")
	require.Greater(t, len(data), 0, "EPUB data should not be empty")

	err = epubReader.Close()
	require.NoError(t, err, "closing reader should succeed")
}

func TestReadFromURL_FileScheme_NonExistent(t *testing.T) {
	reader := NewReader()
	ctx := context.Background()

	u, err := url.Parse("file:///nonexistent/file.epub")
	require.NoError(t, err, "URL parse should succeed")

	_, _, err = reader.ReadFromURL(ctx, u)
	require.Error(t, err, "ReadFromURL should fail for non-existent file")
	require.Contains(t, err.Error(), "failed to stat file", "error message should indicate file stat failure")
}

func TestReadFromURL_FileScheme_Directory(t *testing.T) {
	reader := NewReader()
	ctx := context.Background()

	// Use /tmp which should be a directory
	u, err := url.Parse("file:///tmp")
	require.NoError(t, err, "URL parse should succeed")

	_, _, err = reader.ReadFromURL(ctx, u)
	require.Error(t, err, "ReadFromURL should fail for directory")
	require.Contains(t, err.Error(), "path is a directory", "error message should indicate it's a directory")
}

func TestReadFromURL_FileScheme_InvalidEPUB(t *testing.T) {
	reader := NewReader()
	ctx := context.Background()

	// Create a temporary non-EPUB file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err, "creating temp file should succeed")
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, err = tmpFile.WriteString("This is not an EPUB file")
	require.NoError(t, err, "writing to temp file should succeed")
	_ = tmpFile.Close()

	u, err := url.Parse("file://" + tmpFile.Name())
	require.NoError(t, err, "URL parse should succeed")

	_, _, err = reader.ReadFromURL(ctx, u)
	require.Error(t, err, "ReadFromURL should fail for non-EPUB file")
	require.Contains(t, err.Error(), "failed to open epub as zip", "error message should indicate zip failure")
}

func TestReadFromURL_UnsupportedScheme(t *testing.T) {
	reader := NewReader()
	ctx := context.Background()

	u, err := url.Parse("ftp://example.com/file.epub")
	require.NoError(t, err, "URL parse should succeed")

	_, _, err = reader.ReadFromURL(ctx, u)
	require.Error(t, err, "ReadFromURL should fail for unsupported scheme")
	require.Contains(t, err.Error(), "unsupported url scheme", "error message should indicate unsupported scheme")
}

func TestReadFromFile(t *testing.T) {
	reader := NewReader()

	t.Run("successfully reads file", func(t *testing.T) {
		testDataPath := filepath.Join("..", "..", "..", "cli", "savetoink", "e2e", "testdata", "article.orig.epub")
		if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
			t.Skip("test EPUB file not found")
		}

		u, err := url.Parse("file://" + testDataPath)
		require.NoError(t, err, "URL parse should succeed")

		data, err := reader.readFromFile(u)
		require.NoError(t, err, "readFromFile should succeed")
		require.Greater(t, len(data), 0, "data should not be empty")
	})

	t.Run("non-existent file", func(t *testing.T) {
		u, err := url.Parse("file:///nonexistent/file.epub")
		require.NoError(t, err, "URL parse should succeed")

		_, err = reader.readFromFile(u)
		require.Error(t, err, "readFromFile should fail")
		require.Contains(t, err.Error(), "failed to stat file")
	})

	t.Run("directory", func(t *testing.T) {
		u, err := url.Parse("file:///tmp")
		require.NoError(t, err, "URL parse should succeed")

		_, err = reader.readFromFile(u)
		require.Error(t, err, "readFromFile should fail for directory")
		require.Contains(t, err.Error(), "path is a directory")
	})
}

func TestExtractEPUBTitle(t *testing.T) {
	t.Run("extracts title from EPUB with container", func(t *testing.T) {
		testDataPath := filepath.Join("..", "..", "..", "cli", "savetoink", "e2e", "testdata", "article.orig.epub")
		if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
			t.Skip("test EPUB file not found")
		}

		data, err := os.ReadFile(testDataPath) //nolint:gosec // Test file path is trusted
		require.NoError(t, err, "reading test file should succeed")

		// Open as ZIP
		zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err, "creating zip reader should succeed")

		title := extractEPUBTitle(zipReader)
		require.NotEmpty(t, title, "title should be extracted from EPUB")
	})
}

func TestFindFile(t *testing.T) {
	t.Run("finds existing file", func(t *testing.T) {
		testDataPath := filepath.Join("..", "..", "..", "cli", "savetoink", "e2e", "testdata", "article.orig.epub")
		if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
			t.Skip("test EPUB file not found")
		}

		data, err := os.ReadFile(testDataPath) //nolint:gosec // Test file path is trusted
		require.NoError(t, err)

		zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err)

		// Find container.xml which should exist
		file := findFile(zipReader, "META-INF/container.xml")
		require.NotNil(t, file, "container.xml should be found")
	})

	t.Run("returns nil for non-existent file", func(t *testing.T) {
		testDataPath := filepath.Join("..", "..", "..", "cli", "savetoink", "e2e", "testdata", "article.orig.epub")
		if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
			t.Skip("test EPUB file not found")
		}

		data, err := os.ReadFile(testDataPath) //nolint:gosec // Test file path is trusted
		require.NoError(t, err)

		zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err)

		file := findFile(zipReader, "nonexistent/file.xml")
		require.Nil(t, file, "non-existent file should return nil")
	})
}

func TestIndexBytes(t *testing.T) {
	t.Run("finds existing pattern", func(t *testing.T) {
		s := []byte("hello world")
		sep := []byte("world")
		idx := indexBytes(s, sep)
		require.Equal(t, 6, idx)
	})

	t.Run("returns -1 for non-existent pattern", func(t *testing.T) {
		s := []byte("hello world")
		sep := []byte("xyz")
		idx := indexBytes(s, sep)
		require.Equal(t, -1, idx)
	})

	t.Run("handles empty separator", func(t *testing.T) {
		s := []byte("hello world")
		sep := []byte("")
		idx := indexBytes(s, sep)
		require.Equal(t, 0, idx)
	})

	t.Run("handles pattern longer than slice", func(t *testing.T) {
		s := []byte("hi")
		sep := []byte("hello")
		idx := indexBytes(s, sep)
		require.Equal(t, -1, idx)
	})
}
