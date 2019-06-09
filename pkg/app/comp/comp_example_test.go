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
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"log"
	"math/rand"
	"reflect"
)

// declare component functionality
type Foo func() int
type Bar func(int) int

// what does the component provide
type ProvideFoo func() Foo
type ProvideBar func() Bar

// what does the component do
type InvokeFooBar func(lc fx.Lifecycle, foo Foo, bar Bar, logger *zerolog.Logger, shutdowner fx.Shutdowner) error

var (
	ProvideFooOption   = option.NewDesc(option.Provide, reflect.TypeOf(ProvideFoo(nil)))
	ProvideBarOption   = option.NewDesc(option.Provide, reflect.TypeOf(ProvideBar(nil)))
	InvokeFooBarOption = option.NewDesc(option.Invoke, reflect.TypeOf(InvokeFooBar(nil)))
)

// component descriptors
var (
	FooBar = comp.MustNewDesc(
		comp.ID("01DCY6DYT9CMQCAY89W42HWBGG"),
		comp.Name("foobar"),
		comp.Version("0.1.0"),
		app.Package("github.com/oysterpack/partire-k8s/pkg/foobar"),
		// options declare the component functionality - options represent the component blueprint
		ProvideFooOption,
		ProvideBarOption,
		InvokeFooBarOption,
	)
)

func Example() {

	// build the component
	foobar := FooBar.MustNewComp(
		ProvideFooOption.NewOption(func() Foo {
			return func() int {
				return rand.Int()
			}
		}),
		ProvideBarOption.NewOption(func() Bar {
			return func(i int) int {
				return i + 1
			}
		}),
		InvokeFooBarOption.NewOption(func(lc fx.Lifecycle, foo Foo, bar Bar, logger *zerolog.Logger, shutdowner fx.Shutdowner) error {
			compLogger := FooBar.Logger(logger)
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					compLogger.Info().Int("result", bar(foo())).Msg("started")
					if e := shutdowner.Shutdown(); e != nil {
						compLogger.Error().Err(e).Msg("")
					}
					return nil
				},
				OnStop: func(_ context.Context) error {
					compLogger.Info().Msg("stopped")
					return nil
				},
			})
			return nil
		}),
	)

	apptest.InitEnv()
	fxapp := appfx.New(foobar.AppOptions()...)
	go func() {
		if e := fxapp.Run(); e != nil {
			log.Fatal(e)
		}
	}()

	<-fxapp.Done()

	// Output:
	//
}
