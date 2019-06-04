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
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"log"
	"strings"
	"time"
)

// App is a wrapper around fx.App.
//
// The main reason to wrap fx.App is too override fx.App.Run() with App.Run() that returns an error.
// fx.App.Run() would fatally exit the process if a start or stop error occurs. Instead, we want to simply return an
// error, which also makes it easier to test the App. In addition, the following behaviors are added:
// - all start and stop errors are logged using standardized errors
// - a StopSignal event is logged, which logs exactly what type of os.Signal was received
type App struct {
	*fx.App
	logger *zerolog.Logger
}

// Run starts the application, blocks on the signals channel, and then gracefully shuts the application down.
// It uses DefaultTimeout to set a deadline for application startup and shutdown, unless the user has configured
// different timeouts with the StartTimeout or StopTimeout options. It's designed to make typical applications simple to run.
// Start and stop errors are logged. The stop signal that is received is logged.
func (a *App) Run() error {
	startCtx, cancel := context.WithTimeout(context.Background(), a.StartTimeout())
	defer cancel()

	stopChan := a.App.Done()

	if e := a.Start(startCtx); e != nil {
		appStartErr := err.New(AppStartErr, "01DCFMZ5KHESA1E20C7DHMGS9Y").CausedBy(e)
		appStartErr.Log(a.logger).Msg("")
		return appStartErr
	}

	// wait for the app to be signalled to stop
	signal := <-stopChan
	StopSignal.Log(a.logger).Msg(strings.ToUpper(signal.String()))

	stopCtx, cancel := context.WithTimeout(context.Background(), a.StopTimeout())
	defer cancel()

	if e := a.Stop(stopCtx); e != nil {
		appStopErr := err.New(AppStopErr, "01DCFPFAFFDPKVF5GPYEYJ8Y8C").CausedBy(e)
		appStopErr.Log(a.logger).Msg("")
		return appStopErr
	}

	return nil
}

// New constructs a new fx.App.
// It is configured with the following options:
// - app start and stop timeout options are configured from the env - see `LoadTimeouts()`
// - constructor functions for:
//   - Desc - loaded from the env - see `LoadDesc()`
//   - InstanceID
//   - *zerolog.Logger
//     - is used as the fx.App logger, which logs all fx.App log events using debug level
//     - is used as the go std logger
// - lifecycle hook is registered to log app.Start and app.Stop log events
// - fx.ErrorHandler is registered to log invoke errors
func New(options ...fx.Option) *App {
	config, e := app.LoadTimeouts()
	if e != nil {
		log.Panicf("app.LoadTimeouts() failed: %v", e)
	}

	appOptions := []fx.Option{
		fx.StartTimeout(config.StartTimeout),
		fx.StopTimeout(config.StopTimeout),
		fx.Invoke(registerStartStoppedLifecycleEventLoggerHook),
	}

	desc := loadDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	logger := initLogging(instanceID, desc)
	appOptions = append(appOptions, fx.Provide(func() (app.Desc, app.InstanceID, *zerolog.Logger) {
		return desc, instanceID, logger
	}))
	appOptions = append(appOptions, fx.Logger(logger))
	errHandler := &errLogger{logger}
	appOptions = append(appOptions, fx.ErrorHook(errHandler))
	appOptions = append(appOptions, options...)
	appOptions = append(appOptions, fx.Invoke(registerRunningStoppingLifecycleEventLoggerHook))

	return &App{
		App:    fx.New(appOptions...),
		logger: logger,
	}
}

// panics if the Desc fails to load
// - the panic is logged via go std log
func loadDesc() app.Desc {
	desc, e := app.LoadDesc()
	if e != nil {
		log.Panicf("failed to load app.Desc: %v", e)
	}
	return desc
}

// panics if an error occurs while trying to configure zerolog
func initLogging(instanceID app.InstanceID, desc app.Desc) *zerolog.Logger {
	logger := logcfg.NewLogger(instanceID, desc)
	if e := logcfg.ConfigureZerolog(); e != nil {
		log.Panicf("logcfg.ConfigureZerolog() failed: %v", e)
	}
	logcfg.UseAsStandardLoggerOutput(logger)
	return logger
}

func registerStartStoppedLifecycleEventLoggerHook(lc fx.Lifecycle, logger *zerolog.Logger) {
	const PACKAGE app.Package = "github.com/oysterpack/partire-k8s/pkg/app/fx"
	appLogger := logging.PackageLogger(logger, PACKAGE)
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			Start.Log(appLogger).Msg("")
			return nil
		},
		OnStop: func(context.Context) error {
			Stopped.Log(appLogger).Msg("")
			return nil
		},
	})
}

func registerRunningStoppingLifecycleEventLoggerHook(lc fx.Lifecycle, logger *zerolog.Logger) {
	const PACKAGE app.Package = "github.com/oysterpack/partire-k8s/pkg/app/fx"
	appLogger := logging.PackageLogger(logger, PACKAGE)
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			Running.Log(appLogger).Msg("")
			return nil
		},
		OnStop: func(context.Context) error {
			Stop.Log(appLogger).Msg("")
			return nil
		},
	})
}

type errLogger struct {
	*zerolog.Logger
}

// implements fx.ErrorHandler
func (l *errLogger) HandleError(e error) {
	logError(l.Logger, e)
}

func logError(logger *zerolog.Logger, e error) {
	switch e := e.(type) {
	case *err.Instance:
		e.Log(logger).Msg("")
	default:
		err.New(InvokeErr, "01DCFB4PKEBPEBQNWH7SMDXNAZ").CausedBy(e).Log(logger).Msg("")
	}
}
