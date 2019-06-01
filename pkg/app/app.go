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
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
)

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

func (id ID) String() string {
	return ulid.ULID(id).String()
}

// Name is the application name.
type Name string

// ReleaseID is the application release ID.
type ReleaseID ulid.ULID

func (id ReleaseID) String() string {
	return ulid.ULID(id).String()
}

// Decode implements the envconfig.Decoder interface
func (id *ReleaseID) Decode(value string) error {
	uid, err := ulid.Parse(value)
	if err != nil {
		return err
	}
	*id = ReleaseID(uid)
	return nil
}

// Version represents the app version
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

func (v *Version) String() string {
	return (*semver.Version)(v).String()
}

// InstanceID is the unique app instance ID
type InstanceID ulid.ULID

func (id InstanceID) String() string {
	return ulid.ULID(id).String()
}
