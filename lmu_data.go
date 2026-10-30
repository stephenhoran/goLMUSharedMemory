package goLMUSharedMemory

// Functions for reading and parsing telemetry data from LMU shared memory.

import (
	"fmt"
	"math"
	"unsafe"
)

// LMUData provides a high-level interface for reading LMU telemetry data
type LMUData struct {
	memory *SharedMemory
}

// OpenLMUData opens a connection to LMU shared memory using the default name
// Returns an error if LMU is not running or shared memory is not available
func OpenLMUData() (*LMUData, error) {
	return OpenLMUDataWithName(SharedMemoryName)
}

// OpenLMUDataWithName opens a connection to LMU shared memory with a custom name
// This allows testing with different shared memory names if the default doesn't work
// Returns an error if LMU is not running or shared memory is not available
func OpenLMUDataWithName(name string) (*LMUData, error) {
	memory, err := OpenSharedMemoryWithName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open LMU shared memory: %w", err)
	}

	return &LMUData{
		memory: memory,
	}, nil
}

// Close closes the connection to LMU shared memory
func (ld *LMUData) Close() error {
	if ld.memory != nil {
		return ld.memory.Close()
	}
	return nil
}

func (ld *LMUData) CheckInit() error {
	if ld.memory == nil {
		return fmt.Errorf("LMU data is not initialized")
	}
	return nil
}

// ReadTelemetry reads the current telemetry data from shared memory
// Returns the complete SharedMemoryLayout structure
func (ld *LMUData) ReadTelemetry() (*SharedMemoryLayout, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	data, err := ld.memory.ReadData()
	if err != nil {
		return nil, fmt.Errorf("failed to read telemetry data: %w", err)
	}

	return data, nil
}

// GetScoringInfo returns the current scoring information
func (ld *LMUData) GetScoringInfo() (*ScoringInfoV01, error) {
	if ld.memory == nil {
		return nil, fmt.Errorf("LMU data is not initialized")
	}

	rawData := ld.memory.GetRawData()
	if len(rawData) < 548 {
		return nil, fmt.Errorf("shared memory data is too small")
	}

	// Calculate offset manually based on C++ pack(4) layout
	// SharedMemoryLayout.MData (SharedMemoryObjectOut)
	//   .MGeneric (SharedMemoryGeneric)
	//     events[17] uint32 = 17 * 4 = 68 bytes
	//     gameVersion int32 = 4 bytes (total 72)
	//     FFBTorque float32 = 4 bytes (total 76)
	//     appInfo ApplicationStateV01 = HWND(8) + uint32*4(16) + uint8(1) + [31]byte(31) + [204]byte(204) = 260 bytes
	//     Total: 76 + 260 = 336 bytes (but need to check actual padding)
	//   .MPaths (SharedMemoryPathData)
	//     [260]*5 = 1300 bytes
	//   .MScoring (SharedMemoryScoringData)
	//     .MScoringInfo (ScoringInfoV01) <- offset 0 within this struct

	// Calculate the actual byte offset to ScoringInfoV01
	// We need to manually calculate based on C++ pack(4) sizes, not Go's alignment
	// SharedMemoryGeneric: events[17] (68) + gameVersion (4) + FFBTorque (4) + ApplicationStateV01 (~336)
	// But ApplicationStateV01 with pack(4): HWND(8) + 4*uint32(16) + uint8(1) + [31](31) + [204](204) = 264 bytes
	// Total Generic: 68 + 4 + 4 + 264 = 340 bytes (but might need padding)

	// Actually, let's use a simpler approach: find the track name in the raw data
	// The track name should be visible, so we can search for it
	rawBytes := ld.memory.GetRawData()

	// Search for "de la Sarthe" or similar to find the correct offset
	// For now, let's try calculating based on expected sizes
	// Generic: ~340 bytes, Paths: 1300 bytes (5 * 260), so scoring starts around offset 1640
	// But let's use the struct offset as a starting point and adjust

	// Calculate offset manually based on C++ pack(4) struct sizes
	// SharedMemoryLayout.MData = SharedMemoryObjectOut
	//   .MGeneric = SharedMemoryGeneric
	//     events[17] uint32 = 17 * 4 = 68 bytes
	//     gameVersion int32 = 4 bytes (offset 68-71)
	//     FFBTorque float32 = 4 bytes (offset 72-75)
	//     appInfo ApplicationStateV01:
	//       HWND (uintptr) = 8 bytes on 64-bit (offset 76-83, but pack(4) means it can be at 76)
	//       uint32*4 = 16 bytes (offset 84-99)
	//       uint8 = 1 byte (offset 100)
	//       [31]byte = 31 bytes (offset 101-131)
	//       [204]byte = 204 bytes (offset 132-335)
	//       Total ApplicationStateV01 = 336 bytes
	//     Total SharedMemoryGeneric = 68 + 4 + 4 + 336 = 412 bytes
	//   .MPaths = SharedMemoryPathData
	//     5 * [260]byte = 1300 bytes (offset 412-1711)
	//   .MScoring = SharedMemoryScoringData
	//     .MScoringInfo = ScoringInfoV01 (starts at offset 1712)

	// Use helper function to find offset
	useOffset, err := ld.findScoringInfoOffset()
	if err != nil {
		return nil, err
	}

	if int(useOffset)+560 > len(rawBytes) {
		return nil, fmt.Errorf("offset calculation out of range: offset=%d, len=%d", useOffset, len(rawBytes))
	}

	// Read from raw bytes at calculated/found offset
	return ReadScoringInfoV01(rawBytes[useOffset:]), nil
}

