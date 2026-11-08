package rgui

import (
	"image/color"
	"math"
	"slices"
	"strings"
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
	"rgui/draw"
)

// RichTextStyle styles one run of text. A zero color uses the default text color.
type RichTextStyle struct {
	Color           color.RGBA
	Bold, Underline bool
}

// TextRun is a span of text sharing the same style.
type TextRun struct {
	Text  string
	Style RichTextStyle
}

// RichTextBox edits multiline styled text. Ctrl+B and Ctrl+U apply styling,
// and clipboard operations use plain Unicode text.
func RichTextBox(value *Signal[[]TextRun]) Node {
	if value == nil {
		value = NewSignal([]TextRun(nil))
	}
	return &richTextBox{value: value}
}

type richGlyph struct {
	char  rune
	style RichTextStyle
}

type richTextBox struct {
	BaseNode
	value             *Signal[[]TextRun]
	last              []TextRun
	glyphs            []richGlyph
	measureText       func(string) rl.Vector2
	caret, anchor     int
	scroll            float32
	typing            RichTextStyle
	preedit           Composition
	focused, dragging bool
}

func (b *richTextBox) Measure(ctx MeasureContext) rl.Vector2 {
	b.measureText = ctx.MeasureText
	return rl.Vector2{X: min(360, ctx.Available.X), Y: min(160, ctx.Available.Y)}
}

func (b *richTextBox) bind(pending subscriptions) *binding {
	update := &binding{pending: pending, dependencies: make(map[*subscriptions]struct{}), evaluate: func() {
		b.read()
		b.InvalidateVisual()
	}}
	update.evaluateTracked()
	return update
}

func (b *richTextBox) read() {
	if runs := b.value.Get(); !slices.Equal(runs, b.last) {
		b.last = slices.Clone(runs)
		b.glyphs = b.glyphs[:0]
		for _, run := range runs {
			b.glyphs = append(b.glyphs, richGlyphs(run.Text, run.Style)...)
		}
		b.caret, b.anchor = min(b.caret, len(b.glyphs)), min(b.anchor, len(b.glyphs))
		b.preedit = Composition{}
	}
}

func richGlyphs(text string, style RichTextStyle) []richGlyph {
	var result []richGlyph
	for _, ch := range strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n") {
		if !unicode.IsControl(ch) || ch == '\n' {
			result = append(result, richGlyph{char: ch, style: style})
		}
	}
	return result
}

func (b *richTextBox) publish() {
	var runs []TextRun
	var text strings.Builder
	var style RichTextStyle
	for i, glyph := range b.glyphs {
		if i > 0 && style != glyph.style {
			runs = append(runs, TextRun{Text: text.String(), Style: style})
			text.Reset()
		}
		style = glyph.style
		text.WriteRune(glyph.char)
	}
	if text.Len() > 0 {
		runs = append(runs, TextRun{Text: text.String(), Style: style})
	}
	b.last = slices.Clone(runs)
	b.value.Set(runs)
	b.InvalidateVisual()
}

func (b *richTextBox) selection() (int, int) { return min(b.caret, b.anchor), max(b.caret, b.anchor) }

func (b *richTextBox) replace(start, end int, text string) {
	insert := richGlyphs(text, b.typing)
	if text != "" && len(insert) == 0 {
		return
	}
	b.glyphs = slices.Replace(b.glyphs, start, end, insert...)
	b.caret = start + len(insert)
	b.anchor = b.caret
	b.preedit = Composition{}
	b.publish()
}

func (b *richTextBox) TextInput(e *InputEvent) {
	if e.Text == "" {
		return
	}
	b.read()
	start, end := b.selection()
	b.replace(start, end, e.Text)
	e.Handled = true
}

func (b *richTextBox) CompositionChanged(e *InputEvent) {
	b.read()
	b.preedit = e.Composition
	b.InvalidateVisual()
	e.Handled = true
}

func (b *richTextBox) FocusGained(e *InputEvent) {
	if e.Source == b {
		b.focused = true
		b.InvalidateVisual()
	}
}
func (b *richTextBox) FocusLost(e *InputEvent) {
	if e.Source == b {
		b.focused, b.dragging = false, false
		b.preedit = Composition{}
		b.InvalidateVisual()
	}
}

