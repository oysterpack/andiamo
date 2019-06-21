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
	"go.uber.org/fx"
	"go.uber.org/multierr"
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
type App interface {
	Options
	LifeCycle

	// Run will start running the application and blocks until the app is shutdown.
	// It waits to receive a SIGINT or SIGTERM signal to shutdown the app.
	Run() error

	// Err returns any app error that occurred during the app's lifetime.
	// Multiple errors may be returned as a single aggregated error.
	Err() error
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

// AppBuilder is used to construct a new App instance.
type AppBuilder interface {
	Build() (App, error)

	SetStartTimeout(timeout time.Duration) AppBuilder
	SetStopTimeout(timeout time.Duration) AppBuilder

	Constructors(constructors ...interface{}) AppBuilder
	Funcs(funcs ...interface{}) AppBuilder
}

// NewAppBuilder constructs a new AppBuilder
func NewAppBuilder(desc Desc) AppBuilder {
	return &app{
		instanceID:   NewInstanceID(),
		desc:         desc,
		startTimeout: 15 * time.Second,
		stopTimeout:  15 * time.Second,
		starting:     make(chan struct{}),
		started:      make(chan struct{}),
		stopping:     make(chan os.Signal, 1),
		stopped:      make(chan os.Signal, 1),
	}
}

type app struct {
	desc       Desc
	instanceID InstanceID

	startTimeout time.Duration
	stopTimeout  time.Duration

	constructors []interface{}
	funcs        []interface{}

	*fx.App
	starting, started chan struct{}
	stopping, stopped chan os.Signal

	err error
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

	return fmt.Sprintf("App{%v, StartTimeout: %s, StopTimeout: %s, Constructors: %s, Funcs: %s, Err: %v}",
		a.desc,
		a.startTimeout,
		a.stopTimeout,
		funcTypes(a.constructors),
		funcTypes(a.funcs),
		a.Err(),
	)
}

func (a *app) ConstructorTypes() []reflect.Type {
	return types(a.constructors)
}

func (a *app) FuncTypes() []reflect.Type {
	return types(a.funcs)
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

func (a *app) Desc() Desc {
	return a.desc
}

// Build tries to construct and initialize a new App instance.
// All of the app's functions are run as part of the app initialization phase.
func (a *app) Build() (App, error) {
	if a.desc == nil {
		a.err = a.appendErr(a.desc.Validate())
	}
	if len(a.constructors) == 0 && len(a.funcs) == 0 {
		a.err = a.appendErr(errors.New("at least 1 functional option is required"))
	}

	compOptions := make([]fx.Option, 0, len(a.constructors)+len(a.funcs))
	for _, f := range a.constructors {
		compOptions = append(compOptions, fx.Provide(f))
	}
	for _, f := range a.funcs {
		compOptions = append(compOptions, fx.Invoke(f))
	}

	a.App = fx.New(
		fx.Provide(func() Desc { return a.desc }),
		fx.Provide(func() InstanceID { return a.instanceID }),
		fx.StartTimeout(a.startTimeout),
		fx.StopTimeout(a.stopTimeout),
		fx.Options(compOptions...),
	)
	a.err = a.appendErr(a.App.Err())

	if a.err != nil {
		return nil, a.err
	}

	return a, nil
}

func (a *app) SetStartTimeout(timeout time.Duration) AppBuilder {
	a.startTimeout = timeout
	return a
}

func (a *app) SetStopTimeout(timeout time.Duration) AppBuilder {
	a.stopTimeout = timeout
	return a
}

func (a *app) Constructors(constructors ...interface{}) AppBuilder {
	a.constructors = append(a.constructors, constructors...)
	return a
}

func (a *app) Funcs(funcs ...interface{}) AppBuilder {
	a.funcs = append(a.funcs, funcs...)
	return a
}

func (a *app) InstanceID() InstanceID {
	return a.instanceID
}

func (a *app) Run() error {
	select {
	case <-a.starting:
		return errors.New("app cannot be run again after it has already been started")
	default:
		// app has not been started yet
	}

	if a.Err() != nil {
		return a.Err()
	}

	startCtx, cancel := context.WithTimeout(context.Background(), a.StartTimeout())
	defer cancel()
	defer close(a.stopped)

	stopChan := a.App.Done()

	close(a.starting)
	if e := a.Start(startCtx); e != nil {
		return a.appendErr(e)
	}
	close(a.started)

	// wait for the app to be signalled to stop
	signal := <-stopChan
	a.stopping <- signal
	close(a.stopping)
	defer func() {
		a.stopped <- signal
	}()

	stopCtx, cancel := context.WithTimeout(context.Background(), a.StopTimeout())
	defer cancel()

	if e := a.Stop(stopCtx); e != nil {
		return a.appendErr(e)
	}

	return nil
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

func (a *app) Err() error {
	return a.err
}

func (a *app) appendErr(err error) error {
	a.err = multierr.Append(a.err, err)
	return a.err
}
