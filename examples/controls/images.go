package main

import (
	_ "embed"
	"fmt"

	"rgui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed orbit.png
var orbitPNG []byte

func loadGalleryTexture() (rl.Texture2D, error) {
	bitmap := rl.LoadImageFromMemory(".png", orbitPNG, int32(len(orbitPNG)))
	if bitmap == nil {
		return rl.Texture2D{}, fmt.Errorf("decode orbit.png")
	}
	defer rl.UnloadImage(bitmap)
	if !rl.IsImageValid(bitmap) {
		return rl.Texture2D{}, fmt.Errorf("decode orbit.png")
	}
	texture := rl.LoadTextureFromImage(bitmap)
	if !rl.IsTextureValid(texture) {
		return texture, fmt.Errorf("upload orbit.png texture")
	}
	rl.SetTextureFilter(texture, rl.FilterPoint)
	return texture, nil
}

func images(texture rl.Texture2D) rgui.Node {
	card := func(label string, stretch rgui.ImageStretch, classes ...string) rgui.Node {
		return styled(rgui.GroupBox(label, styled(rgui.Image(texture, stretch), "image-preview")), classes...)
	}
	grid := rgui.Grid([]rgui.Track{rgui.Auto(), rgui.Auto()}, []rgui.Track{rgui.Star(1), rgui.Star(1)},
		rgui.Cell{Child: card("UNIFORM", rgui.ImageUniform)},
		rgui.Cell{Column: 1, Child: card("FILL", rgui.ImageFill, "side-panel")},
		rgui.Cell{Row: 1, Child: card("UNIFORM TO FILL", rgui.ImageUniformToFill, "image-bottom")},
		rgui.Cell{Row: 1, Column: 1, Child: card("NONE", rgui.ImageNone, "image-bottom", "side-panel")},
	)
	return rgui.ScrollViewer(grid, nil, false, true)
}
