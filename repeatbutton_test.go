package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
	"time"
)

func TestRepeatButtonStopsOnRelease(t *testing.T) {
	count := 0
	button := RepeatButton("Repeat", time.Second, time.Second, func() { count++ }).(*repeatButton)
	host := componentOf(button)
	button.KeyDown(&InputEvent{Source: button, Key: rl.KeySpace})
	button.tick(button.next.Add(-time.Millisecond), true, host)
	if count != 1 {
		t.Fatal("must respect delay")
	}
	button.tick(button.next, true, host)
	button.KeyUp(&InputEvent{Source: button, Key: rl.KeySpace})
	button.tick(button.next, true, host)
	if count != 2 {
		t.Fatal("release should stop repeats")
	}
}
