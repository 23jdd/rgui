package rgui

// PrintSettings contains the native printer device context and requested
// copies. DeviceContext is valid only during the PrintDialog callback.
type PrintSettings struct {
	DeviceContext uintptr
	Copies        uint32
	Collate       bool
}

// PrintDialog opens the system printer picker. The callback renders the job
// using the selected device context. Cancel leaves it uncalled.
func PrintDialog(label string, result func(PrintSettings, error)) Node {
	return Button(label, func() { showPrintDialog(result) })
}
