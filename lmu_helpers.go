package goLMUSharedMemory

import (
	"fmt"
	"strings"
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

// GetTireCompoundName returns a human-readable tire compound name from the raw name bytes and index
// Priority: Checks name string first (more reliable), then falls back to index mapping
// Mapping: 0=Mediums, 1=Softs, 2=Hards, 3=Wets (but name is preferred)
func GetTireCompoundName(nameBytes []byte, index uint8) string {
	rawName := CStringToString(nameBytes)
	nameLower := strings.ToLower(strings.TrimSpace(rawName))

	// Debug: If name contains "dium", it's likely "Mediums" truncated
	// Check for partial matches that might indicate truncation
	if strings.HasPrefix(nameLower, "dium") || strings.HasSuffix(nameLower, "dium") {
		return "Mediums"
	}
	if strings.HasPrefix(nameLower, "edium") || strings.Contains(nameLower, "edium") {
		return "Mediums"
	}
	if strings.HasPrefix(nameLower, "ium") && !strings.Contains(nameLower, "soft") {
		return "Mediums"
	}

	// Check for tire compound keywords - check "medium" first to avoid matching "soft" incorrectly
	if strings.Contains(nameLower, "medium") {
		return "Mediums"
	}
	// Check if name starts with "soft" to avoid false matches in words containing "soft"
	if strings.HasPrefix(nameLower, "soft") || (strings.Contains(nameLower, "soft") && !strings.Contains(nameLower, "medium")) {
		return "Softs"
	}
	if strings.Contains(nameLower, "wet") {
		return "Wets"
	}
	if strings.Contains(nameLower, "hard") {
		return "Hards"
	}

	// If name exists but doesn't match known patterns, use it as-is (might be custom compound name)
	// But if it looks truncated (too short), prefer index
	if len(rawName) > 0 {
		trimmed := strings.TrimSpace(rawName)
		if len(trimmed) > 3 { // Only use if it's a reasonable length
			return trimmed
		}
	}

	// If no name or name doesn't match, try to use the index
	// LMU mapping: 0=mediums, 1=softs (verified from debug output)
	// Note: Different games may use different mappings, but LMU uses 0 for mediums
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

// GetTireCompoundNameFromTelem is a convenience function that extracts tire compound name from TelemInfoV01
// for either front or rear tires
func GetTireCompoundNameFromTelem(telem *TelemInfoV01, isFront bool) string {
	if telem == nil {
		return "N/A"
	}
	if isFront {
		return GetTireCompoundName(telem.MFrontTireCompoundName[:], telem.MFrontTireCompoundIndex)
	}
	return GetTireCompoundName(telem.MRearTireCompoundName[:], telem.MRearTireCompoundIndex)
}

