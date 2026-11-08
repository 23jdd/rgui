package rgui

import (
	"fmt"
	"slices"
)

// DocumentViewer displays one supplied page node at a time, with scrolling
// for pages larger than the viewport.
func DocumentViewer(pages []Node, index *Signal[int]) Node {
	if index == nil {
		index = NewSignal(0)
	}
	pages = slices.Clone(pages)
	for i, page := range pages {
		pages[i] = ScrollViewer(page, nil, true, true)
	}
	count := func() int { return len(pages) }
	return styleName(documentControls(index, count, newContentSwitch(func() int { return documentIndex(index.Get(), count()) }, pages)), "documentviewer")
}

func documentIndex(index, count int) int {
	if count == 0 {
		return -1
	}
	return max(0, min(index, count-1))
}

func documentControls(index *Signal[int], count func() int, content Node) Node {
	header := ToolBar(
		Button("Previous", func() { index.Set(documentIndex(index.Get()-1, count())) }),
		Text(func() string { return fmt.Sprintf("%d / %d", documentIndex(index.Get(), count())+1, count()) }),
		Button("Next", func() { index.Set(documentIndex(index.Get()+1, count())) }),
	)
	return DockPanel(true, Docked{Edge: DockTop, Child: header}, Docked{Child: content})
}
