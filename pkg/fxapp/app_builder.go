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
	"bytes"
	"errors"
	"fmt"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"go.uber.org/multierr"
	"io"
	"log"
	"os"
	"reflect"
	"time"
)

// Builder is used to construct a new App instance.
type Builder interface {
	// Provide is used to provide dependency injection
	Provide(constructors ...interface{}) Builder
	// Invoke is used to register application functions, which will be invoked to to initialize the app.
	// The functions are invoked in the order that they are registered.
	Invoke(funcs ...interface{}) Builder

	SetStartTimeout(timeout time.Duration) Builder
	SetStopTimeout(timeout time.Duration) Builder

	// LogWriter is used as the zerolog writer.
	//
	// By default, stderr is used.
	LogWriter(w io.Writer) Builder
	LogLevel(level LogLevel) Builder

	// Error handlers
	HandleInvokeError(errorHandlers ...func(error)) Builder
	HandleStartupError(errorHandlers ...func(error)) Builder
	HandleShutdownError(errorHandlers ...func(error)) Builder
	// HandleError will handle any app error, i.e., app function invoke errors, app startup errors, and app shutdown errors.
	HandleError(errorHandlers ...func(error)) Builder

	// Populate sets targets with values from the dependency injection container during application initialization.
	// All targets must be pointers to the values that must be populated.
	// Pointers to structs that embed fx.In are supported, which can be used to populate multiple values in a struct.
	//
	// NOTE: this is useful for unit testing
	Populate(targets ...interface{}) Builder

	ExposePrometheusMetricsViaHTTP(opts *PrometheusHTTPServerOpts) Builder

	Build() (App, error)
}

// NewBuilder constructs a new Builder
func NewBuilder(desc Desc) Builder {
	return &builder{
		instanceID:   NewInstanceID(),
		desc:         desc,
		startTimeout: fx.DefaultTimeout,
		stopTimeout:  fx.DefaultTimeout,

		globalLogLevel: zerolog.InfoLevel,
		logWriter:      os.Stderr,
	}
}

type builder struct {
	instanceID InstanceID
	desc       Desc

	startTimeout time.Duration
	stopTimeout  time.Duration

	constructors    []interface{}
	funcs           []interface{}
	populateTargets []interface{}

	logWriter      io.Writer
	globalLogLevel zerolog.Level

	invokeErrorHandlers, startErrorHandlers, stopErrorHandlers []func(error)

	prometheusHTTPServerOpts *PrometheusHTTPServerOpts
}

func (b *builder) String() string {
	types := func(objs []interface{}) string {
		if len(objs) == 0 {
			return "[]"
		}
		s := new(bytes.Buffer)
		s.WriteString("[")
		s.WriteString(reflect.TypeOf(objs[0]).String())
		for i := 1; i < len(objs); i++ {
			s.WriteString("|")
			s.WriteString(reflect.TypeOf(objs[i]).String())
		}

		s.WriteString("]")
		return s.String()
	}

	return fmt.Sprintf("Builder{%v, StartTimeout: %s, StopTimeout: %s, Provide: %s, Invoke: %s, Populate: %s, InvokeErrHandlerCount: %d, StartErrHandlerCount: %d}",
		b.desc,
		b.startTimeout,
		b.startTimeout,
		types(b.constructors),
		types(b.funcs),
		types(b.populateTargets),
		len(b.invokeErrorHandlers),
		len(b.startErrorHandlers),
	)
}

// Build tries to construct and initialize a new App instance.
// All of the app's functions are run as part of the app initialization phase.
func (b *builder) Build() (App, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}

	var shutdowner fx.Shutdowner
	var logger *zerolog.Logger
	b.populateTargets = append(b.populateTargets, &shutdowner, &logger)
	app := &app{
		instanceID:   b.instanceID,
		desc:         b.desc,
		constructors: b.constructors,
		funcs:        b.funcs,

		startErrorHandlers: b.startErrorHandlers,
		stopErrorHandlers:  b.stopErrorHandlers,

		starting: make(chan struct{}),
		started:  make(chan struct{}),
		stopping: make(chan os.Signal, 1),
		stopped:  make(chan os.Signal, 1),

		App: fx.New(
			fx.StartTimeout(b.startTimeout),
			fx.StopTimeout(b.stopTimeout),
			fx.Options(b.buildOptions()...),
		),

		Shutdowner: shutdowner,
	}
	app.startErrorHandlers = append(app.startErrorHandlers, func(e error) {
		logEvent := StartFailedEventID.NewLogEventer(logger, zerolog.ErrorLevel)
		logEvent(AppFailed{e}, "app start failed")
	})
	app.stopErrorHandlers = append(app.stopErrorHandlers, func(e error) {
		logEvent := StopFailedEventID.NewLogEventer(logger, zerolog.ErrorLevel)
		logEvent(AppFailed{e}, "app stop failed")
	})

	if err := app.Err(); err != nil {
		return nil, err
	}
	app.logger = logger
	app.logAppInitialized()
	return app, nil
}

