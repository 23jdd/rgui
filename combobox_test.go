package rgui

import "testing"

func TestComboBoxSelectionClosesOverlay(t *testing.T) {
	selected := NewSignal(-1)
	root := componentOf(ComboBox([]string{"A", "B"}, selected))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	root.children[0].onClick()
	flushBindings(pending)
	layers := popupLayers(root, nil)
	if len(layers) != 1 {
		t.Fatal("expected one open overlay")
	}
	layers[0].children[0].children[0].children[0].children[1].onClick()
	flushBindings(pending)
	if selected.Get() != 1 || len(popupLayers(root, nil)) != 0 {
		t.Fatal("selection should dismiss the overlay")
	}
}
