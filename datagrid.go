package rgui

// DataGrid displays an editable matrix. Each cell edits its supplied signal.
// Missing cells are empty and read-only.
func DataGrid(columns []GridViewColumn, rows [][]*Signal[string]) Node {
	tracks, header := gridViewHeader(columns)
	items := []Node{header}
	for _, row := range rows {
		cells := make([]Cell, 0, len(columns))
		for col := range columns {
			if col < len(row) && row[col] != nil {
				cells = append(cells, Cell{Column: col, Child: TextBox(row[col])})
			}
		}
		items = append(items, Grid([]Track{Fixed(42)}, tracks, cells...))
	}
	return ScrollViewer(StackPanel(Vertical, 2, items...), nil, true, true)
}
