package goLMUSharedMemory

import (
	"fmt"
	"unsafe"
)

// CStringToString converts a C-style null-terminated byte array to a Go string
func CStringToString(bytes []byte) string {
	for i, b := range bytes {
		if b == 0 {
			return string(bytes[:i])
		}
	}
	return string(bytes)
}

// CDoubleToFloat64 converts a C-style double (8 bytes) to a Go float64
func CDoubleToFloat64(bytes []byte) float64 {
	if len(bytes) < 8 {
		return 0
	}
	return *(*float64)(unsafe.Pointer(&bytes[0]))
}

// CLongLongToUint64 converts a C-style long long (8 bytes) to a Go uint64
func CLongLongToUint64(bytes []byte) uint64 {
	if len(bytes) < 8 {
		return 0
	}
	return *(*uint64)(unsafe.Pointer(&bytes[0]))
}

// GetTireCompoundName returns a human-readable tire compound name from the compound index of a wheel
// Mapping: 0=Mediums, 1=Softs, 2=Hards, 3=Wets
func GetTireCompoundName(wheel *TelemWheelV01) string {
	index := wheel.MCompoundIndex

	if index <= 10 {
		switch index {
		case 0:
			return "Mediums"
		case 1:
			return "Softs"
		case 2:
			return "Hards"
		case 3:
			return "Wets"
		default:
			// Index is in range but not a standard value
			return fmt.Sprintf("Idx:%d", index)
		}
	}

	// No valid data
	return "N/A"
}