// GetPlayerTelemetry returns the telemetry data for the player vehicle
// Returns a copy of the telemetry data to ensure it remains valid
func (ld *LMUData) GetPlayerTelemetry() (*TelemInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	// Find telemetry section offset
	telemetryOffset, err := ld.findTelemetryOffset()
	if err != nil {
		return nil, err
	}

	// Read player vehicle flags from raw bytes
	// MActiveVehicles at offset 0, MPlayerVehicleIdx at offset 1, MPlayerHasVehicle at offset 2
	if int(telemetryOffset)+3 > len(rawData) {
		return nil, fmt.Errorf("telemetry data too small")
	}

	hasVehicle := rawData[telemetryOffset+2] != 0
	if !hasVehicle {
		return nil, fmt.Errorf("player does not have a vehicle")
	}

	idx := int(rawData[telemetryOffset+1])
	if idx < 0 || idx >= 104 {
		return nil, fmt.Errorf("invalid player vehicle index: %d", idx)
	}

	// Manual offset to MTelemInfo array within SharedMemoryTelemtryData
	// With C++ pack(4): MActiveVehicles(1) + MPlayerVehicleIdx(1) + MPlayerHasVehicle(1) = 3 bytes
	// Array needs 4-byte alignment, so MTelemInfo starts at offset 4
	const telemArrayOffsetWithinStruct uintptr = 4
	telemArrayOffset := telemetryOffset + telemArrayOffsetWithinStruct
	// Use unsafe.Sizeof to get struct size (this is safe - just gives us the size, not alignment-dependent)
	telemSize := unsafe.Sizeof(TelemInfoV01{})
	telemEntryOffset := telemArrayOffset + uintptr(idx)*telemSize

	// Validate bounds
	if int(telemEntryOffset)+int(telemSize) > len(rawData) {
		return nil, fmt.Errorf("telemetry data out of bounds")
	}

	// Copy raw bytes to a new struct
	var telemCopy TelemInfoV01
	srcBytes := rawData[telemEntryOffset : telemEntryOffset+telemSize]
	dstBytes := (*[1 << 20]byte)(unsafe.Pointer(&telemCopy))[:telemSize:telemSize]
	copy(dstBytes, srcBytes)

	return &telemCopy, nil
}

