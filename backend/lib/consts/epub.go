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

	// TitleSeparator is the separator used in article titles (space-dash-space).
	TitleSeparator = " - "

	// MinTitleParts is the minimum number of parts needed for title parsing.
	MinTitleParts = 2

	// EPUBExtension is the file extension for EPUB files.
	EPUBExtension = ".epub"
)
