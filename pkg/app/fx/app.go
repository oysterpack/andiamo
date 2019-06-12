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
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"io"
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
//
// Application lifecycle events are logged.
func (a *App) Run() error {
	startCtx, cancel := context.WithTimeout(context.Background(), a.StartTimeout())
	defer cancel()

	stopChan := a.App.Done()

	if e := a.Start(startCtx); e != nil {
		appStartErr := AppStartErr.CausedBy(e)
		appStartErr.Log(a.logger).Msg("")
		return appStartErr
	}

	// wait for the app to be signalled to stop
	signal := <-stopChan
	StopSignal.Log(a.logger).Msg(strings.ToUpper(signal.String()))

	stopCtx, cancel := context.WithTimeout(context.Background(), a.StopTimeout())
	defer cancel()

	if e := a.Stop(stopCtx); e != nil {
		appStopErr := AppStopErr.CausedBy(e)
		appStopErr.Log(a.logger).Msg("")
		return appStopErr
	}

	return nil
}

// MustNewApp constructs a new fx.App with the specified options.
//
// The app is pre-configured with the following options:
//   - app start and stop timeout options are configured from the env - see `LoadTimeouts()`
//   - constructor functions for:
//     - Desc - loaded from the env - see `LoadDesc()`
//     - InstanceID
//     - *zerolog.Logger
//       - is used as the fx.App logger, which logs all fx.App log events using debug level
//       - is used as the go std logger
//     - *err.Registry
//     - *logging.EventRegistry
//     - *comp.Registry
//   - lifecycle hooks are registered to log app lifecycle events
//   - fx.ErrorHandler is registered to log invoke errors
//
// NOTE: Only `provide` and `invoke` options should be specified. `populate` options are useful for unit testing.
func MustNewApp(opt fx.Option, opts ...fx.Option) *App {
	desc := mustLoadDesc()
	timeouts := mustLoadAppStartStopTimeouts()

	var appOptions fx.Option
	if len(opts) > 0 {
		appOptions = fx.Options(opts...)
		appOptions = fx.Options(opt, appOptions)
	} else {
		appOptions = opt
	}
	fxapp, e := NewApp(desc, timeouts, nil, zerolog.NoLevel, appOptions)
	if e != nil {
		log.Panic(e)
	}

	return fxapp
}

// NewApp tries to construct a new App
func NewApp(desc app.Desc, timeouts app.Timeouts,
	logWriter io.Writer, globalLogLevel zerolog.Level,
	opts ...fx.Option) (*App, error) {
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	logger, e := initLogging(instanceID, desc)
	if e != nil {
		return nil, e
	}
	if logWriter != nil {
		customLogger := logger.Output(logWriter)
		logger = &customLogger
	}
	if globalLogLevel != zerolog.NoLevel {
		zerolog.SetGlobalLevel(globalLogLevel)
	}

	appOptions := []fx.Option{
		fx.Invoke(registerStartStoppedLifecycleEventLoggerHook),

		fx.StartTimeout(timeouts.StartTimeout),
		fx.StopTimeout(timeouts.StopTimeout),

		fx.Provide(
			func() app.Desc { return desc },
			func() app.InstanceID { return instanceID },
			func() *zerolog.Logger { return logger },
			newErrorRegistry,
			newEventRegistry,
			comp.NewRegistry,
		),

		fx.Logger(logger),
		fx.ErrorHook(newErrLogger(logger)),

		// application specific options
		fx.Options(opts...),
		fx.Invoke(registerComponents),

		fx.Invoke(registerRunningStoppingLifecycleEventLoggerHook),
	}

	fxapp := &App{
		App:    fx.New(appOptions...),
		logger: logger,
	}
	if e := fxapp.Err(); e != nil {
		return nil, e
	}

	return fxapp, nil
}

func mustLoadAppStartStopTimeouts() app.Timeouts {
	timeouts, e := app.LoadTimeouts()
	if e != nil {
		log.Panicf("app.LoadTimeouts() failed: %v", e)
	}
	return timeouts
}

func newEventRegistry() *logging.EventRegistry {
	registry := logging.NewEventRegistry()
	registry.Register(Start, Running, Stop, Stopped, StopSignal, CompRegistered)
	return registry
}

func newErrorRegistry() (*err.Registry, error) {
	registry := err.NewRegistry()
	if e := registry.Register(InvokeErr, AppStartErr, AppStopErr); e != nil {
		// should never happen - if it does, then it means it is a bug
		return nil, e
	}
	return registry, nil
}

func mustLoadDesc() app.Desc {
	desc, e := app.LoadDesc()
	if e != nil {
		log.Panicf("failed to load app.Desc: %v", e)
	}
	return desc
}

func initLogging(instanceID app.InstanceID, desc app.Desc) (*zerolog.Logger, error) {
	if e := logcfg.ConfigureZerolog(); e != nil {
		return nil, e
	}
	logger := logcfg.NewLogger(instanceID, desc)
	logcfg.UseAsStandardLoggerOutput(logger)
	return logger, nil
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

func newErrLogger(logger *zerolog.Logger) *errLogger {
	return &errLogger{logger}
}

func logError(logger *zerolog.Logger, e error) {
	switch e := e.(type) {
	case *err.Instance:
		e.Log(logger).Msg("")
	default:
		InvokeErr.CausedBy(e).Log(logger).Msg("")
	}
}

type components struct {
	fx.In

	Comps []*comp.Comp `group:"comp.Registry"`
}

//func provideCompRegistry(comps components, logger *zerolog.Logger) (*comp.Registry, error) {
//	registry := comp.NewRegistry()
//	for _, c := range comps.Comps {
//		if e := registry.Register(c); e != nil {
//			return nil, e
//		}
//		logCompRegisteredEvent(c, logger)
//	}
//	return registry, nil
//}

func registerComponents(registry *comp.Registry, comps components, logger *zerolog.Logger) error {
	for _, c := range comps.Comps {
		if e := registry.Register(c); e != nil {
			return e
		}
		logCompRegisteredEvent(c, logger)
	}
	return nil
}

func logCompRegisteredEvent(c *comp.Comp, logger *zerolog.Logger) {
	options := make([]string, len(c.Options))
	for i := 0; i < len(options); i++ {
		optionDesc := c.Options[i].Desc
		options[i] = fmt.Sprintf("%s => %v", optionDesc.Type, optionDesc.FuncType)
	}

	CompRegistered.Log(c.Logger(logger)).
		Dict(logging.Comp.String(), zerolog.Dict().
			Str(logging.CompID.String(), c.ID.String()).
			Str(logging.CompVersion.String(), c.Version.String()).
			Strs(logging.CompOptions.String(), options),
		).Msg("")
}
