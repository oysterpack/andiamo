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
	"github.com/rs/zerolog"
	"go.uber.org/multierr"
	"io"
	"os"
	"strings"
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

	LogWriter io.Writer // defaults to os.stderr
	// GlobalLogLevel is used to get the global log level.
	//
	// If not explicitly set, then it will first try to lookup the log level from an env var ${EnvPrefix}_LOG_LEVEL.
	// If the env var is not set, then `zerolog.InfoLevel` is returned.
	GlobalLogLevel *zerolog.Level // defaults to zerolog.Info
}

func (o Opts) id() (func() ulid.ULID, error) {
	zero := ulid.ULID{}
	if o.ID == zero {
		id, err := ulidFromEnv(o.EnvPrefix, "ID")
		if err != nil {
			return nil, err
		}
		return id, nil
	}
	return func() ulid.ULID { return o.ID }, nil
}

func (o Opts) releaseID() (func() ulid.ULID, error) {
	zero := ulid.ULID{}
	if o.ReleaseID == zero {
		id, err := ulidFromEnv(o.EnvPrefix, "RELEASE_ID")
		if err != nil {
			return nil, err
		}
		return id, nil
	}
	return func() ulid.ULID { return o.ReleaseID }, nil
}

func (o Opts) globalLogLevel() (zerolog.Level, error) {
	if o.GlobalLogLevel == nil {
		levelStr, ok := os.LookupEnv(key(o.EnvPrefix, "LOG_LEVEL"))
		if ok {
			level, err := zerolog.ParseLevel(levelStr)
			if err != nil {
				return zerolog.InfoLevel, err
			}
			return level, nil
		}
		return zerolog.InfoLevel, nil
	}

	return *o.GlobalLogLevel, nil
}

func (o Opts) logWriter() io.Writer {
	if o.LogWriter == nil {
		return os.Stderr
	}

	return o.LogWriter
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
