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

// App represents an application container.
//
// Dependency injection is provided via registered constructors.
// Application workloads are run via registered functions. At least 1 function must be registered.
//
// Application lifecycle states are:
// - Initialized
// - Starting
// - Running
// - Stopping
// - Done
//
// The following are automatically provided for the app:
// - Desc - app descriptor
// - InstanceID - app instance ID
// - *zerolog.Logger - for application logging
//   - the logger is used as the go standard log - logged with Debug level
//   - fx log events are logged with no level, i.e., which means they are always logged
//   - the log event timestamp is in Unix time format
//   - the logger context is augmented with the app ID, release ID, and instance ID. For example:
//
//     {"a":"01DE2GCMX5ZSVZXE2RTY7DCB88","r":"01DE2GCMX570BXG6468XBXNXQT","x":"01DE2GCMX5Q9S44S8166JX10WV","z":"01DE30RAEQGQBS0THBCVKVHFSW","t":1561304912,"m":"[Fx] RUNNING"}
//
//     where a -> app ID
//           r -> app release ID
//           x -> app instance ID
//           z -> event ID
//           t -> timestamp
//           m -> message
//
// - fx.Lifecycle - for components to use to bind to the app lifecycle
// - fx.Shutdowner - used to trigger app shutdown
// - fx.Dotgraph - contains a DOT language visualization of the app dependency graph
type App interface {
	Options
	LifeCycle

	// Run will start running the application and blocks until the app is shutdown.
	// It waits to receive a SIGINT or SIGTERM signal to shutdown the app.
	Run() error

	// Shutdown signals the app to shutdown. This method does not block, i.e., application shutdown occurs async.
	//
	// Shutdown can only be called after the app has been started - otherwise an error is returned.
	Shutdown() error
}

// LifeCycle defines the application lifecycle.
type LifeCycle interface {
	// Starting signals that the app is starting.
	// Closing the channel is the signal.
	Starting() <-chan struct{}
	// Started signals that the app has fully started
	Started() <-chan struct{}
	// Stopping signals that app is stopping.
	// The channel is closed after the stop signal is sent.
	Stopping() <-chan os.Signal
	// Done signals that the app has shutdown.
	// The channel is closed after the stop signal is sent.
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

	// wait for the app to be signalled to stop
	signal := <-stopChan
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
	logEvent := InitializedEventID.NewLogEventFunc(a.logger, zerolog.NoLevel)
	logEvent(AppInitialized{App: a}, "app initialized")
}

func (a *app) logAppStarting() {
	logEvent := StartingEventID.NewLogEventFunc(a.logger, zerolog.NoLevel)
	logEvent(nil, "app starting")
}

func (a *app) logAppStarted(startupTime time.Duration) {
	logEvent := StartedEventID.NewLogEventFunc(a.logger, zerolog.NoLevel)
	logEvent(AppStarted{startupTime}, "app started")
}

func (a *app) logAppStopping() {
	logEvent := StoppingEventID.NewLogEventFunc(a.logger, zerolog.NoLevel)
	logEvent(nil, "app stopping")
}

func (a *app) logAppStopped(shutdownDuration time.Duration) {
	logEvent := StoppedEventID.NewLogEventFunc(a.logger, zerolog.NoLevel)
	logEvent(AppStopped{shutdownDuration}, "app stopped")
}
