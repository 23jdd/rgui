package rgui

import (
	"image/color"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestDefaultWindowSettings(t *testing.T) {
	got, err := (WindowConfig{}).settings()
	if err != nil {
		t.Fatal(err)
	}
	want := windowSettings{
		width: 480, height: 300, minWidth: 240, minHeight: 160,
		flags: rl.FlagWindowHighdpi | rl.FlagWindowResizable, fps: 60,
		background: color.RGBA{R: 20, G: 26, B: 36, A: 255},
	}
	if got != want {
		t.Fatalf("default settings = %+v, want %+v", got, want)
	}
}

func TestCustomWindowSettings(t *testing.T) {
	config := WindowConfig{
		Width: 180, Height: 700,
		MinWidth: 200, MinHeight: 120, MaxHeight: 480,
		FixedSize: true, Borderless: true, AlwaysOnTop: true, VSync: true,
		TargetFPS: -1, Background: &color.RGBA{},
	}
	got, err := config.settings()
	if err != nil {
		t.Fatal(err)
	}
	want := windowSettings{
		width: 200, height: 480, minWidth: 200, minHeight: 120, maxHeight: 480,
		flags: rl.FlagWindowHighdpi | rl.FlagWindowUndecorated | rl.FlagWindowTopmost | rl.FlagVsyncHint,
	}
	if got != want {
		t.Fatalf("custom settings = %+v, want %+v", got, want)
	}
}

func TestInvalidWindowSettings(t *testing.T) {
	for _, tc := range []struct {
		config WindowConfig
		want   string
	}{
		{WindowConfig{Width: -1}, "window width"},
		{WindowConfig{MinHeight: 200, MaxHeight: 199}, "window height"},
		{WindowConfig{MaxWidth: 200}, "window width"},
		{WindowConfig{TargetFPS: -2}, "target FPS"},
	} {
		err := Run(tc.config, func() Node {
			t.Fatal("invalid settings must fail before mounting the view")
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("Run(%+v) = %v, want an error containing %q", tc.config, err, tc.want)
		}
	}
}
