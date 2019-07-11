/*
 * Copyright (c) 2019 OysterPack, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package ulidgen provides functions for ULID generators
package ulids

import (
	"crypto/rand"
	"errors"
	"github.com/oklog/ulid"
	"sync"
)

// MonotonicULIDGenerator returns a function that generates ULID(s) in strictly increasing order.
//   - is safe for concurrent use.
//   - panics if a ULID fails to be generated
func MonotonicULIDGenerator() func() ulid.ULID {
	var m sync.Mutex
	entropy := ulid.Monotonic(rand.Reader, 0)

	return func() (uid ulid.ULID) {
		m.Lock()
		uid = ulid.MustNew(ulid.Now(), entropy)
		m.Unlock()
		return
	}
}

// RandomULIDGenerator returns a function that generates a cryptographically random ULID
//   - this is ~5x slower than MonotonicULIDGenerator functions
//   - panics if a ULID fails to be generated
func RandomULIDGenerator() func() ulid.ULID {
	return func() ulid.ULID {
		return MustNew()
	}
}

// MustNew generates a new crypto/rand based ULID.
//   - panics if a ULID fails to be generated
func MustNew() ulid.ULID {
	return ulid.MustNew(ulid.Now(), rand.Reader)
}

// Parse tries to parse the id into a ULID.
func Parse(id string) (ulid.ULID, error) {
	ulidID, err := ulid.Parse(id)
	if err != nil {
		return ulidID, err
	}
	if IsZero(ulidID) {
		return ulidID, errors.New("ID must not be zero")
	}

	return ulidID, nil
}

// IsZero returns true if the id is a zero value
func IsZero(id ulid.ULID) bool {
	return ulid.ULID{} == id
}
