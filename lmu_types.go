package goLMUSharedMemory

// Shared memory structure definitions for Le Mans Ultimate telemetry data.
// These structures match the layout of LMU's SharedMemoryInterface.hpp and InternalsPlugin.hpp
// All structures are packed to 4-byte alignment to match C++ #pragma pack(push, 4)

import (
	"unsafe"
)

// Constants for shared memory access
const (
	// SharedMemoryName is the name of the Windows shared memory object used by LMU
	// This matches the LMU_Data shared memory segment created by Le Mans Ultimate
	SharedMemoryName = "LMU_Data"
	// SharedMemoryEventName is the name of the Windows event object used by LMU for data updates
	SharedMemoryEventName = "LMU_Data_Event"
)

// SharedMemoryEvent represents the different event types from LMU
type SharedMemoryEvent uint32

const (
	SME_ENTER SharedMemoryEvent = iota
	SME_EXIT
	SME_STARTUP
	SME_SHUTDOWN
	SME_LOAD
	SME_UNLOAD
	SME_START_SESSION
	SME_END_SESSION
	SME_ENTER_REALTIME
	SME_EXIT_REALTIME
	SME_UPDATE_SCORING
	SME_UPDATE_TELEMETRY
	SME_INIT_APPLICATION
	SME_UNINIT_APPLICATION
	SME_SET_ENVIRONMENT
	SME_FFB
	SME_MAX
)

// TelemVect3 represents a 3D vector (double precision)
type TelemVect3 struct {
	X float64
	Y float64
	Z float64
}

// TelemWheelV01 contains telemetry data for a single wheel
// Size: 8 + 8*20 + 16 + 1 + 1 + 1 + 8 + 8 + 8 + 8*3 + 8 + 24 = approximately 248 bytes
type TelemWheelV01 struct {
	MSuspensionDeflection   float64
	MRideHeight             float64
	MSuspForce              float64
	MBrakeTemp              float64
	MBrakePressure          float64
	MRotation               float64
	MLateralPatchVel        float64
	MLongitudinalPatchVel   float64
	MLateralGroundVel       float64
	MLongitudinalGroundVel  float64
	MCamber                 float64
	MLateralForce           float64
	MLongitudinalForce      float64
	MTireLoad               float64
	MGripFract              float64
	MPressure               float64
	MTemperature            [3]float64 // left/center/right in Kelvin
	MWear                   float64
	MTerrainName            [16]byte
	MSurfaceType            uint8  // 0=dry, 1=wet, 2=grass, 3=dirt, 4=gravel, 5=rumblestrip, 6=special
	MFlat                   bool   // whether tire is flat
	MDetached               bool   // whether wheel is detached
	MStaticUndeflectedRadius uint8 // tire radius in centimeters
	MVerticalTireDeflection  float64
	MWheelYLocation          float64
	MToe                    float64
	MTireCarcassTemperature  float64
	MTireInnerLayerTemperature [3]float64
	MExpansion              [24]uint8
}

// TelemInfoV01 is the main telemetry structure for a vehicle
// This matches the C++ struct TelemInfoV01 from InternalsPlugin.hpp
type TelemInfoV01 struct {
	// Time
	MID           int32
	MDeltaTime    float64
	MElapsedTime  float64
	MLapNumber    int32
	MLapStartET   float64
	MVehicleName  [64]byte
	MTrackName    [64]byte

	// Position and derivatives
	MPos         TelemVect3
	MLocalVel    TelemVect3
	MLocalAccel  TelemVect3

	// Orientation and derivatives (3x3 matrix stored as 3 vectors)
	MOri         [3]TelemVect3
	MLocalRot    TelemVect3
	MLocalRotAccel TelemVect3

	// Vehicle status
	MGear          int32
	MEngineRPM     float64
	MEngineWaterTemp float64
	MEngineOilTemp float64
	MClutchRPM     float64

	// Driver input
	MUnfilteredThrottle float64
	MUnfilteredBrake    float64
	MUnfilteredSteering float64
	MUnfilteredClutch   float64

	// Filtered input
	MFilteredThrottle float64
	MFilteredBrake    float64
	MFilteredSteering float64
	MFilteredClutch   float64

	// Misc
	MSteeringShaftTorque float64
	MFront3rdDeflection  float64
	MRear3rdDeflection   float64

	// Aerodynamics
	MFrontWingHeight float64
	MFrontRideHeight float64
	MRearRideHeight  float64
	MDrag            float64
	MFrontDownforce  float64
	MRearDownforce   float64

	// State/damage info
	MFuel            float64
	MEngineMaxRPM    float64
	MScheduledStops  uint8
	MOverheating     bool
	MDetached        bool
	MHeadlights      bool
	MDentSeverity    [8]uint8
	MLastImpactET    float64
	MLastImpactMagnitude float64
	MLastImpactPos   TelemVect3

	// Expanded
	MEngineTorque    float64
	MCurrentSector   int32
	MSpeedLimiter    uint8
	MMaxGears        uint8
	MFrontTireCompoundIndex uint8
	MRearTireCompoundIndex  uint8
	MFuelCapacity    float64
	MFrontFlapActivated uint8
	MRearFlapActivated  uint8
	MRearFlapLegalStatus uint8
	MIgnitionStarter   uint8
	MFrontTireCompoundName [18]byte
	MRearTireCompoundName  [18]byte
	MSpeedLimiterAvailable uint8
	MAntiStallActivated    uint8
	MUnused                [2]uint8
	MVisualSteeringWheelRange float32
	MRearBrakeBias          float64
	MTurboBoostPressure     float64
	MPhysicsToGraphicsOffset [3]float32
	MPhysicalSteeringWheelRange float32
	MDeltaBest               float64
	MBatteryChargeFraction   float64
	MElectricBoostMotorTorque float64
	MElectricBoostMotorRPM     float64
	MElectricBoostMotorTemperature float64
	MElectricBoostWaterTemperature float64
	MElectricBoostMotorState  uint8
	MExpansion               [103]uint8 // 111-8 = 103

	// Wheels (front left, front right, rear left, rear right)
	MWheel [4]TelemWheelV01
}

