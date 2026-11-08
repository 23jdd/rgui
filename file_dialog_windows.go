package rgui

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	commonDialogs = syscall.NewLazyDLL("comdlg32.dll")
	openFileName  = commonDialogs.NewProc("GetOpenFileNameW")
	saveFileName  = commonDialogs.NewProc("GetSaveFileNameW")
	dialogError   = commonDialogs.NewProc("CommDlgExtendedError")
)

type openFileNameW struct {
	size                         uint32
	owner, instance              uintptr
	filter, customFilter         *uint16
	maxCustomFilter, filterIndex uint32
	file                         *uint16
	maxFile                      uint32
	fileTitle                    *uint16
	maxFileTitle                 uint32
	initialDirectory, title      *uint16
	flags                        uint32
	fileOffset, fileExtension    uint16
	defaultExtension             *uint16
	data, hook                   uintptr
	template                     *uint16
	reserved                     uintptr
	reservedFlags, flagsEx       uint32
}

func fileFilterBuffer(filters []FileFilter) ([]uint16, error) {
	if len(filters) == 0 {
		filters = []FileFilter{{Name: "All files", Pattern: "*.*"}}
	}
	var buffer []uint16
	for _, filter := range filters {
		for _, text := range []string{filter.Name, filter.Pattern} {
			if text == "" {
				return nil, fmt.Errorf("file dialog filters require a name and pattern")
			}
			encoded, err := syscall.UTF16FromString(text)
			if err != nil {
				return nil, err
			}
			buffer = append(buffer, encoded...)
		}
	}
	return append(buffer, 0), nil
}

func showFileDialog(options FileDialogOptions, save bool) (string, error) {
	window := uintptr(rl.GetWindowHandle())
	if window == 0 {
		return "", fmt.Errorf("file dialog requires an open window")
	}
	filters, err := fileFilterBuffer(options.Filters)
	if err != nil {
		return "", err
	}
	title, err := dialogString(options.Title)
	if err != nil {
		return "", err
	}
	directory, err := dialogString(options.InitialDirectory)
	if err != nil {
		return "", err
	}
	extension, err := dialogString(strings.TrimPrefix(options.DefaultExtension, "."))
	if err != nil {
		return "", err
	}
	initial, err := syscall.UTF16FromString(options.InitialName)
	if err != nil {
		return "", err
	}
	filename := make([]uint16, 32768)
	if len(initial) > len(filename) {
		return "", fmt.Errorf("initial filename is too long")
	}
	copy(filename, initial)
	const (
		ofnOverwritePrompt = 0x2
		ofnHideReadOnly    = 0x4
		ofnNoChangeDir     = 0x8
		ofnPathMustExist   = 0x800
		ofnFileMustExist   = 0x1000
		ofnExplorer        = 0x80000
	)
	flags := uint32(ofnExplorer | ofnPathMustExist | ofnNoChangeDir | ofnHideReadOnly)
	proc := openFileName
	if save {
		proc = saveFileName
		flags |= ofnOverwritePrompt
	} else {
		flags |= ofnFileMustExist
	}
	dialog := openFileNameW{owner: window, filter: &filters[0], filterIndex: 1, file: &filename[0], maxFile: uint32(len(filename)),
		title: title, initialDirectory: directory, defaultExtension: extension, flags: flags}
	dialog.size = uint32(unsafe.Sizeof(dialog))
	if ok, _, _ := proc.Call(uintptr(unsafe.Pointer(&dialog))); ok != 0 {
		return syscall.UTF16ToString(filename), nil
	}
	if code, _, _ := dialogError.Call(); code != 0 {
		return "", fmt.Errorf("native file dialog failed (0x%X)", code)
	}
	return "", nil
}

func dialogString(value string) (*uint16, error) {
	if value == "" {
		return nil, nil
	}
	return syscall.UTF16PtrFromString(value)
}
