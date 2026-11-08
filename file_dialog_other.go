//go:build !windows

package rgui

import "fmt"

func showFileDialog(FileDialogOptions, bool) (string, error) {
	return "", fmt.Errorf("native file dialogs are available on Windows")
}
