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
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/kelseyhightower/envconfig"
	"github.com/oklog/ulid"
	"go.uber.org/multierr"
	"regexp"
	"strings"
)

var (
	// name constraints:
	// - must be alpha-numeric and can contain the following non-alpha-numeric chars: '_' '-'
	// - must start with an alpha
	// - min len = 3, max len = 50
	nameRegex = regexp.MustCompile(`^[[:alpha:]][a-zA-Z0-9_-]{2,49}$`)
)

// Desc represents the application descriptor
type Desc interface {
	// ID returns the app ID
	ID() ulid.ULID

	// Name returns the app name
	//
	// name constraints:
	// - must be alpha-numeric and can contain the following non-alpha-numeric chars: '_' '-'
	// - must start with an alpha
	// - min len = 3, max len = 50
	Name() string

	// Version returns the app version
	Version() *semver.Version

	// ReleaseID returns the app release ID
	ReleaseID() ulid.ULID

	// Validates checks if the app descriptor is valid
	Validate() error

	fmt.Stringer
}

// DescBuilder constructs a new Desc
type DescBuilder interface {
	SetID(id ulid.ULID) DescBuilder
	SetName(name string) DescBuilder
	SetVersion(version *semver.Version) DescBuilder
	SetReleaseID(id ulid.ULID) DescBuilder

	Build() (Desc, error)
}

// NewDescBuilder constructs a new DescBuilder
func NewDescBuilder() DescBuilder {
	return &desc{}
}

type desc struct {
	id        ulid.ULID
	name      string
	version   *semver.Version
	releaseID ulid.ULID
}

func (d *desc) String() string {
	return fmt.Sprintf("Desc{ID: %s, Name: %s, Version: %v, ReleaseID: %s}", d.id, d.name, d.version, d.releaseID)
}

func (d *desc) Build() (Desc, error) {
	return d, d.Validate()
}

func (d *desc) Validate() error {
	var err error
	zeroULID := ulid.ULID{}
	if d.id == zeroULID {
		err = multierr.Append(err, errors.New("`ID` is required"))
	}
	d.name = strings.TrimSpace(d.name)
	if !nameRegex.MatchString(d.name) {
		err = multierr.Append(err, fmt.Errorf("`Name` failed to match against regex: %q : %q", nameRegex, d.name))
	}
	if d.version == nil {
		err = multierr.Append(err, errors.New("`Version` is required"))
	}
	if d.releaseID == zeroULID {
		err = multierr.Append(err, errors.New("`ReleaseID` is required"))
	}
	return err
}

func (d *desc) ID() ulid.ULID {
	return d.id
}

func (d *desc) SetID(id ulid.ULID) DescBuilder {
	d.id = id
	return d
}

func (d *desc) Name() string {
	return d.name
}

func (d *desc) SetName(name string) DescBuilder {
	d.name = name
	return d
}

func (d *desc) Version() *semver.Version {
	return d.version
}

func (d *desc) SetVersion(version *semver.Version) DescBuilder {
	d.version = version
	return d
}

func (d *desc) ReleaseID() ulid.ULID {
	return d.releaseID
}

func (d *desc) SetReleaseID(releaseID ulid.ULID) DescBuilder {
	d.releaseID = releaseID
	return d
}

type envconfigDesc struct {
	ID        string `required:"true"`
	Name      string `required:"true"`
	Version   string `required:"true"`
	ReleaseID string `required:"true" split_words:"true"`
}

// LoadDescFromEnv tries to load the app descriptor from env vars:
//
//   - APP12X_ID
//   - APP12X_NAME
//   - APP12X_VERSION
//   - APP12X_RELEASE_ID
func LoadDescFromEnv() (Desc, error) {
	var cfg envconfigDesc
	err := envconfig.Process(EnvconfigPrefix, &cfg)
	if err != nil {
		return nil, err
	}

	id, e := ulid.Parse(cfg.ID)
	err = multierr.Append(err, e)

	version, e := semver.NewVersion(cfg.Version)
	err = multierr.Append(err, e)

	releaseID, e := ulid.Parse(cfg.ReleaseID)
	err = multierr.Append(err, e)

	if err != nil {
		return nil, err
	}

	return NewDescBuilder().
		SetID(id).
		SetName(cfg.Name).
		SetVersion(version).
		SetReleaseID(releaseID).
		Build()

}
