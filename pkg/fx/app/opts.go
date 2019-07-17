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

package app

import (
	"github.com/oklog/ulid"
)

// Opts is used to configure the fx module
type Opts struct {
	// EnvPrefix is used to load the app ID and ReleaseID from env vars, using the following naming:
	//
	//	${EnvPrefix}_ID
	//  ${EnvPrefix}_RELEASE_ID
	//
	// If blank, then the default value of "APP12X" will be used - defined by the `EnvPrefix` const
	EnvPrefix string

	ID        ulid.ULID // if set, then it will not be loaded from the env
	ReleaseID ulid.ULID // if set, then it will not be loaded from the env
}
