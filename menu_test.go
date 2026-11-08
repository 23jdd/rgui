package rgui

import "testing"

func TestMenuDismissesAndRunsAction(t *testing.T) {
	calls := 0
	root := componentOf(Menu(MenuGroup{Label: "File", Items: []MenuItem{{Label: "Open", Action: func() { calls++ }}}}))
	pending := make(subscriptions)
	root.mount(pending)
	defer root.unmount()
	root.children[0].children[0].onClick()
	flushBindings(pending)
	layer := popupLayers(root, nil)[0]
	layer.children[0].children[0].children[0].children[0].onClick()
	flushBindings(pending)
	if calls != 1 || len(popupLayers(root, nil)) != 0 {
		t.Fatal("menu action should run and close")
	}
	root.children[0].children[0].onClick()
	flushBindings(pending)
	if !dismissPopups(root, nil, true) {
		t.Fatal("Escape should dismiss the menu")
	}
}
