package rgui

import (
	"os"
	"runtime"
	"slices"
	"syscall"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestIMECharacterFallbackDeduplicatesAndDecodes(t *testing.T) {
	p := &platformIME{
		skipChars: []uint16{'中'},
		messages:  []imeMessage{{commit: "中"}},
	}
	p.commitCharacter('中')
	p.commitCharacter(0xd83d)
	p.commitCharacter(0xde00)
	messages, _ := p.take()
	if !slices.Equal(messages, []imeMessage{{commit: "中"}, {commit: "😀"}}) {
		t.Fatalf("committed messages = %+v", messages)
	}
}

type imeTestField struct {
	BaseNode
	compositions []Composition
	text         string
}

func (*imeTestField) IMECaret() rl.Rectangle {
	return rl.Rectangle{X: 4, Y: 8, Width: 1, Height: 16}
}
func (f *imeTestField) TextInput(e *InputEvent) { f.text += e.Text }
func (f *imeTestField) CompositionChanged(e *InputEvent) {
	f.compositions = append(f.compositions, e.Composition)
}

func TestWindowsIMEMessageBridge(t *testing.T) {
	if os.Getenv("RGUI_WINDOW_TEST") != "1" {
		t.Skip("set RGUI_WINDOW_TEST=1 to exercise native Windows messages")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(96, 72, "rgui IME test")
	if !rl.IsWindowReady() {
		t.Fatal("could not initialize raylib")
	}
	defer rl.CloseWindow()
	platform := openIME()
	if platform.window == 0 {
		t.Fatal("could not subclass the raylib HWND")
	}
	defer platform.close()
	field := &imeTestField{}
	target := componentOf(field)
	target.mount(make(subscriptions))
	defer target.unmount()
	testLayout(target, 96, 72)
	session := imeSession{platform: platform}
	viewport := viewportForPixels(144, 108, rl.Vector2{X: 1.5, Y: 1.5})
	session.sync(target, true, viewport)
	if platform.caret != (imePixelRect{left: 6, top: 12, right: 8, bottom: 36}) {
		t.Fatalf("native caret = %+v", platform.caret)
	}
	send := syscall.NewLazyDLL("user32.dll").NewProc("SendMessageW")
	send.Call(platform.window, wmIMEStart, 0, 0)
	if !session.drain() || !platform.composing {
		t.Fatal("WM_IME_STARTCOMPOSITION did not start the session")
	}
	preedit := Composition{Text: "ni", Cursor: 2, SelectionStart: 2}
	platform.messages = append(platform.messages, imeMessage{composition: preedit})
	session.drain()
	send.Call(platform.window, wmIMEChar, '你', 0)
	send.Call(platform.window, wmIMEEnd, 0, 0)
	if !session.drain() || platform.composing || field.text != "你" ||
		!slices.Equal(field.compositions, []Composition{preedit, {}}) {
		t.Fatalf("unexpected commit: %q, preedit history %+v", field.text, field.compositions)
	}
	platform.messages = append(platform.messages, imeMessage{composition: preedit})
	session.drain()
	platform.messages = append(platform.messages, imeMessage{commit: "stale"})
	session.sync(nil, false, viewport)
	session.drain()
	if field.text != "你" || field.compositions[len(field.compositions)-1] != (Composition{}) {
		t.Fatal("losing focus must clear preedit and discard stale commits")
	}
}
