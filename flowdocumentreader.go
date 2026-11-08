package rgui

// FlowDocumentReader switches between paginated and continuous text views.
func FlowDocumentReader(paragraphs []string) Node {
	return TabControl([]TabItem{
		{Header: "Pages", Content: FlowDocumentPageViewer(paragraphs, nil)},
		{Header: "Scroll", Content: FlowDocumentScrollViewer(paragraphs)},
	}, nil)
}
