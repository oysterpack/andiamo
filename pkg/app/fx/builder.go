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
	"github.com/oysterpack/partire-k8s/pkg/app/metric"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"io"
	"os"
	"time"
)

type disableLogSamplingFlag struct{}

// AppBuilder is used to build a new App
type AppBuilder struct {
	desc                      []app.Desc
	startTimeout, stopTimeout time.Duration
	comps                     []*comp.Comp
	opts                      []fx.Option
	logWriter                 io.Writer
	globalLogLevel            zerolog.Level
	disableLogSampling        *disableLogSamplingFlag
}

// NewAppBuilder returns a new AppBuilder
func NewAppBuilder() *AppBuilder {
	return &AppBuilder{
		globalLogLevel: zerolog.NoLevel,
	}
}

// DisableLogSampling disables log sampling. Log sampling is enabled by default.
func (b *AppBuilder) DisableLogSampling() *AppBuilder {
	b.disableLogSampling = &disableLogSamplingFlag{}
	return b
}

// GlobalLogLevel sets the global log level
func (b *AppBuilder) GlobalLogLevel(level zerolog.Level) *AppBuilder {
	b.globalLogLevel = level
	return b
}

// LogWriter specifies the writer that will be used for application logging.
func (b *AppBuilder) LogWriter(w io.Writer) *AppBuilder {
	b.logWriter = w
	return b
}

// AppDesc sets the app descriptor. If not specified, then the builder will try to load the app descriptor from env vars.
func (b *AppBuilder) AppDesc(desc app.Desc) *AppBuilder {
	b.desc = []app.Desc{desc}
	return b
}

// StartTimeout sets the app start timeout. If not set, then it will try to load the timeout from the env.
// If not specified in the env, then it defaults to 15 sec.
func (b *AppBuilder) StartTimeout(timeout time.Duration) *AppBuilder {
	b.startTimeout = timeout
	return b
}

// StopTimeout sets the app stop timeout. If not set, then it will try to load the timeout from the env.
// If not specified in the env, then it defaults to 15 sec.
func (b *AppBuilder) StopTimeout(timeout time.Duration) *AppBuilder {
	b.stopTimeout = timeout
	return b
}

// Options is used to specify application options.
// Only `provide` and `invoke` options should be specified.
// `populate` options come in handy when unit testing.
func (b *AppBuilder) Options(opts ...fx.Option) *AppBuilder {
	if len(opts) > 0 {
		b.opts = append(b.opts, opts...)
	}
	return b
}

// Comps is used to specifiy the application components, which will be registered with the component registry.
func (b *AppBuilder) Comps(comps ...*comp.Comp) *AppBuilder {
	if len(comps) > 0 {
		b.comps = append(b.comps, comps...)
	}
	return b
}

// Build tries to build the app.
func (b *AppBuilder) Build() (*App, error) {
	if len(b.comps) == 0 && len(b.opts) == 0 {
		return nil, OptionsRequiredErr.New()
	}

	if len(b.desc) == 0 {
		desc, e := app.LoadDesc()
		if e != nil {
			return nil, e
		}
		b.AppDesc(desc)
	}

	if e := b.desc[0].Validate(); e != nil {
		return nil, InvalidDescErr.CausedBy(e)
	}

	timeouts, e := app.LoadTimeouts()
	if e != nil {
		return nil, InvalidTimeoutsErr.CausedBy(e)
	}
	if b.startTimeout != 0 {
		timeouts.StartTimeout = b.startTimeout
	}
	if b.stopTimeout != 0 {
		timeouts.StopTimeout = b.stopTimeout
	}

	for _, c := range b.comps {
		b.opts = append(b.opts, c.FxOptions())
	}

	fxapp, e := newApp(b.desc[0], timeouts, b.logWriter, b.globalLogLevel, b.opts...)
	if b.disableLogSampling != nil {
		zerolog.DisableSampling(true)
	}
	return fxapp, e
}

