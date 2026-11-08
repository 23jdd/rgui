package rgui

import (
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type trackKind uint8

const (
	autoTrack trackKind = iota
	fixedTrack
	starTrack
)

// Track describes a Grid row height or column width.
type Track struct {
	kind       trackKind
	value      float32
	sizeSignal *Signal[float32]
}

// Auto sizes a track to the largest child measured in it.
func Auto() Track { return Track{} }

// Fixed reserves a track size in logical pixels.
func Fixed(size float32) Track { return Track{kind: fixedTrack, value: max(0, size)} }

// FixedFrom creates a fixed track whose size can be changed on the UI thread.
func FixedFrom(size *Signal[float32]) Track { return Track{kind: fixedTrack, sizeSignal: size} }

func (t Track) fixedSize() float32 {
	if t.sizeSignal == nil {
		return t.value
	}
	value := t.sizeSignal.Get()
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return 0
	}
	return max(0, value)
}

// Star allocates a weighted share of the remaining space.
func Star(weight float32) Track { return Track{kind: starTrack, value: max(0, weight)} }

// Cell places one child in a zero-based row and column.
type Cell struct {
	Row, Column         int
	RowSpan, ColumnSpan int // Zero means one track.
	Child               Node
}

// Grid arranges cells on explicit tracks. An empty axis has one Star(1) track.
// A cell outside the defined tracks panics during construction.
func Grid(rows, columns []Track, cells ...Cell) Node {
	g := &gridLayout{rows: gridTracks(rows), columns: gridTracks(columns)}
	children := make([]Node, 0, len(cells))
	for _, cell := range cells {
		if cell.RowSpan == 0 {
			cell.RowSpan = 1
		}
		if cell.ColumnSpan == 0 {
			cell.ColumnSpan = 1
		}
		if cell.Row < 0 || cell.Row >= len(g.rows) || cell.Column < 0 || cell.Column >= len(g.columns) ||
			cell.RowSpan < 1 || cell.RowSpan > len(g.rows)-cell.Row ||
			cell.ColumnSpan < 1 || cell.ColumnSpan > len(g.columns)-cell.Column {
			panic("rgui: grid cell is outside the defined tracks")
		}
		children = append(children, cell.Child)
		cell.Child = nil
		g.positions = append(g.positions, cell)
	}
	return panel(g, children)
}

func gridTracks(tracks []Track) []Track {
	if len(tracks) == 0 {
		return []Track{Star(1)}
	}
	return slices.Clone(tracks)
}

type gridLayout struct {
	rows, columns          []Track
	positions              []Cell
	rowContent, colContent []float32
}

func (g *gridLayout) bindLayout(owner *component, pending subscriptions) *binding {
	for _, tracks := range [][]Track{g.rows, g.columns} {
		for _, track := range tracks {
			if track.sizeSignal != nil {
				update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}),
					evaluate: func() {
						for _, group := range [][]Track{g.rows, g.columns} {
							for _, item := range group {
								item.fixedSize()
							}
						}
						owner.invalidateMeasure()
					},
				}
				update.evaluateTracked()
				return update
			}
		}
	}
	return nil
}
func (g *gridLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	g.colContent = make([]float32, len(g.columns))
	g.rowContent = make([]float32, len(g.rows))
	columnNeeds := make([]trackNeed, len(children))
	rowNeeds := make([]trackNeed, len(children))
	for i, child := range children {
		cell := g.positions[i]
		size := child.measure(rl.Vector2{
			X: probeSpan(g.columns[cell.Column : cell.Column+cell.ColumnSpan]),
			Y: probeSpan(g.rows[cell.Row : cell.Row+cell.RowSpan]),
		}, text)
		columnNeeds[i] = trackNeed{cell.Column, cell.ColumnSpan, size.X}
	}
	growTracks(g.columns, g.colContent, columnNeeds, available.X)
	columns, width := allocateTracks(g.columns, g.colContent, available.X)
	for i, child := range children {
		cell := g.positions[i]
		size := child.measure(rl.Vector2{
			X: trackSpan(columns, cell.Column, cell.ColumnSpan),
			Y: probeSpan(g.rows[cell.Row : cell.Row+cell.RowSpan]),
		}, text)
		rowNeeds[i] = trackNeed{cell.Row, cell.RowSpan, size.Y}
	}
	growTracks(g.rows, g.rowContent, rowNeeds, available.Y)
	rows, height := allocateTracks(g.rows, g.rowContent, available.Y)
	for i, child := range children {
		cell := g.positions[i]
		child.measure(rl.Vector2{X: trackSpan(columns, cell.Column, cell.ColumnSpan), Y: trackSpan(rows, cell.Row, cell.RowSpan)}, text)
	}
	return rl.Vector2{X: width, Y: height}
}

func probeSpan(tracks []Track) float32 {
	var size float32
	for _, track := range tracks {
		if track.kind != fixedTrack {
			return float32(math.Inf(1))
		}
		size += track.fixedSize()
	}
	return size
}

type trackNeed struct {
	start, span int
	size        float32
}

func growTracks(tracks []Track, content []float32, needs []trackNeed, available float32) {
	unbounded := math.IsInf(float64(available), 1)
	slices.SortStableFunc(needs, func(a, b trackNeed) int { return a.span - b.span })
	for _, need := range needs {
		var current float32
		var eligible []int
		for i := need.start; i < need.start+need.span; i++ {
			track := tracks[i]
			switch {
			case track.kind == fixedTrack:
				current += track.fixedSize()
			case track.kind == autoTrack || (unbounded && track.kind == starTrack && track.value > 0):
				current += content[i]
				eligible = append(eligible, i)
			}
		}
		if need.size > current && len(eligible) != 0 {
			extra := (need.size - current) / float32(len(eligible))
			for _, i := range eligible {
				content[i] += extra
			}
		}
	}
}

func trackSpan(sizes []float32, start, span int) float32 {
	var size float32
	for _, track := range sizes[start : start+span] {
		size += track
	}
	return size
}

func allocateTracks(tracks []Track, content []float32, available float32) ([]float32, float32) {
	sizes := make([]float32, len(tracks))
	var used, weights float32
	unbounded := math.IsInf(float64(available), 1)
	for i, track := range tracks {
		switch track.kind {
		case fixedTrack:
			sizes[i] = track.fixedSize()
		case autoTrack:
			sizes[i] = content[i]
		case starTrack:
			if unbounded && track.value > 0 {
				sizes[i] = content[i]
			} else {
				weights += track.value
			}
		}
		used += sizes[i]
	}
	if !unbounded && weights > 0 {
		remainder := max(0, available-used)
		for i, track := range tracks {
			if track.kind == starTrack {
				sizes[i] = remainder * (track.value / weights)
			}
		}
		used += remainder
	}
	return sizes, used
}

func (g *gridLayout) arrange(children []*component, area rl.Rectangle) {
	rows, _ := allocateTracks(g.rows, g.rowContent, area.Height)
	columns, _ := allocateTracks(g.columns, g.colContent, area.Width)
	x, y := trackOffsets(columns, area.X), trackOffsets(rows, area.Y)
	for i, child := range children {
		cell := g.positions[i]
		child.arrange(rl.Rectangle{
			X: x[cell.Column], Y: y[cell.Row],
			Width: trackSpan(columns, cell.Column, cell.ColumnSpan), Height: trackSpan(rows, cell.Row, cell.RowSpan),
		})
	}
}

func trackOffsets(sizes []float32, origin float32) []float32 {
	offsets := make([]float32, len(sizes))
	for i, size := range sizes {
		offsets[i] = origin
		origin += size
	}
	return offsets
}
