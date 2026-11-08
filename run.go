package rgui

import (
	"fmt"
	"os"
	"runtime"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// WindowConfig describes the window and optional font used by Run.
type WindowConfig struct {
	Title       string
	FontPath    string
	ExtraGlyphs string // Characters to load in addition to printable ASCII.
}

// Run opens a window, mounts view once, and runs the UI until the window closes.
// Call Run on the main goroutine. View and click callbacks run on the UI thread.
func Run(config WindowConfig, view func() Node) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rl.SetConfigFlags(rl.FlagWindowHighdpi | rl.FlagWindowResizable)
	rl.InitWindow(480, 300, config.Title)
	if !rl.IsWindowReady() {
		return fmt.Errorf("open raylib window")
	}
	defer rl.CloseWindow()
	rl.SetWindowMinSize(240, 160)
	rl.SetTargetFPS(60)

	font := rl.GetFontDefault()
	if config.FontPath != "" {
		if _, err := os.Stat(config.FontPath); err != nil {
			return fmt.Errorf("load font: %w", err)
		}
		glyphs := make([]rune, 0, 95+len(config.ExtraGlyphs))
		for ch := ' '; ch <= '~'; ch++ {
			glyphs = append(glyphs, ch)
		}
		glyphs = append(glyphs, []rune(config.ExtraGlyphs)...)
		loaded := rl.LoadFontEx(config.FontPath, fontSize, glyphs)
		if loaded.Texture.ID == font.Texture.ID || !rl.IsFontValid(loaded) {
			return fmt.Errorf("raylib could not load font %q", config.FontPath)
		}
		font = loaded
		defer rl.UnloadFont(font)
	}

	root := view().component
	if root == nil {
		return fmt.Errorf("view returned an empty node")
	}
	pending := make(subscriptions)
	defer root.unmount()
	root.mount(pending)

	var click clickState
	var previousViewport rl.Vector2
	style := textStyle(font)
	measure := func(value string) rl.Vector2 { return draw.MeasureText(value, style) }
	for !rl.WindowShouldClose() {
		viewport := rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
		if viewport != previousViewport {
			layoutRoot(root, viewport, measure)
			previousViewport = viewport
		}
		mouse := rl.GetMousePosition()
		hovered := windowButtonAt(root, mouse, viewport)
		click.update(hovered, rl.IsMouseButtonPressed(rl.MouseButtonLeft), rl.IsMouseButtonReleased(rl.MouseButtonLeft))
		if len(pending) > 0 {
			flushBindings(pending)
			layoutRoot(root, viewport, measure)
			hovered = windowButtonAt(root, mouse, viewport)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(23, 28, 38, 255))
		root.draw(font, hovered, click.pressed)
		rl.EndDrawing()
	}
	return nil
}

func windowButtonAt(root *component, mouse, viewport rl.Vector2) *component {
	if mouse.X < 0 || mouse.Y < 0 || mouse.X >= viewport.X || mouse.Y >= viewport.Y {
		return nil
	}
	return root.buttonAt(mouse)
}

type clickState struct {
	pressed *component
}

func (c *clickState) update(hovered *component, down, up bool) {
	if down {
		c.pressed = hovered
	}
	if up {
		pressed := c.pressed
		c.pressed = nil
		if pressed != nil && pressed == hovered {
			pressed.onClick()
		}
	}
}

func flushBindings(pending subscriptions) {
	for b := range pending {
		delete(pending, b)
		b.evaluateTracked()
	}
}