func (b *builder) validate() error {
	var err error
	if b.desc == nil {
		err = multierr.Append(err, errors.New("app descriptor is required"))
	} else {
		err = multierr.Append(err, b.desc.Validate())
	}
	if len(b.constructors) == 0 && len(b.funcs) == 0 {
		err = multierr.Append(err, errors.New("at least 1 functional option is required"))
	}
	return err
}

func (b *builder) buildOptions() []fx.Option {
	compOptions := make([]fx.Option, 0, 5+len(b.invokeErrorHandlers))

	instanceID := b.instanceID
	desc := b.desc
	logger := b.initZerolog()
	compOptions = append(compOptions, fx.Provide(
		func() Desc { return desc },
		func() InstanceID { return instanceID },
		func() *zerolog.Logger { return logger },
		newMetricRegistry,
	))

	compOptions = append(compOptions, fx.Provide(b.constructors...))
	compOptions = append(compOptions, fx.Invoke(b.funcs...))
	compOptions = append(compOptions, fx.Populate(b.populateTargets...))
	compOptions = append(compOptions, fx.Logger(newFxLogger(logger)))

	if b.prometheusHTTPServerOpts != nil {
		compOptions = append(compOptions, fx.Invoke(PrometheusHTTPServerRunner(b.prometheusHTTPServerOpts)))
	}

	for _, f := range b.invokeErrorHandlers {
		compOptions = append(compOptions, fx.ErrorHook(errorHandler(f)))
	}
	compOptions = append(compOptions, fx.ErrorHook(errorHandler(func(err error) {
		logEvent := InitFailedEventID.NewLogEventer(logger, zerolog.ErrorLevel)
		logEvent(AppFailed{err}, "app init failed")
	})))

	return compOptions
}

type fxlogger struct {
	*zerolog.Logger
}

func newFxLogger(logger *zerolog.Logger) fxlogger {
	return fxlogger{ComponentLogger(logger, "fx")}
}

func (l fxlogger) Printf(msg string, params ...interface{}) {
	l.Log().Msgf(msg, params...)
}

func (b *builder) initZerolog() *zerolog.Logger {
	zerolog.SetGlobalLevel(b.globalLogLevel)
	var newEventID = ulidgen.MonotonicULIDGenerator()
	logger := zerolog.New(b.logWriter).
		Hook(zerolog.HookFunc(func(e *zerolog.Event, _ zerolog.Level, _ string) {
			e.Str(EventLabel, newEventID().String())
		})).
		With().
		Timestamp().
		Str(AppIDLabel, b.desc.ID().String()).
		Str(AppReleaseIDLabel, b.desc.ReleaseID().String()).
		Str(AppInstanceIDLabel, b.instanceID.String()).
		Logger()

	// use the logger as the go standard log output
	log.SetFlags(0)
	log.SetOutput(&logger)

	return &logger
}

func newMetricRegistry(appDesc Desc, instanceID InstanceID) (prometheus.Gatherer, prometheus.Registerer) {
	registry := prometheus.NewRegistry()
	regsisterer := prometheus.WrapRegistererWith(
		prometheus.Labels{
			AppIDLabel:         appDesc.ID().String(),
			AppReleaseIDLabel:  appDesc.ReleaseID().String(),
			AppInstanceIDLabel: instanceID.String(),
		},
		registry,
	)
	regsisterer.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{ReportErrors: true}),
	)

	return registry, regsisterer
}

func (b *builder) SetStartTimeout(timeout time.Duration) Builder {
	b.startTimeout = timeout
	return b
}

func (b *builder) SetStopTimeout(timeout time.Duration) Builder {
	b.stopTimeout = timeout
	return b
}

func (b *builder) Provide(constructors ...interface{}) Builder {
	b.constructors = append(b.constructors, constructors...)
	return b
}

func (b *builder) Invoke(funcs ...interface{}) Builder {
	b.funcs = append(b.funcs, funcs...)
	return b
}

func (b *builder) Populate(targets ...interface{}) Builder {
	b.populateTargets = append(b.populateTargets, targets...)
	return b
}

func (b *builder) HandleInvokeError(errorHandlers ...func(error)) Builder {
	b.invokeErrorHandlers = append(b.invokeErrorHandlers, errorHandlers...)
	return b
}

func (b *builder) HandleStartupError(errorHandlers ...func(error)) Builder {
	b.startErrorHandlers = append(b.startErrorHandlers, errorHandlers...)
	return b
}

func (b *builder) HandleShutdownError(errorHandlers ...func(error)) Builder {
	b.stopErrorHandlers = append(b.stopErrorHandlers, errorHandlers...)
	return b
}

func (b *builder) HandleError(errorHandlers ...func(error)) Builder {
	b.HandleInvokeError(errorHandlers...)
	b.HandleStartupError(errorHandlers...)
	b.HandleShutdownError(errorHandlers...)
	return b
}

func (b *builder) LogWriter(w io.Writer) Builder {
	b.logWriter = w
	return b
}

func (b *builder) LogLevel(level LogLevel) Builder {
	b.globalLogLevel = level.ZerologLevel()
	return b
}

func (b *builder) ExposePrometheusMetricsViaHTTP(opts *PrometheusHTTPServerOpts) Builder {
	b.prometheusHTTPServerOpts = opts
	return b
}
