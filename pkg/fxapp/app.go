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
	"context"
	"errors"
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"os"
	"reflect"
	"time"
)

// InstanceID is used to assign an app instance a unique ULID.
// The instance ID can be used to identify the app instance in logs, metrics, events, etc.
type InstanceID ulid.ULID

// NewInstanceID returns a new unique InstanceID
func NewInstanceID() InstanceID {
	return InstanceID(ulidgen.MustNew())
}

// ULID returns the InstanceID's underlying ULID
func (id InstanceID) ULID() ulid.ULID {
	return ulid.ULID(id)
}

func (id InstanceID) String() string {
	return id.ULID().String()
}

// used to implement the fx.ErrorHandler interface
type errorHandler func(err error)

func (f errorHandler) HandleError(err error) {
	f(err)
}

// App represents a functional application container, leveraging fx (https://godoc.org/go.uber.org/fx) as the underlying
// framework. Functional means, the application behavior is defined via functions.
//
// The key is understanding the application life cycle. The application transitions through the following lifecycle states:
//	1. Initialized
//	2. Starting
//	3. Started
//	4. Ready
//	5. Stopping
//	6. Done
//
// When building an application, functions are registered which specify how to:
//  - initialize the application
//  - register services that are bound to the application life cycle, via `fx.Lifecycle` (https://godoc.org/go.uber.org/fx#Lifecycle)
//
// Function arguments are provided via dependency injection by registering provider constructor functions with the application.
// Provider constructor functions are lazily invoked when needed inject function dependencies.
//
// Application Descriptor
//
// The application descriptor is another way to say application metadata (see `Desc`). Every application has the following
// metadata:
// 	- ID - represented as a ULID
//  - name
//	- version
//	- release ID - an application has many versions, but not all versions are released.
//    - can be used to look up additional release artifacts, e.g., release notes, test reports, etc
//
// Application Logging
//
// Zerolog (https://godoc.org/github.com/rs/zerolog) is used as the structured JSON logging framework. A `*zerolog.Logger`
// is automatically provided when building the application and available for dependency injection. The application logger
// context is augmented with application metadata and an event ID, e.g.,
//
//		{"a":"01DE2GCMX5ZSVZXE2RTY7DCB88","r":"01DE2GCMX570BXG6468XBXNXQT","x":"01DE2GCMX5Q9S44S8166JX10WV","z":"01DE30RAEQGQBS0THBCVKVHFSW","t":1561304912,"m":"[Fx] RUNNING"}
//
//		where a -> app ID
//			  r -> app release ID
//			  x -> app instance ID
//			  z -> event ID
//			  t -> timestamp - in Unix time format
//			  m -> message
//
// The zerolog application logger is plugged in as the go standard log, where log events are logged with no level and logged
// using a component logger named 'log' ("c":"log")
//
// Prometheus Metrics
//
// Health Checks
//
// The application provides support to register health checks, which will be automatically run on a schedule.
//  - Health checks are integrated with the readiness and liveliness probes. Any Red health checks will cause the probes to fail.
//  - Health check results are logged
//  - Health checks are integrated with metrics. A gauge is created for each health check, using the health check status
//    as the gauge value.
//  - health check HTTP endpoints:
//    - health check descriptors
//    - health checks
//    - health check results
//
// Readiness Probe
//
// Liveliness Probe
//
// HTTP server support
//
// If HTTPHandler(s) are discovered, i.e., they have been provided, then the app will run an HTTP server.
// HTTP server settings can be provided via an *http.Server (NOTE: http.Server.Handler will be overwritten using
// http handlers that are provided by the app). If no *http.Server is discovered, then the app will automatically
// create an HTTP server with the following settings:
// 	- Addr:              ":8008",
//	- ReadHeaderTimeout: time.Second,
//	- MaxHeaderBytes:    1024,
//
// NOTE: when exposing Prometheus metrics via HTTP will provide an HTTP handler, and thus cause the HTTP server to run.
//
// Automatically Provided
//	- Desc
//	- InstanceID
//	- *zerolog.Logger
//	- fx.Lifecycle - for components to use to bind to the app lifecycle
//	- fx.Shutdowner - used to trigger app shutdown
//	- fx.Dotgraph - contains a DOT language visualization of the app dependency graph
//	- ReadinessWaitGroup - the readiness probe uses the ReadinessWaitGroup to know when the application is ready to serve requests
//	- prometheus.Gatherer
//	- prometheus.Registerer
//  - health.Registry
//  - health.Scheduler
type App interface {
	Options
	LifeCycle

	// Run will start running the application and blocks until the app is shutdown.
	// It waits to receive a SIGINT or SIGTERM signal to shutdown the app.
	Run() error

	// StopAsync signals the app to shutdown. This method does not block, i.e., application shutdown occurs async.
	//
	// StopAsync can only be called after the app has been started - otherwise an error is returned.
	Shutdown() error
}

