package rgui

import (
	"fmt"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// WindowConfig describes the window and optional font used by Run.
type WindowConfig struct {
	Title       string
	FontPath    string
	ExtraGlyphs string // Characters to load in addition to printable ASCII.

	// Dimensions are in logical pixels. Zero uses the default on that axis.
	Width, Height       int32 // Default: 480 × 300.
	MinWidth, MinHeight int32 // Default: 240 × 160.
	MaxWidth, MaxHeight int32 // Zero leaves that axis unrestricted.

	FixedSize   bool // Disable user resizing.
	Borderless  bool // Hide the native window decorations.
	AlwaysOnTop bool
	VSync       bool
	TargetFPS   int32       // Zero uses 60 FPS; -1 removes the frame-rate cap.
	Background  *color.RGBA // Nil uses the default dark background.

	// RetainedRendering reuses the last painted scene. Custom nodes should
	// invalidate their visual when their paint output changes.
	RetainedRendering bool
	OnFrame           func(FrameStats)
}

type windowSettings struct {
	width, height       int32
	minWidth, minHeight int32
	maxWidth, maxHeight int32
	flags               uint32
	fps                 int32
	background          color.RGBA
}

func (config WindowConfig) settings() (windowSettings, error) {
	width, minWidth, err := windowAxis(config.Width, config.MinWidth, config.MaxWidth, 480, 240)
	if err != nil {
		return windowSettings{}, fmt.Errorf("window width: %w", err)
	}
	height, minHeight, err := windowAxis(config.Height, config.MinHeight, config.MaxHeight, 300, 160)
	if err != nil {
		return windowSettings{}, fmt.Errorf("window height: %w", err)
	}
	if config.TargetFPS < -1 {
		return windowSettings{}, fmt.Errorf("window target FPS must be -1, zero, or positive")
	}
	settings := windowSettings{
		width: width, height: height,
		minWidth: minWidth, minHeight: minHeight,
		maxWidth: config.MaxWidth, maxHeight: config.MaxHeight,
		flags:      rl.FlagWindowHighdpi,
		fps:        config.TargetFPS,
		background: rl.NewColor(23, 28, 38, 255),
	}
	if !config.FixedSize {
		settings.flags |= rl.FlagWindowResizable
	}
	if config.Borderless {
		settings.flags |= rl.FlagWindowUndecorated
	}
	if config.AlwaysOnTop {
		settings.flags |= rl.FlagWindowTopmost
	}
	if config.VSync {
		settings.flags |= rl.FlagVsyncHint
	}
	switch config.TargetFPS {
	case 0:
		settings.fps = 60
	case -1:
		settings.fps = 0
	}
	if config.Background != nil {
		settings.background = *config.Background
	}
	return settings, nil
}

func windowAxis(initial, minimum, maximum, defaultInitial, defaultMinimum int32) (size, lower int32, err error) {
	if initial < 0 || minimum < 0 || maximum < 0 {
		return 0, 0, fmt.Errorf("dimensions must be nonnegative")
	}
	if minimum == 0 {
		minimum = defaultMinimum
	}
	if maximum > 0 && maximum < minimum {
		return 0, 0, fmt.Errorf("maximum %d is smaller than minimum %d", maximum, minimum)
	}
	if initial == 0 {
		initial = defaultInitial
	}
	initial = max(initial, minimum)
	if maximum > 0 {
		initial = min(initial, maximum)
	}
	return initial, minimum, nil
}