// VehicleScoringInfoV01 contains scoring information for a vehicle
type VehicleScoringInfoV01 struct {
	MID               int32
	MDriverName       [32]byte
	MVehicleName      [64]byte
	MTotalLaps        int16
	MSector           int8
	MFinishStatus     int8
	MLapDist          float64
	MPathLateral      float64
	MTrackEdge        float64
	MBestSector1      float64
	MBestSector2      float64
	MBestLapTime      float64
	MLastSector1      float64
	MLastSector2      float64
	MLastLapTime      float64
	MCurSector1       float64
	MCurSector2       float64
	MNumPitstops      int16
	MNumPenalties     int16
	MIsPlayer         bool
	MControl          int8
	MInPits           bool
	MPlace            uint8
	MVehicleClass     [32]byte
	MTimeBehindNext   float64
	MLapsBehindNext   int32
	MTimeBehindLeader float64
	MLapsBehindLeader int32
	MLapStartET       float64
	MPos              TelemVect3
	MLocalVel         TelemVect3
	MLocalAccel       TelemVect3
	MOri              [3]TelemVect3
	MLocalRot         TelemVect3
	MLocalRotAccel    TelemVect3
	MHeadlights       uint8
	MPitState         uint8
	MServerScored     uint8
	MIndividualPhase  uint8
	MQualification    int32
	MTimeIntoLap      float64
	MEstimatedLapTime float64
	MPitGroup         [24]byte
	MFlag             uint8
	MUnderYellow      bool
	MCountLapFlag     uint8
	MInGarageStall    bool
	MUpgradePack      [16]uint8
	MPitLapDist       float32
	MBestLapSector1   float32
	MBestLapSector2   float32
	MSteamID          uint64
	MVehFilename      [32]byte
	MAttackMode       int16
	MFuelFraction     uint8
	MDRSState         bool
	MExpansion        [4]uint8
}

