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

package health

// Status is used to define a health error status
type Status uint8

// Status enum
const (
	Green Status = iota
	// Yellow indicates the health check is triggering a warning - usually to signal a degraded state.
	Yellow
	Red
)

func (e Status) String() string {
	switch e {
	case Green:
		return "Green"
	case Yellow:
		return "Yellow"
	default:
		return "Red"
	}
}
