package rgui

import (
	"time"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// RepeatButton activates immediately on press and repeats after delay.
func RepeatButton(label string, delay, interval time.Duration, activate func()) Node {
	if delay <= 0 {
		delay = 400 * time.Millisecond
	}
	if interval <= 0 {
		interval = 80 * time.Millisecond
	}
	node := &repeatButton{label: label, delay: delay, interval: interval, activate: activate}
	node.Input.Focusable = activate != nil
	return node
}

type repeatButton struct {
	BaseNode
	label           string
	delay, interval time.Duration
	activate        func()
	next            time.Time
	pointer, inside bool
	key             int32
}

func (b *repeatButton) Measure(ctx MeasureContext) rl.Vector2 {
	size := ctx.MeasureText(b.label)
	return rl.Vector2{X: size.X + 24, Y: max(40, size.Y)}
}

func (b *repeatButton) begin() {
	if b.activate != nil {
		b.next = time.Now().Add(b.delay)
		b.InvalidateVisual()
		b.activate()
	}
}

func (b *repeatButton) PointerDown(e *InputEvent) {
	if e.Source == b && e.Button == PrimaryButton && b.activate != nil && e.CapturePointer() {
		b.pointer, b.inside = true, true
		b.begin()
	}
}
func (b *repeatButton) PointerMove(e *InputEvent) {
	if b.pointer {
		b.inside = containsPoint(b.host.visualRect(b.Bounds()), e.Position)
	}
}
func (b *repeatButton) PointerUp(e *InputEvent) {
	if e.Button == PrimaryButton {
		b.stopPointer()
	}
}
func (b *repeatButton) PointerCaptureLost(*InputEvent) { b.stopPointer() }
func (b *repeatButton) stopPointer()                   { b.pointer = false; b.InvalidateVisual() }

func (b *repeatButton) KeyDown(e *InputEvent) {
	if e.Source == b && b.activate != nil && (e.Key == rl.KeySpace || e.Key == rl.KeyEnter || e.Key == rl.KeyKpEnter) {
		if b.key == 0 {
			b.key = e.Key
			b.begin()
		}
		e.Handled = true
	}
}
func (b *repeatButton) KeyUp(e *InputEvent) {
	if b.key != 0 && e.Key == b.key {
		b.key = 0
		b.InvalidateVisual()
		e.Handled = true
	}
}

func (b *repeatButton) tick(now time.Time, active bool, focus *component) {
	if !active || !b.host.inputEnabled() {
		if b.pointer || b.key != 0 {
			b.pointer = false
			b.key = 0
			b.InvalidateVisual()
		}
		return
	}
	if b.key != 0 && focus != b.host {
		b.key = 0
		b.InvalidateVisual()
	}
	if b.activate != nil && (b.key != 0 || b.pointer && b.inside) && !now.Before(b.next) {
		b.next = now.Add(b.interval)
		b.activate()
	}
}

func (b *repeatButton) Primitives(ctx PaintContext) []Node {
	bounds := b.Bounds()
	size := draw.MeasureText(b.label, ctx.TextStyle())
	return []Node{
		&TextNode{Position: rl.Vector2{X: bounds.X + (bounds.Width-size.X)/2, Y: bounds.Y + (bounds.Height-size.Y)/2},
			Text: b.label, Style: ctx.TextStyle()},
	}
}

func tickControls(root *component, now time.Time, active bool, focus *component) {
	if control, ok := root.custom.(interface {
		tick(time.Time, bool, *component)
	}); ok {
		control.tick(now, active, focus)
	}
	for _, child := range root.children {
		tickControls(child, now, active, focus)
	}
}
