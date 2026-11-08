package draw

import (
	"slices"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestTriangulateConcavePolygon(t *testing.T) {
	points := []rl.Vector2{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 4}, {X: 2, Y: 2}, {X: 0, Y: 4}}
	for range 2 {
		triangles := triangulate(points)
		if len(triangles) != 3 {
			t.Fatalf("got %d triangles, want 3", len(triangles))
		}
		var area float64
		for _, triangle := range triangles {
			signedArea := cross(triangle[0], triangle[1], triangle[2]) / 2
			if signedArea >= 0 {
				t.Fatal("wrong winding for raylib")
			}
			area -= signedArea
		}
		if area != 12 {
			t.Fatalf("filled area = %v, want 12", area)
		}
		slices.Reverse(points)
	}
}

func TestDashContinuesAcrossSegments(t *testing.T) {
	points := []rl.Vector2{{X: 0}, {X: 5}, {X: 14}}
	var got [][2]float32
	walkStroke(points, false, 7, 3, func(from, to rl.Vector2) {
		got = append(got, [2]float32{from.X, to.X})
	})
	want := [][2]float32{{0, 5}, {5, 7}, {10, 14}}
	if !slices.Equal(got, want) {
		t.Fatalf("dash intervals = %v, want %v", got, want)
	}
}

func TestWrapTextPreservesNewlinesAndUnicode(t *testing.T) {
	got := wrapText("one two three\r\n\n中文换行", 7, func(s string) float32 {
		var width float32
		for _, ch := range s {
			width++
			if ch > 127 {
				width += 2
			}
		}
		return width
	})
	want := []string{"one two", "three", "", "中文", "换行"}
	if !slices.Equal(got, want) {
		t.Fatalf("wrapped lines = %q, want %q", got, want)
	}
}

func TestIntersectClips(t *testing.T) {
	outer := rl.Rectangle{X: 10, Y: 10, Width: 30, Height: 20}
	inner := rl.Rectangle{X: 25, Y: 0, Width: 30, Height: 20}
	want := rl.Rectangle{X: 25, Y: 10, Width: 15, Height: 10}
	if got := intersect(outer, inner); got != want {
		t.Fatalf("intersection = %+v, want %+v", got, want)
	}
	if got := intersect(outer, rl.Rectangle{X: 50, Width: 3, Height: 20}); got.Width != 0 {
		t.Fatalf("disjoint clips should have zero width: %+v", got)
	}
}
