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
	"github.com/oysterpack/andiamo/pkg/eventlog"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"log"
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
			return opts.id()
		},
		func() (ReleaseID, error) {
			return opts.releaseID()
		},
		func() InstanceID {
			id := ulids.MustNew()
			return func() ulid.ULID { return id }
		},
		provideEventLogger(opts),
	)
}

// application ID labels
//
// - used to add app IDs to log events, e.g.,
//
// 	 {"a":"01DG138TTVDX5JH5F4GMNC3V67","r":"01DG138TTVK4MVW3B5TJGDSKHR","x":"01DG138TTVYGSN7QWBFT9660SS","n":"foo","z":"01DG138TTVBHCXQW29QTQAWPNM","t":1563405085,"m":"bar"}
const (
	IDLabel         = "a"
	ReleaseIDLabel  = "r"
	InstanceIDLabel = "x"
)

func provideEventLogger(opts Opts) func(id ID, releaseID ReleaseID, instanceID InstanceID) (Logger, error) {
	setGlobalLogLevel := func(opts Opts) error {
		level, err := opts.globalLogLevel()
		if err != nil {
			return err
		}
		zerolog.SetGlobalLevel(level)
		return nil
	}

	return func(id ID, releaseID ReleaseID, instanceID InstanceID) (Logger, error) {
		if err := setGlobalLogLevel(opts); err != nil {
			return nil, err
		}

		logger := eventlog.NewZeroLogger(opts.logWriter()).
			With().
			Str(IDLabel, ulid.ULID(id()).String()).
			Str(ReleaseIDLabel, ulid.ULID(releaseID()).String()).
			Str(InstanceIDLabel, ulid.ULID(instanceID()).String()).
			Logger()

		// use the logger as the go standard log output
		log.SetFlags(0)
		log.SetOutput(eventlog.ForComponent(&logger, "log"))

		return func(event string, level zerolog.Level) eventlog.Logger {
			return eventlog.NewLogger(event, &logger, level)
		}, nil
	}
}

//type fxPrinter eventlog.Logger
//
//func (p fxPrinter) Printf(msg string, args ...interface{}) {
//	switch {
//	case len(args) == 0:
//		p(nil, msg)
//	default:
//		p(nil, fmt.Sprintf(msg, args...))
//	}
//}
//
//func provideFxLogger(opts Opts) fx.Printer {
//	return nil
//}
