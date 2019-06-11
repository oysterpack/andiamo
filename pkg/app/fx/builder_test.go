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

package fx_test

import (
	"context"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"reflect"
	"testing"
)

func TestAppBuilder(t *testing.T) {
	t.Run("run app with minimal required and no env vars", testRunAppWithMinimalRequired)
	t.Run("run app with minimal required and app.Desc loaded from env", testRunAppWithMinimalRequiredAppDescFromEnv)
	t.Run("build app with no app.Desc specified and env vars not set", testBuildAppWithNoAppDescAndEnvVarsNotSet)

	t.Run("build app using components", testBuildAppUsingComps)
}

func testBuildAppUsingComps(t *testing.T) {
	type Command func()

	var (
		CommandOptionDesc = option.NewDesc(option.Invoke, reflect.TypeOf(Command(nil)))
		CommandOption     = CommandOptionDesc.NewOption(func() { t.Log("ciao") })

		FooCompDesc = comp.MustNewDesc(
			comp.ID(ulidgen.MustNew().String()),
			comp.Name("Foo"),
			comp.Version("0.0.1"),
			app.Package("github.com/oysterpack/partire-k8s/pkg/foo"),
			CommandOptionDesc,
		)

		FooComp = FooCompDesc.MustNewComp(CommandOption)
	)

	// Given the app.Desc env vars are set
	apptest.InitEnv()

	// When the app is built, app.Desc will be loaded from the env
	fxapp, e := appfx.NewAppBuilder().
		// And a component is plugged in
		Comps(FooComp).
		Build()

	if e != nil {
		t.Fatal(e)
	}

	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}

	if e := fxapp.Stop(context.Background()); e != nil {
		t.Fatal(e)
	}

	// Then the app runs successfully
}

func testBuildAppWithNoAppDescAndEnvVarsNotSet(t *testing.T) {
	apptest.ClearAppEnvSettings()

	// When the app is built, app.Desc will be loaded from the env
	_, e := appfx.NewAppBuilder().
		Options(fx.Invoke(func(lc fx.Lifecycle, l *zerolog.Logger, s fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					l.Info().Msg("shutting down ...")
					if e := s.Shutdown(); e != nil {
						t.Error(e)
					}
					return nil
				},
			})
		})).
		Build()

	if e == nil {
		t.Fatal("App should have failed to build because the app.Desc env vars were not set")
	}

	t.Log(e)
}

func testRunAppWithMinimalRequiredAppDescFromEnv(t *testing.T) {
	// Given the app.Desc env vars are set
	apptest.InitEnv()

	// When the app is built, app.Desc will be loaded from the env
	fxapp, e := appfx.NewAppBuilder().
		Options(fx.Invoke(func(lc fx.Lifecycle, l *zerolog.Logger, s fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					l.Info().Msg("shutting down ...")
					if e := s.Shutdown(); e != nil {
						t.Error(e)
					}
					return nil
				},
			})
		})).
		Build()

	if e != nil {
		t.Fatal(e)
	}

	go func() {
		if e := fxapp.Run(); e != nil {
			t.Error(e)
		}
	}()

	<-fxapp.Done()
}

func testRunAppWithMinimalRequired(t *testing.T) {
	desc := apptest.InitEnv()
	apptest.ClearAppEnvSettings()

	fxapp, e := appfx.NewAppBuilder().
		// app.Desc is required
		AppDesc(desc).
		Options(fx.Invoke(func(lc fx.Lifecycle, l *zerolog.Logger, s fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					l.Info().Msg("shutting down ...")
					if e := s.Shutdown(); e != nil {
						t.Error(e)
					}
					return nil
				},
			})
		})).
		Build()

	if e != nil {
		t.Fatal(e)
	}

	go func() {
		if e := fxapp.Run(); e != nil {
			t.Error(e)
		}
	}()

	<-fxapp.Done()
}
