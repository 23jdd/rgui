package rgui

import (
	"slices"
	"syscall"
	"testing"
)

func TestFileFilterEncoding(t *testing.T) {
	got, err := fileFilterBuffer([]FileFilter{{"文档", "*.txt"}})
	name, _ := syscall.UTF16FromString("文档")
	pattern, _ := syscall.UTF16FromString("*.txt")
	if err != nil || !slices.Equal(got, append(append(name, pattern...), 0)) {
		t.Fatalf("filter=%v err=%v", got, err)
	}
}
