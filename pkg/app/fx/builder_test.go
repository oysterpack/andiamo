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
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"reflect"
	"testing"
	"time"
)

func TestAppBuilder(t *testing.T) {
	// The minimum required to build an app is:
	// - app.Desc
	// - at least 1 option specified
	t.Run("run app with minimal required and no env vars", testRunAppWithMinimalRequired)
	t.Run("run app with minimal required and app.Desc loaded from env", testRunAppWithMinimalRequiredAppDescFromEnv)
	// Given no app.Desc
	// Then the app will fail to build
	t.Run("build app with no app.Desc specified and env vars not set", testBuildAppWithNoAppDescAndEnvVarsNotSet)
	// Given an invalid app.Desc
	// Then the app will fail to build
	t.Run("build app with invalid app.Desc", testBuildAppWithInvalidAppDesc)

	t.Run("build app using components", testBuildAppUsingComps)
	t.Run("build app using duplicate components", testBuildAppUsingDupComps)

	t.Run("build app specifying timeouts", buildAppSpecifyingTimeouts)
}

func buildAppSpecifyingTimeouts(t *testing.T) {
	// Given the app.Desc env vars are set
	apptest.InitEnv()

	startTimeout := 30 * time.Second
	stopTimeout := time.Minute

	fxapp, e := appfx.NewAppBuilder().
		StartTimeout(startTimeout).
		StopTimeout(stopTimeout).
		Options(fx.Invoke(func(shutdowner fx.Shutdowner) { shutdowner.Shutdown() })).
		Build()

	if e != nil {
		t.Fatal(e)
	}

	if fxapp.StartTimeout() != startTimeout {
		t.Errorf("StartTimeout did not match: %v", fxapp.StartTimeout())
	}
	if fxapp.StopTimeout() != stopTimeout {
		t.Errorf("StopTimeout did not match: %v", fxapp.StopTimeout())
	}
}

func testBuildAppUsingDupComps(t *testing.T) {
	type Command func()

	var (
		CommandOptionDesc = option.NewDesc(option.Invoke, reflect.TypeOf(Command(nil)))

		FooCompDesc = comp.MustNewDesc(
			comp.ID(ulidgen.MustNew().String()),
			comp.Name("foo"),
			comp.Version("0.0.1"),
			app.Package("github.com/oysterpack/partire-k8s/pkg/foo"),
			CommandOptionDesc,
		)

		FooComp = FooCompDesc.MustNewComp(CommandOptionDesc.NewOption(func() { t.Log("foo: ciao") }))

		BarCompDesc = comp.MustNewDesc(
			comp.ID(ulidgen.MustNew().String()),
			comp.Name("bar"),
			comp.Version("0.0.1"),
			app.Package("github.com/oysterpack/partire-k8s/pkg/bar"),
			CommandOptionDesc,
		)

		BarComp = BarCompDesc.MustNewComp(CommandOptionDesc.NewOption(func() { t.Log("bar: ciao") }))
	)

	// Given the app.Desc env vars are set
	apptest.InitEnv()

	// When the app is built, app.Desc will be loaded from the env
	_, e := appfx.NewAppBuilder().
		Comps(
			FooComp,
			BarComp,
			FooComp, // dup
		).
		Build()

	// Then app should fail to build
	if e == nil {
		t.Fatal("app should have failed to build because duplicate components were specified")
	}
	t.Log(e)
}

func testBuildAppWithInvalidAppDesc(t *testing.T) {
	var desc app.Desc
	apptest.ClearAppEnvSettings()

	_, e := appfx.NewAppBuilder().
		// app.Desc is required
		AppDesc(desc).
		// at least 1 option is required
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
		t.Fatal("App should have failed to build because app.Desc is not valid")
	}

	t.Log(e)

	buildErr := e.(*err.Instance)
	if buildErr.SrcID != appfx.InvalidDescErr.SrcID {
		t.Errorf("unexpected error instance: %v", appfx.InvalidDescErr.SrcID)
	}

}

func testBuildAppUsingComps(t *testing.T) {
	type Command func()

	var (
		CommandOptionDesc = option.NewDesc(option.Invoke, reflect.TypeOf(Command(nil)))

		FooCompDesc = comp.MustNewDesc(
			comp.ID(ulidgen.MustNew().String()),
			comp.Name("foo"),
			comp.Version("0.0.1"),
			app.Package("github.com/oysterpack/partire-k8s/pkg/foo"),
			CommandOptionDesc,
		)

		FooComp = FooCompDesc.MustNewComp(CommandOptionDesc.NewOption(func() { t.Log("foo: ciao") }))

		BarCompDesc = comp.MustNewDesc(
			comp.ID(ulidgen.MustNew().String()),
			comp.Name("bar"),
			comp.Version("0.0.1"),
			app.Package("github.com/oysterpack/partire-k8s/pkg/bar"),
			CommandOptionDesc,
		)

		BarComp = BarCompDesc.MustNewComp(CommandOptionDesc.NewOption(func() { t.Log("bar: ciao") }))
	)

	// Given the app.Desc env vars are set
	apptest.InitEnv()

	// When the app is built, app.Desc will be loaded from the env
	fxapp, e := appfx.NewAppBuilder().
		// And components are plugged in
		Comps(
			FooComp,
			BarComp,
		).
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
		// at least 1 option is required
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

	if fxapp.StartTimeout() != fx.DefaultTimeout && fxapp.StopTimeout() != fx.DefaultTimeout {
		t.Errorf("app start/stop timeouts do not match the expected defaults: %v : %v", fxapp.StartTimeout(), fxapp.StopTimeout())
	}

	go func() {
		if e := fxapp.Run(); e != nil {
			t.Error(e)
		}
	}()

	<-fxapp.Done()
}
