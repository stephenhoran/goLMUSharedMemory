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

// Custom types to represent fields that can not use native Go types
// directly due to the shared memory pack(4) alignment
type (
	CDouble   [8]byte
	CUintPtr  [8]byte
	CLongLong [8]byte
)

// Generic CString type for different fixed-size character arrays used in LMU shared memory
type CString[T ~[16]byte | ~[18]byte | ~[24]byte | ~[30]byte | ~[31]byte | ~[32]byte | ~[64]byte | ~[260]byte] struct {
	Data T
}

// TelemVect3 represents a 3D vector (double precision)
type TelemVect3 struct {
	X CDouble
	Y CDouble
	Z CDouble
}

// IPVehicleClass represents the class of a vehicle in LMU
type IPVehicleClass uint8

const (
	Hypercar     IPVehicleClass = 0x00
	LMP2_ELMS    IPVehicleClass = 0x02
	LMP2         IPVehicleClass = 0x03
	LMP3         IPVehicleClass = 0x04
	GTE          IPVehicleClass = 0x05
	GT3          IPVehicleClass = 0x06
	PaceCar      IPVehicleClass = 0x08
	UnknownClass IPVehicleClass = 0xFF
)

// IPVehicleChampionship represents the championship of a vehicle in LMU
type IPVehicleChampionship uint8

const (
	WEC_2023            IPVehicleChampionship = 0x00
	WEC_2024            IPVehicleChampionship = 0x01
	WEC_2025            IPVehicleChampionship = 0x02
	WEC_2026            IPVehicleChampionship = 0x03
	ELMS_2025           IPVehicleChampionship = 0x10
	ELMS_2026           IPVehicleChampionship = 0x11
	UnknownChampionship IPVehicleChampionship = 0xFF
)

// TyreCompound represents the type of tire compound used on a wheel
type TyreCompound uint8

const (
	Soft TyreCompound = 0x00
	Medium TyreCompound = 0x01
	Hard TyreCompound = 0x02
	Wet TyreCompound = 0x03
)

const (
	LMU_MAX_VEHICLES = 104
)

// TelemWheelV01 contains telemetry data for a single wheel
// This matches the C++ struct TelemWheelV01 from InternalsPlugin.hpp
// Size: 260 bytes
type TelemWheelV01 struct {
	MSuspensionDeflection   CDouble
	MRideHeight             CDouble
	MSuspForce              CDouble
	MBrakeTemp              CDouble
	MBrakePressure          CDouble
	MRotation               CDouble
	MLateralPatchVel        CDouble
	MLongitudinalPatchVel   CDouble
	MLateralGroundVel       CDouble
	MLongitudinalGroundVel  CDouble
	MCamber                 CDouble
	MLateralForce           CDouble
	MLongitudinalForce      CDouble
	MTireLoad               CDouble
	MGripFract              CDouble
	MPressure               CDouble
	MTemperature            [3]CDouble // left/center/right in Kelvin
	MWear                   CDouble
	MTerrainName            CString[[16]byte]
	MSurfaceType            uint8  // 0=dry, 1=wet, 2=grass, 3=dirt, 4=gravel, 5=rumblestrip, 6=special
	MFlat                   bool   // whether tire is flat
	MDetached               bool   // whether wheel is detached
	MStaticUndeflectedRadius uint8 // tire radius in centimeters
	MVerticalTireDeflection  CDouble
	MWheelYLocation          CDouble
	MToe                    CDouble
	MTireCarcassTemperature  CDouble
	MTireInnerLayerTemperature [3]CDouble
	MOptimalTemp            float32
	MCompoundIndex          TyreCompound
	MCompoundType           TyreCompound
	MExpansion              [18]uint8
}

