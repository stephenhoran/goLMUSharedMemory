# goLMUSharedMemory

**⚠️ Alpha Release** - This library is in active development and may have breaking changes. Use with caution.
* Currently supports LMU v1.4.2

Go library for accessing Le Mans Ultimate's built-in shared memory interface to read telemetry data.

## About

This package provides a Go interface to read telemetry data from Le Mans Ultimate (LMU) via Windows shared memory. It replicates the functionality of the Python [pyLMUSharedMemory](https://github.com/s-victor/pyLMUSharedMemory) package.

The library handles the complex memory alignment differences between Go's default struct layout and C++'s `#pragma pack(4)` alignment used by LMU, ensuring accurate data reading through the use of custom variable types and size validation.

## Acknowledgement

This library is based on:
- LMU SharedMemoryInterface header file by S397, found in the game's `Support\SharedMemoryInterface` folder
- [pyLMUSharedMemory](https://github.com/s-victor/pyLMUSharedMemory) by s-victor (used as reference for struct sizes and offsets)
- [pyRfactor2SharedMemory](https://github.com/TonyWhitley/pyRfactor2SharedMemory) by Tony Whitley

## Features

- Read session information (track, state, flags, timing, weather)
- Read vehicle telemetry (speed, RPM, gear, inputs, position, tire compounds)
- Read wheel data (temperatures, pressure, wear, brake temperature)
- Access environmental data (track temperature, weather, time of day)
- Helper functions for tire compound detection (Softs, Mediums, Hards, Wets)
- Thread-safe shared memory access
- Tested struct size matching to ensure correct unpacking
- Clean, idiomatic Go API

## Requirements

- Go 1.19 or later
- Windows (shared memory implementation is Windows-specific)
- Le Mans Ultimate with shared memory enabled

## Installation

```bash
go get github.com/stephenhoran/goLMUSharedMemory
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/stephenhoran/goLMUSharedMemory"
)

func main() {
    // Open connection to LMU shared memory
    lmu, err := goLMUSharedMemory.OpenLMUData()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer lmu.Close()

    // Access session information
    scoring, err := lmu.GetScoringInfo()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Track: %s\n", scoring.MTrackName.ToString())
    fmt.Printf("Session Time: %.2f s\n", scoring.MCurrentET.ToFloat64())

    // Access player vehicle telemetry
    playerTelem, err := lmu.GetPlayerTelemetry()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    if playerTelem == nil {
        fmt.Printf("Player telemetry not available\n")
        return
    }
    
    speed := playerTelem.MLocalVel.Z.ToFloat64()
    if speed < 0 {
        speed = -speed
    }
    fmt.Printf("Speed: %.1f km/h\n", speed*3.6)
    fmt.Printf("RPM: %.0f\n", playerTelem.MEngineRPM.ToFloat64())
    fmt.Printf("Gear: %d\n", playerTelem.MGear)
    
    fmt.Printf("Tires: FL=%s FR=%s RL=%s RR=%s\n", 
        playerTelem.MWheel[0].MCompoundIndex.ToString(), playerTelem.MWheel[1].MCompoundIndex.ToString(),
        playerTelem.MWheel[2].MCompoundIndex.ToString(), playerTelem.MWheel[3].MCompoundIndex.ToString())

    fmt.Printf("Vehicle Model: %s\n", playerTelem.MVehicleModel.ToString())
    fmt.Printf("Vehicle Championship: %s\n", playerTelem.MVehicleChampionship.ToString())

    // Access other vehicles' data
    allVehicles, err := lmu.GetAllVehicleScoring()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    if allVehicles != nil {
        fmt.Printf("\nTotal vehicles: %d\n", len(allVehicles))
        // Display first 3 cars in list
        for i := 0; i < 3 && i < len(allVehicles); i++ {
            v := allVehicles[i]
            fmt.Printf("Position %d: %s (%s)\n", 
				v.MPlace, v.MDriverName.ToString(), v.MVehicleName.ToString())
        }
    }
}
```

## Usage Examples

### Displaying Tire Compounds for All Vehicles

```go
package main

import (
    "fmt"
    "github.com/stephenhoran/goLMUSharedMemory"
)

func main() {
    lmu, err := goLMUSharedMemory.OpenLMUData()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer lmu.Close()
    
    // Get all vehicle telemetry
    allTelemetry, err := lmu.GetAllTelemetry()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    if allTelemetry == nil {
        fmt.Println("No telemetry data available")
        return
    }

    // Get all vehicle scoring for positions
    allScoring, err := lmu.GetAllVehicleScoring()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    for _, telem := range allTelemetry {
        // Find corresponding scoring info
        var scoring *goLMUSharedMemory.VehicleScoringInfoV01
        for _, s := range allScoring {
            if s.MID == telem.MID {
                scoring = s
                break
            }
        }
        
        if scoring == nil {
            continue
        }
        
        fmt.Printf("Pos %d: %s - FL: %s, FR: %s, RL: %s, RR: %s\n",
            scoring.MPlace, scoring.MDriverName.ToString(), 
            telem.MWheel[0].MCompoundIndex.ToString(), telem.MWheel[1].MCompoundIndex.ToString(), 
            telem.MWheel[2].MCompoundIndex.ToString(), telem.MWheel[3].MCompoundIndex.ToString())
    }
}
```

### Building a Leaderboard

```go
package main

import (
    "fmt"
    "sort"
    "github.com/stephenhoran/goLMUSharedMemory"
)

func main() {
    lmu, err := goLMUSharedMemory.OpenLMUData()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer lmu.Close()

    allVehicles, err := lmu.GetAllVehicleScoring()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    if allVehicles == nil {
        fmt.Println("No vehicles data available")
        return
    }

    // Sort by position
    sort.Slice(allVehicles, func(i, j int) bool {
        return allVehicles[i].MPlace < allVehicles[j].MPlace
    })

    fmt.Printf("%-4s %-25s %-50s %-12s %-12s %-12s\n",
        "Pos", "Driver", "Vehicle", "Best Lap", "Last Lap", "Gap to Leader")
    
    var leaderBestLap float64
    for i, v := range allVehicles {
        if i == 0 {
            leaderBestLap = v.MBestLapTime.ToFloat64()
        }
        
        gap := v.MBestLapTime.ToFloat64() - leaderBestLap
        
        fmt.Printf("%-4d %-25s %-50s %-12.3f %-12.3f %-12.3f\n",
            v.MPlace, v.MDriverName.ToString(), 
            v.MVehicleName.ToString(), v.MBestLapTime.ToFloat64(), 
            v.MLastLapTime.ToFloat64(), gap)
    }
}
```

### Monitoring Player Telemetry in Real-Time

```go
package main

import (
    "fmt"
    "time"
    "github.com/stephenhoran/goLMUSharedMemory"
)

func main() {
    lmu, err := goLMUSharedMemory.OpenLMUData()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer lmu.Close()

    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for range ticker.C {
        if !lmu.IsConnected() {
            fmt.Println("Lost connection to LMU")
            break
        }

        playerTelem, err := lmu.GetPlayerTelemetry()
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        if playerTelem == nil {
            continue
        }

        playerScoring, err := lmu.GetPlayerVehicleScoring()
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            continue
        }
        if playerScoring == nil {
            continue
        }

        // Clear screen (Windows)
        fmt.Print("\033[2J\033[H")

        speed := playerTelem.MLocalVel.Z.ToFloat64()
        if speed < 0 {
            speed = -speed
        }

        fmt.Printf("=== Player Telemetry ===\n")
        fmt.Printf("Vehicle: %s\n", playerTelem.MVehicleName.ToString())
        fmt.Printf("Driver: %s\n", playerScoring.MDriverName.ToString())
        fmt.Printf("Speed: %.1f km/h\n", speed*3.6)
        fmt.Printf("RPM: %.0f / %.0f\n", playerTelem.MEngineRPM.ToFloat64(), playerTelem.MEngineMaxRPM.ToFloat64())
        fmt.Printf("Gear: %d / %d\n", playerTelem.MGear, playerTelem.MMaxGears)
        fmt.Printf("Fuel: %.1f / %.1f L\n", playerTelem.MFuel.ToFloat64(), playerTelem.MFuelCapacity.ToFloat64())
        fmt.Printf("VE: %.1f %%\n", playerTelem.MVirtualEnergy*100)
        fmt.Printf("Lap: %d\n", playerTelem.MLapNumber)
        fmt.Printf("Best Lap: %.3f\n", playerScoring.MBestLapTime.ToFloat64())
        fmt.Printf("Last Lap: %.3f\n", playerScoring.MLastLapTime.ToFloat64())
        
        // Tire compounds
        fmt.Printf("Tires: FL=%s FR=%s, RL=%s RR=%s\n", 
			playerTelem.MWheel[0].MCompoundIndex.ToString(), playerTelem.MWheel[1].MCompoundIndex.ToString(),
			playerTelem.MWheel[2].MCompoundIndex.ToString(), playerTelem.MWheel[3].MCompoundIndex.ToString())
    }
}
```

## API Reference

### Opening a Connection

```go
func OpenLMUData() (*LMUData, error)
```

Opens a connection to LMU shared memory using the default name (`LMU_Data`). Returns an error if LMU is not running or shared memory is not available.

```go
func OpenLMUDataWithName(name string) (*LMUData, error)
```

Opens a connection to LMU shared memory with a custom name. Useful if LMU uses a different shared memory name than the default.

### Reading Telemetry and Session Data

```go
func (ld *LMUData) ReadTelemetry() (*SharedMemoryLayout, error)
```

Reads the complete shared memory layout structure.

```go
func (ld *LMUData) GetScoringInfo() (*ScoringInfoV01, error)
```

Returns the session scoring information (track, session type, timing, weather, etc.).

### Vehicle Data Access

```go
func (ld *LMUData) IsConnected() bool
```

Checks if the shared memory connection is still valid. Returns `false` if LMU has disconnected or the shared memory segment is no longer available.

```go
func (ld *LMUData) GetTelemetry(vehicleIdx int) (*TelemInfoV01, error)
```

Returns telemetry data for a specific vehicle by index (0-103). Returns `nil` if the vehicle index is invalid or the vehicle is not active.

```go
func (ld *LMUData) GetPlayerTelemetry() (*TelemInfoV01, error)
```

Returns telemetry data for the player's vehicle. Returns an error if player telemetry is not available.

```go
func (ld *LMUData) GetAllTelemetry() ([]*TelemInfoV01, error)
```

Returns telemetry data for all active vehicles in the session. The slice may be empty if no vehicles are active.

```go
func (ld *LMUData) GetNumVehicles() (int, error)
```

Returns the number of active vehicles in the session.

```go
func (ld *LMUData) GetVehicleScoring(vehicleIdx int) (*VehicleScoringInfoV01, error)
```

Returns scoring information (position, lap times, sector times, etc.) for a specific vehicle by index.

```go
func (ld *LMUData) GetAllVehicleScoring() ([]*VehicleScoringInfoV01, error)
```

Returns scoring information for all vehicles in the session. The slice may be empty if no vehicles are active.

```go
func (ld *LMUData) GetPlayerVehicleScoring() (*VehicleScoringInfoV01, error)
```

Returns scoring information for the player's vehicle (position, lap times, sector times, etc.). Returns an error if player vehicle scoring is not available.

```go
func (ld *LMUData) GetPlayerVehicleIndex() (int, error)
```

Returns vehicle index for the player's vehicle. Returns an error if the vehicle index is out of range.

```go
func (ld *LMUData) Close() error
```

Closes the connection to shared memory. Always call this when done to release system resources.

### Helper Functions

```go
func (c *CString[T]) ToString() string
```

Converts a custom type CString variable into a Go string. This is useful for converting driver names, vehicle names, track names, and other string fields from the shared memory structures.

```go
func (c *CDouble) ToFloat64() float64
```

Converts a custom type CDouble variable into a Go float64.

```go
func (c *CLongLong) ToUint64() uint64
```

Converts a custom type CLongLong variable into a Go uint64.

```go
func (c *IPVehicleClass) ToString() string
```

Converts a custom type IPVehicleClass into a human readable Go string.

Returns "N/A" if no valid class is found.

```go
func (c *IPVehicleChampionship) ToString() string
```

Converts a custom type IPVehicleChampionship into a human readable Go string.

Returns "N/A" if no valid championship is found.

```go
func (t *TyreCompound) ToString() string
```

Converts a custom type TyreCompound into a human readable Go string.

Returns "N/A" if no valid compound is found.

## Data Structures

### ScoringInfoV01

Contains session-related information:
- `MSession`: Session state (Invalid, Replay, Live, Paused)
- `MSessionPhase`: Current session phase (Driving, Pits, Formation, etc.)
- `MCurrentET`: Current session elapsed time in seconds
- `MEndET`: Session end time in seconds
- `MMaxLaps`: Maximum laps for the session
- `MLapDist`: Track length in meters
- `MTrackName`: Name of the current track (byte array)
- `MTrackString`: Track name string (byte array)
- `MTrackTemp`: Track temperature in Celsius
- `MAmbientTemp`: Ambient temperature in Celsius
- `MRainDensity`: Rain density (0.0 - 1.0)
- `MWindSpeed`: Wind speed in m/s
- Flags: `MYellowFlagState`, `MRedFlagState`, `MSafetyCarState`
- `MAvgPathWetness`: Average track wetness (0.0 - 1.0)
- `MSessionTimeRemaining`: Seconds until the end of the current session
- `MTimeOfDay`: Time of day in session
- `MIsFixedSetup`: Is the current session forcing fixed setups
- `MTrackGripLevel`: Track grip level index (0-4)
- `MCloudCoverage`: Average cloud coverage across the track (0.0 - 1.0)
- `MTrackLimitsStepsPerPenalty`: Amount of track limits steps allowed before a penalty
- `MTrackLimitsStepsPerPoint`: Amount of track limits steps per track limit point

**Track Grip Level Mapping:**
- `0` = Green
- `1` = Medium Grip
- `2` = High Grip
- `3` = Saturated Grip

### TelemInfoV01

Contains detailed vehicle telemetry data for any vehicle:
- Position and orientation: `MPos`, `MOri`, `MLocalVel`, `MLocalAccel`, `MLocalRot`
- Engine: `MEngineRPM`, `MEngineMaxRPM`, `MEngineWaterTemp`, `MEngineOilTemp`
- Transmission: `MGear`, `MMaxGears`, `MClutchRPM`
- Driver inputs: `MUnfilteredThrottle`, `MUnfilteredBrake`, `MUnfilteredSteering`, `MUnfilteredClutch`
- Filtered inputs: `MFilteredThrottle`, `MFilteredBrake`, `MFilteredSteering`, `MFilteredClutch`
- Fuel and energy: `MFuel`, `MFuelCapacity`, `MVirtualEnergy`, `MStateOfCharge`  
- Wheels: Array of `MWheel[4]` with detailed data (temperatures, pressures, wear, brake temp, compound index, etc.)
- **Tire Compounds**: 
  - `MFrontTireCompoundName`: Front tire compound name (byte array)
  - `MFrontTireCompoundIndex`: Front tire compound index (0-3)
  - `MRearTireCompoundName`: Rear tire compound name (byte array)
  - `MRearTireCompoundIndex`: Rear tire compound index (0-3)
- Timing: `MLapNumber`, `MLapStartET`, `MElapsedTime`, `MDeltaBest`, `MLapInvalidated`, `MTimeGapCarAhead`, `MTimeGapCarBehind`, `MTimeGapPlaceAhead`, `MTimeGapPlaceBehind` 
- Vehicle state: `MHeadlights`, `MOverheating`, `MDetached`, `MWiperState`, `MSpeedLimiterActive`, etc.
- In-car adjustments: `MABSActive`, `MTCActive`, , `MTC`, `MTCMax`, `MTCSlip`, `MTCSlipMax`, `MTCCut`, `MTCCutMax`, `MABS`, `MABSMax`, `MMotorMap`, `MMotorMapMax`, `MMigration`, `MMigrationMax`, `MFrontAntiSway`, `MFrontAntiSwayMax`, `MRearAntiSway`, `MRearAntiSwayMax`,
- Vehicle identification: `MID` (0-103)
- Vehicle description: `MVehicleMode` (byte array), `MVehicleClass`, `MVehicleChampionship` 

### VehicleScoringInfoV01

Contains race/scoring information for any vehicle:
- Identity: `MID`, `MDriverName`, `MVehicleName`, `MVehicleClass`
- Position: `MPlace`, `MPathLateral`, `MTrackEdge`, `MPos`
- Lap information: `MTotalLaps`, `MBestLapTime`, `MLastLapTime`, `MCurSector1`, `MCurSector2`
- Sector times: `MBestSector1`, `MBestSector2`, `MLastSector1`, `MLastSector2`
- Race gaps: `MTimeBehindNext`, `MLapsBehindNext`, `MTimeBehindLeader`, `MLapsBehindLeader`
- Status: `MInPits`, `MFinishStatus`, `MIsPlayer`, `MPitState`
- Flags: `MFlag`, `MUnderYellow`, `MCountLapFlag`
- Other: `MNumPitstops`, `MNumPenalties`, `MQualification`, `MFuelFraction`, `MDRSState`

### WheelData

Contains data for a single wheel (FL, FR, RL, RR):
- `Rotation`: Wheel rotation speed in rad/s
- `BrakeTemp`: Brake temperature in Celsius
- `Pressure`: Tire pressure in bar
- `Temperature`: Tire temperatures (left, center, right) in Celsius
- `Wear`: Tire wear (0.0 = new, 1.0 = worn)
- `Compound`: Tire compound index and type

## Memory Alignment

This library implements robust struct unpacking to handle the C++'s `#pragma pack(4)` alignment used by LMU, by avoiding the use of any Go native 64-bit type, which would force an 8-byte alignment, and incorrectly assign values to the respective fields. 

By making sure that any value above 32-bits is unpacked to an array of bytes, which does not force an 8-byte struct alignment in Go, and then converted after unpacked, this guarantees that data is correctly unpacked as long as the native LMU shared memory organization does not change.

All struct size calculations were done with the help and cross validation of the Python `pyLMUSharedMemory` library.

This ensures accurate data reading and prevents crashes from invalid memory access.

## Platform Notes

This package is currently Windows-specific as it uses Windows shared memory APIs (`CreateFileMapping`, `MapViewOfFile`). The implementation uses Go's `syscall` package to interface with Windows system calls.

## Thread Safety

The shared memory reading operations are safe to call from multiple goroutines. The shared memory data itself changes over time, however this is only a concern when reading raw data directly, as the ReadTelemetry() method already returns an atomic copy of the data when called.

## Troubleshooting

### "The system cannot find the file specified" Error

If you get an error like:
```
failed to open shared memory mapping Local\LMUSharedMem (error 2: The system cannot find the file specified)
```

This usually means one of the following:

1. **LMU is not running**: The shared memory segment is only created when LMU is running. Make sure LMU is started and you're in a session (not just at the main menu).

2. **Shared memory is not enabled**: Check LMU's settings to ensure shared memory output is enabled. This option is typically found in the game's settings or configuration files.

3. **Incorrect shared memory name**: If LMU uses a different shared memory name than the default (`Local\LMUSharedMem`), you can try opening with a custom name:

```go
// Try alternative shared memory names
alternativeNames := []string{
    "Local\\LMUSharedMem",
    "Local\\$rf2data$",
    "Local\\rf2data",
    // Add other possible names here
}

for _, name := range alternativeNames {
    lmu, err := goLMUSharedMemory.OpenLMUDataWithName(name)
    if err == nil {
        fmt.Printf("Successfully connected using name: %s\n", name)
        defer lmu.Close()
        // Use lmu...
        break
    }
}
```

4. **Permissions issue**: Try running your application as administrator to rule out permission problems.

5. **Check LMU documentation**: Refer to LMU's `Support\SharedMemoryInterface` folder for the exact shared memory name and size used by the game.

### Finding the Correct Shared Memory Name

To determine the actual shared memory name used by LMU:

1. Check the `SharedMemoryInterface.h` header file in LMU's installation directory (usually in `Support\SharedMemoryInterface`).
2. Look for the `#define` statement that specifies the shared memory name.
3. Update the `SharedMemoryName` constant in `lmu_types.go` or use `OpenLMUDataWithName()` with the correct name.

### Data Corruption or Incorrect Values

If you're seeing corrupted data or incorrect values:

1. **Check LMU version**: The shared memory structure may change between game versions. Ensure you're using a compatible version of this library.

2. **Verify connection**: Call `IsConnected()` to ensure the connection is still valid.

3. **Check for nil returns**: Always check for `nil` return values from getter functions before accessing fields.

4. **File an issue**: If you believe you've found a bug, please file an issue with:
   - LMU version
   - Library version
   - Sample code that reproduces the issue
   - Expected vs. actual values

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

**Note:** This library is in Alpha state. The API may change in future releases.

## Disclaimer

This package is not affiliated with or endorsed by Studio 397 or Le Mans Ultimate. Use at your own risk.
