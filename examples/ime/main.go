package main

import (
	"log"
	"os"
	"path/filepath"

	"rgui"
	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type field struct {
	rgui.BaseNode
	text        []rune
	preedit     rgui.Composition
	measureText func(string) rl.Vector2
	focused     bool
}

func (f *field) Measure(ctx rgui.MeasureContext) rl.Vector2 {
	f.measureText = ctx.MeasureText
	return rl.Vector2{X: min(360, ctx.Available.X), Y: 48}
}

func (f *field) FocusGained(*rgui.InputEvent) { f.focused = true }
func (f *field) FocusLost(*rgui.InputEvent)   { f.focused = false }

func (f *field) CompositionChanged(e *rgui.InputEvent) {
	f.preedit = e.Composition
	e.Handled = true
}

func (f *field) TextInput(e *rgui.InputEvent) {
	f.text = append(f.text, []rune(e.Text)...)
	e.Handled = true
}

func (f *field) KeyDown(e *rgui.InputEvent)   { f.backspace(e) }
func (f *field) KeyRepeat(e *rgui.InputEvent) { f.backspace(e) }

func (f *field) backspace(e *rgui.InputEvent) {
	if e.Key == rl.KeyBackspace && len(f.text) != 0 {
		f.text = f.text[:len(f.text)-1]
		e.Handled = true
	}
}

func (f *field) textPositions() (start, preedit, caret float32) {
	uncommitted := []rune(f.preedit.Text)
	cursor := min(max(0, f.preedit.Cursor), len(uncommitted))
	committedWidth := f.measureText(string(f.text)).X
	caretWidth := f.measureText(string(f.text) + string(uncommitted[:cursor])).X
	scroll := max(0, caretWidth+24-f.Bounds().Width)
	start = 12 - scroll
	return start, start + committedWidth, start + caretWidth
}

func (f *field) IMECaret() rl.Rectangle {
	_, _, caret := f.textPositions()
	return rl.Rectangle{X: caret, Y: 12, Width: 1, Height: 24}
}

func (f *field) Primitives(ctx rgui.PaintContext) []rgui.Node {
	bounds := f.Bounds()
	border := rl.Gray
	if f.focused {
		border = rl.SkyBlue
	}
	start, preedit, caret := f.textPositions()
	y := bounds.Y + 12
	nodes := []rgui.Node{
		&rgui.RectNode{Rect: bounds, Radius: 6, Style: draw.Style{
			Fill: rl.DarkGray, Stroke: draw.Stroke{Color: border, Width: 2},
		}},
		&rgui.TextNode{Position: rl.Vector2{X: bounds.X + start, Y: y}, Text: string(f.text),
			Style: draw.TextStyle{Font: ctx.Font, Size: 24, Spacing: 1, Color: rl.RayWhite}},
	}
	if f.preedit.Text != "" {
		nodes = append(nodes, &rgui.TextNode{
			Position: rl.Vector2{X: bounds.X + preedit, Y: y}, Text: f.preedit.Text,
			Style: draw.TextStyle{Font: ctx.Font, Size: 24, Spacing: 1, Color: rl.SkyBlue},
		})
	}
	if f.focused {
		nodes = append(nodes, &rgui.RectNode{
			Rect:  rl.Rectangle{X: bounds.X + caret, Y: y, Width: 1, Height: 24},
			Style: draw.Style{Fill: rl.RayWhite},
		})
	}
	return nodes
}

func view() rgui.Node {
	return rgui.StackPanel(rgui.Vertical, 12,
		rgui.Text(func() string { return "Click the field and type with an IME." }),
		&field{},
		rgui.Button("Move focus here", func() {}),
	)
}

func main() {
	config := rgui.WindowConfig{
		Title: "rgui Windows IME", Width: 600,
		FontPath:    filepath.Join(os.Getenv("WINDIR"), "Fonts", "Deng.ttf"),
		ExtraGlyphs: "你好世界中文输入测试拼音候选窗口我爱北京上海广州深圳中华人民共和国",
	}
	if err := rgui.Run(config, view); err != nil {
		log.Fatal(err)
	}
}
