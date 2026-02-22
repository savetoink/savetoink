package epub

import (
	"testing"
	"time"

	"github.com/shaftoe/savetoink/internal/model"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
}

func TestGenerate_Success(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:   "Test Article",
		Content: "<p>This is test content</p>",
		Author:  "Test Author",
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	if len(data) == 0 {
		t.Error("Generate() expected non-empty data")
	}
}

func TestGenerate_WithMetadata(t *testing.T) {
	gen := NewGenerator()
	publishedAt := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	article := &model.Article{
		Title:              "Test Article",
		Content:            "<p>This is test content</p>",
		Author:             "Test Author",
		SiteName:           "Example Site",
		SourceDomain:       "example.com",
		ReadingTimeMinutes: 5,
		PublishedAt:        &publishedAt,
		ContentType:        "article",
		CreatedAt:          time.Now().UTC(),
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}

	if len(data) == 0 {
		t.Error("Generate() expected non-empty data")
	}
}

func TestGenerate_EmptyTitle(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:   "",
		Content: "<p>This is test content</p>",
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}

func TestGenerate_EmptyContent(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:   "Test Article",
		Content: "",
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}

func TestGenerate_WithLanguage(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:    "Test Article",
		Content:  "<p>This is test content</p>",
		Language: "en",
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}

func TestGenerate_WithExcerpt(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:   "Test Article",
		Content: "<p>This is test content</p>",
		Excerpt: "This is a test excerpt",
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}

func TestGenerate_WithImage(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:    "Test Article",
		Content:  "<p>This is test content</p>",
		ImageURL: "https://example.com/image.jpg",
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}

func TestGenerate_ZeroReadingTime(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:              "Test Article",
		Content:            "<p>This is test content</p>",
		ReadingTimeMinutes: 0,
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}

func TestGenerate_NilPublishedAt(t *testing.T) {
	gen := NewGenerator()
	article := &model.Article{
		Title:       "Test Article",
		Content:     "<p>This is test content</p>",
		PublishedAt: nil,
	}

	data, err := gen.Generate(article)

	if err != nil {
		t.Fatalf("Generate() unexpected error = %v", err)
	}

	if data == nil {
		t.Fatal("Generate() expected data, got nil")
	}
}
