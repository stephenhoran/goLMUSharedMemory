package goLMUSharedMemory

// Functions for reading and parsing telemetry data from LMU shared memory.

import (
	"fmt"
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

	dataCopy := *data
	return &dataCopy, nil
}

// GetScoringInfo returns the current scoring information
// Returns the ScoringInfoV01 structure
func (ld *LMUData) GetScoringInfo() (*ScoringInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	return &data.MData.MScoring.MScoringInfo, nil
}

// GetPlayerTelemetry returns the telemetry data for the player vehicle
// Returns a copy of the telemetry data to ensure it remains valid
func (ld *LMUData) GetPlayerTelemetry() (*TelemInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	playerIdx := int(data.MData.MTelemetry.MPlayerVehicleIdx)
	if playerIdx < 0 || playerIdx >= LMU_MAX_VEHICLES {
		return nil, fmt.Errorf("player vehicle index out of range: %d", playerIdx)
	}

	return &data.MData.MTelemetry.MTelemInfo[playerIdx], nil
}

// GetTelemetry returns telemetry data for a specific vehicle index
// Returns a copy of the telemetry data to ensure it remains valid
func (ld *LMUData) GetTelemetry(vehicleIdx int) (*TelemInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	if vehicleIdx < 0 || vehicleIdx >= LMU_MAX_VEHICLES {
		return nil, fmt.Errorf("vehicle index out of range: %d", vehicleIdx)
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	// Return the vehicle scoring info for the specified index
	return &data.MData.MTelemetry.MTelemInfo[vehicleIdx], nil
}

// GetAllTelemetry returns telemetry data for all active vehicles
// Returns a slice of telemetry data, one per active vehicle (up to activeVehicles count)
// Returns copies of the telemetry data to ensure they remain valid
func (ld *LMUData) GetAllTelemetry() ([]*TelemInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	// Read active vehicles count
	numVehicles := int(data.MData.MTelemetry.MActiveVehicles)
	if numVehicles > LMU_MAX_VEHICLES {
		numVehicles = LMU_MAX_VEHICLES
	}
	if numVehicles < 0 {
		numVehicles = 0
	}

	result := make([]*TelemInfoV01, 0, numVehicles)
	for i := 0; i < numVehicles; i++ {
		result = append(result, &data.MData.MTelemetry.MTelemInfo[i])
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

// GetVehicleScoring returns scoring information for a specific vehicle index
func (ld *LMUData) GetVehicleScoring(vehicleIdx int) (*VehicleScoringInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	if vehicleIdx < 0 || vehicleIdx >= LMU_MAX_VEHICLES {
		return nil, fmt.Errorf("vehicle index out of range: %d", vehicleIdx)
	}

	// Return the vehicle scoring info for the specified index
	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	return &data.MData.MScoring.MVehScoringInfo[vehicleIdx], nil
}

// GetAllVehicleScoring returns scoring information for all vehicles
// Returns a slice of scoring data for all vehicles in the session
func (ld *LMUData) GetAllVehicleScoring() ([]*VehicleScoringInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	numVehicles := int(data.MData.MTelemetry.MActiveVehicles)
	if numVehicles > LMU_MAX_VEHICLES {
		numVehicles = LMU_MAX_VEHICLES
	}
	if numVehicles < 0 {
		numVehicles = 0
	}

	result := make([]*VehicleScoringInfoV01, 0, numVehicles)
	for i := 0; i < numVehicles; i++ {
		result = append(result, &data.MData.MScoring.MVehScoringInfo[i])
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
// Returns an empty string if the track name is not available
func (ld *LMUData) GetTrackName() (string, error) {
	scoring, err := ld.GetScoringInfo()
	if err != nil {
		return "", err
	}
	return CStringToString(scoring.MTrackName[:]), nil
}

// GetPlayerName returns the current player name as a string from scoring info
// Returns an empty string if the player name is not available
func (ld *LMUData) GetPlayerName() (string, error) {
	scoring, err := ld.GetScoringInfo()
	if err != nil {
		return "", err
	}
	return CStringToString(scoring.MPlayerName[:]), nil
}

// GetPlayerVehicleScoring returns the scoring information for the player's vehicle
// Returns an error if the player vehicle index is out of range or if the data cannot be read
func (ld *LMUData) GetPlayerVehicleScoring() (*VehicleScoringInfoV01, error) {
	if err := ld.CheckInit(); err != nil {
		return nil, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return nil, err
	}

	playerIdx := int(data.MData.MTelemetry.MPlayerVehicleIdx)
	if playerIdx < 0 || playerIdx >= LMU_MAX_VEHICLES {
		return nil, fmt.Errorf("player vehicle index out of range: %d", playerIdx)
	}

	return &data.MData.MScoring.MVehScoringInfo[playerIdx], nil
}

// GetPlayerVehicleIndex returns the index of the player's vehicle
// Returns an error if the player vehicle index is out of range or if the data cannot be read
func (ld *LMUData) GetPlayerVehicleIndex() (int, error) {
	if err := ld.CheckInit(); err != nil {
		return -1, err
	}

	data, err := ld.ReadTelemetry()
	if err != nil {
		return -1, err
	}

	playerIdx := int(data.MData.MTelemetry.MPlayerVehicleIdx)
	if playerIdx < 0 || playerIdx >= LMU_MAX_VEHICLES {
		return -1, fmt.Errorf("player vehicle index out of range: %d", playerIdx)
	}

	return playerIdx, nil
}