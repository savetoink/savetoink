package consts

// EPUB and HTML formatting constants.
const (
	// DefaultChapterTitle is the default title for single-chapter EPUBs.
	DefaultChapterTitle = "Chapter 1"

	// DefaultChapterFilename is the default filename for a chapter in single-chapter EPUBs.
	DefaultChapterFilename = "chapter1.xhtml"

	// HTMLTagP is the opening paragraph HTML tag.
	HTMLTagP = "<p>"

	// HTMLTagPEnd is the closing paragraph HTML tag.
	HTMLTagPEnd = "</p>"

	// HTMLTagDiv is the opening div HTML tag.
	HTMLTagDiv = "<div>"

	// HTMLTagDivEnd is the closing div HTML tag.
	HTMLTagDivEnd = "</div>"

	// HTMLTagBr is the line break HTML tag.
	HTMLTagBr = "<br>"

	// HTMLTagBrSelfClosing is the self-closing line break HTML tag.
	HTMLTagBrSelfClosing = "<br />"

	// HTMLTagStrongStart is the opening strong HTML tag.
	HTMLTagStrongStart = "<strong>"

	// HTMLTagStrongEnd is the closing strong HTML tag.
	HTMLTagStrongEnd = "</strong>"

	// TitleSeparator is the separator used in article titles (space-dash-space).
	TitleSeparator = " - "

	// DateYearMonthDay is the RFC3339 year-month-day date format.
	DateYearMonthDay = "2006-01-02"

	// RequestIDFormat is the format for request IDs (YYYYMMDD-HHMMSS.mmm).
	RequestIDFormat = "20060102-150405.000"

	// MinTitleParts is the minimum number of parts needed for title parsing.
	MinTitleParts = 2
)
