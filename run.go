package rgui

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"rgui/draw"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Run opens a window, mounts view once, and runs the UI until the window closes.
// Call Run on the main goroutine. View and click callbacks run on the UI thread.
func Run(config WindowConfig, view func() Node) error {
	settings, err := config.settings()
	if err != nil {
		return err
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rl.SetConfigFlags(settings.flags)
	rl.InitWindow(settings.width, settings.height, config.Title)
	if !rl.IsWindowReady() {
		return fmt.Errorf("open raylib window")
	}
	defer rl.CloseWindow()
	// Escape belongs to the focused input while the IME is composing.
	rl.SetExitKey(rl.KeyNull)
	// Raylib retains its window flags across CloseWindow.
	defer rl.ClearWindowState(settings.flags &^ rl.FlagWindowHighdpi)
	// Clear old limits before applying a new minimum.
	rl.SetWindowMaxSize(0, 0)
	rl.SetWindowMinSize(int(settings.minWidth), int(settings.minHeight))
	rl.SetWindowMaxSize(int(settings.maxWidth), int(settings.maxHeight))
	rl.SetTargetFPS(settings.fps)

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

	root := componentOf(view())
	if root == nil {
		return fmt.Errorf("view returned an empty node")
	}
	if root.parent != nil || root.pending != nil {
		return fmt.Errorf("view root already belongs to a tree")
	}
	pending := make(subscriptions)
	defer root.unmount()
	root.mount(pending)
	ime := imeSession{platform: openIME()}
	defer ime.platform.close()
	defer ime.setTarget(nil)

	var click clickState
	var focus focusState
	var input inputRouter
	var previousViewport windowViewport
	var previousPaint paintState
	var retained retainedFrame
	defer retained.close()
	style := textStyle(font)
	measure := func(value string) rl.Vector2 { return draw.MeasureText(value, style) }
	for !rl.WindowShouldClose() {
		var stats FrameStats
		layoutUpdated := false
		updateLayout := func(viewport rl.Vector2) {
			start := time.Now()
			layoutRoot(root, viewport, measure)
			stats.Layout += time.Since(start)
			layoutUpdated = true
		}
		pruneInput(&click, &focus)
		input.prune()
		window := viewportForPixels(int32(rl.GetRenderWidth()), int32(rl.GetRenderHeight()), rl.GetWindowScaleDPI())
		viewport := window.size
		if window != previousViewport || root.layoutChanged() {
			root.layoutScale = window.scale
			updateLayout(viewport)
			previousViewport = window
		}
		// raylib already reports the pointer in logical coordinates.
		mouse := rl.GetMousePosition()
		active := rl.IsWindowFocused()
		ime.sync(focus.button, active, window)
		imeBusy := ime.drain()
		var hovered *component
		if active {
			if !focus.button.focusable(viewport) {
				focus.set(nil)
			}
			hit := hitWindow(root, mouse, viewport)
			input.updateHover(hit, mouse)
			if hit != nil && hit.pending == nil {
				hit = nil
			}
			hovered = nearestButton(hit)
			mouseDown := rl.IsMouseButtonPressed(rl.MouseButtonLeft)
			if mouseDown {
				if !input.pointer(PointerDown, hit, mouse, rl.Vector2{}) {
					if hit != nil && hit.pending == nil {
						hit, hovered = nil, nil
					}
					focus.cancelPress()
					focus.set(nearestFocusable(hit, viewport))
					click.update(hovered, true, false)
				}
			}
			if mouse != input.point {
				input.pointer(PointerMove, hit, mouse, rl.Vector2{})
			}
			if wheel := rl.GetMouseWheelMoveV(); wheel != (rl.Vector2{}) {
				input.pointer(PointerWheel, hit, mouse, wheel)
			}
			if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
				if input.pointer(PointerUp, hit, mouse, rl.Vector2{}) {
					click.pressed = nil
				} else {
					click.update(hovered, false, true)
				}
			}
			if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
				input.secondary(PointerDown, hit, mouse)
			}
			if rl.IsMouseButtonReleased(rl.MouseButtonRight) {
				input.secondary(PointerUp, hit, mouse)
			}
			if focus.readKeyboard(root, viewport, imeBusy) {
				break
			}
		} else {
			click.pressed = nil
			focus.cancelPress()
			focus.keys = nil
			input.cancel()
			input.updateHover(nil, mouse)
		}
		bindingsChanged := len(pending) > 0
		flushBindings(pending)
		pruneInput(&click, &focus)
		input.prune()
		if bindingsChanged || root.layoutChanged() {
			updateLayout(viewport)
			if active {
				hit := hitWindow(root, mouse, viewport)
				input.updateHover(hit, mouse)
				hovered = nearestButton(hit)
			}
		}
		pressed := click.pressed
		if pressed != hovered {
			pressed = nil
		}
		var focused *component
		if active && focus.button.focusable(viewport) {
			focused = focus.button
			if focus.key != 0 {
				pressed = focused
			}
		}
		ime.sync(focused, active, window)

		state := paintState{hovered: hovered, pressed: pressed, focused: focused}
		rl.BeginDrawing()
		start := time.Now()
		var cached, fresh bool
		if config.RetainedRendering {
			cached, fresh = retained.prepare(window)
		}
		if cached {
			damage := unionRect(root.takeVisualDamage(), state.damage(previousPaint))
			if layoutUpdated || fresh {
				damage = window.fullDamage()
			}
			pixels := window.pixelDamage(damage)
			if nonempty(pixels) {
				stats.Damage = window.logicalDamage(pixels)
				stats.Painted = true
				stats.PaintedPixels = int64(pixels.Width) * int64(pixels.Height)
				retained.update(stats.Damage, settings.background, func(canvas *draw.Canvas) {
					root.paint(canvas, font, hovered, pressed, focused, &stats.Damage)
				})
			}
		} else {
			stats.Painted = true
			stats.Damage = window.fullDamage()
			stats.PaintedPixels = int64(window.width) * int64(window.height)
			rl.ClearBackground(settings.background)
			canvas := draw.NewCanvas(window.scale, window.height)
			root.paint(&canvas, font, hovered, pressed, focused, nil)
			rl.DrawRenderBatchActive()
		}
		stats.Paint = time.Since(start)
		start = time.Now()
		if cached {
			retained.present()
		}
		rl.EndDrawing()
		stats.Present = time.Since(start)
		previousPaint = state
		if config.OnFrame != nil {
			config.OnFrame(stats)
		}
	}
	return nil
}

func windowButtonAt(root *component, mouse, viewport rl.Vector2) *component {
	return nearestButton(hitWindow(root, mouse, viewport))
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
		if pressed == hovered && pressed.enabledButton() {
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