// TelemInfoV01 is the main telemetry structure for a vehicle
// This matches the C++ struct TelemInfoV01 from InternalsPlugin.hpp
// Size: 1888 bytes
type TelemInfoV01 struct {
	// Time
	MID           int32
	MDeltaTime    CDouble
	MElapsedTime  CDouble
	MLapNumber    int32
	MLapStartET   CDouble
	MVehicleName  CString[[64]byte]
	MTrackName    CString[[64]byte]

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
	MEngineRPM     CDouble
	MEngineWaterTemp CDouble
	MEngineOilTemp CDouble
	MClutchRPM     CDouble

	// Driver input
	MUnfilteredThrottle CDouble
	MUnfilteredBrake    CDouble
	MUnfilteredSteering CDouble
	MUnfilteredClutch   CDouble

	// Filtered input
	MFilteredThrottle CDouble
	MFilteredBrake    CDouble
	MFilteredSteering CDouble
	MFilteredClutch   CDouble

	// Misc
	MSteeringShaftTorque CDouble
	MFront3rdDeflection  CDouble
	MRear3rdDeflection   CDouble

	// Aerodynamics
	MFrontWingHeight CDouble
	MFrontRideHeight CDouble
	MRearRideHeight  CDouble
	MDrag            CDouble
	MFrontDownforce  CDouble
	MRearDownforce   CDouble

	// State/damage info
	MFuel            CDouble
	MEngineMaxRPM    CDouble
	MScheduledStops  uint8
	MOverheating     bool
	MDetached        bool
	MHeadlights      bool
	MDentSeverity    [8]uint8
	MLastImpactET    CDouble
	MLastImpactMagnitude CDouble
	MLastImpactPos   TelemVect3

	// Expanded
	MEngineTorque    CDouble
	MCurrentSector   int32
	MSpeedLimiter    uint8
	MMaxGears        uint8
	MFrontTireCompoundIndex uint8
	MRearTireCompoundIndex  uint8
	MFuelCapacity    CDouble
	MFrontFlapActivated uint8
	MRearFlapActivated  uint8
	MRearFlapLegalStatus uint8
	MIgnitionStarter   uint8
	MFrontTireCompoundName CString[[18]byte]
	MRearTireCompoundName  CString[[18]byte]
	MSpeedLimiterAvailable uint8
	MAntiStallActivated    uint8
	MUnused                [2]uint8
	MVisualSteeringWheelRange float32
	MRearBrakeBias          CDouble
	MTurboBoostPressure     CDouble
	MPhysicsToGraphicsOffset [3]float32
	MPhysicalSteeringWheelRange float32
	MDeltaBest               CDouble
	MBatteryChargeFraction   CDouble
	MElectricBoostMotorTorque CDouble
	MElectricBoostMotorRPM     CDouble
	MElectricBoostMotorTemperature CDouble
	MElectricBoostWaterTemperature CDouble
	MElectricBoostMotorState  uint8
	MLapInvalidated           bool
	MABSActive                bool
	MTCActive                 bool
	MSpeedLimiterActive       bool
	MWiperState               uint8
	MTC                       uint8
	MTCMax                    uint8
	MTCSlip                   uint8
	MTCSlipMax                uint8
	MTCCut                    uint8
	MTCCutMax				  uint8
	MABS                      uint8
	MABSMax                   uint8
	MMotorMap                 uint8
	MMotorMapMax              uint8
	MMigration                uint8
	MMigrationMax             uint8
	MFrontAntiSway            uint8
	MFrontAntiSwayMax         uint8
	MRearAntiSway             uint8
	MRearAntiSwayMax          uint8
	MLiftAndCoastProgress     uint8
	MTrackLimitsSteps         uint8
	MRegen                    float32
	MStateOfCharge            float32
	MVirtualEnergy            float32
	MTimeGapCarAhead          float32
	MTimeGapCarBehind         float32
	MTimeGapPlaceAhead        float32
	MTimeGapPlaceBehind       float32
	MVehicleModel             CString[[30]byte]
	MVehicleClass             IPVehicleClass
	MVehicleChampionship      IPVehicleChampionship
	MExpansion               [20]uint8

	// Wheels (front left, front right, rear left, rear right)
	MWheel [4]TelemWheelV01
}

// VehicleScoringInfoV01 contains scoring information for a vehicle
// This matches the C++ struct VehicleScoringInfoV01 from InternalsPlugin.hpp
// Size: 584 bytes
type VehicleScoringInfoV01 struct {
	MID               int32
	MDriverName       CString[[32]byte]
	MVehicleName      CString[[64]byte]
	MTotalLaps        int16
	MSector           int8
	MFinishStatus     int8
	MLapDist          CDouble
	MPathLateral      CDouble
	MTrackEdge        CDouble
	MBestSector1      CDouble
	MBestSector2      CDouble
	MBestLapTime      CDouble
	MLastSector1      CDouble
	MLastSector2      CDouble
	MLastLapTime      CDouble
	MCurSector1       CDouble
	MCurSector2       CDouble
	MNumPitstops      int16
	MNumPenalties     int16
	MIsPlayer         bool
	MControl          int8
	MInPits           bool
	MPlace            uint8
	MVehicleClass     CString[[32]byte]
	MTimeBehindNext   CDouble
	MLapsBehindNext   int32
	MTimeBehindLeader CDouble
	MLapsBehindLeader int32
	MLapStartET       CDouble
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
	MTimeIntoLap      CDouble
	MEstimatedLapTime CDouble
	MPitGroup         CString[[24]byte]
	MFlag             uint8
	MUnderYellow      bool
	MCountLapFlag     uint8
	MInGarageStall    bool
	MUpgradePack      [16]uint8
	MPitLapDist       float32
	MBestLapSector1   float32
	MBestLapSector2   float32
	MSteamID          CLongLong
	MVehFilename      CString[[32]byte]
	MAttackMode       int16
	MFuelFraction     uint8
	MDRSState         bool
	MExpansion        [4]uint8
}