// GetTelemetry returns telemetry data for a specific vehicle index
// Returns a copy of the telemetry data to ensure it remains valid
func (ld *LMUData) GetTelemetry(vehicleIdx int) (*TelemInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	if vehicleIdx < 0 || vehicleIdx >= 104 {
		return nil, fmt.Errorf("vehicle index out of range: %d", vehicleIdx)
	}

	rawData := ld.memory.GetRawData()
	if len(rawData) < 10000 {
		return nil, fmt.Errorf("shared memory data is too small")
	}

	// Find telemetry section offset
	telemetryOffset, err := ld.findTelemetryOffset()
	if err != nil {
		return nil, err
	}

	// Manual offset to MTelemInfo array within SharedMemoryTelemtryData
	// With C++ pack(4): MActiveVehicles(1) + MPlayerVehicleIdx(1) + MPlayerHasVehicle(1) = 3 bytes
	// Array needs 4-byte alignment, so MTelemInfo starts at offset 4
	const telemArrayOffsetWithinStruct uintptr = 4
	telemArrayOffset := telemetryOffset + telemArrayOffsetWithinStruct
	// Use unsafe.Sizeof to get struct size (this is safe - just gives us the size, not alignment-dependent)
	telemSize := unsafe.Sizeof(TelemInfoV01{})
	telemEntryOffset := telemArrayOffset + uintptr(vehicleIdx)*telemSize

	// Validate bounds
	if int(telemEntryOffset)+int(telemSize) > len(rawData) {
		return nil, fmt.Errorf("telemetry data out of bounds")
	}

	// Copy raw bytes to a new struct
	var telemCopy TelemInfoV01
	srcBytes := rawData[telemEntryOffset : telemEntryOffset+telemSize]
	dstBytes := (*[1 << 20]byte)(unsafe.Pointer(&telemCopy))[:telemSize:telemSize]
	copy(dstBytes, srcBytes)

	return &telemCopy, nil
}

