//go:build !windows

package rgui

import "fmt"

func showPrintDialog(result func(PrintSettings, error)) {
	if result != nil {
		result(PrintSettings{}, fmt.Errorf("native print dialog is available on Windows"))
	}
}
