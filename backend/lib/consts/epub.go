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

	// EPUBStylesheetFilename is the filename for the embedded CSS stylesheet in the EPUB.
	EPUBStylesheetFilename = "styles.css"
)

// EPUBStylesheet contains the CSS stylesheet embedded in generated EPUB files.
// It provides styling for code blocks and inline code to ensure readability
// and proper whitespace preservation on e-readers.
const EPUBStylesheet = `pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: monospace;
  font-size: 0.9em;
  line-height: 1.4;
  background-color: #f4f4f4;
  padding: 1em;
  margin: 1em 0;
  border: 1px solid #ddd;
  overflow-x: auto;
}
code {
  font-family: monospace;
  font-size: 0.9em;
}
pre code {
  background: none;
  padding: 0;
  border: none;
}
p code, li code {
  background-color: #f4f4f4;
  padding: 0.2em 0.4em;
  border-radius: 3px;
}
`
