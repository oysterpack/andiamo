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

package fxapp

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/oklog/ulid"
	"github.com/oysterpack/andiamo/pkg/ulids"
)

// ID corresponds to an application
type ID ulid.ULID

// ReleaseID corresponds to an application release
type ReleaseID ulid.ULID

// InstanceID corresponds to an application instance
type InstanceID ulid.ULID

// LoadIDsFromEnv tries to load the app descriptor from env vars:
//
//   - APP12X_ID
//   - APP12X_RELEASE_ID
func LoadIDsFromEnv() (ID, ReleaseID, error) {
	type desc struct {
		ID        string `required:"true"`                    // ULID
		ReleaseID string `required:"true" split_words:"true"` // ULID
	}

	var cfg desc
	err := envconfig.Process(EnvconfigPrefix, &cfg)
	if err != nil {
		return ID(ulid.ULID{}), ReleaseID(ulid.ULID{}), err
	}

	id, err := ulids.Parse(cfg.ID)
	if err != nil {
		return ID(ulid.ULID{}), ReleaseID(ulid.ULID{}), err
	}

	releaseID, err := ulids.Parse(cfg.ReleaseID)
	if err != nil {
		return ID(id), ReleaseID(ulid.ULID{}), err
	}

	return ID(id), ReleaseID(releaseID), nil
}
