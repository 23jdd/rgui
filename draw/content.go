package draw

import (
	"image/color"
	"strings"
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TextStyle holds the font parameters shared by measurement and drawing.
type TextStyle struct {
	Font    rl.Font
	Size    float32
	Spacing float32
	Color   color.RGBA
}

// MeasureText returns raylib's dimensions for text with the given style.
func MeasureText(text string, style TextStyle) rl.Vector2 {
	return rl.MeasureTextEx(style.Font, text, style.Size, style.Spacing)
}

// Text draws text at its top-left position.
func Text(at rl.Vector2, text string, style TextStyle) {
	rl.DrawTextEx(style.Font, text, at, style.Size, style.Spacing, style.Color)
}

// TextLayout stores measured lines. Create a new layout when the text or width changes.
type TextLayout struct {
	lines      []string
	style      TextStyle
	size       rl.Vector2
	lineHeight float32
}

// TextBlock wraps at whitespace, falling back to rune boundaries for long words
// and text without spaces. A nonpositive width only honors explicit line breaks.
func TextBlock(text string, width, lineGap float32, style TextStyle) TextLayout {
	measure := func(line string) float32 { return MeasureText(line, style).X }
	layout := TextLayout{
		lines: WrapLines(text, width, measure), style: style,
		lineHeight: style.Size + max(0, lineGap),
	}
	for _, line := range layout.lines {
		layout.size.X = max(layout.size.X, measure(line))
	}
	layout.size.Y = style.Size + float32(len(layout.lines)-1)*layout.lineHeight
	return layout
}

// Size returns the width of the longest line and the height of the full block.
func (layout TextLayout) Size() rl.Vector2 { return layout.size }

// Draw uses the same lines, font and spacing used to measure the block.
func (layout TextLayout) Draw(at rl.Vector2) {
	for _, line := range layout.lines {
		Text(at, line, layout.style)
		at.Y += layout.lineHeight
	}
}

// WrapLines breaks text at whitespace or rune boundaries using the supplied
// width metric. A nonpositive width only honors explicit line breaks.
func WrapLines(text string, width float32, measure func(string) float32) []string {
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if width <= 0 {
		return paragraphs
	}
	var lines []string
	for _, paragraph := range paragraphs {
		remaining := []rune(paragraph)
		if len(remaining) == 0 {
			lines = append(lines, "")
		}
		for len(remaining) > 0 {
			end, space := 1, 0
			for end < len(remaining) && measure(string(remaining[:end+1])) <= width {
				if unicode.IsSpace(remaining[end]) {
					space = end
				}
				end++
			}
			if end < len(remaining) && space > 0 && !unicode.IsSpace(remaining[end]) {
				end = space
			}
			lines = append(lines, strings.TrimRightFunc(string(remaining[:end]), unicode.IsSpace))
			remaining = remaining[end:]
			for len(remaining) > 0 && unicode.IsSpace(remaining[0]) {
				remaining = remaining[1:]
			}
		}
	}
	return lines
}

// Image draws a region of a caller-owned texture into dest.
func Image(texture rl.Texture2D, source, dest rl.Rectangle, tint color.RGBA) {
	rl.DrawTexturePro(texture, source, dest, rl.Vector2{}, 0, tint)
}

// NinePatch stretches the center and edges of a caller-owned texture.
// Draw into a destination at least as large as the combined border sizes.
func NinePatch(texture rl.Texture2D, patch rl.NPatchInfo, dest rl.Rectangle, tint color.RGBA) {
	patch.Layout = rl.NPatchNinePatch
	rl.DrawTextureNPatch(texture, patch, dest, rl.Vector2{}, 0, tint)
}
