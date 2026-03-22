// Package epub provides EPUB file generation and reading functionality.
package epub

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Reader handles reading and parsing EPUB files.
type Reader struct {
	client *http.Client
}

// NewReader creates a new EPUB reader instance.
func NewReader() *Reader {
	return &Reader{
		client: &http.Client{},
	}
}

// ReadFromURL reads an EPUB file from a URL (file:// or http:// or https://)
// and returns the file reader and extracted title.
func (r *Reader) ReadFromURL(ctx context.Context, u *url.URL) (io.ReadCloser, string, error) {
	var data []byte
	var err error

	switch u.Scheme {
	case "file":
		data, err = r.readFromFile(u)
	case "http", "https":
		data, err = r.readFromHTTP(ctx, u)
	default:
		return nil, "", fmt.Errorf("unsupported url scheme: %s", u.Scheme)
	}

	if err != nil {
		return nil, "", err
	}

	// Open as ZIP to extract metadata
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to open epub as zip: %w", err)
	}

	title := extractEPUBTitle(zipReader)
	if title == "" {
		return nil, "", errors.New("failed to extract title from epub file")
	}

	return io.NopCloser(bytes.NewReader(data)), title, nil
}

// readFromFile reads an EPUB file from the local filesystem.
func (r *Reader) readFromFile(u *url.URL) ([]byte, error) {
	filePath := filepath.Clean(u.Path)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open epub file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// readFromHTTP reads an EPUB file from an HTTP/HTTPS URL.
func (r *Reader) readFromHTTP(ctx context.Context, u *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch epub: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return data, nil
}

// opfMetadata represents the metadata section of an EPUB OPF file.
type opfMetadata struct {
	XMLName  xml.Name `xml:"metadata"`
	Title    string   `xml:"title"`
	Language string   `xml:"language"`
}

// opfPackage represents the root package element of an EPUB OPF file.
type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Metadata opfMetadata `xml:"metadata"`
}

// extractEPUBTitle extracts the title from an EPUB file.
func extractEPUBTitle(zipReader *zip.Reader) string {
	// EPUB files typically contain a META-INF/container.xml that points to the OPF file
	// The OPF file contains the metadata including the title
	containerPath := "META-INF/container.xml"
	containerFile := findFile(zipReader, containerPath)
	if containerFile == nil {
		return ""
	}

	// Read container.xml to find the OPF path
	containerReader, err := containerFile.Open()
	if err != nil {
		return ""
	}
	defer func() { _ = containerReader.Close() }()

	containerData, err := io.ReadAll(containerReader)
	if err != nil {
		return ""
	}

	// Parse container.xml to get the OPF path
	var container struct {
		XMLName xml.Name `xml:"container"`
		Root    struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if unmarshalErr := xml.Unmarshal(containerData, &container); unmarshalErr != nil {
		return ""
	}

	opfPath := container.Root.FullPath
	if opfPath == "" {
		return ""
	}

	// Read the OPF file
	opfFile := findFile(zipReader, opfPath)
	if opfFile == nil {
		return ""
	}

	opfReader, err := opfFile.Open()
	if err != nil {
		return ""
	}
	defer func() { _ = opfReader.Close() }()

	opfData, err := io.ReadAll(opfReader)
	if err != nil {
		return ""
	}

	// Parse the OPF file to extract the title
	var pkg opfPackage
	if unmarshalErr := xml.Unmarshal(opfData, &pkg); unmarshalErr != nil {
		// Try parsing with a common namespace
		return extractEPUBTitleWithNamespace(opfData)
	}

	if pkg.Metadata.Title != "" {
		return pkg.Metadata.Title
	}

	return ""
}

// extractEPUBTitleWithNamespace tries to extract the title with namespace support.
//
//nolint:gocognit,nestif // Complex XML parsing requires nested conditionals
func extractEPUBTitleWithNamespace(data []byte) string {
	type opfPackageNS struct {
		XMLName  xml.Name    `xml:"package"`
		Metadata opfMetadata `xml:"metadata"`
	}

	var pkg opfPackageNS
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		if se, ok := token.(xml.StartElement); ok {
			if se.Name.Local == "package" {
				for _, attr := range se.Attr {
					if attr.Name.Local == "xmlns" || attr.Name.Local == "xmlns:opf" {
						// Namespace found, try to unmarshal
						if decodeErr := decoder.DecodeElement(&pkg, &se); decodeErr == nil {
							if pkg.Metadata.Title != "" {
								return pkg.Metadata.Title
							}
						}
					}
				}
			}
		}
	}

	// Fallback: try to find <dc:title> element
	dcTitle := "<dc:title>"
	startIdx := indexBytes(data, []byte(dcTitle))
	if startIdx != -1 {
		startIdx += len(dcTitle)
		endIdx := indexBytes(data[startIdx:], []byte("</dc:title>"))
		if endIdx != -1 {
			title := string(bytes.TrimSpace(data[startIdx : startIdx+endIdx]))
			if title != "" {
				return title
			}
		}
	}

	return ""
}

// findFile finds a file in the ZIP archive by path.
func findFile(zipReader *zip.Reader, path string) *zip.File {
	for i := range zipReader.File {
		if zipReader.File[i].Name == path {
			return zipReader.File[i]
		}
	}
	return nil
}

// indexBytes is a helper to find the index of a byte slice in another byte slice.
func indexBytes(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-n; i++ {
		if bytes.Equal(s[i:i+n], sep) {
			return i
		}
	}
	return -1
}
