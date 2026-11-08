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
	kind  trackKind
	value float32
}

// Auto sizes a track to the largest child measured in it.
func Auto() Track { return Track{} }

// Fixed reserves a track size in logical pixels.
func Fixed(size float32) Track { return Track{kind: fixedTrack, value: max(0, size)} }

// Star allocates a weighted share of the remaining space.
func Star(weight float32) Track { return Track{kind: starTrack, value: max(0, weight)} }

// Cell places one child in a zero-based row and column.
type Cell struct {
	Row, Column int
	Child       Node
}

// Grid arranges cells on explicit tracks. An empty axis has one Star(1) track.
// A cell outside the defined tracks panics during construction.
func Grid(rows, columns []Track, cells ...Cell) Node {
	g := &gridLayout{rows: gridTracks(rows), columns: gridTracks(columns)}
	children := make([]Node, 0, len(cells))
	for _, cell := range cells {
		if cell.Row < 0 || cell.Row >= len(g.rows) || cell.Column < 0 || cell.Column >= len(g.columns) {
			panic("rgui: grid cell is outside the defined tracks")
		}
		g.positions = append(g.positions, [2]int{cell.Row, cell.Column})
		children = append(children, cell.Child)
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
	positions              [][2]int
	rowContent, colContent []float32
}

func (g *gridLayout) measure(children []*component, available rl.Vector2, text measureText) rl.Vector2 {
	g.colContent = make([]float32, len(g.columns))
	g.rowContent = make([]float32, len(g.rows))
	for i, child := range children {
		row, col := g.positions[i][0], g.positions[i][1]
		size := child.measure(rl.Vector2{X: g.columns[col].probe(), Y: g.rows[row].probe()}, text)
		g.colContent[col] = max(g.colContent[col], size.X)
	}
	columns, width := allocateTracks(g.columns, g.colContent, available.X)
	for i, child := range children {
		row, col := g.positions[i][0], g.positions[i][1]
		size := child.measure(rl.Vector2{X: columns[col], Y: g.rows[row].probe()}, text)
		g.rowContent[row] = max(g.rowContent[row], size.Y)
	}
	rows, height := allocateTracks(g.rows, g.rowContent, available.Y)
	for i, child := range children {
		row, col := g.positions[i][0], g.positions[i][1]
		child.measure(rl.Vector2{X: columns[col], Y: rows[row]}, text)
	}
	return rl.Vector2{X: width, Y: height}
}

func (t Track) probe() float32 {
	if t.kind == fixedTrack {
		return t.value
	}
	return float32(math.Inf(1))
}

func allocateTracks(tracks []Track, content []float32, available float32) ([]float32, float32) {
	sizes := make([]float32, len(tracks))
	var used, weights float32
	unbounded := math.IsInf(float64(available), 1)
	for i, track := range tracks {
		switch track.kind {
		case fixedTrack:
			sizes[i] = track.value
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
		row, col := g.positions[i][0], g.positions[i][1]
		child.arrange(rl.Rectangle{X: x[col], Y: y[row], Width: columns[col], Height: rows[row]})
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