// findTelemetryOffset finds the byte offset to SharedMemoryTelemtryData
// Calculates relative to the known working scoring info offset to avoid alignment issues
func (ld *LMUData) findTelemetryOffset() (uintptr, error) {
	rawBytes := ld.memory.GetRawData()
	if len(rawBytes) < 10000 {
		return 0, fmt.Errorf("shared memory data is too small")
	}

	// Find the vehicle array offset - we know this works
	vehicleArrayOffset, err := ld.findVehicleArrayOffset()
	if err != nil {
		return 0, fmt.Errorf("failed to find vehicle array offset: %w", err)
	}

	// Calculate where telemetry starts:
	// After the vehicle array (104 * VehicleScoringInfoV01) comes MScoringStream[65536]
	// VehicleScoringInfoV01 is approximately 584 bytes with pack(4)
	// Note: scoringStreamSize is 12 bytes (ubyte[12]), not 8 bytes
	vehicleArraySize := uintptr(104 * 584) // 60736 bytes
	scoringStreamSize := uintptr(65536)    // 65536 bytes

	// Telemetry starts after vehicle array + scoring stream
	// The vehicle array starts at: scoringInfoOffset + 560 (ScoringInfoV01) + 12 (scoringStreamSize)
	// So we can calculate: vehicleArrayOffset + vehicleArraySize + scoringStreamSize
	approxTelemetryOffset := vehicleArrayOffset + vehicleArraySize + scoringStreamSize

	// Verify by checking MActiveVehicles field and first telemetry entry's MID
	// Manual offset to MTelemInfo array within SharedMemoryTelemtryData (with C++ pack(4))
	const telemArrayOffsetWithinStruct uintptr = 4

	verifyTelemetryOffset := func(offset uintptr) bool {
		if int(offset)+1000 > len(rawBytes) {
			return false
		}
		// Check active vehicles count (should be 0-104)
		activeVehicles := rawBytes[offset]
		if activeVehicles > 104 {
			return false
		}
		// Allow 0 for validation purposes, we'll check telemetry entries

		// Check player vehicle index (should be 0-103 if player has vehicle)
		playerIdx := rawBytes[offset+1]
		if playerIdx > 103 {
			return false
		}

		// Check if telemetry entries have valid MIDs and reasonable data
		telemArrayStart := offset + telemArrayOffsetWithinStruct
		telemSize := unsafe.Sizeof(TelemInfoV01{})

		if activeVehicles == 0 {
			// If no active vehicles, just verify the structure is there
			return true
		}

		validCount := 0
		checkCount := activeVehicles
		if checkCount > 5 {
			checkCount = 5 // Check first 5 entries more strictly
		}

		for i := 0; i < int(checkCount); i++ {
			telemEntryOffset := telemArrayStart + uintptr(i)*telemSize
			if int(telemEntryOffset)+100 > len(rawBytes) {
				break
			}

			// Check MID (offset 0 in TelemInfoV01)
			mid := *(*int32)(unsafe.Pointer(&rawBytes[telemEntryOffset]))
			if mid < -1 || mid > 103 {
				continue // Invalid MID
			}

			// Check DeltaTime (offset 8) - should be reasonable (0-100 seconds typically)
			deltaTime := *(*float64)(unsafe.Pointer(&rawBytes[telemEntryOffset+8]))
			if math.IsNaN(deltaTime) || math.IsInf(deltaTime, 0) || deltaTime < 0 || deltaTime > 100 {
				continue // Invalid delta time
			}

			// Check EngineRPM (offset varies, but should be after several fields)
			// EngineRPM is at offset 108 (after MID, MDeltaTime, MElapsedTime, MLapNumber, MLapStartET, MVehicleName, MTrackName, MPos, MLocalVel, MLocalAccel, MOri, MLocalRot, MLocalRotAccel, MGear)
			// Actually, let's just check that we have a reasonable MID and deltaTime
			validCount++
		}

		// At least one entry should be valid if there are active vehicles
		return validCount > 0 || activeVehicles == 0
	}

	// Try approximate offset first
	if verifyTelemetryOffset(approxTelemetryOffset) {
		return approxTelemetryOffset, nil
	}

	// Search around the approximate offset with wider range
	// Since we're calculating from vehicle array, we should be close, but sizes might be slightly off
	searchStart := int(approxTelemetryOffset) - 5000
	searchEnd := int(approxTelemetryOffset) + 5000
	if searchStart < 0 {
		searchStart = 0
	}
	if searchEnd > len(rawBytes) {
		searchEnd = len(rawBytes)
	}

	// Try with 4-byte alignment first
	bestOffset := uintptr(0)
	bestScore := 0

	for testOffset := searchStart; testOffset < searchEnd && testOffset+1000 < len(rawBytes); testOffset += 4 {
		if verifyTelemetryOffset(uintptr(testOffset)) {
			// Score this offset by checking how many valid telemetry entries we find
			activeVehicles := rawBytes[testOffset]
			if activeVehicles > 104 {
				activeVehicles = 104
			}

			telemArrayStart := uintptr(testOffset) + telemArrayOffsetWithinStruct
			telemSize := unsafe.Sizeof(TelemInfoV01{})

			validCount := 0
			for i := 0; i < int(activeVehicles) && i < 20; i++ {
				telemEntryOffset := telemArrayStart + uintptr(i)*telemSize
				if int(telemEntryOffset)+4 > len(rawBytes) {
					break
				}
				mid := *(*int32)(unsafe.Pointer(&rawBytes[telemEntryOffset]))
				if mid >= 0 && mid <= 103 {
					validCount++
				}
			}

			if validCount > bestScore {
				bestScore = validCount
				bestOffset = uintptr(testOffset)
				// If we found a really good match, return it immediately
				if validCount >= int(activeVehicles)/2 && activeVehicles > 5 {
					return uintptr(testOffset), nil
				}
			}
		}
	}

	if bestScore > 0 {
		return bestOffset, nil
	}

	return 0, fmt.Errorf("could not find telemetry offset in shared memory")
}

