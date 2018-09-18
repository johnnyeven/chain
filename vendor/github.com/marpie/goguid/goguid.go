// Copyright 2014 Markus Piéton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package goguid is a GUID generation library based on noeqd / snowflake.
// All credit belongs to them!
//
// GUID generation and guarantees (GUID size = 64 bit)
//
//  - time            - 42 bit - millisecond precision (~ 100 years)
//  - machine id      - 10 bit - max. 1024 machines
//  - sequence number - 12 bit - rolls over after 4096 id's in one millisecond
//
// Performance
//   On a DELL Latitude (i5 2.4GHz) notebook it generates approx. 4 million GUID per second.
//
// Usage
//   // Initialize the package
//   InitGUID(0, 0)
//
//   // Generate IDs...
//   for i := 0; i < 10; i++ {
//       id := GetGUID()
//       if id != 0 {
//         print(id)
//       }
//   }
//
package guid

import (
	"sync"
	"time"
)

const (
	customEpoch = 1325376000000 // -> 2012-01-01 00:00:00 --> msec

	sequenceBits  = uint64(12)
	machineIdBits = uint64(10)

	machineIdShift = sequenceBits
	timestampShift = machineIdShift + machineIdBits

	sequenceMask = int64(-1) ^ (int64(-1) << sequenceBits)
)

var (
	lastTimestamp  int64
	machineId      int64
	sequenceNumber int64
	mutex          sync.Mutex
)

func customTimeInMilliseconds() int64 {
	return (time.Now().UnixNano() / 1e6)
}

// InitGUID must be called to initialize the GUID engine.
func InitGUID(machineIdentifier int64, lastUsedTimestamp int64) {
	machineId = int64(machineIdentifier << machineIdShift)
	lastTimestamp = int64(lastUsedTimestamp)
}

// GetGUID generates the next GUID.
// The function uses a mutex to ensure unique values.
func GetGUID() int64 {
	timestamp := customTimeInMilliseconds()
	mutex.Lock()
	defer mutex.Unlock()
	if lastTimestamp == timestamp {
		sequenceNumber = (sequenceNumber + 1) & sequenceMask
		if sequenceNumber == 0 {
			// sequence wrapped ... wait for the next millisecond
			for timestamp <= lastTimestamp {
				timestamp = customTimeInMilliseconds()
			}
		}
	} else {
		sequenceNumber = 0
	}

	if timestamp < lastTimestamp {
		return 0
	}
	lastTimestamp = timestamp

	id := (timestamp << timestampShift) | machineId | sequenceNumber

	return id
}

// GetLastTimestamp returns the last timestamp used by GetGUID.
func GetLastTimestamp() int64 {
	mutex.Lock()
	defer mutex.Unlock()
	return lastTimestamp
}
