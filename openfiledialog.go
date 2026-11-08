package rgui

// FileFilter is a native dialog filter, such as {"Text files", "*.txt"}.
type FileFilter struct{ Name, Pattern string }

type FileDialogOptions struct {
	Title, InitialDirectory, InitialName, DefaultExtension string
	Filters                                                []FileFilter
}

// OpenFileDialog creates a button that opens a native file picker. The
// callback receives a path or an error. Cancel leaves it uncalled.
func OpenFileDialog(label string, options FileDialogOptions, result func(string, error)) Node {
	return fileDialogButton(label, options, false, result)
}

func fileDialogButton(label string, options FileDialogOptions, save bool, result func(string, error)) Node {
	return Button(label, func() {
		path, err := showFileDialog(options, save)
		if result != nil && (path != "" || err != nil) {
			result(path, err)
		}
	})
}
