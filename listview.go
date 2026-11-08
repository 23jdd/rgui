package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// GridViewColumn configures a ListView or DataGrid column. A nonpositive
// width uses the default of 120 logical pixels.
type GridViewColumn struct {
	Header string
	Width  float32
}

// ListView displays read-only rows with aligned columns and single selection.
func ListView(columns []GridViewColumn, rows [][]string, selected *Signal[int]) Node {
	if selected == nil {
		selected = NewSignal(-1)
	}
	tracks, header := gridViewHeader(columns)
	items := []Node{header}
	for i, row := range rows {
		cells := make([]Cell, 0, len(columns))
		for col := range columns {
			label := ""
			if col < len(row) {
				label = row[col]
			}
			cells = append(cells, Cell{Column: col, Child: Text(func() string { return label })})
		}
		item := &listViewRow{index: i, selected: selected}
		item.Height = 36
		item.Input.Focusable = true
		items = append(items, WithChildren(item, Grid([]Track{Fixed(36)}, tracks, cells...)))
	}
	return styleName(ScrollViewer(StackPanel(Vertical, 2, items...), nil, true, true), "listview")
}

func gridViewHeader(columns []GridViewColumn) ([]Track, Node) {
	if len(columns) == 0 {
		panic("rgui: at least one column is required")
	}
	tracks := make([]Track, len(columns))
	cells := make([]Cell, len(columns))
	for col, column := range columns {
		width := column.Width
		if width <= 0 {
			width = 120
		}
		tracks[col] = Fixed(width)
		cells[col] = Cell{Column: col, Child: Text(func() string { return column.Header })}
	}
	return tracks, styleName(Grid([]Track{Fixed(36)}, tracks, cells...), "tableheader")
}

type listViewRow struct {
	BaseNode
	index    int
	selected *Signal[int]
}

func (r *listViewRow) bind(pending subscriptions) *binding {
	return bindVisual(&r.BaseNode, r.selected, pending)
}
func (r *listViewRow) Tapped(e *InputEvent) { r.selected.Set(r.index); e.Handled = true }
func (r *listViewRow) KeyDown(e *InputEvent) {
	if e.Key == rl.KeyEnter || e.Key == rl.KeySpace {
		r.selected.Set(r.index)
		e.Handled = true
	}
}