// GetAllTelemetry returns telemetry data for all active vehicles
// Returns a slice of telemetry data, one per active vehicle (up to activeVehicles count)
// Returns copies of the telemetry data to ensure they remain valid
func (ld *LMUData) GetAllTelemetry() ([]*TelemInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	// Find telemetry section offset
	telemetryOffset, err := ld.findTelemetryOffset()
	if err != nil {
		return nil, err
	}

	// Read active vehicles count from raw bytes (offset 0 within SharedMemoryTelemtryData)
	activeVehicles := int(rawData[telemetryOffset])
	if activeVehicles > 104 {
		activeVehicles = 104
	}
	if activeVehicles < 0 {
		activeVehicles = 0
	}

	if activeVehicles == 0 {
		return []*TelemInfoV01{}, nil
	}

	// Manual offset to MTelemInfo array within SharedMemoryTelemtryData
	// With C++ pack(4): MActiveVehicles(1) + MPlayerVehicleIdx(1) + MPlayerHasVehicle(1) = 3 bytes
	// Array needs 4-byte alignment, so MTelemInfo starts at offset 4
	const telemArrayOffsetWithinStruct uintptr = 4
	telemArrayOffset := telemetryOffset + telemArrayOffsetWithinStruct

	// Use unsafe.Sizeof to get the size - this should match if structs are correctly aligned
	telemSize := unsafe.Sizeof(TelemInfoV01{})

	result := make([]*TelemInfoV01, 0, activeVehicles)

	for i := 0; i < activeVehicles; i++ {
		// Calculate offset for this telemetry entry
		telemEntryOffset := telemArrayOffset + uintptr(i)*telemSize

		// Validate bounds
		if int(telemEntryOffset)+int(telemSize) > len(rawData) {
			break
		}

		// Copy raw bytes to a new struct
		var telemCopy TelemInfoV01
		srcBytes := rawData[telemEntryOffset : telemEntryOffset+telemSize]
		dstBytes := (*[1 << 20]byte)(unsafe.Pointer(&telemCopy))[:telemSize:telemSize]
		copy(dstBytes, srcBytes)

		// Validate that the copied data looks reasonable
		// MID should be in valid range (0-103, or possibly -1)
		if telemCopy.MID >= -1 && telemCopy.MID <= 103 {
			result = append(result, &telemCopy)
		}
		// Skip entries with invalid MID
	}

	return result, nil
}

// GetNumVehicles returns the number of active vehicles in the session
func (ld *LMUData) GetNumVehicles() (int, error) {
	if err := ld.CheckInit(); err != nil {
		return 0, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return 0, err
	}
	return int(data.MData.MTelemetry.MActiveVehicles), nil
}