func (b *richTextBox) format(start, end int, bold bool) {
	style := b.typing
	if start < end {
		style = b.glyphs[start].style
	}
	enabled := !style.Underline
	if bold {
		enabled = !style.Bold
		b.typing.Bold = enabled
	} else {
		b.typing.Underline = enabled
	}
	for i := start; i < end; i++ {
		if bold {
			b.glyphs[i].style.Bold = enabled
		} else {
			b.glyphs[i].style.Underline = enabled
		}
	}
	if start < end {
		b.publish()
	}
}

func (b *richTextBox) editShortcut(e *InputEvent, start, end int) bool {
	switch e.Key {
	case rl.KeyA:
		b.anchor, b.caret = 0, len(b.glyphs)
	case rl.KeyB, rl.KeyU:
		b.format(start, end, e.Key == rl.KeyB)
	case rl.KeyC, rl.KeyX:
		if start != end {
			var text strings.Builder
			for _, glyph := range b.glyphs[start:end] {
				text.WriteRune(glyph.char)
			}
			rl.SetClipboardText(text.String())
			if e.Key == rl.KeyX && rl.GetClipboardText() == text.String() {
				b.replace(start, end, "")
			}
		}
	case rl.KeyV:
		if text := rl.GetClipboardText(); text != "" {
			b.replace(start, end, text)
		}
	default:
		return false
	}
	return true
}

func (b *richTextBox) KeyDown(e *InputEvent) { b.editKey(e) }
func (b *richTextBox) KeyRepeat(e *InputEvent) {
	if !e.Ctrl {
		b.editKey(e)
	}
}

