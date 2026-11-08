package rgui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestImageUniformAndCrop(t *testing.T) {
	texture := rl.Texture2D{Width: 200, Height: 100}
	node := Image(texture, ImageUniform).(*imageNode)
	node.Place(rl.Rectangle{Width: 100, Height: 100})
	source, dest := node.geometry()
	if source.Width != 200 || dest.Y != 25 || dest.Height != 50 {
		t.Fatalf("uniform source=%+v dest=%+v", source, dest)
	}
	node.stretch = ImageUniformToFill
	source, dest = node.geometry()
	if source.X != 50 || source.Width != 100 || dest.Width != 100 {
		t.Fatalf("crop source=%+v dest=%+v", source, dest)
	}
}
