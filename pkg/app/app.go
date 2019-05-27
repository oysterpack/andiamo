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
	"crypto/rand"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"go.uber.org/fx"
	"time"
)

// ENV_PREFIX is used as the environment variable name prefix to load config.
// "APPX12" was chosen to represent 12-factor apps.
const ENV_PREFIX = "APPX12"

// New construct a new fx.App with the following options:
// - app start and stop timeout options are configured from the env - see `LoadConfigFromEnv()`
// - constructor functions for:
//   - Desc - loaded from the env - see `LoadDescFromEnv()`
//   - InstanceID
func New(options ...fx.Option) *fx.App {
	config := LoadConfigFromEnv()
	options = append(options, fx.StartTimeout(config.StartTimeout))
	options = append(options, fx.StopTimeout(config.StopTimeout))
	options = append(options, fx.Provide(LoadDescFromEnv))
	options = append(options, fx.Provide(func() InstanceID {
		return InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	}))
	return fx.New(options...)
}

// ID is the unique application ID.
type ID ulid.ULID

// Decode implements the envconfig.Decoder interface
func (id *ID) Decode(value string) error {
	uid, err := ulid.Parse(value)
	if err != nil {
		return err
	}
	*id = ID(uid)
	return nil
}

// Name is the application name.
type Name string

// ReleaseID is the application release ID.
type ReleaseID ulid.ULID

// Decode implements the envconfig.Decoder interface
func (id *ReleaseID) Decode(value string) error {
	uid, err := ulid.Parse(value)
	if err != nil {
		return err
	}
	*id = ReleaseID(uid)
	return nil
}

type Version semver.Version

// Decode implements the envconfig.Decoder interface
func (v *Version) Decode(value string) error {
	version, err := semver.NewVersion(value)
	if err != nil {
		return err
	}
	*v = Version(*version)
	return nil
}

// InstanceID is the unique app instance ID
type InstanceID ulid.ULID
