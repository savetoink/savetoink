package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaftoe/savetoink/backend/lib/internal/content"
	"github.com/shaftoe/savetoink/backend/lib/model"
	"golang.org/x/net/html"
)

// Fetch fetches HTML content from a URL.
func (s *Service) Fetch(ctx context.Context, u *url.URL) (*content.FetchedContent, error) {
	result, err := s.fetcher.Fetch(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url: %w", err)
	}

	return result, nil
}

// ParseHTMLFromSource fetches or reads HTML content from a URL or local file.
func (s *Service) ParseHTMLFromSource(ctx context.Context, u *url.URL) (*html.Node, error) {
	var fetched *content.FetchedContent
	var err error

	if u.Scheme == "file" {
		fetched, err = s.fetchFromFile(u)
	} else {
		fetched, err = s.Fetch(ctx, u)
	}
	if err != nil {
		return nil, err
	}

	doc, err := s.ParseHTML(ctx, fetched)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *Service) fetchFromFile(u *url.URL) (*content.FetchedContent, error) {
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
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &content.FetchedContent{
		HTML: file,
		URL:  u,
		Type: content.FetcherTypeGo,
	}, nil
}

// ParseHTML parses HTML content from fetched content into a DOM node.
func (s *Service) ParseHTML(ctx context.Context, fetched *content.FetchedContent) (*html.Node, error) {
	defer func() {
		_ = fetched.HTML.Close()
	}()

	doc, err := s.extractor.Extract(ctx, fetched.HTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	return doc, nil
}

// Clean extracts article content from a DOM node.
func (s *Service) Clean(ctx context.Context, doc *html.Node, u *url.URL) (*model.Article, error) {
	article, err := s.cleaner.Clean(ctx, doc, u)
	if err != nil {
		return nil, fmt.Errorf("failed to clean content: %w", err)
	}

	return article, nil
}

// GenerateEPUB generates an EPUB from an article.
func (s *Service) GenerateEPUB(article *model.Article) (io.ReadCloser, error) {
	epubReader, err := s.publisher.GenerateEPUB(article)
	if err != nil {
		return nil, fmt.Errorf("failed to generate epub: %w", err)
	}
	return epubReader, nil
}

// ReadEPUB reads an EPUB file from the given path and returns the file reader and title.
func (s *Service) ReadEPUB(ctx context.Context, path string) (io.ReadCloser, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open epub file: %w", err)
	}

	// Check if it's a valid EPUB (ZIP archive)
	fileInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Read file content to memory for title extraction
	epubData, err := io.ReadAll(file)
	if err != nil {
		_ = file.Close()
		return nil, "", fmt.Errorf("failed to read epub file: %w", err)
	}

	// Reopen as ZIP to extract metadata
	zipReader, err := zip.NewReader(bytes.NewReader(epubData), fileInfo.Size())
	if err != nil {
		_ = file.Close()
		return nil, "", fmt.Errorf("failed to open epub as zip: %w", err)
	}

	// Find and parse the content.opf file to extract title
	title := extractEPUBTitle(zipReader)
	if title == "" {
		_ = file.Close()
		return nil, "", fmt.Errorf("failed to extract title from epub file")
	}

	// Return a new reader for the file (seek back to beginning)
	return io.NopCloser(bytes.NewReader(epubData)), title, nil
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
	defer containerReader.Close()

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
	if err := xml.Unmarshal(containerData, &container); err != nil {
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
	defer opfReader.Close()

	opfData, err := io.ReadAll(opfReader)
	if err != nil {
		return ""
	}

	// Parse the OPF file to extract the title
	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		// Try parsing with a common namespace
		return extractEPUBTitleWithNamespace(opfData)
	}

	if pkg.Metadata.Title != "" {
		return pkg.Metadata.Title
	}

	return ""
}

// extractEPUBTitleWithNamespace tries to extract the title with namespace support.
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
						if err := decoder.DecodeElement(&pkg, &se); err == nil {
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
	startIdx := strings.Index(string(data), dcTitle)
	if startIdx != -1 {
		startIdx += len(dcTitle)
		endIdx := strings.Index(string(data)[startIdx:], "</dc:title>")
		if endIdx != -1 {
			title := strings.TrimSpace(string(data)[startIdx : startIdx+endIdx])
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
