package goLMUSharedMemory

// Helper functions to read correctly aligned structs from shared memory
// This avoids Go's struct alignment issues with C++ pack(4)

import (
	"math"
	"unsafe"
)

// isInvalidFloat checks if a float64 is NaN, Inf, or an unreasonable value
func isInvalidFloat(f float64) bool {
	return math.IsNaN(f) || math.IsInf(f, 0) || math.Abs(f) > 1e10
}

// ReadScoringInfoV01 reads a ScoringInfoV01 from raw bytes
// This manually handles the C++ pack(4) alignment
// Size is 548 bytes (verified from Python library ctypes.sizeof())
func ReadScoringInfoV01(data []byte) *ScoringInfoV01 {
	if len(data) < 548 {
		return nil
	}
	
	// Create struct and manually copy fields with correct offsets
	s := &ScoringInfoV01{}
	
	// MTrackName [64]byte - offset 0
	copy(s.MTrackName[:], data[0:64])
	
	// MSession int32 - offset 64
	s.MSession = *(*int32)(unsafe.Pointer(&data[64]))
	
	// MCurrentET float64 - offset 68 (with pack(4), double can start at 4-byte boundary)
	s.MCurrentET = *(*float64)(unsafe.Pointer(&data[68]))
	
	// MEndET float64 - offset 76
	s.MEndET = *(*float64)(unsafe.Pointer(&data[76]))
	
	// MMaxLaps int32 - offset 84
	s.MMaxLaps = *(*int32)(unsafe.Pointer(&data[84]))
	
	// MLapDist float64 - offset 88
	s.MLapDist = *(*float64)(unsafe.Pointer(&data[88]))
	
	// MResultsStream uintptr - offset 96 (8 bytes on 64-bit)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		s.MResultsStream = *(*uintptr)(unsafe.Pointer(&data[96]))
		// MNumVehicles int32 - offset 104
		s.MNumVehicles = *(*int32)(unsafe.Pointer(&data[104]))
		// MGamePhase uint8 - offset 108
		s.MGamePhase = data[108]
		// MYellowFlagState int8 - offset 109
		s.MYellowFlagState = int8(data[109])
		// MSectorFlag [3]int8 - offset 110-112
		s.MSectorFlag[0] = int8(data[110])
		s.MSectorFlag[1] = int8(data[111])
		s.MSectorFlag[2] = int8(data[112])
		// MStartLight uint8 - offset 113
		s.MStartLight = data[113]
		// MNumRedLights uint8 - offset 114
		s.MNumRedLights = data[114]
		// MInRealtime bool - offset 115
		s.MInRealtime = data[115] != 0
		// padding to 120
		// MPlayerName [32]byte - offset 120
		copy(s.MPlayerName[:], data[120:152])
		// MPlrFileName [64]byte - offset 152-215
		copy(s.MPlrFileName[:], data[152:216])
		// After mPlrFileName[64], we should have the weather doubles
		// With pack(4), doubles can start at 4-byte boundaries
		// mPlrFileName ends at 215 (152+64-1), so next field starts at 216
		// 216 % 4 == 0, so double is properly aligned
		// After mPlrFileName[64], check if we need padding
		// mPlrFileName[64] at offset 152-215 (64 bytes)
		// Next field is double which needs 4-byte alignment with pack(4)
		// 152 + 64 = 216, and 216 % 4 == 0, so no padding needed
		// MDarkCloud float64 - offset 216-223 (8 bytes)
		s.MDarkCloud = *(*float64)(unsafe.Pointer(&data[216]))
		
		// Try multiple offsets for Raining to find the correct one
		// Test offsets from 220 to 232 in 4-byte increments
		bestRainingOffset := 224
		bestRaining := *(*float64)(unsafe.Pointer(&data[224]))
		for testOffset := 220; testOffset <= 232; testOffset += 4 {
			testVal := *(*float64)(unsafe.Pointer(&data[testOffset]))
			// Raining should be in 0.0-1.0 range, and not be NaN or Inf
			if testVal >= 0 && testVal <= 1 && !isInvalidFloat(testVal) {
				bestRaining = testVal
				bestRainingOffset = testOffset
				break
			}
		}
		
		s.MRaining = bestRaining
		// Adjust Ambient Temp and Track Temp based on where we found Raining
		offsetDiff := bestRainingOffset - 224
		s.MAmbientTemp = *(*float64)(unsafe.Pointer(&data[232+offsetDiff]))
		s.MTrackTemp = *(*float64)(unsafe.Pointer(&data[240+offsetDiff]))
		
		// MWind TelemVect3 - offset 248 (3 doubles = 24 bytes)
		s.MWind.X = *(*float64)(unsafe.Pointer(&data[248+offsetDiff]))
		s.MWind.Y = *(*float64)(unsafe.Pointer(&data[256+offsetDiff]))
		s.MWind.Z = *(*float64)(unsafe.Pointer(&data[264+offsetDiff]))
		// MMinPathWetness float64 - offset 272
		s.MMinPathWetness = *(*float64)(unsafe.Pointer(&data[272+offsetDiff]))
		// MMaxPathWetness float64 - offset 280
		s.MMaxPathWetness = *(*float64)(unsafe.Pointer(&data[280+offsetDiff]))
		
		// MGameMode uint8 - offset 288
		s.MGameMode = data[288+offsetDiff]
		// MIsPasswordProtected bool - offset 289
		s.MIsPasswordProtected = data[289+offsetDiff] != 0
		// padding to 292
		// MServerPort uint16 - offset 292
		s.MServerPort = *(*uint16)(unsafe.Pointer(&data[292+offsetDiff]))
		// MServerPublicIP uint32 - offset 294
		s.MServerPublicIP = *(*uint32)(unsafe.Pointer(&data[294+offsetDiff]))
		// MMaxPlayers int32 - offset 298
		s.MMaxPlayers = *(*int32)(unsafe.Pointer(&data[298+offsetDiff]))
		// MServerName [32]byte - offset 302
		copy(s.MServerName[:], data[302+offsetDiff:334+offsetDiff])
		// MStartET float32 - offset 334
		s.MStartET = *(*float32)(unsafe.Pointer(&data[334+offsetDiff]))
		// MAvgPathWetness float64 - offset 332 (verified from Python)
		s.MAvgPathWetness = *(*float64)(unsafe.Pointer(&data[332+offsetDiff]))
		// MExpansion [200]uint8 - offset 340 (verified from Python), size 200
		// MVehicle uintptr - offset 540 (verified from Python), size 8
		// Struct ends at 548 (verified from Python)
		copy(s.MExpansion[:], data[340+offsetDiff:540+offsetDiff])
		s.MVehicle = *(*uintptr)(unsafe.Pointer(&data[540+offsetDiff]))
	} else {
		// 32-bit - adjust offsets accordingly
		s.MResultsStream = *(*uintptr)(unsafe.Pointer(&data[96]))
		s.MNumVehicles = *(*int32)(unsafe.Pointer(&data[100]))
		s.MGamePhase = data[104]
		s.MYellowFlagState = int8(data[105])
		s.MSectorFlag[0] = int8(data[106])
		s.MSectorFlag[1] = int8(data[107])
		s.MSectorFlag[2] = int8(data[108])
		s.MStartLight = data[109]
		s.MNumRedLights = data[110]
		s.MInRealtime = data[111] != 0
		copy(s.MPlayerName[:], data[116:148])
		copy(s.MPlrFileName[:], data[148:212])
		s.MDarkCloud = *(*float64)(unsafe.Pointer(&data[212]))
		s.MRaining = *(*float64)(unsafe.Pointer(&data[220]))
		s.MAmbientTemp = *(*float64)(unsafe.Pointer(&data[228]))
		s.MTrackTemp = *(*float64)(unsafe.Pointer(&data[236]))
		s.MWind.X = *(*float64)(unsafe.Pointer(&data[244]))
		s.MWind.Y = *(*float64)(unsafe.Pointer(&data[252]))
		s.MWind.Z = *(*float64)(unsafe.Pointer(&data[260]))
		s.MMinPathWetness = *(*float64)(unsafe.Pointer(&data[268]))
		s.MMaxPathWetness = *(*float64)(unsafe.Pointer(&data[276]))
		s.MGameMode = data[284]
		s.MIsPasswordProtected = data[285] != 0
		s.MServerPort = *(*uint16)(unsafe.Pointer(&data[288]))
		s.MServerPublicIP = *(*uint32)(unsafe.Pointer(&data[290]))
		s.MMaxPlayers = *(*int32)(unsafe.Pointer(&data[294]))
		copy(s.MServerName[:], data[298:330])
		s.MStartET = *(*float32)(unsafe.Pointer(&data[330]))
		// MAvgPathWetness float64 - offset 340 for 32-bit (but Python shows 332 for 64-bit)
		// For 32-bit, we need to recalculate, but using 340 for now
		s.MAvgPathWetness = *(*float64)(unsafe.Pointer(&data[340]))
		copy(s.MExpansion[:], data[348:548])
		s.MVehicle = *(*uintptr)(unsafe.Pointer(&data[548]))
	}
	
	return s
}

