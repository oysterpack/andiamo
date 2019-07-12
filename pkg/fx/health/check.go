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

import (
	"time"
)

// Check defines a health check
type Check struct {
	// ID format is ULID
	ID          string
	Description string
	// RedImpact describes what is the application impact of the health check status is red.
	RedImpact string
	// YellowImpact describes what is the application impact of the health check status is yellow.
	// YellowImpact is optional because some health checks may not have a yellow state.
	YellowImpact string // optional
	// Tags are used to categorize related health checks.
	// Tags are ULIDs because naming is hard and we want to avoid accidental collision.
	Tags []string // optional
}

// Checker performs the health check.
// Return a `YellowError` to indicate that the health check failed with a `Yellow` status.
//
// NOTE: health checks must be designed to run as fast and as efficient as possible.
type Checker func() error

// checker constraints
const (
	// Health checks should be designed to run fast
	MaxTimeout = 10 * time.Second
	// Health checks should not be scheduled to run more frequently than every second
	MinRunInterval = time.Second

	// MaxCheckParallelism is used to configure the max number of health checks that can run concurrently
	MaxCheckParallelism uint8 = 1
)

// checker defaults
const (
	DefaultTimeout     = 5 * time.Second
	DefaultRunInterval = 15 * time.Second
)

// CheckerOpts is used to configure Checker run Module.
// Zero values imply using the system default values.
type CheckerOpts struct {
	// Timeout must not be zero
	Timeout time.Duration
	// Used to schedule health checks to be run on an interval
	RunInterval time.Duration
}

// RegisteredCheck represents a registered health check.
//
// NOTE: when a health check is registered the following augmentations are applied:
//  - Check fields are trimmed during registration
//  - Checker function is wrapped when registered to enforce the run timeout policy.
//	- defaults are applied to CheckerOpts zero value fields
type RegisteredCheck struct {
	Check
	CheckerOpts
	Checker
}
