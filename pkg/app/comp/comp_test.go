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

package comp_test

import (
	"context"
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"go.uber.org/fx"
	"reflect"
	"testing"
)

type components struct {
	fx.In

	Comps []*comp.Comp `group:"comp.Registry"`
}

func TestComp(t *testing.T) {
	// define some app options
	type Greeter func() string
	type ProvideGreeter func() Greeter
	type LogGreeting func(Greeter)

	var (
		// option descriptors
		GreeterDesc     = option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))
		LogGreetingDesc = option.NewDesc(option.Invoke, reflect.TypeOf(LogGreeting(nil)))

		// component descriptor
		FooDesc = comp.NewDescBuilder().
			ID("01DCVEMPTQRNED7XDKGB30V2CF").
			Name("Foo").
			Version("0.1.0").
			Package(Package).
			// Specify the component's option descriptors
			Options(GreeterDesc, LogGreetingDesc).
			MustBuild()

		provideGreeter ProvideGreeter = func() Greeter {
			return func() string {
				return "greetings"
			}
		}
		logGreeting LogGreeting = func(greeter Greeter) {
			t.Log(greeter())
		}
	)

	// Given a component
	FooComp := FooDesc.MustNewComp(
		LogGreetingDesc.NewOption(logGreeting),
		GreeterDesc.NewOption(provideGreeter),
	)

	t.Log(FooComp)

	fxapp := fx.New(
		FooComp.FxOptions(),
		fx.Invoke(func(comps components) {
			// components should self-register with the app
			found := false
			for _, c := range comps.Comps {
				if c.ID == FooComp.ID {
					t.Logf("component was found: %v", c)
					found = true
					break
				}
			}
			if !found {
				t.Error("component was not found")
			}
		}),
	)
	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Fatal(e)
	}
}

func TestComp_Logger(t *testing.T) {
	logger := apptest.NewAppTestLogger()

	type Greeter func() string
	type ProvideGreeter func() Greeter
	type LogGreeting func(Greeter)

	var (
		// option descriptors
		GreeterDesc     = option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))
		LogGreetingDesc = option.NewDesc(option.Invoke, reflect.TypeOf(LogGreeting(nil)))

		// component descriptor
		FooDesc = comp.NewDescBuilder().
			ID("01DCVEMPTQRNED7XDKGB30V2CF").
			Name("Foo").
			Version("0.1.0").
			Package(Package).
			// Specify the component's option descriptors
			Options(GreeterDesc, LogGreetingDesc).
			MustBuild()

		provideGreeter ProvideGreeter = func() Greeter {
			return func() string {
				return "greetings"
			}
		}
		logGreeting LogGreeting = func(greeter Greeter) {
			t.Log(greeter())
		}
	)

	// Given a component
	FooComp := FooDesc.MustNewComp(
		LogGreetingDesc.NewOption(logGreeting),
		GreeterDesc.NewOption(provideGreeter),
	)
	// When a comp logger is created
	compLogger := FooComp.Logger(logger.Logger)
	// And a comp event is logged
	compLogger.Info().Msg("")

	// parse the JSON log event
	var logEvent apptest.LogEvent
	t.Log(logger.Buf.String())
	if e := json.Unmarshal([]byte(logger.Buf.String()), &logEvent); e != nil {
		t.Fatal(e)
	}
	// Then the log event contains the package field
	if logEvent.Package != string(Package) {
		t.Error("package field is missing from the log event")
	}
	// And the log event contains the component field
	if logEvent.Component != FooComp.Name {
		t.Error("component field is missing from the log event")
	}
}

func NewComp(t *testing.T, id, name, version string) *comp.Comp {
	// define some app options
	type Greeter func() string
	type ProvideGreeter func() Greeter
	type LogGreeting func(Greeter)

	var (
		// option descriptors
		GreeterDesc     = option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))
		LogGreetingDesc = option.NewDesc(option.Invoke, reflect.TypeOf(LogGreeting(nil)))

		// component descriptor
		FooDesc = comp.NewDescBuilder().
			ID(id).
			Name(name).
			Version(version).
			Package(Package).
			// Specify the component's option descriptors
			Options(GreeterDesc, LogGreetingDesc).
			MustBuild()

		provideGreeter ProvideGreeter = func() Greeter {
			return func() string {
				return "greetings"
			}
		}
		logGreeting LogGreeting = func(greeter Greeter) {
			t.Log(greeter())
		}
	)

	// Given a component
	return FooDesc.MustNewComp(
		LogGreetingDesc.NewOption(logGreeting),
		GreeterDesc.NewOption(provideGreeter),
	)
}
