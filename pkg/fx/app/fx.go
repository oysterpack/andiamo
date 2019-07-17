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
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"go.uber.org/fx"
	"go.uber.org/multierr"
	"os"
	"strings"
)

// envconfig related constants
const (
	// EnvPrefix is the standard env var name prefix.
	// "APP12X" was chosen to represent 12-factor apps.
	EnvPrefix = "APP12X"
)

// Module returns the module's fx options
func Module(opts Opts) fx.Option {
	return fx.Provide(
		func() (ID, error) {
			zero := ulid.ULID{}
			if opts.ID == zero {
				return ulidFromEnv(opts.EnvPrefix, "ID")
			}
			return func() ulid.ULID { return opts.ID }, nil
		},
		func() (ReleaseID, error) {
			zero := ulid.ULID{}
			if opts.ReleaseID == zero {
				return ulidFromEnv(opts.EnvPrefix, "RELEASE_ID")
			}
			return func() ulid.ULID { return opts.ReleaseID }, nil
		},
		func() InstanceID {
			id := ulids.MustNew()
			return func() ulid.ULID { return id }
		},
	)
}

// ulidFromEnv will try to read a ULID from an env var using the following naming convention:
//
// 	${prefix}_ID
//
// prefix will get trimmed and uppercased. If prefix is blank then "APP12X" default value will be used
func ulidFromEnv(prefix, name string) (func() ulid.ULID, error) {
	id, ok := os.LookupEnv(key(prefix, name))
	if !ok {
		return nil, fmt.Errorf("env var is not defined: %q", key(prefix, name))
	}
	appID, err := ulids.Parse(id)
	if err != nil {
		return nil, multierr.Append(fmt.Errorf("failed to parse env var as ULID: %q", key(prefix, name)), err)
	}
	return func() ulid.ULID { return appID }, nil
}

func key(prefix, name string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = EnvPrefix
	}
	return strings.ToUpper(prefix + "_" + name)
}