// LifeCycle defines the application lifecycle.
type LifeCycle interface {
	// Starting signals that the app is starting.
	// Closing the channel is the signal.
	Starting() <-chan struct{}
	// Started signals that the app has fully started
	Started() <-chan struct{}
	// Ready means the app is ready to serve requests
	Ready() <-chan struct{}
	// Stopping signals that app is stopping.
	// The channel is closed after the stop signal is sent.
	Stopping() <-chan os.Signal
	// Done signals that the app has shutdown.
	// The channel is closed after the stop signal is sent.
	// If the app fails to startup, then the channel is simply closed, i.e., no stop signal will be sent on the channel.
	Done() <-chan os.Signal
}

// Options represent application options that were used to configure and build app.
type Options interface {
	// Desc returns the app descriptor
	Desc() Desc

	// InstanceID returns the app unique instance ID
	InstanceID() InstanceID

	// StartTimeout returns the app start timeout. If the app takes longer than the specified timeout, then the app will
	// fail to run.
	StartTimeout() time.Duration
	// StopTimeout returns the app shutdown timeout. If the app takes longer than the specified timeout, then the app shutdown
	// will be aborted.
	StopTimeout() time.Duration

	// ConstructorTypes returns the registered constructor types
	ConstructorTypes() []reflect.Type
	// FuncTypes returns the registered function types
	FuncTypes() []reflect.Type
}

type app struct {
	desc       Desc
	instanceID InstanceID

	constructors []interface{}
	funcs        []interface{}

	startErrorHandlers, stopErrorHandlers []func(error)

	*fx.App
	fx.Shutdowner
	starting, started chan struct{}
	readiness         ReadinessWaitGroup
	stopping, stopped chan os.Signal

	logger *zerolog.Logger
}

func (a *app) String() string {
	funcTypes := func(funcs []interface{}) string {
		if len(funcs) == 0 {
			return "[]"
		}
		s := new(bytes.Buffer)
		s.WriteString("[")
		s.WriteString(reflect.TypeOf(funcs[0]).String())
		for i := 1; i < len(funcs); i++ {
			s.WriteString("|")
			s.WriteString(reflect.TypeOf(funcs[i]).String())
		}

		s.WriteString("]")
		return s.String()
	}

	switch {
	case a.App != nil:
		return fmt.Sprintf("App{%v, StartTimeout: %s, StopTimeout: %s, Provide: %s, Invoke: %s, Err: %v}",
			a.desc,
			a.StartTimeout(),
			a.StopTimeout(),
			funcTypes(a.constructors),
			funcTypes(a.funcs),
			a.Err(),
		)
	default:
		return fmt.Sprintf("App{%v, StartTimeout: %s, StopTimeout: %s, Provide: %s, Invoke: %s, Err: %v}",
			a.desc,
			fx.DefaultTimeout,
			fx.DefaultTimeout,
			funcTypes(a.constructors),
			funcTypes(a.funcs),
			nil,
		)
	}
}

func (a *app) Desc() Desc {
	return a.desc
}

func (a *app) InstanceID() InstanceID {
	return a.instanceID
}