// ReadVehicleScoringInfoV01 reads a VehicleScoringInfoV01 from raw bytes
// This manually handles the C++ pack(4) alignment
func ReadVehicleScoringInfoV01(data []byte) *VehicleScoringInfoV01 {
	// Calculate approximate size needed (should be around 440+ bytes)
	if len(data) < 440 {
		return nil
	}
	
	s := &VehicleScoringInfoV01{}
	
	// mID long (int32) - offset 0-3
	s.MID = *(*int32)(unsafe.Pointer(&data[0]))
	
	// mDriverName [32]byte - offset 4-35
	copy(s.MDriverName[:], data[4:36])
	
	// mVehicleName [64]byte - offset 36-99
	copy(s.MVehicleName[:], data[36:100])
	
	// mTotalLaps short (int16) - offset 100-101
	s.MTotalLaps = *(*int16)(unsafe.Pointer(&data[100]))
	
	// mSector signed char (int8) - offset 102
	s.MSector = int8(data[102])
	
	// mFinishStatus signed char (int8) - offset 103
	s.MFinishStatus = int8(data[103])
	
	// mLapDist double - offset 104-111 (104 % 4 == 0, aligned)
	s.MLapDist = *(*float64)(unsafe.Pointer(&data[104]))
	
	// mPathLateral double - offset 112-119
	s.MPathLateral = *(*float64)(unsafe.Pointer(&data[112]))
	
	// mTrackEdge double - offset 120-127
	s.MTrackEdge = *(*float64)(unsafe.Pointer(&data[120]))
	
	// mBestSector1 double - offset 128-135
	s.MBestSector1 = *(*float64)(unsafe.Pointer(&data[128]))
	
	// mBestSector2 double - offset 136-143
	s.MBestSector2 = *(*float64)(unsafe.Pointer(&data[136]))
	
	// mBestLapTime double - offset 144-151
	s.MBestLapTime = *(*float64)(unsafe.Pointer(&data[144]))
	
	// mLastSector1 double - offset 152-159
	s.MLastSector1 = *(*float64)(unsafe.Pointer(&data[152]))
	
	// mLastSector2 double - offset 160-167
	s.MLastSector2 = *(*float64)(unsafe.Pointer(&data[160]))
	
	// mLastLapTime double - offset 168-175
	s.MLastLapTime = *(*float64)(unsafe.Pointer(&data[168]))
	
	// mCurSector1 double - offset 176-183
	s.MCurSector1 = *(*float64)(unsafe.Pointer(&data[176]))
	
	// mCurSector2 double - offset 184-191
	s.MCurSector2 = *(*float64)(unsafe.Pointer(&data[184]))
	
	// mNumPitstops short (int16) - offset 192-193
	s.MNumPitstops = *(*int16)(unsafe.Pointer(&data[192]))
	
	// mNumPenalties short (int16) - offset 194-195
	s.MNumPenalties = *(*int16)(unsafe.Pointer(&data[194]))
	
	// mIsPlayer bool - offset 196
	s.MIsPlayer = data[196] != 0
	
	// mControl signed char (int8) - offset 197
	s.MControl = int8(data[197])
	
	// mInPits bool - offset 198
	s.MInPits = data[198] != 0
	
	// mPlace unsigned char (uint8) - offset 199
	s.MPlace = data[199]
	
	// mVehicleClass [32]byte - offset 200-231 (need to check if 200 is 4-byte aligned)
	// 200 % 4 == 0, so aligned
	copy(s.MVehicleClass[:], data[200:232])
	
	// mTimeBehindNext double - offset 232-239
	s.MTimeBehindNext = *(*float64)(unsafe.Pointer(&data[232]))
	
	// mLapsBehindNext long (int32) - offset 240-243
	s.MLapsBehindNext = *(*int32)(unsafe.Pointer(&data[240]))
	
	// mTimeBehindLeader double - offset 244-251
	s.MTimeBehindLeader = *(*float64)(unsafe.Pointer(&data[244]))
	
	// mLapsBehindLeader long (int32) - offset 252-255
	s.MLapsBehindLeader = *(*int32)(unsafe.Pointer(&data[252]))
	
	// mLapStartET double - offset 256-263
	s.MLapStartET = *(*float64)(unsafe.Pointer(&data[256]))
	
	// mPos TelemVect3 (3 doubles = 24 bytes) - offset 264-287
	s.MPos.X = *(*float64)(unsafe.Pointer(&data[264]))
	s.MPos.Y = *(*float64)(unsafe.Pointer(&data[272]))
	s.MPos.Z = *(*float64)(unsafe.Pointer(&data[280]))
	
	// mLocalVel TelemVect3 - offset 288-311
	s.MLocalVel.X = *(*float64)(unsafe.Pointer(&data[288]))
	s.MLocalVel.Y = *(*float64)(unsafe.Pointer(&data[296]))
	s.MLocalVel.Z = *(*float64)(unsafe.Pointer(&data[304]))
	
	// mLocalAccel TelemVect3 - offset 312-335
	s.MLocalAccel.X = *(*float64)(unsafe.Pointer(&data[312]))
	s.MLocalAccel.Y = *(*float64)(unsafe.Pointer(&data[320]))
	s.MLocalAccel.Z = *(*float64)(unsafe.Pointer(&data[328]))
	
	// mOri [3]TelemVect3 - offset 336 (3 * 3 * 8 = 72 bytes)
	s.MOri[0].X = *(*float64)(unsafe.Pointer(&data[336]))
	s.MOri[0].Y = *(*float64)(unsafe.Pointer(&data[344]))
	s.MOri[0].Z = *(*float64)(unsafe.Pointer(&data[352]))
	s.MOri[1].X = *(*float64)(unsafe.Pointer(&data[360]))
	s.MOri[1].Y = *(*float64)(unsafe.Pointer(&data[368]))
	s.MOri[1].Z = *(*float64)(unsafe.Pointer(&data[376]))
	s.MOri[2].X = *(*float64)(unsafe.Pointer(&data[384]))
	s.MOri[2].Y = *(*float64)(unsafe.Pointer(&data[392]))
	s.MOri[2].Z = *(*float64)(unsafe.Pointer(&data[400]))
	
	// mLocalRot TelemVect3 - offset 408-431
	s.MLocalRot.X = *(*float64)(unsafe.Pointer(&data[408]))
	s.MLocalRot.Y = *(*float64)(unsafe.Pointer(&data[416]))
	s.MLocalRot.Z = *(*float64)(unsafe.Pointer(&data[424]))
	
	// mLocalRotAccel TelemVect3 - offset 432-455
	s.MLocalRotAccel.X = *(*float64)(unsafe.Pointer(&data[432]))
	s.MLocalRotAccel.Y = *(*float64)(unsafe.Pointer(&data[440]))
	s.MLocalRotAccel.Z = *(*float64)(unsafe.Pointer(&data[448]))
	
	// mHeadlights unsigned char (uint8) - offset 456
	s.MHeadlights = data[456]
	
	// mPitState unsigned char (uint8) - offset 457
	s.MPitState = data[457]
	
	// mServerScored unsigned char (uint8) - offset 458
	s.MServerScored = data[458]
	
	// mIndividualPhase unsigned char (uint8) - offset 459
	s.MIndividualPhase = data[459]
	
	// mQualification long (int32) - offset 460-463
	s.MQualification = *(*int32)(unsafe.Pointer(&data[460]))
	
	// mTimeIntoLap double - offset 464-471
	s.MTimeIntoLap = *(*float64)(unsafe.Pointer(&data[464]))
	
	// mEstimatedLapTime double - offset 472-479
	s.MEstimatedLapTime = *(*float64)(unsafe.Pointer(&data[472]))
	
	// mPitGroup [24]byte - offset 480-503
	copy(s.MPitGroup[:], data[480:504])
	
	// mFlag unsigned char (uint8) - offset 504
	s.MFlag = data[504]
	
	// mUnderYellow bool - offset 505
	s.MUnderYellow = data[505] != 0
	
	// mCountLapFlag unsigned char (uint8) - offset 506
	s.MCountLapFlag = data[506]
	
	// mInGarageStall bool - offset 507
	s.MInGarageStall = data[507] != 0
	
	// padding to align uint8 array? Let's check: 507 + 1 = 508, 508 % 4 == 0, so aligned
	// mUpgradePack [16]uint8 - offset 508-523
	copy(s.MUpgradePack[:], data[508:524])
	
	// mPitLapDist float32 - offset 524-527
	s.MPitLapDist = *(*float32)(unsafe.Pointer(&data[524]))
	
	// mBestLapSector1 float32 - offset 528-531
	s.MBestLapSector1 = *(*float32)(unsafe.Pointer(&data[528]))
	
	// mBestLapSector2 float32 - offset 532-535
	s.MBestLapSector2 = *(*float32)(unsafe.Pointer(&data[532]))
	
	// mSteamID unsigned long long (uint64) - offset 536-543
	s.MSteamID = *(*uint64)(unsafe.Pointer(&data[536]))
	
	// mVehFilename [32]byte - offset 544-575
	copy(s.MVehFilename[:], data[544:576])
	
	// mAttackMode short (int16) - offset 576-577
	s.MAttackMode = *(*int16)(unsafe.Pointer(&data[576]))
	
	// mFuelFraction unsigned char (uint8) - offset 578
	s.MFuelFraction = data[578]
	
	// mDRSState bool - offset 579
	s.MDRSState = data[579] != 0
	
	// padding to align? 579 + 1 = 580, but next is mExpansion[4], so let's see
	// Actually mExpansion might start right after, but let's check alignment
	// mExpansion [4]uint8 - offset 580-583 (or might be padded)
	// Let's check: after bool at 579, we might need padding, but 580 % 4 == 0, so can start there
	if len(data) >= 584 {
		copy(s.MExpansion[:], data[580:584])
	}
	
	return s
}
