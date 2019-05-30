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

package fx

import (
	"context"
	"crypto/rand"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"log"
	"time"
)

// New constructs a new fx.App.
// It is configured with the following options:
// - app start and stop timeout options are configured from the env - see `LoadConfig()`
// - constructor functions for:
//   - Desc - loaded from the env - see `LoadDesc()`
//   - InstanceID
//   - *zerolog.Logger
//     - is used as the fx.App logger, which logs all fx.App log events using debug level
//     - is used as the go std logger
// - lifecycle hook is registered to log app.Start and app.Stop log events
func New(options ...fx.Option) *fx.App {
	config := app.LoadConfig()
	options = append(options, fx.StartTimeout(config.StartTimeout))
	options = append(options, fx.StopTimeout(config.StopTimeout))

	desc := loadDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	logger := initLogging(instanceID, desc)
	options = append(options, fx.Provide(func() (app.Desc, app.InstanceID, *zerolog.Logger) {
		return desc, instanceID, logger
	}))
	options = append(options, fx.Logger(logger))
	options = append(options, fx.Invoke(registerLifecycleEventLoggerHook))

	return fx.New(options...)
}

// panics if the Desc fails to load
// - the panic is logged via go std log
func loadDesc() app.Desc {
	desc, err := app.LoadDesc()
	if err != nil {
		log.Panicf("failed to load app.Desc: %v", err)
	}
	return desc
}

// panics if an error occurs while trying to configure zerolog
func initLogging(instanceID app.InstanceID, desc app.Desc) *zerolog.Logger {
	logger := app.NewLogger(instanceID, desc)
	if err := app.ConfigureZerolog(); err != nil {
		panic(err)
	}
	app.UseAsStandardLoggerOutput(logger)
	return logger
}

func registerLifecycleEventLoggerHook(lc fx.Lifecycle, logger *zerolog.Logger) {
	const PACKAGE app.Package = "github.com/oysterpack/partire-k8s/pkg/app/fx"
	appLogger := PACKAGE.Logger(logger)
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			app.Start.Log(appLogger).Msg("")
			return nil
		},
		OnStop: func(context.Context) error {
			app.Stop.Log(appLogger).Msg("")
			return nil
		},
	})
}
