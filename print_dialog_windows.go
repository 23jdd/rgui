package rgui

import (
	"fmt"
	"syscall"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	printDialogEx   = commonDialogs.NewProc("PrintDlgExW")
	freePrintMemory = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalFree")
	deletePrinterDC = syscall.NewLazyDLL("gdi32.dll").NewProc("DeleteDC")
)

type printDialogExW struct {
	size                                                         uint32
	owner, devMode, devNames, deviceContext                      uintptr
	flags, flags2, exclusionFlags, pageRangeCount, maxPageRanges uint32
	pageRanges                                                   uintptr
	minPage, maxPage, copies                                     uint32
	instance, templateName, callback                             uintptr
	propertyPageCount                                            uint32
	propertyPages                                                uintptr
	startPage, result                                            uint32
}

func showPrintDialog(result func(PrintSettings, error)) {
	if result == nil {
		return
	}
	window := uintptr(rl.GetWindowHandle())
	if window == 0 {
		result(PrintSettings{}, fmt.Errorf("print dialog requires an open window"))
		return
	}
	const (
		pdNoSelection     = 0x4
		pdNoPageNums      = 0x8
		pdCollate         = 0x10
		pdReturnDC        = 0x100
		pdHidePrintToFile = 0x100000
		pdResultPrint     = 1
	)
	dialog := printDialogExW{
		owner: window, flags: pdNoSelection | pdNoPageNums | pdReturnDC | pdHidePrintToFile,
		minPage: 1, maxPage: 1, copies: 1, startPage: ^uint32(0),
	}
	dialog.size = uint32(unsafe.Sizeof(dialog))
	status, _, _ := printDialogEx.Call(uintptr(unsafe.Pointer(&dialog)))
	if dialog.devMode != 0 {
		defer freePrintMemory.Call(dialog.devMode)
	}
	if dialog.devNames != 0 {
		defer freePrintMemory.Call(dialog.devNames)
	}
	if dialog.deviceContext != 0 {
		defer deletePrinterDC.Call(dialog.deviceContext)
	}
	if int32(status) < 0 {
		result(PrintSettings{}, fmt.Errorf("native print dialog failed (0x%08X)", uint32(status)))
		return
	}
	if dialog.result == pdResultPrint {
		result(PrintSettings{DeviceContext: dialog.deviceContext, Copies: max(1, dialog.copies), Collate: dialog.flags&pdCollate != 0}, nil)
	}
}
