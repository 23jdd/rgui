package rgui

// SaveFileDialog creates a button that chooses a destination path with an
// overwrite prompt. The callback is responsible for writing the file.
func SaveFileDialog(label string, options FileDialogOptions, result func(string, error)) Node {
	return fileDialogButton(label, options, true, result)
}
