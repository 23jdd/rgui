package rgui

import (
	"syscall"
	"unicode/utf16"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	wmIMEStart       = 0x010d
	wmIMEEnd         = 0x010e
	wmIMEComposition = 0x010f
	wmIMESetContext  = 0x0281
	wmIMENotify      = 0x0282
	wmIMEChar        = 0x0286
	imnOpenCandidate = 0x0005
	iscShowPreedit   = 0x80000000
	gcsCompStr       = 0x0008
	gcsCompAttr      = 0x0010
	gcsCursorPos     = 0x0080
	gcsResultStr     = 0x0800
	cfsPoint         = 0x0002
	cfsExclude       = 0x0080
	niCompositionStr = 0x0015
	cpsCancel        = 0x0004
)

var (
	imm32                   = syscall.NewLazyDLL("imm32.dll")
	immGetContext           = imm32.NewProc("ImmGetContext")
	immReleaseContext       = imm32.NewProc("ImmReleaseContext")
	immGetCompositionString = imm32.NewProc("ImmGetCompositionStringW")
	immSetCandidateWindow   = imm32.NewProc("ImmSetCandidateWindow")
	immSetCompositionWindow = imm32.NewProc("ImmSetCompositionWindow")
	immNotifyIME            = imm32.NewProc("ImmNotifyIME")
	comctl32                = syscall.NewLazyDLL("comctl32.dll")
	setWindowSubclass       = comctl32.NewProc("SetWindowSubclass")
	removeWindowSubclass    = comctl32.NewProc("RemoveWindowSubclass")
	defSubclassProc         = comctl32.NewProc("DefSubclassProc")
	imeSubclass             = syscall.NewCallback(imeWindowProc)
	// Run and the subclass callback execute on the same window thread.
	imeWindows = make(map[uintptr]*platformIME)
)

type platformIME struct {
	window        uintptr
	enabled       bool
	composing     bool
	resetting     bool
	hadIMEMessage bool
	caret         imePixelRect
	messages      []imeMessage
	skipChars     []uint16
	highSurrogate uint16
}

func openIME() *platformIME {
	p := &platformIME{}
	window := uintptr(rl.GetWindowHandle())
	if window != 0 {
		ok, _, _ := setWindowSubclass.Call(window, imeSubclass, 1, 0)
		if ok != 0 {
			p.window = window
			imeWindows[window] = p
		}
	}
	return p
}

func (p *platformIME) close() {
	if p.window != 0 {
		p.reset(false)
		removeWindowSubclass.Call(p.window, imeSubclass, 1)
		delete(imeWindows, p.window)
	}
}

func (p *platformIME) reset(enabled bool) {
	if p.window == 0 {
		return
	}
	if p.enabled && p.composing {
		p.resetting = true
		p.withContext(func(context uintptr) { immNotifyIME.Call(context, niCompositionStr, cpsCancel, 0) })
		p.resetting = false
	}
	p.enabled = enabled
	p.composing, p.hadIMEMessage = false, false
	p.messages = nil
	p.skipChars = nil
	p.highSurrogate = 0
	p.caret = imePixelRect{}
}

func (p *platformIME) take() ([]imeMessage, bool) {
	messages, busy := p.messages, p.composing || p.hadIMEMessage
	p.messages, p.hadIMEMessage = nil, false
	p.skipChars = nil
	return messages, busy
}

func (p *platformIME) withContext(use func(uintptr)) {
	context, _, _ := immGetContext.Call(p.window)
	if context != 0 {
		defer immReleaseContext.Call(p.window, context)
		use(context)
	}
}

type imePoint struct{ x, y int32 }

type candidateForm struct {
	index, style uint32
	point        imePoint
	area         imePixelRect
}

type compositionForm struct {
	style uint32
	point imePoint
	area  imePixelRect
}

func (p *platformIME) position(caret imePixelRect) {
	if !p.enabled || p.caret == caret {
		return
	}
	p.caret = caret
	p.placeCandidates(1)
}