// findScoringInfoOffset finds the byte offset to ScoringInfoV01 in the raw data
// Returns the offset or error if not found
func (ld *LMUData) findScoringInfoOffset() (uintptr, error) {
	rawBytes := ld.memory.GetRawData()
	if len(rawBytes) < 548 {
		return 0, fmt.Errorf("shared memory data is too small")
	}

	// Calculate offset manually based on C++ pack(4) layout (same as GetScoringInfo)
	genericSize := uintptr(68 + 4 + 4 + 336) // events + gameVersion + FFBTorque + ApplicationStateV01
	pathsSize := uintptr(5 * 260)            // 5 * MAX_PATH
	// SharedMemoryLayout.MData is the only field, so offset is 0
	const baseOffset uintptr = 0
	totalOffset := baseOffset + genericSize + pathsSize

	// Verify the calculated offset
	verifyOffset := func(offset uintptr) bool {
		if int(offset)+100 > len(rawBytes) {
			return false
		}
		testData := rawBytes[offset:]

		// Check if track name looks valid (has null terminator within 64 bytes)
		trackNameValid := false
		for i := 0; i < 64 && i < len(testData); i++ {
			if testData[i] == 0 {
				trackNameValid = true
				break
			}
			if i == 63 && testData[i] != 0 {
				break
			}
		}
		if !trackNameValid {
			return false
		}

		// Check if session is reasonable (0-20 is typical)
		if len(testData) >= 68 {
			session := *(*int32)(unsafe.Pointer(&testData[64]))
			if session < 0 || session > 20 {
				return false
			}
		}

		// Check if current time is reasonable (positive, not NaN/Inf)
		if len(testData) >= 76 {
			currentTime := *(*float64)(unsafe.Pointer(&testData[68]))
			if currentTime < 0 || currentTime > 86400*10 || math.IsNaN(currentTime) || math.IsInf(currentTime, 0) {
				return false
			}
		}

		return true
	}

	// Try calculated offset first
	useOffset := totalOffset

	// If calculated offset doesn't verify, search nearby
	if !verifyOffset(useOffset) {
		searchStart := int(totalOffset) - 100
		searchEnd := int(totalOffset) + 100
		if searchStart < 0 {
			searchStart = 0
		}
		if searchEnd > len(rawBytes) {
			searchEnd = len(rawBytes)
		}

		found := false
		for testOffset := searchStart; testOffset < searchEnd && testOffset+100 < len(rawBytes); testOffset += 4 {
			if verifyOffset(uintptr(testOffset)) {
				useOffset = uintptr(testOffset)
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("could not find ScoringInfoV01 offset in shared memory")
		}
	}

	return useOffset, nil
}

// findVehicleArrayOffset finds the byte offset to the VehicleScoringInfoV01 array
// Uses the known working scoring info offset and calculates directly
func (ld *LMUData) findVehicleArrayOffset() (uintptr, error) {
	rawBytes := ld.memory.GetRawData()
	if len(rawBytes) < 10000 {
		return 0, fmt.Errorf("shared memory data is too small")
	}

	// Find ScoringInfoV01 offset first (we know this works)
	scoringInfoOffset, err := ld.findScoringInfoOffset()
	if err != nil {
		return 0, fmt.Errorf("failed to find scoring info offset: %w", err)
	}

	// ScoringInfoV01 size is 548 bytes (verified from Python library)
	// In LMUScoringData, vehScoringInfo starts at offset 560 (verified from Python)
	// So: vehicleArrayStart = scoringInfoOffset + 560
	vehicleArrayStart := scoringInfoOffset + 560

	// Try a few byte offsets in case there's unexpected padding
	// Test offsets: -4, -3, -2, -1, exact, +1, +2, +3, +4 (around calculated offset)
	testOffsets := []uintptr{
		vehicleArrayStart - 4,
		vehicleArrayStart - 3,
		vehicleArrayStart - 2,
		vehicleArrayStart - 1,
		vehicleArrayStart,
		vehicleArrayStart + 1,
		vehicleArrayStart + 2,
		vehicleArrayStart + 3,
		vehicleArrayStart + 4,
	}

	// Validate the calculated offset by checking first few vehicles
	if int(vehicleArrayStart)+600 > len(rawBytes) {
		return 0, fmt.Errorf("vehicle array offset out of range")
	}

	// Try each test offset and find the best one
	// We'll validate by checking MID first (must be 0-103), then names
	bestOffset := vehicleArrayStart
	bestScore := 0

	for _, testOffset := range testOffsets {
		// Skip negative offsets (would be out of bounds)
		if testOffset > uintptr(len(rawBytes)) || int(testOffset)+600 > len(rawBytes) || int(testOffset) < 0 {
			continue
		}

		testData := rawBytes[testOffset:]

		// First check MID (offset 0-3) - must be in valid range
		if len(testData) < 4 {
			continue
		}
		mid := *(*int32)(unsafe.Pointer(&testData[0]))
		if mid < -1 || mid > 103 {
			continue // Invalid MID means wrong offset
		}

		// Check driver name (offset 4-35) - look for complete name with null terminator
		driverNameScore := 0
		driverNameStart := -1
		for i := 4; i < 36 && i < len(testData); i++ {
			if testData[i] == 0 && i > 4 {
				// Found null terminator - count printable characters
				driverNameStart = 4
				for j := 4; j < i; j++ {
					if testData[j] >= 32 && testData[j] < 127 {
						driverNameScore++
					} else if testData[j] != 0 {
						// Non-printable, non-null character suggests wrong offset
						driverNameScore = -1000
						break
					}
				}
				break
			}
		}

		// Check vehicle name (offset 36-99) - look for complete name
		vehicleNameScore := 0
		for i := 36; i < 100 && i < len(testData); i++ {
			if testData[i] == 0 && i > 36 {
				// Found null terminator - count printable characters
				for j := 36; j < i; j++ {
					if testData[j] >= 32 && testData[j] < 127 {
						vehicleNameScore++
					} else if testData[j] != 0 {
						vehicleNameScore = -1000
						break
					}
				}
				break
			}
		}

		// Check position (offset 199) - should be 0-104
		place := uint8(255)
		if len(testData) >= 200 {
			place = testData[199]
		}

		// Score this offset: prefer longer, complete names and valid MID
		// Penalize if driver name starts with non-printable (suggests offset is wrong)
		score := driverNameScore + vehicleNameScore
		if mid >= 0 && mid <= 103 {
			score += 20 // Big bonus for valid MID
		}
		if place > 0 && place <= 104 {
			score += 10 // Bonus for valid position
		}
		if driverNameStart == 4 && testData[4] >= 32 && testData[4] < 127 {
			score += 5 // Bonus if driver name starts with printable char (not cut off)
		}

		if score > bestScore {
			bestScore = score
			bestOffset = testOffset
		}
	}

	// If we found a good offset with valid MID and names, return it
	if bestScore > 10 {
		return bestOffset, nil
	}

	// If direct calculation didn't validate, fall back to search
	// Verify a vehicle entry looks valid
	verifyVehicleOffset := func(offset uintptr) bool {
		if int(offset)+600 > len(rawBytes) {
			return false
		}
		testData := rawBytes[offset:]

		// Check multiple things to ensure this is a valid vehicle entry:

		// 1. Check if position (mPlace at offset 199) is reasonable (0-104)
		// Position 0 might be valid (unpositioned), but >104 is definitely wrong
		if len(testData) < 200 {
			return false
		}
		place := testData[199]
		if place > 104 {
			return false // Position must be 0-104
		}

		// 2. Check if driver name looks valid (has null terminator within 32 bytes and contains printable chars)
		hasValidDriverName := false
		for i := 4; i < 36 && i < len(testData); i++ {
			if testData[i] == 0 && i > 4 {
				// Check if there's at least one printable character before null
				printableCount := 0
				for j := 4; j < i; j++ {
					if testData[j] >= 32 && testData[j] < 127 {
						printableCount++
					}
				}
				if printableCount > 0 {
					hasValidDriverName = true
				}
				break
			}
		}

		// 3. Check if laps (mTotalLaps at offset 100-101, signed short) is reasonable
		if len(testData) < 102 {
			return false
		}
		totalLaps := int16(*(*int16)(unsafe.Pointer(&testData[100])))
		if totalLaps < -1 || totalLaps > 10000 {
			return false
		}

		// 4. Check if best lap time is reasonable (0 or positive, not NaN/Inf, reasonable range)
		if len(testData) < 152 {
			return false
		}
		bestLapTime := *(*float64)(unsafe.Pointer(&testData[144]))
		if math.IsNaN(bestLapTime) || math.IsInf(bestLapTime, 0) {
			return false
		}
		// Best lap time should be 0 (no lap yet) or between 10-600 seconds
		// Allow slightly wider range for validation: 0 to 600
		if bestLapTime < 0 || (bestLapTime > 0 && bestLapTime > 600) {
			return false
		}

		// 5. Check if lap distance (mLapDist at offset 104) is reasonable (0 to track length)
		if len(testData) < 112 {
			return false
		}
		lapDist := *(*float64)(unsafe.Pointer(&testData[104]))
		if math.IsNaN(lapDist) || math.IsInf(lapDist, 0) || lapDist < 0 || lapDist > 50000 {
			return false
		}

		// If we get here, we've validated position, laps, and best lap time
		// Also require driver name to be valid for stricter validation
		return hasValidDriverName && bestLapTime >= 0 && bestLapTime <= 600
	}

	// Try calculated start first
	if verifyVehicleOffset(vehicleArrayStart) {
		return vehicleArrayStart, nil
	}

	// Search in a smaller range around calculated offset
	// The vehicle array should be very close to our calculation
	searchStart := int(vehicleArrayStart) - 100
	searchEnd := int(vehicleArrayStart) + 100
	if searchStart < 0 {
		searchStart = 0
	}
	if searchEnd > len(rawBytes) {
		searchEnd = len(rawBytes)
	}

	// Try offsets at 4-byte boundaries (pack(4) alignment)
	// Reuse bestOffset and bestScore from above, but reset for search
	searchBestOffset := uintptr(0)
	searchBestScore := 0

	for testOffset := searchStart; testOffset < searchEnd && testOffset+600 < len(rawBytes); testOffset += 4 {
		testOffsetUint := uintptr(testOffset)
		if verifyVehicleOffset(testOffsetUint) {
			// If we find a valid offset, verify the next few vehicles too to ensure we're at the array start
			validCount := 1
			vehicleSize := 584
			firstVehicleData := rawBytes[testOffset:]
			positions := []uint8{firstVehicleData[199]} // Store positions to check for reasonable sequence

			for i := 1; i < 10 && testOffset+vehicleSize*(i+1)+600 < len(rawBytes); i++ {
				nextOffset := testOffsetUint + uintptr(i*vehicleSize)
				if int(nextOffset)+200 >= len(rawBytes) {
					break
				}
				nextData := rawBytes[nextOffset:]

				// Quick validation for this vehicle
				nextPlace := nextData[199]
				if nextPlace > 104 {
					break // Invalid position, stop
				}

				// Check driver name quickly
				hasDriverName := false
				for j := 4; j < 36 && j < len(nextData); j++ {
					if nextData[j] == 0 && j > 4 {
						hasDriverName = true
						break
					}
				}

				if !hasDriverName {
					break // No driver name, stop
				}

				validCount++
				positions = append(positions, nextPlace)
			}

			// Check that positions are reasonable - they should mostly be in range 1-104,
			// and not all be the same (unless it's early in race)
			validPositions := 0
			for _, pos := range positions {
				if pos > 0 && pos <= 104 {
					validPositions++
				}
			}

			// Require at least 70% of positions to be in valid range
			if validCount > searchBestScore && validPositions*10 >= validCount*7 {
				searchBestScore = validCount
				searchBestOffset = uintptr(testOffset)
				// Require at least 5 consecutive valid vehicles with good positions
				if validCount >= 5 && validPositions*10 >= validCount*7 {
					return uintptr(testOffset), nil
				}
			}
		}
	}

	if searchBestOffset > 0 {
		return searchBestOffset, nil
	}

	// If we can't find it by validation, return the calculated offset
	// Based on Python library structure, this should be correct
	return vehicleArrayStart, nil
}

// GetVehicleScoring returns scoring information for a specific vehicle index
func (ld *LMUData) GetVehicleScoring(vehicleIdx int) (*VehicleScoringInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	if vehicleIdx < 0 || vehicleIdx >= 104 {
		return nil, fmt.Errorf("vehicle index out of range: %d", vehicleIdx)
	}

	rawBytes := ld.memory.GetRawData()
	if len(rawBytes) < 1000 {
		return nil, fmt.Errorf("shared memory data is too small")
	}

	// Find vehicle array offset using the validated offset finder
	vehicleArrayStart, err := ld.findVehicleArrayOffset()
	if err != nil {
		return nil, fmt.Errorf("failed to find vehicle array offset: %w", err)
	}

	// VehicleScoringInfoV01 size is 584 bytes with pack(4) alignment
	// This is calculated from ReadVehicleScoringInfoV01 which ends at offset 584
	vehicleSize := uintptr(584)
	vehicleOffset := vehicleArrayStart + uintptr(vehicleIdx)*vehicleSize

	if int(vehicleOffset)+int(vehicleSize) > len(rawBytes) {
		return nil, fmt.Errorf("vehicle offset out of range: offset=%d, size=%d, len=%d", vehicleOffset, vehicleSize, len(rawBytes))
	}

	// Use helper function to read with correct alignment
	return ReadVehicleScoringInfoV01(rawBytes[vehicleOffset:]), nil
}

// GetAllVehicleScoring returns scoring information for all vehicles
// Returns a slice of scoring data for all vehicles in the session
func (ld *LMUData) GetAllVehicleScoring() ([]*VehicleScoringInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	numVehicles := int(scoring.MNumVehicles)
	if numVehicles > 104 {
		numVehicles = 104
	}
	if numVehicles < 0 {
		numVehicles = 0
	}

	result := make([]*VehicleScoringInfoV01, numVehicles)
	for i := 0; i < numVehicles; i++ {
		vehicle, err := ld.GetVehicleScoring(i)
		if err != nil {
			return nil, fmt.Errorf("error reading vehicle %d: %w", i, err)
		}
		result[i] = vehicle
	}

	return result, nil
}

// IsConnected checks if the shared memory connection is still valid
func (ld *LMUData) IsConnected() bool {
	if ld.memory == nil {
		return false
	}
	// Try to read data to verify connection
	_, err := ld.ReadTelemetry()
	return err == nil
}

// GetTrackName returns the current track name as a string from scoring info
func (ld *LMUData) GetTrackName() (string, error) {
	scoring, err := ld.GetScoringInfo()
	if err != nil {
		return "", err
	}
	return CStringToString(scoring.MTrackName[:]), nil
}

// Helper function to convert C-style null-terminated byte array to Go string
// Deprecated: Use CStringToString from lmu_helpers.go instead
func cStringToString(bytes []byte) string {
	return CStringToString(bytes)
}
