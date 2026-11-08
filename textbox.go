package rgui

import (
	"slices"
	"strings"
	"unicode"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextBox creates a single-line editor backed by value. A nil signal creates
// private state. Edit on the UI thread.
func TextBox(value *Signal[string]) Node {
	if value == nil {
		value = NewSignal("")
	}
	return &textBox{value: value}
}

type textBox struct {
	BaseNode
	value             *Signal[string]
	measureText       func(string) rl.Vector2
	last              string
	caret, anchor     int
	scroll            float32
	focused, dragging bool
	preedit           Composition
	masked            bool
}

const editInset = 10

func (b *textBox) Measure(ctx MeasureContext) rl.Vector2 {
	b.measureText = ctx.MeasureText
	return rl.Vector2{X: min(300, ctx.Available.X), Y: max(42, b.textMetrics().size+12)}
}

func (b *textBox) FocusGained(e *InputEvent) {
	if e.Source == b {
		b.focused = true
		b.InvalidateVisual()
	}
}

func (b *textBox) FocusLost(e *InputEvent) {
	if e.Source == b {
		b.focused, b.dragging = false, false
		b.preedit = Composition{}
		b.InvalidateVisual()
	}
}

func (b *textBox) read() []rune {
	value := b.value.Get()
	text := []rune(value)
	if b.last != value {
		b.caret = min(b.caret, len(text))
		b.anchor = min(b.anchor, len(text))
		b.preedit = Composition{}
		b.last = value
	}
	return text
}

func (b *textBox) bind(pending subscriptions) *binding {
	update := &binding{
		pending: pending, dependencies: make(map[*subscriptions]struct{}),
		evaluate: func() {
			b.read()
			b.InvalidateVisual()
		},
	}
	update.evaluateTracked()
	return update
}

func (b *textBox) selection() (int, int) {
	return min(b.anchor, b.caret), max(b.anchor, b.caret)
}

func (b *textBox) replace(text []rune, start, end int, value string) {
	insert := []rune(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return -1
		}
		return r
	}, value))
	if value != "" && len(insert) == 0 {
		return
	}
	b.last = string(slices.Replace(text, start, end, insert...))
	b.caret = start + len(insert)
	b.anchor = b.caret
	b.preedit = Composition{}
	b.value.Set(b.last)
	b.InvalidateVisual()
}

func (b *textBox) TextInput(e *InputEvent) {
	if e.Text == "" {
		return
	}
	text := b.read()
	start, end := b.selection()
	b.replace(text, start, end, e.Text)
	e.Handled = true
}

func (b *textBox) CompositionChanged(e *InputEvent) {
	b.read()
	if !b.masked {
		b.preedit = e.Composition
	}
	b.InvalidateVisual()
	e.Handled = true
}

func (b *textBox) KeyDown(e *InputEvent)   { b.editKey(e) }
func (b *textBox) KeyRepeat(e *InputEvent) { b.editKey(e) }

func (b *textBox) editKey(e *InputEvent) {
	text := b.read()
	start, end := b.selection()
	if e.Ctrl {
		switch e.Key {
		case rl.KeyA:
			b.anchor, b.caret = 0, len(text)
		case rl.KeyC, rl.KeyX:
			if !b.masked && start != end {
				selected := string(text[start:end])
				rl.SetClipboardText(selected)
				if e.Key == rl.KeyX && rl.GetClipboardText() == selected {
					b.replace(text, start, end, "")
				}
			}
		case rl.KeyV:
			if value := rl.GetClipboardText(); value != "" {
				b.replace(text, start, end, value)
			}
		default:
			return
		}
	} else {
		switch e.Key {
		case rl.KeyLeft:
			b.caret = max(0, b.caret-1)
			if !e.Shift && start != end {
				b.caret = start
			}
		case rl.KeyRight:
			b.caret = min(len(text), b.caret+1)
			if !e.Shift && start != end {
				b.caret = end
			}
		case rl.KeyHome:
			b.caret = 0
		case rl.KeyEnd:
			b.caret = len(text)
		case rl.KeyBackspace:
			if start == end {
				start = max(0, start-1)
			}
			b.replace(text, start, end, "")
		case rl.KeyDelete:
			if start == end {
				end = min(len(text), end+1)
			}
			b.replace(text, start, end, "")
		default:
			return
		}
		if !e.Shift {
			b.anchor = b.caret
		}
	}
	b.InvalidateVisual()
	e.Handled = true
}

