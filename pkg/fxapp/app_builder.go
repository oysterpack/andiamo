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

	HandleInvokeError(errorHandlers ...func(error)) AppBuilder
}

// NewAppBuilder constructs a new AppBuilder
func NewAppBuilder(desc Desc) AppBuilder {
	return &appBuilder{
		desc:         desc,
		startTimeout: 15 * time.Second,
		stopTimeout:  15 * time.Second,
	}
}

type appBuilder struct {
	desc Desc

	startTimeout time.Duration
	stopTimeout  time.Duration

	constructors        []interface{}
	funcs               []interface{}
	invokeErrorHandlers []func(error)
}

func (a *appBuilder) String() string {
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

	return fmt.Sprintf("AppBuilder{%v, StartTimeout: %s, StopTimeout: %s, Provide: %s, Invoke: %s}",
		a.desc,
		a.startTimeout,
		a.startTimeout,
		funcTypes(a.constructors),
		funcTypes(a.funcs),
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
	app := &app{
		instanceID:   instanceID,
		desc:         a.desc,
		constructors: a.constructors,
		funcs:        a.funcs,

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
	return compOptions
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

func (a *appBuilder) HandleInvokeError(errorHandlers ...func(error)) AppBuilder {
	a.invokeErrorHandlers = append(a.invokeErrorHandlers, errorHandlers...)
	return a
}