// ScoringInfoV01 contains session scoring information
// This matches the C++ struct ScoringInfoV01 from InternalsPlugin.hpp
// Size: 548 bytes
type ScoringInfoV01 struct {
	MTrackName       CString[[64]byte]
	MSession         int32
	MCurrentET       CDouble
	MEndET           CDouble
	MMaxLaps         int32
	MLapDist         CDouble
	MResultsStream   CUintPtr
	MNumVehicles     int32
	MGamePhase       uint8
	MYellowFlagState int8
	MSectorFlag      [3]int8
	MStartLight      uint8
	MNumRedLights    uint8
	MInRealtime      bool
	MPlayerName      CString[[32]byte]
	MPlrFileName     CString[[64]byte]
	MDarkCloud       CDouble
	MRaining         CDouble
	MAmbientTemp     CDouble
	MTrackTemp       CDouble
	MWind            TelemVect3
	MMinPathWetness  CDouble
	MMaxPathWetness  CDouble
	MGameMode        uint8
	MIsPasswordProtected bool
	MServerPort      uint16
	MServerPublicIP  uint32
	MMaxPlayers      int32
	MServerName      CString[[32]byte]
	MStartET         float32
	MAvgPathWetness  CDouble
	MSessionTimeRemaining       float32
	MTimeOfDay                  float32
	MIsFixedSetup               bool
	MTrackGripLevel             uint8
	MCloudCoverage              uint8
	MTrackLimitsStepsPerPenalty uint8
	MTrackLimitsStepsPerPoint   uint8
	MExpansion       [187]uint8
	MVehicle         CUintPtr
}

// ApplicationStateV01 contains application state information
// This matches the C++ struct ApplicationStateV01 from InternalsPlugin.hpp
// Size: 260 bytes
type ApplicationStateV01 struct {
	MAppWindow       CUintPtr // HWND
	MWidth           uint32
	MHeight          uint32
	MRefreshRate     uint32
	MWindowed        uint32
	MOptionsLocation uint8
	MOptionsPage     CString[[31]byte]
	MExpansion       [204]uint8
}

// SharedMemoryGeneric contains generic shared memory data
// This matches the C++ struct SharedMemoryGeneric from SharedMemoryInterface.hpp
// Size: (16 * 4) + 4 + 4 + 260 = 332 bytes
type SharedMemoryGeneric struct {
	MEvents    [SME_MAX]uint32 // SharedMemoryEvent array stored as uint32
	MGameVersion int32
	MFFBTorque   float32
	MAppInfo     ApplicationStateV01
}

// SharedMemoryPathData contains path information
// This matches the C++ struct SharedMemoryPathData from SharedMemoryInterface.hpp
// Size: 260 * 5 = 1300 bytes
type SharedMemoryPathData struct {
	MUserData        CString[[260]byte] // MAX_PATH
	MCustomVariables CString[[260]byte] // MAX_PATH
	MStewardResults  CString[[260]byte] // MAX_PATH
	MPlayerProfile   CString[[260]byte] // MAX_PATH
	MPluginsFolder   CString[[260]byte] // MAX_PATH
}

// SharedMemoryScoringData contains scoring data
// This matches the C++ struct SharedMemoryScoringData from SharedMemoryInterface.hpp
// Size: 548 + 12 + (LMU_MAX_VEHICLES * 584) + 65536 = 126832 bytes
type SharedMemoryScoringData struct {
	MScoringInfo       ScoringInfoV01
	MScoringStreamSize [12]byte
	MVehScoringInfo    [LMU_MAX_VEHICLES]VehicleScoringInfoV01 // MUST NOT BE MOVED!
	MScoringStream     [65536]byte
}

// SharedMemoryTelemetryData contains telemetry data for all vehicles
// This matches the C++ struct SharedMemoryTelemetryData from SharedMemoryInterface.hpp
// Size: 1 + 1 + 1 + 1 + (LMU_MAX_VEHICLES * 1888) = 196356 bytes
type SharedMemoryTelemetryData struct {
	MActiveVehicles   uint8
	MPlayerVehicleIdx uint8
	MPlayerHasVehicle bool
	_                 [1]byte // Padding to match C++ struct alignment
	MTelemInfo        [LMU_MAX_VEHICLES]TelemInfoV01
}

// SharedMemoryObjectOut is the main data structure in shared memory
// This matches the C++ struct SharedMemoryObjectOut from SharedMemoryInterface.hpp
// Size: 332 + 1300 + 126832 + 196356 = 324820 bytes
type SharedMemoryObjectOut struct {
	MGeneric   SharedMemoryGeneric
	MPaths     SharedMemoryPathData
	MScoring   SharedMemoryScoringData
	MTelemetry SharedMemoryTelemetryData
}

// SharedMemoryLayout wraps the shared memory data
// This is what's actually in the shared memory segment
type SharedMemoryLayout struct {
	MData SharedMemoryObjectOut
}

// Size returns the size of SharedMemoryLayout in bytes (should be 324820 as of LMU V1.4.2)
func (s *SharedMemoryLayout) Size() int {
	return int(unsafe.Sizeof(*s))
}