func types(values []interface{}) []reflect.Type {
	if len(values) == 0 {
		return nil
	}
	valueTypes := make([]reflect.Type, 0, len(values))
	for _, value := range values {
		valueTypes = append(valueTypes, reflect.TypeOf(value))
	}

	return valueTypes
}

func (a *app) ConstructorTypes() []reflect.Type {
	return types(a.constructors)
}

func (a *app) FuncTypes() []reflect.Type {
	return types(a.funcs)
}

func (a *app) Run() error {
	select {
	case <-a.starting:
		return errors.New("app cannot be run again after it has already been started")
	default:
		// app has not been started yet
	}
	a.logAppStarting()

	startCtx, cancel := context.WithTimeout(context.Background(), a.StartTimeout())
	defer cancel()
	defer close(a.stopped)

	stopChan := a.App.Done()

	close(a.starting)
	startingTime := time.Now()
	if e := a.Start(startCtx); e != nil {
		return a.handleStartError(e)
	}
	a.logAppStarted(time.Since(startingTime))
	close(a.started)
	a.readiness.Done() // the app has started

	// wait for the app to be ready to service requests
	select {
	case <-a.readiness.Ready():
		a.logAppReady()
		return a.shutdown(<-stopChan) // shutdown on stop signal
	case signal := <-stopChan: // wait for the app to be signalled to stop
		return a.shutdown(signal)
	}
}

func (a *app) shutdown(signal os.Signal) error {
	a.stopping <- signal
	close(a.stopping)
	defer func() {
		a.stopped <- signal
	}()

	a.logAppStopping()

	stopCtx, cancel := context.WithTimeout(context.Background(), a.StopTimeout())
	defer cancel()
	stoppingTime := time.Now()
	defer func() { a.logAppStopped(time.Since(stoppingTime)) }()
	if e := a.Stop(stopCtx); e != nil {
		return a.handleStopError(e)
	}
	return nil
}

func (a *app) handleStartError(err error) error {
	for _, f := range a.startErrorHandlers {
		f(err)
	}
	return err
}

func (a *app) handleStopError(err error) error {
	for _, f := range a.stopErrorHandlers {
		f(err)
	}
	return err
}

func (a *app) Starting() <-chan struct{} {
	return a.starting
}

func (a *app) Started() <-chan struct{} {
	return a.started
}

func (a *app) Ready() <-chan struct{} {
	return a.readiness.Ready()
}

func (a *app) Stopping() <-chan os.Signal {
	return a.stopping
}

func (a *app) Done() <-chan os.Signal {
	return a.stopped
}

func (a *app) Shutdown() error {
	select {
	case <-a.started:
		return a.Shutdowner.Shutdown()
	default:
		return errors.New("app can only be shutdown after it has started")
	}

}

func (a *app) logAppInitialized() {
	logEvent := InitializedEventID.NewLogEventer(a.logger, zerolog.NoLevel)
	logEvent(AppInitialized{App: a}, "app initialized")
}

func (a *app) logAppStarting() {
	logEvent := StartingEventID.NewLogEventer(a.logger, zerolog.NoLevel)
	logEvent(nil, "app starting")
}

func (a *app) logAppStarted(startupTime time.Duration) {
	logEvent := StartedEventID.NewLogEventer(a.logger, zerolog.NoLevel)
	logEvent(AppStarted{startupTime}, "app started")
}

func (a *app) logAppReady() {
	logEvent := ReadyEventID.NewLogEventer(a.logger, zerolog.NoLevel)
	logEvent(nil, "app is ready to service requests")
}

func (a *app) logAppStopping() {
	logEvent := StoppingEventID.NewLogEventer(a.logger, zerolog.NoLevel)
	logEvent(nil, "app stopping")
}

func (a *app) logAppStopped(shutdownDuration time.Duration) {
	logEvent := StoppedEventID.NewLogEventer(a.logger, zerolog.NoLevel)
	logEvent(AppStopped{shutdownDuration}, "app stopped")
}