func (p *platformIME) placeCandidates(mask uintptr) {
	caret := p.caret
	if caret == (imePixelRect{}) {
		return
	}
	p.withContext(func(context uintptr) {
		composition := compositionForm{style: cfsPoint, point: imePoint{caret.left, caret.top}}
		immSetCompositionWindow.Call(context, uintptr(unsafe.Pointer(&composition)))
		for index := uint32(0); index < 32; index++ {
			if mask&(1<<index) == 0 {
				continue
			}
			candidate := candidateForm{index: index, style: cfsExclude, point: imePoint{caret.left, caret.bottom}, area: caret}
			immSetCandidateWindow.Call(context, uintptr(unsafe.Pointer(&candidate)))
		}
	})
}

func imeWindowProc(window, message, wParam, lParam, _, _ uintptr) uintptr {
	p := imeWindows[window]
	if p != nil && p.enabled {
		switch message {
		case wmIMESetContext:
			lParam &^= iscShowPreedit
		case wmIMEStart:
			if !p.resetting {
				p.skipChars = nil
				p.highSurrogate = 0
				p.composing, p.hadIMEMessage = true, true
				p.placeCandidates(1)
			}
			return 0
		case wmIMEComposition:
			if !p.resetting {
				p.hadIMEMessage = true
				p.readComposition(lParam)
			}
			return 0
		case wmIMEEnd:
			if !p.resetting {
				p.composing, p.hadIMEMessage = false, true
				p.messages = append(p.messages, imeMessage{})
			}
			return 0
		case wmIMEChar:
			if !p.resetting {
				p.hadIMEMessage = true
				p.commitCharacter(uint16(wParam))
			}
			return 0
		case wmIMENotify:
			if wParam == imnOpenCandidate && !p.resetting {
				p.placeCandidates(lParam)
			}
		}
	}
	result, _, _ := defSubclassProc.Call(window, message, wParam, lParam)
	return result
}

func (p *platformIME) readComposition(flags uintptr) {
	p.withContext(func(context uintptr) {
		if flags&gcsResultStr != 0 {
			text := imeString(context, gcsResultStr)
			result := string(utf16.Decode(text))
			p.skipChars = text
			p.messages = append(p.messages, imeMessage{commit: result})
		}
		if flags&(gcsCompStr|gcsCompAttr|gcsCursorPos) != 0 {
			text := imeString(context, gcsCompStr)
			cursor, _, _ := immGetCompositionString.Call(context, gcsCursorPos, 0, 0)
			attributes := imeAttributes(context)
			p.messages = append(p.messages, imeMessage{composition: decodeComposition(text, int(int32(cursor)), attributes)})
		} else if flags == 0 {
			p.messages = append(p.messages, imeMessage{})
		}
	})
}

func (p *platformIME) commitCharacter(unit uint16) {
	if len(p.skipChars) > 0 && p.skipChars[0] == unit {
		p.skipChars = p.skipChars[1:]
		return
	}
	p.skipChars = nil
	if unit >= 0xd800 && unit <= 0xdbff {
		p.highSurrogate = unit
		return
	}
	character := rune(unit)
	if p.highSurrogate != 0 {
		if unit >= 0xdc00 && unit <= 0xdfff {
			character = utf16.DecodeRune(rune(p.highSurrogate), rune(unit))
		} else {
			p.messages = append(p.messages, imeMessage{commit: "\uFFFD"})
		}
		p.highSurrogate = 0
	} else if unit >= 0xdc00 && unit <= 0xdfff {
		character = '\uFFFD'
	}
	p.messages = append(p.messages, imeMessage{commit: string(character)})
}

func imeString(context, kind uintptr) []uint16 {
	size, _, _ := immGetCompositionString.Call(context, kind, 0, 0)
	if int32(size) <= 0 {
		return nil
	}
	text := make([]uint16, size/2)
	if len(text) == 0 {
		return nil
	}
	n, _, _ := immGetCompositionString.Call(context, kind, uintptr(unsafe.Pointer(&text[0])), size)
	if int32(n) <= 0 || n > size {
		return nil
	}
	return text[:n/2]
}

func imeAttributes(context uintptr) []byte {
	size, _, _ := immGetCompositionString.Call(context, gcsCompAttr, 0, 0)
	if int32(size) <= 0 {
		return nil
	}
	attributes := make([]byte, size)
	n, _, _ := immGetCompositionString.Call(context, gcsCompAttr, uintptr(unsafe.Pointer(&attributes[0])), size)
	if int32(n) <= 0 || n > size {
		return nil
	}
	return attributes[:n]
}
