package e2e_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const epubCloseWarnFormat = "warning: failed to close epub entry %s: %v"

// binaryPath holds the path to the compiled CLI binary, set once in TestMain.
var binaryPath string

// TestMain compiles the CLI binary once before all tests run.
func TestMain(m *testing.M) {
	bin, compileErr := compileBinary()
	if compileErr != nil {
		// Use panic so the error is visible; log.Fatal suppresses the stack.
		panic("failed to compile CLI binary: " + compileErr.Error())
	}
	binaryPath = bin
	exitCode := m.Run()
	if removeErr := os.Remove(binaryPath); removeErr != nil {
		slog.Warn("failed to remove binary", slog.Any("path", binaryPath), slog.Any("error", removeErr))
	}
	os.Exit(exitCode)
}

// compileBinary builds the CLI and returns the path to the resulting binary.
func compileBinary() (string, error) {
	const pkg = "github.com/shaftoe/savetoink/cli/savetoink"

	tmp, createErr := os.CreateTemp("", "savetoink-e2e-*")
	if createErr != nil {
		return "", fmt.Errorf("create temp file: %w", createErr)
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return "", fmt.Errorf("close temp file: %w", closeErr)
	}

	out := tmp.Name()
	if runtime.GOOS == "windows" {
		out += ".exe"
	}

	//nolint:gosec // G204: This is a controlled test build with hardcoded arguments
	cmd := exec.CommandContext(context.Background(), "go", "build", "-buildvcs=false", "-o", out, pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		return "", fmt.Errorf("build binary: %w", runErr)
	}
	return out, nil
}

// fixtureServer starts an httptest.Server serving HTML from the testdata directory.
// Each file under testdata/pages/ is accessible at /<filename>.
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("testdata/pages")))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// runCLI invokes the compiled binary with the given args and returns stdout,
// stderr, and the exit error (nil on success).
func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	//nolint:gosec // G204: binaryPath is a controlled test binary
	cmd := exec.CommandContext(context.Background(), binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

// -----------------------------------------------------------------
// Helpers for EPUB inspection.
// -----------------------------------------------------------------.

// openZipReaders opens two zip files and returns their readers with cleanup function.
func openZipReaders(t *testing.T, path1, path2 string) (r1, r2 *zip.ReadCloser) {
	t.Helper()
	r1, err1 := zip.OpenReader(path1)
	if err1 != nil {
		t.Fatalf("open epub %s: %v", path1, err1)
	}
	r2, err2 := zip.OpenReader(path2)
	if err2 != nil {
		_ = r1.Close()
		t.Fatalf("open epub %s: %v", path2, err2)
	}
	return r1, r2
}

// closeZipReaders safely closes both zip readers with warning logging.
func closeZipReaders(t *testing.T, path1, path2 string, r1, r2 *zip.ReadCloser) {
	t.Helper()
	if closeErr1 := r1.Close(); closeErr1 != nil {
		t.Logf("warning: failed to close epub %s: %v", path1, closeErr1)
	}
	if closeErr2 := r2.Close(); closeErr2 != nil {
		t.Logf(epubCloseWarnFormat, path2, closeErr2)
	}
}

// compareZipMetadata checks if two zip files have the same file structure.
func compareZipMetadata(r1, r2 *zip.ReadCloser) bool {
	if len(r1.File) != len(r2.File) {
		return false
	}
	for i := range r1.File {
		if r1.File[i].Name != r2.File[i].Name {
			return false
		}
	}
	return true
}

// compareFileContent compares the content of a single file entry between two zips.
func compareFileContent(t *testing.T, f1, f2 *zip.File) bool {
	t.Helper()
	rc1, err1 := f1.Open()
	if err1 != nil {
		t.Fatalf("read epub entry %s: %v", f1.Name, err1)
	}
	defer func() {
		if closeErr := rc1.Close(); closeErr != nil {
			t.Logf(epubCloseWarnFormat, f1.Name, closeErr)
		}
	}()

	rc2, err2 := f2.Open()
	if err2 != nil {
		t.Fatalf("read epub entry %s: %v", f2.Name, err2)
	}
	defer func() {
		if closeErr := rc2.Close(); closeErr != nil {
			t.Logf(epubCloseWarnFormat, f2.Name, closeErr)
		}
	}()

	var buf1, buf2 bytes.Buffer
	if _, readErr1 := buf1.ReadFrom(rc1); readErr1 != nil {
		t.Fatalf("read epub entry content %s: %v", f1.Name, readErr1)
	}
	if _, readErr2 := buf2.ReadFrom(rc2); readErr2 != nil {
		t.Fatalf("read epub entry content %s: %v", f2.Name, readErr2)
	}

	return bytes.Equal(buf1.Bytes(), buf2.Bytes())
}

// compareEpubs compares two EPUB files for equality.
func compareEpubs(t *testing.T, path1, path2 string) bool {
	t.Helper()
	r1, r2 := openZipReaders(t, path1, path2)
	defer closeZipReaders(t, path1, path2, r1, r2)

	if !compareZipMetadata(r1, r2) {
		return false
	}

	for i := range r1.File {
		if !compareFileContent(t, r1.File[i], r2.File[i]) {
			return false
		}
	}
	return true
}

// TestEPUBFileConversion verifies that an EPUB file can be passed through
// the convert command and output unchanged.
func TestEPUBFileConversion(t *testing.T) {
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "output.epub")
	testFile := filepath.Join("testdata", "article.orig.epub")

	_, stderr, err := runCLI(t,
		"convert", testFile,
		"--output", outFile,
	)
	if err != nil {
		t.Fatalf("CLI exited with error: %v\nstderr: %s", err, stderr)
	}

	_, err = os.Stat(outFile)
	if os.IsNotExist(err) {
		t.Fatalf("expected output file %s was not created", outFile)
	}

	// Verify the output is identical to the input
	if !compareEpubs(t, outFile, testFile) {
		t.Fatalf("output EPUB does not match input EPUB")
	}
}

// -----------------------------------------------------------------
// Tests
// -----------------------------------------------------------------

// TestBasicArticle verifies that a standard article page produces a valid EPUB
// containing the expected EPUB document previously created.
func TestBasicArticle(t *testing.T) {
	srv := fixtureServer(t)
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "article.epub")
	testFile := filepath.Join("testdata", "article.orig.epub")

	_, stderr, err := runCLI(t,
		"convert", srv.URL+"/article.html",
		"--output", outFile,
	)
	if err != nil {
		t.Fatalf("CLI exited with error: %v\nstderr: %s", err, stderr)
	}

	_, err = os.Stat(outFile)
	if os.IsNotExist(err) {
		t.Fatalf("expected output file %s was not created", outFile)
	}

	_, err = os.Stat(testFile)
	if os.IsNotExist(err) {
		t.Fatalf("expected test file %s was not found", testFile)
	}

	if compareEpubs(t, outFile, testFile) {
		t.Fatalf("output EPUB does not match test file")
	}
}
