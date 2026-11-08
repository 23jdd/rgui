package rgui

import "testing"

func TestFlowDocumentReflowsByWidth(t *testing.T) {
	root := componentOf(FlowDocumentPageViewer([]string{"aa bb cc dd"}, NewSignal(0)))
	page := root.children[1].custom.(*flowPage)
	testLayout(root, 58, 115)
	if page.count.Get() < 2 {
		t.Fatalf("narrow page count = %d", page.count.Get())
	}
	testLayout(root, 300, 115)
	if page.count.Get() != 1 {
		t.Fatalf("wide page count = %d", page.count.Get())
	}
}
