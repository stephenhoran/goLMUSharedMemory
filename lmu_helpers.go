package goLMUSharedMemory

import (
	"unsafe"
)

// CString.ToString converts a CString custom type to a Go string
func (c *CString[T]) ToString() string {
    b := unsafe.Slice((*byte)(unsafe.Pointer(&c.Data)), len(c.Data))
    for i, ch := range b {
        if ch == 0 {
            return string(b[:i])
        }
    }
    return string(b)
}

// CDouble.ToFloat64 converts a CDouble custom type (8 bytes) to a Go float64
func (c *CDouble) ToFloat64() float64 {
	if unsafe.Sizeof(c) < 8 {
		return 0
	}
	return *(*float64)(unsafe.Pointer(c))
}

// CLongLong.ToUint64 converts a CLongLong custom type (8 bytes) to a Go uint64
func (c *CLongLong) ToUint64() uint64 {
	if unsafe.Sizeof(c) < 8 {
		return 0
	}
	return *(*uint64)(unsafe.Pointer(c))
}

// IPVehicleClass.ToString converts an IPVehicleClass custom type to a human-readable string
func (c *IPVehicleClass) ToString() string {
	switch *c {
	case Hypercar:
		return "Hypercar"
	case LMP2_ELMS:
		return "LMP2 ELMS"
	case LMP2:
		return "LMP2"
	case LMP3:
		return "LMP3"
	case GTE:
		return "GTE"
	case GT3:
		return "GT3"
	case PaceCar:
		return "Pace Car"
	default:
		return "N/A"
	}
}

func (c *IPVehicleChampionship) ToString() string {
	switch *c {
	case WEC_2023:
		return "WEC 2023"
	case WEC_2024:
		return "WEC 2024"
	case WEC_2025:
		return "WEC 2025"
	case WEC_2026:
		return "WEC 2026"
	case ELMS_2025:
		return "ELMS 2025"
	case ELMS_2026:
		return "ELMS 2026"
	default:
		return "N/A"
	}
}

// TyreCompound.ToString returns a human-readable tire compound name from the compound index of a wheel
func (t *TyreCompound) ToString() string {
	switch *t {
	case Soft:
		return "Soft"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	case Wet:
		return "Wet"
	default:
		return "N/A"
	}
}