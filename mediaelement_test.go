package rgui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"slices"
	"testing"
	"time"
)

type testMedia struct{ calls []string }

func (m *testMedia) Play()                    { m.calls = append(m.calls, "play") }
func (m *testMedia) Pause()                   { m.calls = append(m.calls, "pause") }
func (m *testMedia) Stop()                    { m.calls = append(m.calls, "stop") }
func (m *testMedia) Update(time.Time) bool    { m.calls = append(m.calls, "update"); return true }
func (*testMedia) VideoTexture() rl.Texture2D { return rl.Texture2D{} }

func TestMediaElementTransitionsAndStopsOnUnmount(t *testing.T) {
	backend := &testMedia{}
	state := NewSignal(MediaStopped)
	root := componentOf(MediaElement(backend, state))
	root.mount(make(subscriptions))
	state.Set(MediaPlaying)
	tickControls(root, time.Now(), true, nil)
	state.Set(MediaPaused)
	tickControls(root, time.Now(), true, nil)
	root.unmount()
	if !slices.Equal(backend.calls, []string{"play", "update", "pause", "stop"}) {
		t.Fatalf("calls = %v", backend.calls)
	}
}