// newApp tries to construct a new App
func newApp(desc app.Desc, timeouts app.Timeouts, logWriter io.Writer, globalLogLevel zerolog.Level, opts ...fx.Option) (*App, error) {
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

		fx.Logger(logger),
		fx.ErrorHook(newErrLogger(logger)),

		fx.Provide(
			func() app.Desc { return desc },
			func() app.InstanceID { return instanceID },
			func() *zerolog.Logger { return logger },
			newErrorRegistry,
			newEventRegistry,
			comp.NewRegistry,
			newMetricRegistry,
		),

		// application specific options
		fx.Options(opts...),
		fx.Invoke(registerComponents),

		fx.Invoke(registerRunningStoppingLifecycleEventLoggerHook),
	}

	fxapp := &App{
		App:     fx.New(appOptions...),
		logger:  logger,
		stopped: make(chan os.Signal, 1),
	}
	if e := fxapp.Err(); e != nil {
		return nil, e
	}

	return fxapp, nil
}

func newMetricRegistry(appDesc app.Desc, instanceID app.InstanceID) (prometheus.Gatherer, prometheus.Registerer) {
	registry := prometheus.NewRegistry()
	regsisterer := prometheus.WrapRegistererWith(
		prometheus.Labels{
			metric.AppID.String():         appDesc.ID.String(),
			metric.AppReleaseID.String():  appDesc.ReleaseID.String(),
			metric.AppInstanceID.String(): instanceID.String(),
		},
		registry,
	)
	regsisterer.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{ReportErrors: true}),
	)

	return registry, regsisterer
}

func newEventRegistry() *logging.EventRegistry {
	registry := logging.NewEventRegistry()
	registry.Register(Start, Running, Stop, Stopped, StopSignal, CompRegistered)
	return registry
}

func newErrorRegistry() (*err.Registry, error) {
	registry := err.NewRegistry()
	if e := registry.Register(err.RegistryConflictErr, InvokeErr, AppStartErr, AppStopErr); e != nil {
		// should never happen - if it does, then it means it is a bug
		return nil, e
	}
	return registry, nil
}

func initLogging(instanceID app.InstanceID, desc app.Desc) (*zerolog.Logger, error) {
	if e := logcfg.ConfigureZerolog(); e != nil {
		return nil, e
	}
	logger := logcfg.NewLogger(instanceID, desc)
	logcfg.UseAsStandardLoggerOutput(logger)
	return logger, nil
}

// this is the very first lifecycle hook registered. This means its `OnStart` hook is called first, i.e., before all other
// application specific lifecycle `OnStart` hooks. Its `OnStop` hook get called last, i.e., after all other application
// specific lifecycle `OnStop` hook.
func registerStartStoppedLifecycleEventLoggerHook(lc fx.Lifecycle, logger *zerolog.Logger) {
	appLogger := logging.PackageLogger(logger, pkg)
	New.Log(appLogger).Msg("")
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logStartEvent(appLogger)
			return nil
		},
		OnStop: func(context.Context) error {
			Stopped.Log(appLogger).Msg("")
			return nil
		},
	})
}

func registerRunningStoppingLifecycleEventLoggerHook(lc fx.Lifecycle, logger *zerolog.Logger) {
	appLogger := logging.PackageLogger(logger, pkg)
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
	Initialized.Log(appLogger).Msg("")
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

	// NOTE: the group name must match the constant: comp.CompRegistry
	Comps []*comp.Comp `group:"comp.Registry"`
}

func registerComponents(registry *comp.Registry, comps components, logger *zerolog.Logger, eventRegistry *logging.EventRegistry, errRegistry *err.Registry) error {
	for _, c := range comps.Comps {
		if e := registry.Register(c); e != nil {
			return e
		}
		logCompRegisteredEvent(c, logger)
		eventRegistry.Register(c.EventRegistry.Events()...)
		if e := errRegistry.Register(c.ErrorRegistry.Errs()...); e != nil {
			return e
		}
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
