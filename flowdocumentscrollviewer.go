package rgui

import "strings"

// FlowDocumentScrollViewer wraps and scrolls text paragraphs continuously.
func FlowDocumentScrollViewer(paragraphs []string) Node {
	content := strings.Join(paragraphs, "\n\n")
	return styleName(ScrollViewer(WithLayout(TextBlock(func() string { return content }, Start), Box{Padding: UniformInsets(8)}), nil, false, true), "flowdocumentscrollviewer")
}
