package e2e_test

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

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
		fmt.Printf("warning: failed to remove binary %s: %v\n", binaryPath, removeErr)
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
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", out, pkg)
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
func compareEpubs(t *testing.T, path1, path2 string) bool {
	t.Helper()
	r1, openErr1 := zip.OpenReader(path1)
	if openErr1 != nil {
		t.Fatalf("open epub %s: %v", path1, openErr1)
	}
	defer func() {
		if closeErr1 := r1.Close(); closeErr1 != nil {
			t.Logf("warning: failed to close epub %s: %v", path1, closeErr1)
		}
	}()
	r2, openErr2 := zip.OpenReader(path2)
	if openErr2 != nil {
		t.Fatalf("open epub %s: %v", path2, openErr2)
	}
	defer func() {
		if closeErr2 := r2.Close(); closeErr2 != nil {
			t.Logf("warning: failed to close epub %s: %v", path2, closeErr2)
		}
	}()

	if len(r1.File) != len(r2.File) {
		return false
	}
	for i := range r1.File {
		if r1.File[i].Name != r2.File[i].Name {
			return false
		}
		rc1, errOpen1 := r1.File[i].Open()
		if errOpen1 != nil {
			t.Fatalf("read epub entry %s: %v", r1.File[i].Name, errOpen1)
		}
		rc2, errOpen2 := r2.File[i].Open()
		if errOpen2 != nil {
			if closeErr1 := rc1.Close(); closeErr1 != nil {
				t.Logf("warning: failed to close epub entry %s: %v", r1.File[i].Name, closeErr1)
			}
			t.Fatalf("read epub entry %s: %v", r2.File[i].Name, errOpen2)
		}
		var buf1, buf2 bytes.Buffer
		if _, readErr1 := buf1.ReadFrom(rc1); readErr1 != nil {
			_ = rc1.Close()
			_ = rc2.Close()
			t.Fatalf("read epub entry content %s: %v", r1.File[i].Name, readErr1)
		}
		if _, readErr2 := buf2.ReadFrom(rc2); readErr2 != nil {
			_ = rc1.Close()
			_ = rc2.Close()
			t.Fatalf("read epub entry content %s: %v", r2.File[i].Name, readErr2)
		}
		if closeErr1 := rc1.Close(); closeErr1 != nil {
			t.Logf("warning: failed to close epub entry %s: %v", r1.File[i].Name, closeErr1)
		}
		if closeErr2 := rc2.Close(); closeErr2 != nil {
			t.Logf("warning: failed to close epub entry %s: %v", r2.File[i].Name, closeErr2)
		}
		if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
			return false
		}
	}
	return true
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
