package rgui

import "testing"

func TestProgressBarClampsFill(t *testing.T) {
	value := NewSignal(30.0)
	bar := ProgressBar(value, 20, 40)
	testLayout(componentOf(bar), 100, 16)
	width := func() float32 { return bar.Primitives(PaintContext{})[1].(*RectNode).Rect.Width }
	if width() != 50 {
		t.Fatal("expected half fill")
	}
	value.Set(100)
	if width() != 100 {
		t.Fatal("expected full fill")
	}
	value.Set(-1)
	if width() != 0 {
		t.Fatal("expected empty fill")
	}
}