// ScoringInfoV01 contains session scoring information
// NOTE: This struct uses explicit padding to match C++ #pragma pack(4) alignment
// With pack(4), all fields align to 4-byte boundaries (not their natural alignment)
type ScoringInfoV01 struct {
	MTrackName       [64]byte // offset 0-63
	MSession         int32    // offset 64-67 (no padding needed, 64 is 4-byte aligned)
	MCurrentET       float64  // offset 68-75 (with pack(4), double can start at offset 68)
	MEndET           float64  // offset 76-83
	MMaxLaps         int32    // offset 84-87
	MLapDist         float64  // offset 88-95 (with pack(4), double can start at offset 88)
	MResultsStream   uintptr  // offset 96-103 (64-bit) or 96-99 (32-bit)
	MNumVehicles     int32    // offset 104-107 (64-bit) or 100-103 (32-bit)
	MGamePhase       uint8    // offset 108 (64-bit) or 104 (32-bit)
	MYellowFlagState int8     // offset 109 (64-bit) or 105 (32-bit)
	MSectorFlag      [3]int8  // offset 110-112 (64-bit) or 106-108 (32-bit)
	MStartLight      uint8    // offset 113 (64-bit) or 109 (32-bit)
	MNumRedLights    uint8    // offset 114 (64-bit) or 110 (32-bit)
	MInRealtime      bool     // offset 115 (64-bit) or 111 (32-bit)
	_                [1]byte  // padding to 4-byte boundary
	MPlayerName      [32]byte // offset 120 (64-bit) or 116 (32-bit) - padded to 4-byte boundary
	MPlrFileName     [64]byte // offset 152 (64-bit) or 148 (32-bit)
	MDarkCloud       float64  // offset 216 (64-bit) or 212 (32-bit)
	MRaining         float64  // offset 224 (64-bit) or 220 (32-bit)
	MAmbientTemp     float64  // offset 232 (64-bit) or 228 (32-bit)
	MTrackTemp       float64  // offset 240 (64-bit) or 236 (32-bit)
	MWind            TelemVect3 // offset 248 (64-bit) or 244 (32-bit) - 3 doubles = 24 bytes
	MMinPathWetness  float64  // offset 272 (64-bit) or 268 (32-bit)
	MMaxPathWetness  float64  // offset 280 (64-bit) or 276 (32-bit)
	MGameMode        uint8    // offset 288 (64-bit) or 284 (32-bit)
	MIsPasswordProtected bool  // offset 289 (64-bit) or 285 (32-bit)
	_                [2]byte  // padding to align uint16
	MServerPort      uint16   // offset 292 (64-bit) or 288 (32-bit)
	MServerPublicIP  uint32   // offset 294 (64-bit) or 290 (32-bit)
	MMaxPlayers      int32    // offset 298 (64-bit) or 294 (32-bit)
	MServerName      [32]byte // offset 302 (64-bit) or 298 (32-bit)
	MStartET         float32  // offset 334 (64-bit) or 330 (32-bit)
	_                [4]byte  // padding to align double to 4-byte boundary
	MAvgPathWetness  float64  // offset 344 (64-bit) or 340 (32-bit)
	MExpansion       [200]uint8 // offset 352 (64-bit) or 348 (32-bit)
	MVehicle         uintptr  // offset 552 (64-bit) or 548 (32-bit)
}

// ApplicationStateV01 contains application state information
type ApplicationStateV01 struct {
	MAppWindow       uintptr // HWND
	MWidth           uint32
	MHeight          uint32
	MRefreshRate     uint32
	MWindowed        uint32
	MOptionsLocation uint8
	MOptionsPage     [31]byte
	MExpansion       [204]uint8
}

// SharedMemoryGeneric contains generic shared memory data
type SharedMemoryGeneric struct {
	MEvents    [SME_MAX]uint32 // SharedMemoryEvent array stored as uint32
	MGameVersion int32
	MFFBTorque   float32
	MAppInfo     ApplicationStateV01
}

// SharedMemoryPathData contains path information
type SharedMemoryPathData struct {
	MUserData        [260]byte // MAX_PATH
	MCustomVariables [260]byte // MAX_PATH
	MStewardResults  [260]byte // MAX_PATH
	MPlayerProfile   [260]byte // MAX_PATH
	MPluginsFolder   [260]byte // MAX_PATH
}

// SharedMemoryScoringData contains scoring data
// NOTE: scoringStreamSize is actually 12 bytes (ubyte[12]) in C++, not uint64
// This is based on the Python mapping which uses ctypes.c_ubyte*12
type SharedMemoryScoringData struct {
	MScoringInfo       ScoringInfoV01
	MScoringStreamSize [12]byte // Changed from uint64 to [12]byte to match C++ pack(4) layout
	MVehScoringInfo    [104]VehicleScoringInfoV01 // MUST NOT BE MOVED!
	MScoringStream     [65536]byte
}

// SharedMemoryTelemtryData contains telemetry data for all vehicles
type SharedMemoryTelemtryData struct {
	MActiveVehicles   uint8
	MPlayerVehicleIdx uint8
	MPlayerHasVehicle bool
	MTelemInfo        [104]TelemInfoV01
}

// SharedMemoryObjectOut is the main data structure in shared memory
type SharedMemoryObjectOut struct {
	MGeneric   SharedMemoryGeneric
	MPaths     SharedMemoryPathData
	MScoring   SharedMemoryScoringData
	MTelemetry SharedMemoryTelemtryData
}

// SharedMemoryLayout wraps the shared memory data
// This is what's actually in the shared memory segment
type SharedMemoryLayout struct {
	MData SharedMemoryObjectOut
}

// Size returns the size of SharedMemoryLayout in bytes
func (s *SharedMemoryLayout) Size() int {
	return int(unsafe.Sizeof(*s))
}
