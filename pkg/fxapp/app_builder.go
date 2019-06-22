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
	"go.uber.org/fx"
	"go.uber.org/multierr"
	"os"
	"reflect"
	"time"
)

// AppBuilder is used to construct a new App instance.
type AppBuilder interface {
	Build() (App, error)

	SetStartTimeout(timeout time.Duration) AppBuilder
	SetStopTimeout(timeout time.Duration) AppBuilder

	Provide(constructors ...interface{}) AppBuilder
	Invoke(funcs ...interface{}) AppBuilder

	// Populate sets targets with values from the dependency injection container during application initialization.
	// All targets must be pointers to the values that must be populated.
	// Pointers to structs that embed fx.In are supported, which can be used to populate multiple values in a struct.
	Populate(targets ...interface{}) AppBuilder

	Logger(logger func(msg string, args ...interface{})) AppBuilder

	HandleInvokeError(errorHandlers ...func(error)) AppBuilder
	HandleStartupError(errorHandlers ...func(error)) AppBuilder
	HandleShutdownError(errorHandlers ...func(error)) AppBuilder
	// HandleError will handle any app error, i.e., app function invoke errors, app startup errors, and app shutdown errors.
	HandleError(errorHandlers ...func(error)) AppBuilder
}

// NewAppBuilder constructs a new AppBuilder
func NewAppBuilder(desc Desc) AppBuilder {
	return &appBuilder{
		desc:         desc,
		startTimeout: fx.DefaultTimeout,
		stopTimeout:  fx.DefaultTimeout,
	}
}

type appBuilder struct {
	desc Desc

	startTimeout time.Duration
	stopTimeout  time.Duration

	constructors    []interface{}
	funcs           []interface{}
	populateTargets []interface{}

	logger func(msg string, args ...interface{})

	invokeErrorHandlers, startErrorHandlers, stopErrorHandlers []func(error)
}

func (a *appBuilder) String() string {
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

	return fmt.Sprintf("AppBuilder{%v, StartTimeout: %s, StopTimeout: %s, Provide: %s, Invoke: %s, Populate: %s, InvokeErrHandlerCount: %d, StartErrHandlerCount: %d}",
		a.desc,
		a.startTimeout,
		a.startTimeout,
		types(a.constructors),
		types(a.funcs),
		types(a.populateTargets),
		len(a.invokeErrorHandlers),
		len(a.startErrorHandlers),
	)
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

// Build tries to construct and initialize a new App instance.
// All of the app's functions are run as part of the app initialization phase.
func (a *appBuilder) Build() (App, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}

	instanceID := NewInstanceID()
	var shutdowner fx.Shutdowner
	a.populateTargets = append(a.populateTargets, &shutdowner)
	app := &app{
		instanceID:   instanceID,
		desc:         a.desc,
		constructors: a.constructors,
		funcs:        a.funcs,

		startErrorHandlers: a.startErrorHandlers,
		stopErrorHandlers:  a.stopErrorHandlers,

		starting: make(chan struct{}),
		started:  make(chan struct{}),
		stopping: make(chan os.Signal, 1),
		stopped:  make(chan os.Signal, 1),

		App: fx.New(
			fx.Provide(func() Desc { return a.desc }),
			fx.Provide(func() InstanceID { return instanceID }),
			fx.StartTimeout(a.startTimeout),
			fx.StopTimeout(a.stopTimeout),
			fx.Options(a.buildOptions()...),
		),

		Shutdowner: shutdowner,
	}

	if err := app.Err(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *appBuilder) validate() error {
	var err error
	if a.desc == nil {
		err = multierr.Append(err, a.desc.Validate())
	}
	if len(a.constructors) == 0 && len(a.funcs) == 0 {
		err = multierr.Append(err, errors.New("at least 1 functional option is required"))
	}
	return err
}

func (a *appBuilder) buildOptions() []fx.Option {
	compOptions := make([]fx.Option, 0, len(a.constructors)+len(a.funcs))
	for _, f := range a.constructors {
		compOptions = append(compOptions, fx.Provide(f))
	}
	for _, f := range a.funcs {
		compOptions = append(compOptions, fx.Invoke(f))
	}
	for _, f := range a.invokeErrorHandlers {
		compOptions = append(compOptions, fx.ErrorHook(errorHandler(f)))
	}
	for _, target := range a.populateTargets {
		compOptions = append(compOptions, fx.Populate(target))
	}
	if a.logger != nil {
		compOptions = append(compOptions, fx.Logger(logger(a.logger)))
	}
	return compOptions
}

type logger func(msg string, args ...interface{})

func (l logger) Printf(msg string, args ...interface{}) {
	l(msg, args...)
}

func (a *appBuilder) SetStartTimeout(timeout time.Duration) AppBuilder {
	a.startTimeout = timeout
	return a
}

func (a *appBuilder) SetStopTimeout(timeout time.Duration) AppBuilder {
	a.stopTimeout = timeout
	return a
}

func (a *appBuilder) Provide(constructors ...interface{}) AppBuilder {
	a.constructors = append(a.constructors, constructors...)
	return a
}

func (a *appBuilder) Invoke(funcs ...interface{}) AppBuilder {
	a.funcs = append(a.funcs, funcs...)
	return a
}

func (a *appBuilder) Populate(targets ...interface{}) AppBuilder {
	a.populateTargets = append(a.populateTargets, targets...)
	return a
}

func (a *appBuilder) HandleInvokeError(errorHandlers ...func(error)) AppBuilder {
	a.invokeErrorHandlers = append(a.invokeErrorHandlers, errorHandlers...)
	return a
}

func (a *appBuilder) HandleStartupError(errorHandlers ...func(error)) AppBuilder {
	a.startErrorHandlers = append(a.startErrorHandlers, errorHandlers...)
	return a
}

func (a *appBuilder) HandleShutdownError(errorHandlers ...func(error)) AppBuilder {
	a.stopErrorHandlers = append(a.stopErrorHandlers, errorHandlers...)
	return a
}

func (a *appBuilder) HandleError(errorHandlers ...func(error)) AppBuilder {
	a.HandleInvokeError(errorHandlers...)
	a.HandleStartupError(errorHandlers...)
	a.HandleShutdownError(errorHandlers...)
	return a
}

func (a *appBuilder) Logger(logger func(msg string, args ...interface{})) AppBuilder {
	a.logger = logger
	return a
}
