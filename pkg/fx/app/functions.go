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

import "github.com/oklog/ulid"

// ID returns the application ID, i.e., it corresponds to an application
type ID func() ulid.ULID

// ReleaseID returns the application release ID, i.e., it corresponds to an applicaiton release mapped to a specific version
type ReleaseID func() ulid.ULID

// InstanceID returns the application instance ID, i.e., it corresponds to an application instance
type InstanceID func() ulid.ULID
