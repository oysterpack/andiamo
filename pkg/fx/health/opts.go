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

import "time"

// Opts are used to configure the fx module.
type Opts struct {
	MinRunInterval time.Duration
	MaxTimeout     time.Duration

	DefaultTimeout     time.Duration
	DefaultRunInterval time.Duration

	MaxCheckParallelism uint8

	// FailFastOnStartup means the app will fail fast if any health checks fail to pass on app start up.
	//
	// default = false
	FailFastOnStartup bool
}

// DefaultOpts constructs a new Opts using recommended default values.
func DefaultOpts() Opts {
	return Opts{
		MinRunInterval: MinRunInterval,
		MaxTimeout:     MaxTimeout,

		DefaultRunInterval: DefaultRunInterval,
		DefaultTimeout:     DefaultTimeout,

		MaxCheckParallelism: MaxCheckParallelism,
	}
}
