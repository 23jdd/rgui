package rgui

import (
	"math"
	"unicode/utf16"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// IMEClient provides the caret rectangle in the node's content coordinates.
// Implement TextInputHandler for committed text and CompositionHandler for
// preedit. The node must also be able to receive focus.
type IMEClient interface {
	IMECaret() rl.Rectangle
}

// Composition is the uncommitted IME text. Offsets count Unicode code points.
// An empty Text clears preedit without changing the committed text.
type Composition struct {
	Text            string
	Cursor          int
	SelectionStart  int
	SelectionLength int
}

type imeMessage struct {
	composition Composition
	commit      string
}

type imeSession struct {
	platform    *platformIME
	target      *component
	composition Composition
}

func (s *imeSession) setTarget(target *component) {
	if target == s.target {
		return
	}
	previous := s.target
	s.target = target
	s.platform.reset(target != nil)
	if s.composition != (Composition{}) {
		s.composition = Composition{}
		routeEvent(previous, &InputEvent{Kind: CompositionChanged})
	}
}

func (s *imeSession) sync(target *component, active bool, viewport windowViewport) {
	var client IMEClient
	if active && target.inputEnabled() && target.pending != nil {
		node := target.eventNode()
		if private, ok := node.(interface{ suppressIME() bool }); ok && private.suppressIME() {
			s.setTarget(nil)
			s.platform.suspend(true)
			return
		}
		client, _ = node.(IMEClient)
	}
	s.platform.suspend(false)
	if client == nil {
		s.setTarget(nil)
		return
	}
	s.setTarget(target)
	s.platform.position(imeCaretPixels(client.IMECaret(), target.contentBounds(), viewport.scale))
}

func (s *imeSession) drain() bool {
	messages, busy := s.platform.take()
	for _, message := range messages {
		if s.target == nil || s.target.pending == nil || !s.target.inputEnabled() {
			break
		}
		if s.composition != message.composition {
			s.composition = message.composition
			routeEvent(s.target, &InputEvent{Kind: CompositionChanged, Composition: s.composition})
		}
		if message.commit != "" && s.target.pending != nil && s.target.inputEnabled() {
			routeEvent(s.target, &InputEvent{Kind: TextInput, Text: message.commit})
		}
	}
	return busy
}

type imePixelRect struct{ left, top, right, bottom int32 }

func imeCaretPixels(caret, content rl.Rectangle, scale rl.Vector2) imePixelRect {
	x, y := content.X+caret.X, content.Y+caret.Y
	left := int32(math.Floor(float64(x * scale.X)))
	top := int32(math.Floor(float64(y * scale.Y)))
	return imePixelRect{
		left: left, top: top,
		right:  max(left+1, int32(math.Ceil(float64((x+caret.Width)*scale.X)))),
		bottom: max(top+1, int32(math.Ceil(float64((y+caret.Height)*scale.Y)))),
	}
}

func decodeComposition(text []uint16, cursor int, attributes []byte) Composition {
	result := Composition{Text: string(utf16.Decode(text))}
	runes := func(offset int) int {
		offset = min(max(0, offset), len(text))
		if offset > 0 && offset < len(text) && text[offset] >= 0xdc00 && text[offset] <= 0xdfff &&
			text[offset-1] >= 0xd800 && text[offset-1] <= 0xdbff {
			offset--
		}
		return len(utf16.Decode(text[:offset]))
	}
	result.Cursor = runes(cursor)
	result.SelectionStart = result.Cursor
	limit := min(len(text), len(attributes))
	target := func(value byte) bool { return value == 1 || value == 3 }
	for start := 0; start < limit; start++ {
		if !target(attributes[start]) {
			continue
		}
		end := start + 1
		for end < limit && target(attributes[end]) {
			end++
		}
		result.SelectionStart = runes(start)
		result.SelectionLength = runes(end) - result.SelectionStart
		break
	}
	return result
}
