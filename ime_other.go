//go:build !windows

package rgui

type platformIME struct{}

func openIME() *platformIME                     { return &platformIME{} }
func (*platformIME) close()                     {}
func (*platformIME) reset(bool)                 {}
func (*platformIME) position(imePixelRect)      {}
func (*platformIME) take() ([]imeMessage, bool) { return nil, false }
