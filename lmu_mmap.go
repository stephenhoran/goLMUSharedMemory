package goLMUSharedMemory

// Windows shared memory mapping functionality for accessing LMU telemetry data.

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procOpenFileMapping      = kernel32.NewProc("OpenFileMappingW")
	procMapViewOfFile        = kernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile      = kernel32.NewProc("UnmapViewOfFile")
	procCloseHandle          = kernel32.NewProc("CloseHandle")
	procGetLastError         = kernel32.NewProc("GetLastError")
	procFormatMessage        = kernel32.NewProc("FormatMessageW")
)

const (
	FORMAT_MESSAGE_FROM_SYSTEM = 0x00001000
	FORMAT_MESSAGE_IGNORE_INSERTS = 0x00000200
)

const (
	FILE_MAP_READ       = 0x0004
	FILE_MAP_WRITE      = 0x0002
	FILE_MAP_ALL_ACCESS = 0xF001F // Standard rights (0xF0000) | SYNCHRONIZE (0x100000) | FILE_MAP_ALL_ACCESS (0x1F)
)

// SharedMemory represents a mapped shared memory segment
type SharedMemory struct {
	handle  syscall.Handle
	data    []byte
	dataPtr uintptr // Store original pointer for unmapping
	closed  bool
}

// getLastError retrieves the last Windows error code
func getLastError() uint32 {
	ret, _, _ := procGetLastError.Call()
	return uint32(ret)
}

// formatWindowsError formats a Windows error code into a readable message
func formatWindowsError(code uint32) string {
	buf := make([]uint16, 256)
	ret, _, _ := procFormatMessage.Call(
		FORMAT_MESSAGE_FROM_SYSTEM|FORMAT_MESSAGE_IGNORE_INSERTS,
		0,
		uintptr(code),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0,
	)
	if ret == 0 {
		return fmt.Sprintf("error code %d", code)
	}
	// Remove trailing \r\n
	msg := syscall.UTF16ToString(buf)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	if len(msg) > 0 && msg[len(msg)-1] == '\r' {
		msg = msg[:len(msg)-1]
	}
	return msg
}

// OpenSharedMemory opens and maps the LMU shared memory segment
// Returns a SharedMemory instance that must be closed when done
// Uses the default shared memory name from SharedMemoryName constant
func OpenSharedMemory() (*SharedMemory, error) {
	return OpenSharedMemoryWithName(SharedMemoryName)
}

// OpenSharedMemoryWithName opens and maps a shared memory segment with the specified name
// This allows testing with different shared memory names if the default doesn't work
// Returns a SharedMemory instance that must be closed when done
func OpenSharedMemoryWithName(name string) (*SharedMemory, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("failed to convert memory name to UTF16: %w", err)
	}

	// Open the existing file mapping
	// Use FILE_MAP_ALL_ACCESS like the LMU example does
	handle, _, _ := procOpenFileMapping.Call(
		uintptr(FILE_MAP_ALL_ACCESS), // dwDesiredAccess
		0,                            // bInheritHandle
		uintptr(unsafe.Pointer(namePtr)), // lpName
	)
	
	if handle == 0 {
		errCode := getLastError()
		errMsg := formatWindowsError(errCode)
		return nil, fmt.Errorf("failed to open shared memory mapping %s (error %d: %s). Make sure LMU is running and shared memory is enabled", name, errCode, errMsg)
	}

	// Map the view of the file
	// Use FILE_MAP_ALL_ACCESS like the LMU example does
	// Pass 0 as size to map the entire shared memory segment
	// This avoids alignment/size issues between Go and C++ structs
	dataPtr, _, _ := procMapViewOfFile.Call(
		uintptr(handle),           // hFileMappingObject
		uintptr(FILE_MAP_ALL_ACCESS), // dwDesiredAccess
		0,                         // dwFileOffsetHigh
		0,                         // dwFileOffsetLow
		0,                         // dwNumberOfBytesToMap - 0 means map entire segment
	)

	if dataPtr == 0 {
		// Get error immediately - must be before any other API calls
		errCode := getLastError()
		procCloseHandle.Call(uintptr(handle))
		
		if errCode == 0 {
			// This can happen if the shared memory exists but is not properly initialized
			// or if the size we're requesting is larger than what was created
			return nil, fmt.Errorf("failed to map view of file: MapViewOfFile returned NULL with no error code. This may indicate the shared memory size is incorrect or LMU hasn't fully initialized it yet")
		}
		errMsg := formatWindowsError(errCode)
		return nil, fmt.Errorf("failed to map view of file (error %d: %s). The shared memory might not be initialized yet", errCode, errMsg)
	}

	// Convert pointer to byte slice
	// Since we mapped the entire segment with size 0, use the expected struct size
	// This gives us a byte slice view of the mapped memory
	layoutSize := unsafe.Sizeof(SharedMemoryLayout{})
	data := (*[1 << 30]byte)(unsafe.Pointer(dataPtr))[:layoutSize:layoutSize]

	return &SharedMemory{
		handle:  syscall.Handle(handle),
		data:    data,
		dataPtr: dataPtr,
		closed:  false,
	}, nil
}

// Close unmaps and closes the shared memory segment
func (sm *SharedMemory) Close() error {
	if sm.closed {
		return nil
	}

	// Unmap the view of the file (use original pointer)
	ret, _, _ := procUnmapViewOfFile.Call(sm.dataPtr)
	if ret == 0 {
		return fmt.Errorf("failed to unmap view of file")
	}

	// Close the handle
	ret, _, _ = procCloseHandle.Call(uintptr(sm.handle))
	if ret == 0 {
		return fmt.Errorf("failed to close handle")
	}

	sm.closed = true
	return nil
}

// ReadData reads the shared memory data into a SharedMemoryLayout structure
func (sm *SharedMemory) ReadData() (*SharedMemoryLayout, error) {
	if sm.closed {
		return nil, fmt.Errorf("shared memory is closed")
	}

	// Validate shared memory data size against the expected struct size
	if len(sm.data) < (&SharedMemoryLayout{}).Size() {
		return nil, fmt.Errorf("shared memory data is too small")
	}

	// Cast the byte slice to the struct
	data := (*SharedMemoryLayout)(unsafe.Pointer(&sm.data[0]))
	return data, nil
}

// GetRawData returns the raw byte slice of the shared memory
// This is useful for debugging or custom parsing
func (sm *SharedMemory) GetRawData() ([]byte, error) {
	if sm.closed {
		return nil, fmt.Errorf("shared memory is closed")
	}

	return sm.data, nil
}

