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
	"encoding/json"
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
	"strings"
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

	t.Run("build app using components", testBuildAppUsingComps)
	t.Run("build app specifying timeouts", buildAppSpecifyingTimeouts)

	t.Run("build app with log writer", buildAppWithLogWriter)
	t.Run("build with global log level", buildWithGlobalLogLevel)
	t.Run("build with log sampling disabled", buildWithLogSamplingDisabled)

	// error test cases
	t.Run("build app with invalid app.Desc", testBuildAppWithInvalidAppDesc)
	t.Run("build app using duplicate components", testBuildAppUsingDupComps)
	t.Run("build app with no options", testBuildAppWithNoOptions)
	t.Run("build app with an invalid timeout env setting", buildAppWithInvalidTimeoutEnvSetting)
}

func buildAppWithInvalidTimeoutEnvSetting(t *testing.T) {
	apptest.InitEnv()

	apptest.Setenv(apptest.StartTimeout, "INVALID")

	_, e := appfx.NewAppBuilder().
		Options(fx.Invoke(func() {})).
		Build()
	if e == nil {
		t.Fatal("app should have failed to build because no options or components were specified")
	}

	appErr := e.(*err.Instance)
	if appErr.SrcID != appfx.InvalidTimeoutsErr.SrcID {
		t.Errorf("unexpected error: %v", appErr)
	}
}

func testBuildAppWithNoOptions(t *testing.T) {
	apptest.InitEnv()

	_, e := appfx.NewAppBuilder().Build()
	if e == nil {
		t.Fatal("app should have failed to build because no options or components were specified")
	}

	appErr := e.(*err.Instance)
	if appErr.SrcID != appfx.OptionsRequiredErr.SrcID {
		t.Errorf("unexpected error: %v", appErr)
	}
}

func buildWithLogSamplingDisabled(t *testing.T) {
	apptest.InitEnv()

	buf := new(strings.Builder)

	key, value := ulidgen.MustNew(), ulidgen.MustNew()

	fxapp, e := (appfx.NewAppBuilder().
		LogWriter(buf).
		DisableLogSampling().
		Options(fx.Invoke(func(logger *zerolog.Logger) {
			logger.Info().
				Str(key.String(), value.String()).
				Msgf("buildAppWithLogWriter: %v", value)
		})).
		Build())

	if e != nil {
		t.Error(e)
	}

	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Fatal(e)
	}

	logEvents := buf.String()
	t.Log(logEvents)
	if len(logEvents) == 0 {
		t.Error("no events were logged")
	}

	var foundLogEvent bool
	var logEvent apptest.LogEvent
	for _, line := range strings.Split(logEvents, "\n") {
		if len(line) == 0 {
			break
		}
		if e := json.Unmarshal([]byte(line), &logEvent); e != nil {
			t.Error(e)
		} else {
			if strings.Contains(logEvent.Message, value.String()) {
				foundLogEvent = true
			}
		}
	}
	if !foundLogEvent {
		t.Error("buildAppWithLogWriter log event was not found")
	}
}

func buildWithGlobalLogLevel(t *testing.T) {
	apptest.InitEnv()

	buf := new(strings.Builder)

	key, value := ulidgen.MustNew(), ulidgen.MustNew()

	fxapp, e := (appfx.NewAppBuilder().
		LogWriter(buf).
		GlobalLogLevel(zerolog.WarnLevel).
		Options(fx.Invoke(func(logger *zerolog.Logger) {
			logger.Warn().
				Str(key.String(), value.String()).
				Msgf("buildAppWithLogWriter: %v", value)
			logger.Info().
				Str(key.String(), value.String()).
				Msgf("buildAppWithLogWriter: %v", value)
		})).
		Build())

	if e != nil {
		t.Error(e)
	}

	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Fatal(e)
	}

	logEvents := buf.String()
	t.Log(logEvents)
	if len(logEvents) == 0 {
		t.Error("no events were logged")
	}

	var logEventCount int
	for _, line := range strings.Split(logEvents, "\n") {
		if len(line) == 0 {
			break
		}
		var logEvent apptest.LogEvent
		if e := json.Unmarshal([]byte(line), &logEvent); e != nil {
			t.Error(e)
		} else {
			if strings.Contains(logEvent.Message, value.String()) {
				t.Logf("matched: %v : %v", logEvent.Message, line)
				logEventCount += 1
			}
		}
	}
	if logEventCount != 1 {
		t.Error("only the warn event should have been logged")
	}
}

func buildAppWithLogWriter(t *testing.T) {
	apptest.InitEnv()

	buf := new(strings.Builder)

	key, value := ulidgen.MustNew(), ulidgen.MustNew()

	fxapp, e := (appfx.NewAppBuilder().
		LogWriter(buf).
		Options(fx.Invoke(func(logger *zerolog.Logger) {
			logger.Info().
				Str(key.String(), value.String()).
				Msgf("buildAppWithLogWriter: %v", value)
		})).
		Build())

	if e != nil {
		t.Error(e)
	}

	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Fatal(e)
	}

	logEvents := buf.String()
	t.Log(logEvents)
	if len(logEvents) == 0 {
		t.Error("no events were logged")
	}

	var foundLogEvent bool
	var logEvent apptest.LogEvent
	for _, line := range strings.Split(logEvents, "\n") {
		if len(line) == 0 {
			break
		}
		if e := json.Unmarshal([]byte(line), &logEvent); e != nil {
			t.Error(e)
		} else {
			if strings.Contains(logEvent.Message, value.String()) {
				foundLogEvent = true
			}
		}
	}
	if !foundLogEvent {
		t.Error("buildAppWithLogWriter log event was not found")
	}
}

func buildAppSpecifyingTimeouts(t *testing.T) {
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

	<-fxapp.Stopped()
}

func testRunAppWithMinimalRequired(t *testing.T) {
	desc := apptest.InitEnv()
	apptest.ClearAppEnvSettings()

	fxapp, e := appfx.NewAppBuilder().
		// app.Desc is required
		AppDesc(desc).
		// at least 1 option is required
		Options(fx.Invoke(func(lc fx.Lifecycle, l *zerolog.Logger, appDesc app.Desc, appInstanceID app.InstanceID, s fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					l.Info().Msgf("%v : %v : shutting down ...", appDesc, appInstanceID)
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

	<-fxapp.Stopped()
}