func (b *richTextBox) editKey(e *InputEvent) {
	b.read()
	start, end := b.selection()
	if e.Ctrl {
		if !b.editShortcut(e, start, end) {
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
			b.caret = min(len(b.glyphs), b.caret+1)
			if !e.Shift && start != end {
				b.caret = end
			}
		case rl.KeyUp, rl.KeyDown, rl.KeyHome, rl.KeyEnd:
			positions := b.positions(b.glyphs)
			target := positions[b.caret]
			switch e.Key {
			case rl.KeyUp:
				target.Y -= b.textMetrics().size
			case rl.KeyDown:
				target.Y += b.textMetrics().size
			case rl.KeyHome:
				target.X = 0
			case rl.KeyEnd:
				target.X = b.Bounds().Width
			}
			b.caret = nearestTextPosition(positions, target)
		case rl.KeyBackspace:
			if start == end {
				start = max(0, start-1)
			}
			b.replace(start, end, "")
		case rl.KeyDelete:
			if start == end {
				end = min(len(b.glyphs), end+1)
			}
			b.replace(start, end, "")
		case rl.KeyEnter, rl.KeyKpEnter:
			b.replace(start, end, "\n")
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

func (b *richTextBox) glyphWidth(ch rune) float32 {
	if b.measureText == nil || ch == '\n' {
		return 0
	}
	return b.measureText(string(ch)).X + b.textMetrics().spacing
}

func (b *richTextBox) positions(glyphs []richGlyph) []rl.Vector2 {
	limit := max(1, b.Bounds().Width-2*editInset)
	size := b.textMetrics().size
	positions := make([]rl.Vector2, len(glyphs)+1)
	var current rl.Vector2
	for i, glyph := range glyphs {
		width := b.glyphWidth(glyph.char)
		if glyph.char != '\n' && current.X > 0 && current.X+width > limit {
			current = rl.Vector2{Y: current.Y + size}
		}
		positions[i] = current
		if glyph.char == '\n' {
			current = rl.Vector2{Y: current.Y + size}
		} else {
			current.X += width
		}
	}
	positions[len(glyphs)] = current
	return positions
}

func nearestTextPosition(positions []rl.Vector2, target rl.Vector2) int {
	best, dy, dx := 0, math.Inf(1), math.Inf(1)
	for i, point := range positions {
		row := math.Abs(float64(point.Y - target.Y))
		column := math.Abs(float64(point.X - target.X))
		if row < dy || row == dy && column < dx {
			best, dy, dx = i, row, column
		}
	}
	return best
}

func (b *richTextBox) positionPointer(point rl.Vector2) {
	b.read()
	local := b.localPointer(point)
	b.caret = nearestTextPosition(b.positions(b.glyphs), rl.Vector2{X: local.X - editInset, Y: local.Y - editInset + b.scroll - b.textMetrics().size/2})
	b.InvalidateVisual()
}

func (b *richTextBox) PointerDown(e *InputEvent) {
	if e.Button == PrimaryButton {
		b.positionPointer(e.Position)
		b.anchor = b.caret
		b.dragging = e.CapturePointer()
	}
}
func (b *richTextBox) PointerMove(e *InputEvent) {
	if b.dragging {
		b.positionPointer(e.Position)
	}
}
func (b *richTextBox) PointerUp(*InputEvent)          { b.dragging = false }
func (b *richTextBox) PointerCaptureLost(*InputEvent) { b.dragging = false }

func (b *richTextBox) display() ([]richGlyph, int, int, int) {
	b.read()
	start, end := b.selection()
	if b.preedit.Text == "" {
		return b.glyphs, b.caret, start, end
	}
	style := b.typing
	style.Underline = true
	temporary := richGlyphs(b.preedit.Text, style)
	cursor := start + min(max(0, b.preedit.Cursor), len(temporary))
	highlight := start + min(max(0, b.preedit.SelectionStart), len(temporary))
	highlightEnd := min(start+len(temporary), highlight+max(0, b.preedit.SelectionLength))
	return slices.Replace(slices.Clone(b.glyphs), start, end, temporary...), cursor, highlight, highlightEnd
}

func (b *richTextBox) caretPosition(positions []rl.Vector2, index int) rl.Vector2 {
	caret := positions[index]
	size := b.textMetrics().size
	height := max(size, b.Bounds().Height-2*editInset)
	b.scroll = max(0, min(max(0, positions[len(positions)-1].Y+size-height), max(caret.Y+size-height, min(caret.Y, b.scroll))))
	return rl.Vector2{X: editInset + caret.X, Y: editInset + caret.Y - b.scroll}
}

func (b *richTextBox) IMECaret() rl.Rectangle {
	glyphs, cursor, _, _ := b.display()
	point := b.caretPosition(b.positions(glyphs), cursor)
	return rl.Rectangle{X: point.X, Y: point.Y, Width: 1, Height: b.textMetrics().size}
}

func (b *richTextBox) Primitives(ctx PaintContext) []Node {
	glyphs, cursor, start, end := b.display()
	positions := b.positions(glyphs)
	caret := b.caretPosition(positions, cursor)
	bounds := b.Bounds()
	text := ctx.TextStyle()
	selection := ctx.color("--selection", rl.DarkBlue)
	var nodes []Node
	for i, glyph := range glyphs {
		y := editInset + positions[i].Y - b.scroll
		if y+text.Size <= 0 || y >= bounds.Height {
			continue
		}
		x := bounds.X + editInset + positions[i].X
		y += bounds.Y
		width := b.glyphWidth(glyph.char)
		if b.focused && i >= start && i < end {
			nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: x, Y: y, Width: max(3, width), Height: text.Size}, Style: draw.Style{Fill: selection}})
		}
		if glyph.char == '\n' {
			continue
		}
		style := text
		if glyph.style.Color != (color.RGBA{}) {
			style.Color = glyph.style.Color
		}
		nodes = append(nodes, &TextNode{Position: rl.Vector2{X: x, Y: y}, Text: string(glyph.char), Style: style})
		if glyph.style.Bold {
			nodes = append(nodes, &TextNode{Position: rl.Vector2{X: x + 1, Y: y}, Text: string(glyph.char), Style: style})
		}
		if glyph.style.Underline {
			nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: x, Y: y + text.Size - 1, Width: width, Height: 1}, Style: draw.Style{Fill: style.Color}})
		}
	}
	if b.focused {
		nodes = append(nodes, &RectNode{Rect: rl.Rectangle{X: bounds.X + caret.X, Y: bounds.Y + caret.Y, Width: 1, Height: text.Size}, Style: draw.Style{Fill: ctx.color("caret-color", text.Color)}})
	}
	return nodes
}