func (b *textBox) width(text []rune) float32 {
	if b.measureText == nil || len(text) == 0 {
		return 0
	}
	return b.measureText(string(text)).X
}

func (b *textBox) positionCaret(x float32) {
	text := b.mask(b.read())
	x += b.scroll - editInset
	b.caret = len(text)
	var previous float32
	for i := range text {
		width := b.width(text[:i+1])
		if x < (width+previous)/2 {
			b.caret = i
			break
		}
		previous = width
	}
	b.InvalidateVisual()
}

func (b *textBox) PointerDown(e *InputEvent) {
	if e.Button != PrimaryButton {
		return
	}
	b.positionCaret(b.localPointer(e.Position).X)
	b.anchor = b.caret
	b.dragging = e.CapturePointer()
}

func (b *textBox) PointerMove(e *InputEvent) {
	if b.dragging {
		b.positionCaret(b.localPointer(e.Position).X)
	}
}

func (b *textBox) PointerUp(*InputEvent)          { b.dragging = false }
func (b *textBox) PointerCaptureLost(*InputEvent) { b.dragging = false }

func (b *textBox) suppressIME() bool { return b.masked }

func (b *textBox) mask(text []rune) []rune {
	if b.masked {
		for i := range text {
			text[i] = '*'
		}
	}
	return text
}

func (b *textBox) display(text []rune) (line []rune, cursor, start, end int) {
	text = b.mask(text)
	if b.preedit.Text == "" {
		start, end = b.selection()
		return text, b.caret, start, end
	}
	preedit := []rune(b.preedit.Text)
	first, last := b.selection()
	line = slices.Replace(text, first, last, preedit...)
	cursor = first + min(max(0, b.preedit.Cursor), len(preedit))
	start = first + min(max(0, b.preedit.SelectionStart), len(preedit))
	end = start + min(max(0, b.preedit.SelectionLength), first+len(preedit)-start)
	return line, cursor, start, end
}

func (b *textBox) scrollToCaret(caret, total float32) {
	space := max(1, b.Bounds().Width-2*editInset)
	b.scroll = max(0, min(max(0, total-space), max(caret-space, min(caret, b.scroll))))
}

func (b *textBox) IMECaret() rl.Rectangle {
	size := b.textMetrics().size
	line, cursor, _, _ := b.display(b.read())
	x := b.width(line[:cursor])
	b.scrollToCaret(x, b.width(line))
	return rl.Rectangle{X: editInset + x - b.scroll, Y: (b.Bounds().Height - size) / 2, Width: 1, Height: size}
}

func (b *textBox) Primitives(ctx PaintContext) []Node {
	text := b.read()
	before, _ := b.selection()
	line, cursor, start, end := b.display(text)
	bounds := b.Bounds()
	style := ctx.TextStyle()
	caret := b.width(line[:cursor])
	b.scrollToCaret(caret, b.width(line))
	x, y := bounds.X+editInset-b.scroll, bounds.Y+(bounds.Height-style.Size)/2
	var nodes []Node
	if b.focused && start != end {
		left, right := b.width(line[:start]), b.width(line[:end])
		nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: x + left, Y: y, Width: right - left, Height: style.Size},
			Style: draw.Style{Fill: ctx.color("--selection", rl.DarkBlue)}})
	}
	nodes = append(nodes, &TextNode{Position: rl.Vector2{X: x, Y: y}, Text: string(line), Style: style})
	if b.preedit.Text != "" {
		left := b.width(line[:before])
		right := b.width(line[:before+len([]rune(b.preedit.Text))])
		nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: x + left, Y: y + style.Size - 1, Width: right - left, Height: 1},
			Style: draw.Style{Fill: ctx.color("--accent", rl.SkyBlue)}})
	}
	if b.focused {
		nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: x + caret, Y: y, Width: 1, Height: style.Size},
			Style: draw.Style{Fill: ctx.color("caret-color", style.Color)}})
	}
	return nodes
}
