//go:build !windows

package rgui

import rl "github.com/gen2brain/raylib-go/raylib"

func resizeScale() rl.Vector2 { return rl.Vector2{X: 1, Y: 1} }
