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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/comp"
	"github.com/oysterpack/partire-k8s/pkg/app/err"
	appfx "github.com/oysterpack/partire-k8s/pkg/app/fx"
	"github.com/oysterpack/partire-k8s/pkg/app/fx/option"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMustNewApp(t *testing.T) {
	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	t.Run("using default settings", testNewAppWithDefaultSettings)

	t.Run("using overidden app start and stop timeouts", testNewAppWithCustomAppTimeouts)

	t.Run("using invalid app start/stop timeouts", testNewAppWithInvalidTimeouts)

	t.Run("using invalid log config", testNewAppWithInvalidLogConfig)

	t.Run("app.Desc env vars not defined", testDescNotDefinedInEnv)
}

func testNewAppWithInvalidLogConfig(t *testing.T) {
	apptest.InitEnv()
	apptest.Setenv(apptest.LogGlobalLevel, "--")
	defer func() {
		if e := recover(); e == nil {
			t.Error("fx.MustNewApp() should have because the app global log level was misconfigured")
		} else {
			t.Logf("as expected, fx.MustNewApp() failed because of: %v", e)
		}
	}()
	appfx.MustNewApp(fx.Invoke(func() {}))
}

func testNewAppWithInvalidTimeouts(t *testing.T) {
	apptest.InitEnv()
	apptest.Setenv(apptest.StartTimeout, "--")
	defer func() {
		if e := recover(); e == nil {
			t.Error("fx.MustNewApp() should have because the app start timeout was misconfigured")
		} else {
			t.Logf("as expected, fx.MustNewApp() failed because of: %v", e)
		}
	}()
	appfx.MustNewApp(fx.Invoke(func() {}))
}

func testNewAppWithCustomAppTimeouts(t *testing.T) {
	apptest.Setenv(apptest.StartTimeout, "30s")
	apptest.Setenv(apptest.StopTimeout, "60s")
	fxapp := appfx.MustNewApp(fx.Invoke(func() {}))
	if fxapp.StartTimeout() != 30*time.Second {
		t.Error("StartTimeout did not match the default")
	}
	if fxapp.StopTimeout() != 60*time.Second {
		t.Error("StopTimeout did not match the default")
	}
	if e := fxapp.Start(context.Background()); e != nil {
		panic(e)
	}
	defer func() {
		if e := fxapp.Stop(context.Background()); e != nil {
			t.Errorf("fxapp.Stop error: %v", e)
		}
	}()
}

func testNewAppWithDefaultSettings(t *testing.T) {
	// Given the env is initialized
	expectedDesc := apptest.InitEnv()

	// When the fx.App is created
	var desc app.Desc
	var instanceID app.InstanceID
	fxapp := appfx.MustNewApp(
		fx.Populate(&desc),
		fx.Populate(&instanceID),
		fx.Invoke(logTestEvents),
	)
	if fxapp.StartTimeout() != 15*time.Second {
		t.Error("StartTimeout did not match the default")
	}
	if fxapp.StopTimeout() != 15*time.Second {
		t.Error("StopTimeout did not match the default")
	}

	// Then it starts with no errors
	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if e := fxapp.Stop(context.Background()); e != nil {
			t.Errorf("fxapp.Stop error: %v", e)
		}
	}()

	// And app.Desc is provided in the fx.App context
	t.Logf("Desc specified in the env: %s", &expectedDesc)
	t.Logf("Desc loaded via fx app   : %s", &desc)
	apptest.CheckDescsAreEqual(t, desc, expectedDesc)

	// And the app.InstanceID is defined
	t.Logf("app InstanceID: %s", ulid.ULID(instanceID))
	var zeroULID ulid.ULID
	if zeroULID == ulid.ULID(instanceID) {
		t.Error("instanceID was not initialized")
	}
}

type empty struct{}

const (
	LogTestEventLogEventName = "LogTestEvents"
	LogTestEventOnStartMsg   = "OnStart"
	LogTestEventOnStopMsg    = "OnStop"
)

func logTestEvents(logger *zerolog.Logger, lc fx.Lifecycle) {
	logger = logging.PackageLogger(logger, app.GetPackage(empty{}))
	foo := logging.MustNewEvent(LogTestEventLogEventName, zerolog.InfoLevel)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			foo.Log(logger).Msg(LogTestEventOnStartMsg)
			return nil
		},
		OnStop: func(_ context.Context) error {
			foo.Log(logger).Msg(LogTestEventOnStopMsg)
			return nil
		},
	})

	log.Printf("logging using go std log")
}

// Feature: The app logs lifecycle events at the appropriate times.
//
// Given that the app log is captured
// When the app is started
// Then it logs the Start event as the first lifecycle OnStart hook
// And then after all other OnStart hooks are run, the Running event is logged
// When the app is stopped
// Then the Stop event is logged as the first OnStop hook
// Then the Stopped event is logged as the last OnStop hook
func TestAppLifecycleEvents(t *testing.T) {
	checkLifecycleEvents := func(t *testing.T, logFile io.Reader) {
		events := make([]string, 0, 6)
		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			logEventJSON := scanner.Text()
			t.Log(logEventJSON)

			var logEvent apptest.LogEvent
			e := json.Unmarshal([]byte(logEventJSON), &logEvent)
			if e != nil {
				t.Fatal(e)
			}

			switch logEvent.Event {
			case appfx.Start.Name, appfx.Running.Name, appfx.Stop.Name, appfx.Stopped.Name:
				events = append(events, logEvent.Event)
			case LogTestEventLogEventName:
				events = append(events, logEvent.Message)
			}
		}

		expectedEvents := []string{
			appfx.Start.Name,
			LogTestEventOnStartMsg,
			appfx.Running.Name,
			appfx.Stop.Name,
			LogTestEventOnStopMsg,
			appfx.Stopped.Name,
		}

		t.Log(events)
		t.Log(expectedEvents)

		if len(expectedEvents) != len(events) {
			t.Fatalf("the expected number of events did not match: %v != %v", len(expectedEvents), len(events))
		}

		for i, event := range events {
			if expectedEvents[i] != event {
				t.Errorf("event did not match: %v != %v", expectedEvents[i], event)
			}
		}
	}

	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// redirect Stderr to a log file
	stderrBackup := os.Stderr
	logFilePath := path.Join(os.TempDir(), ulidgen.MustNew().String()) + ".log"
	t.Logf("log file: %v", logFilePath)
	logFile, e := os.Create(logFilePath)
	if e != nil {
		t.Fatalf("failed to create log file: %v", e)
	}
	os.Stderr = logFile
	defer func() {
		// restore stderr
		os.Stderr = stderrBackup
		checkLogEvents(t, logFilePath, logFile, checkLifecycleEvents)
	}()

	// Given that the app log is captured
	apptest.InitEnv()
	fxapp := appfx.MustNewApp(fx.Invoke(logTestEvents))

	// When the app is started
	if e := fxapp.Start(context.Background()); e != nil {
		t.Fatal(e)
	}

	// Then it logs the Start event as the first lifecycle OnStart hook
	// And then after all other OnStart hooks are run, the Running event is logged
	// When the app is stopped
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Errorf("fxapp.Stop error: %v", e)
	}

	// Then the Stop event is logged as the first OnStop hook
	// Then the Stopped event is logged as the last OnStop hook
}

func checkLogEvents(t *testing.T, logFilePath string, logFile *os.File, checker func(t *testing.T, logFile io.Reader)) {
	// close the log file to ensure it is flushed to disk
	if e := logFile.Close(); e != nil {
		t.Fatalf("failed to close log file: %v", e)
	}
	logFile, e := os.Open(logFilePath)
	if e != nil {
		t.Fatalf("failed to open the log file for reading: %v", e)
	}
	checker(t, logFile)
	if e := logFile.Close(); e != nil {
		t.Fatalf("failed to close log file: %v", e)
	}
	if e := os.Remove(logFilePath); e != nil {
		t.Logf("failed to delete log file: %v", e)
	}
}

var (
	TestErr  = err.MustNewDesc("01DCF9FYQMKKM6MA3RAYZWEVTR", "TestError", "test error")
	TestErr1 = err.New(TestErr, "01DC9JRXD98HS9BEXJ1MBXWWM8")
)

// Feature: Errors produced by app functions that are invoked by fx will be logged automatically
//
// Scenario: invoked func return an error of type *err.Instance
//
// Given that the app log is captured
// When the app is started
// Then the app will fail to start because the invoked test function fails
// And the error will be logged
func TestAppInvokeErrorHandling(t *testing.T) {
	checkErrorEvents := func(t *testing.T, logFile io.Reader) {
		scanner := bufio.NewScanner(logFile)
		errorLogged := false
		for scanner.Scan() {
			logEventJSON := scanner.Text()
			t.Log(logEventJSON)

			var logEvent apptest.LogEvent
			e := json.Unmarshal([]byte(logEventJSON), &logEvent)
			if e != nil {
				t.Fatal(e)
			}

			if logEvent.Level == zerolog.ErrorLevel.String() {
				if logEvent.Error.ID == TestErr.ID.String() {
					errorLogged = true
				}
			}
		}

		if !errorLogged {
			t.Error("Error was not logged")
		}
	}

	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// Given that the app log is captured

	// redirect Stderr to a log file
	stderrBackup := os.Stderr
	logFile, logFilePath := apptest.CreateLogFile(t)
	os.Stderr = logFile

	defer func() {
		// restore stderr
		os.Stderr = stderrBackup
		checkLogEvents(t, logFilePath, logFile, checkErrorEvents)
	}()

	// When the app is created with a function that fails and returns an error when invoked
	apptest.InitEnv()
	func() {
		defer func() {
			e := recover()
			if e == nil {
				t.Fatal("app should have failed to be created because the invoked func returned an error")
			}
			t.Log(e)
		}()
		appfx.MustNewApp(fx.Invoke(func() error {
			t.Log("test func has been invoked ...")
			return TestErr1.New()
		}))
	}()
}

// Feature: Errors produced by app functions that are invoked by fx will be logged automatically
//
// Scenario: invoked func return an error that is a non-standard type, i.e. not of type *err.Instance
//
// Given that the app log is captured
// When the app is started
// Then the app will fail to start because the invoked test function fails
// And the error will be logged
func TestAppInvokeErrorHandlingForNonStandardError(t *testing.T) {
	checkErrorEvents := func(t *testing.T, logFile io.Reader) {
		scanner := bufio.NewScanner(logFile)
		errorLogged := false
		for scanner.Scan() {
			logEventJSON := scanner.Text()
			t.Log(logEventJSON)

			var logEvent apptest.LogEvent
			e := json.Unmarshal([]byte(logEventJSON), &logEvent)
			if e != nil {
				t.Fatal(e)
			}

			if logEvent.Level == zerolog.ErrorLevel.String() {
				if logEvent.Error.ID == appfx.InvokeErrClass.ID.String() {
					errorLogged = true
				}
			}
		}

		if !errorLogged {
			t.Error("Error was not logged")
		}
	}

	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// Given that the app log is captured

	// redirect Stderr to a log file
	stderrBackup := os.Stderr
	logFile, logFilePath := apptest.CreateLogFile(t)
	os.Stderr = logFile
	defer func() {
		// restore stderr
		os.Stderr = stderrBackup
		checkLogEvents(t, logFilePath, logFile, checkErrorEvents)
	}()

	// When the app is created with a function that fails and returns an error when invoked
	apptest.InitEnv()
	func() {
		defer func() {
			e := recover()
			if e == nil {
				t.Fatal("app should have failed to be created because the invoked func returned an error")
			}
			t.Log(e)
		}()
		appfx.MustNewApp(fx.Invoke(func() error {
			t.Log("test func has been invoked ...")
			return errors.New("non standard error")
		}))
	}()
}

// Feature: Errors produced by app functions that are invoked by fx will be logged automatically
//
// Scenario: hook OnStart handler results in an error of type *err.Instance
//
// Given that the app log is captured
// When the app is started
// Then the app will fail to start because a hook OnStart function returns an error
// And the error will be logged
func TestAppHookOnStartErrorHandling(t *testing.T) {
	checkErrorEvents := func(t *testing.T, logFile io.Reader) {
		scanner := bufio.NewScanner(logFile)
		errorLogged := false
		for scanner.Scan() {
			logEventJSON := scanner.Text()
			t.Log(logEventJSON)

			var logEvent apptest.LogEvent
			e := json.Unmarshal([]byte(logEventJSON), &logEvent)
			if e != nil {
				t.Fatal(e)
			}

			if logEvent.Level == zerolog.ErrorLevel.String() {
				if logEvent.Error.ID == appfx.AppStartErrClass.ID.String() {
					errorLogged = true
				}
			}
		}

		if !errorLogged {
			t.Error("Error was not logged")
		}
	}

	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// Given that the app log is captured

	// redirect Stderr to a log file
	stderrBackup := os.Stderr
	logFile, logFilePath := apptest.CreateLogFile(t)
	os.Stderr = logFile
	defer func() {
		// restore stderr
		os.Stderr = stderrBackup
		checkLogEvents(t, logFilePath, logFile, checkErrorEvents)
	}()

	// When the app is created with a function that fails and returns an error when invoked
	apptest.InitEnv()
	fxapp := appfx.MustNewApp(fx.Invoke(func(lc fx.Lifecycle) error {
		t.Log("test func has been invoked ...")
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				t.Log("OnStart is about to fail ...")
				return TestErr1.New()
			},
		})
		return nil
	}))

	e := fxapp.Run()
	if e == nil {
		t.Fatal("Expected the app to fail to start up")
	}
	t.Logf("as expected, app failed to start: %v", e)
}

// Feature: Errors produced by app functions that are invoked by fx will be logged automatically
//
// Scenario: hook OnStart handler results in an error of type *err.Instance
//
// Given that the app log is captured
// When the app is signalled to stop
// Then the app will fail to stop cleanly because a hook OnStop function returns an error
// And the error will be logged
func TestAppHookOnStopErrorHandling(t *testing.T) {
	checkErrorEvents := func(t *testing.T, logFile io.Reader) {
		scanner := bufio.NewScanner(logFile)
		errorLogged := false
		for scanner.Scan() {
			logEventJSON := scanner.Text()
			t.Log(logEventJSON)

			var logEvent apptest.LogEvent
			e := json.Unmarshal([]byte(logEventJSON), &logEvent)
			if e != nil {
				t.Fatal(e)
			}

			if logEvent.Level == zerolog.ErrorLevel.String() {
				if logEvent.Error.ID == appfx.AppStopErrClass.ID.String() {
					errorLogged = true
				}
			}
		}

		if !errorLogged {
			t.Error("Error was not logged")
		}
	}

	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// Given that the app log is captured

	// redirect Stderr to a log file
	stderrBackup := os.Stderr
	logFile, logFilePath := apptest.CreateLogFile(t)
	os.Stderr = logFile
	defer func() {
		// restore stderr
		os.Stderr = stderrBackup
		checkLogEvents(t, logFilePath, logFile, checkErrorEvents)
	}()

	apptest.InitEnv()
	fxapp := appfx.MustNewApp(
		// When the app is configured with an OnStop hook that will fail
		fx.Invoke(func(lc fx.Lifecycle) error {
			t.Log("test func has been invoked ...")
			lc.Append(fx.Hook{
				OnStop: func(context.Context) error {
					t.Log("OnStop is about to fail ...")
					return TestErr1.New()
				},
			})
			return nil
		}),
		// And the app will stop itself right after it starts
		fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					fmt.Println("App will be shutdown ...")
					if e := shutdowner.Shutdown(); e != nil {
						t.Fatalf("shutdowner.Shutdown() failed: %v", e)
					}
					fmt.Println("App has been signalled to shutdown ...")
					return nil
				},
			})

		}),
	)

	errChan := make(chan error)
	go func() {
		e := fxapp.Run()
		if e != nil {
			errChan <- e
		}
		close(errChan)
	}()
	// wait for the app to stop
	e := <-errChan
	if e == nil {
		t.Fatal("Expected the app to fail to start up")
	}
	t.Logf("as expected, app failed to start: %v", e)
}

// Feature: App will run until it is signalled to shutdown
//
// Scenario: the app will signal itself to shutdown as soon as it starts up
//
// When the app starts, it shuts itself down
// Then the app shuts down cleanly
func TestApp_Run(t *testing.T) {
	// reset the std logger when the test is done because the app will configure the std logger to use zerolog
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	apptest.InitEnv()
	fxapp := appfx.MustNewApp(
		// And the app will stop itself right after it starts
		fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					if e := shutdowner.Shutdown(); e != nil {
						t.Fatalf("shutdowner.Shutdown() failed: %v", e)
					}
					return nil
				},
			})

		}),
	)

	errChan := make(chan error)
	go func() {
		e := fxapp.Run()
		if e != nil {
			errChan <- e
		}
		close(errChan)
	}()
	// wait for the app to stop
	if e := <-errChan; e != nil {
		t.Errorf("App failed to run: %v", e)
	}
}

func TestErrRegistryIsProvided(t *testing.T) {
	// reset the std logger when the test is done because the app will configure the std logger to use zerolog
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	apptest.InitEnv()
	fxapp := appfx.MustNewApp(fx.Invoke(func(errRegistry *err.Registry, logger *zerolog.Logger, shutdowner fx.Shutdowner, lc fx.Lifecycle) {
		logger.Info().Msgf("registered errors: %v", errRegistry.Errs())

		// all of the standard app errors should be registered
		errs := []*err.Err{appfx.InvokeErr, appfx.AppStartErr, appfx.AppStopErr}
		for _, e := range errs {
			if !errRegistry.Registered(e.SrcID) {
				t.Errorf("error is not registered: %v", e)
			}
		}

		// when the app starts, shut it down
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				if e := shutdowner.Shutdown(); e != nil {
					t.Fatalf("shutdowner.Shutdown() failed: %v", e)
				}
				return nil
			},
		})
	}))
	errChan := make(chan error)
	go func() {
		e := fxapp.Run()
		if e != nil {
			errChan <- e
		}
		close(errChan)
	}()
	// wait for the app to stop
	if e := <-errChan; e != nil {
		t.Errorf("App failed to run: %v", e)
	}

}

func TestEventRegistryIsProvided(t *testing.T) {
	// reset the std logger when the test is done because the app will configure the std logger to use zerolog
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	apptest.InitEnv()
	fxapp := appfx.MustNewApp(fx.Invoke(func(registry *logging.EventRegistry, logger *zerolog.Logger, shutdowner fx.Shutdowner, lc fx.Lifecycle) {
		logger.Info().Msgf("registered events: %v", registry.Events())

		// all of the standard app events should be registered
		events := []*logging.Event{appfx.Start, appfx.Running, appfx.Stop, appfx.Stopped, appfx.StopSignal, appfx.CompRegistered}
		for _, e := range events {
			if !registry.Registered(e) {
				t.Errorf("event is not registered: %v", e)
			}
		}

		// when the app starts, shut it down
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				if e := shutdowner.Shutdown(); e != nil {
					t.Fatalf("shutdowner.Shutdown() failed: %v", e)
				}
				return nil
			},
		})
	}))
	errChan := make(chan error)
	go func() {
		e := fxapp.Run()
		if e != nil {
			errChan <- e
		}
		close(errChan)
	}()
	// wait for the app to stop
	if e := <-errChan; e != nil {
		t.Errorf("App failed to run: %v", e)
	}
}

func testDescNotDefinedInEnv(t *testing.T) {
	defer func() {
		if e := recover(); e == nil {
			t.Fatal("loading Desc should have failed and triggered a panic")
		} else {
			t.Logf("panic is expected: %v", e)
		}
	}()
	apptest.ClearAppEnvSettings()
	appfx.MustNewApp(fx.Invoke(func() {}))
}

type RandomNumberGenerator func() int
type ProvideRandomNumberGenerator func() RandomNumberGenerator

type Greeter func() string
type ProvideGreeter func() Greeter

var (
	ProvideRandomNumberGeneratorOption = option.NewDesc(option.Provide, reflect.TypeOf(ProvideRandomNumberGenerator(nil)))
	ProvideGreeterOption               = option.NewDesc(option.Provide, reflect.TypeOf(ProvideGreeter(nil)))

	FooComp = comp.MustNewDesc(
		comp.ID("01DCYBFQBQVXG8PZ758AM9JJCD"),
		comp.Name("foo"),
		comp.Version("0.0.1"),
		app.Package("github.com/oysterpack/partire-k8s/pkg/foo"),
		ProvideRandomNumberGeneratorOption,
	)

	BarComp = comp.MustNewDesc(
		comp.ID("01DCYD1X7FMSRJMVMA8RWK7HMB"),
		comp.Name("bar"),
		comp.Version("0.0.1"),
		app.Package("github.com/oysterpack/partire-k8s/pkg/bar"),
		ProvideGreeterOption,
	)
)

func TestCompRegistryIsProvided(t *testing.T) {
	// reset the std logger when the test is done because the app will configure the std logger to use zerolog
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	t.Run("with 2 comps injectd", testCompRegistryWithCompsRegistered)

	t.Run("with 0 comps injected", testEmptyComponentRegistry)

	// error scenarios
	t.Run("with duplicate comps injected", testComponentRegistryWithDuplicateComps)
	t.Run("with comps that have conflicting errors registered", testCompRegistryWithCompsContainingConflictingErrors)
}

func testComponentRegistryWithDuplicateComps(t *testing.T) {
	// Given 2 components conflict because they have the same ID
	FooComp := comp.MustNewDesc(
		comp.ID("01DCYBFQBQVXG8PZ758AM9JJCD"),
		comp.Name("foo"),
		comp.Version("0.0.1"),
		app.Package("github.com/oysterpack/partire-k8s/pkg/foo"),
		ProvideRandomNumberGeneratorOption,
	)
	BarComp := comp.MustNewDesc(
		comp.ID(FooComp.ID.String()), // dup comp.ID will cause comp registration to fail
		comp.Name("bar"),
		comp.Version("0.0.1"),
		app.Package("github.com/oysterpack/partire-k8s/pkg/bar"),
		ProvideGreeterOption,
	)

	foo := FooComp.MustNewComp(
		ProvideRandomNumberGeneratorOption.NewOption(func() RandomNumberGenerator {
			return rand.Int
		}),
	)
	bar := BarComp.MustNewComp(ProvideGreeterOption.NewOption(func() Greeter {
		return func() string { return "greetings" }
	}))

	apptest.InitEnv()
	func() {
		defer func() {
			e := recover()
			if e == nil {
				t.Fatal("app should have failed to be created because the invoked func returned an error")
			}
			t.Log(e)
		}()
		appfx.MustNewApp(foo.FxOptions(), bar.FxOptions(), fx.Invoke(func(r *comp.Registry, l *zerolog.Logger) {
			// triggers the comp.Registry to be constructed, which should then trigger the error when comps register
			l.Info().Msgf("%v", r.Comps())
		}))
	}()
}

// app components are optional, i.e., in order for an app to run, it requires at least 1 fx.Option
//
// When an app is created with no explicit components, but has options defined
// Then the app starts up fine
func testEmptyComponentRegistry(t *testing.T) {
	apptest.InitEnv()
	// When the app is created with no components
	var compRegistry *comp.Registry
	fxapp := appfx.MustNewApp(fx.Populate(&compRegistry))
	// Then the app starts up just fine
	if e := fxapp.Start(context.Background()); e != nil {
		t.Errorf("failed to start app: %v", e)
	}
	if e := fxapp.Stop(context.Background()); e != nil {
		t.Errorf("failed to start app: %v", e)
	}
	t.Logf("registered components: %v", compRegistry.Comps())
}

// When components are registered
// And components expose errors and events
// Then events are logged when the components are registered
// And the component's events are registered with the app event registry
// And the component's errors are registered with the app error registry
func testCompRegistryWithCompsRegistered(t *testing.T) {
	apptest.InitEnv()

	event1 := logging.MustNewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)
	event2 := logging.MustNewEvent(ulidgen.MustNew().String(), zerolog.InfoLevel)

	errDesc1 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc1")
	err1 := err.New(errDesc1, ulidgen.MustNew().String())
	err2 := err.New(errDesc1, ulidgen.MustNew().String())

	// Given 2 components
	foo := FooComp.MustNewComp(
		ProvideRandomNumberGeneratorOption.NewOption(func() RandomNumberGenerator {
			return rand.Int
		}),
	)
	// And the component exposes events
	foo.EventRegistry.Register(event1, event2)
	// And the component exposes errors
	if e := foo.ErrorRegistry.Register(err1, err2); e != nil {
		t.Fatal(e)
	}
	bar := BarComp.MustNewComp(ProvideGreeterOption.NewOption(func() Greeter {
		return func() string { return "greetings" }
	}))

	// redirect Stderr to a log file so that we can read the log
	stderrBackup := os.Stderr
	logFile, logFilePath := apptest.CreateLogFile(t)
	os.Stderr = logFile
	defer func() {
		// restore stderr
		os.Stderr = stderrBackup
	}()

	// When the app is created with the 2 components
	var compRegistry *comp.Registry
	var eventRegistry *logging.EventRegistry
	var errRegistry *err.Registry
	fxapp := appfx.MustNewApp(
		foo.FxOptions(),
		bar.FxOptions(),
		fx.Populate(&compRegistry),
		fx.Populate(&eventRegistry),
		fx.Populate(&errRegistry),
	)
	if e := fxapp.Start(context.Background()); e != nil {
		t.Errorf("failed to start app: %v", e)
	}

	// Then the components are registered
	for _, c := range []*comp.Comp{foo, bar} {
		if compRegistry.FindByID(c.ID) == nil {
			t.Errorf("*** component was not found in the registry: %v", c)
		}
	}

	for _, event := range []*logging.Event{event1, event2} {
		if !eventRegistry.Registered(event) {
			t.Errorf("*** event is not registered: %v", event)
		}
	}

	for _, e := range []*err.Err{err1, err2} {
		if !errRegistry.Registered(e.SrcID) {
			t.Errorf("*** error is not registered: %v", e)
		}
	}

	if e := fxapp.Stop(context.Background()); e != nil {
		t.Errorf("*** failed to start app: %v", e)
	}

	defer func() {
		checkCompRegisteredEvents := func(t *testing.T, log io.Reader) {
			compRegisteredEvents := apptest.CollectLogEvents(t, log, func(logEvent *apptest.LogEvent) bool {
				return logEvent.Event == appfx.CompRegistered.Name
			})

			if len(compRegisteredEvents) == 0 {
				t.Errorf("no %q events were logged", appfx.CompRegistered.Name)
			} else {
				t.Logf("len(compRegisteredEvents) = %d", len(compRegisteredEvents))
				checkCompRegisteredEvents(t, []*comp.Comp{foo, bar}, compRegisteredEvents)
			}
		}

		// And comp registration events are logged
		checkLogEvents(t, logFilePath, logFile, checkCompRegisteredEvents)
	}()
}

// When components are registered that conflict
// Then the app construction will fail
func testCompRegistryWithCompsContainingConflictingErrors(t *testing.T) {
	appDesc := apptest.InitEnv()

	errDesc1 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc1")
	errDesc2 := err.MustNewDesc(ulidgen.MustNew().String(), ulidgen.MustNew().String(), "errDesc2")

	err1 := err.New(errDesc1, ulidgen.MustNew().String())
	err2 := err.New(errDesc2, err1.SrcID.String()) // will fail error registration

	// Given 2 components
	foo := FooComp.MustNewComp(
		ProvideRandomNumberGeneratorOption.NewOption(func() RandomNumberGenerator {
			return rand.Int
		}),
	)
	bar := BarComp.MustNewComp(ProvideGreeterOption.NewOption(func() Greeter {
		return func() string { return "greetings" }
	}))

	// And the component exposes errors, but they conflict
	if e := foo.ErrorRegistry.Register(err1); e != nil {
		t.Fatal(e)
	}
	if e := bar.ErrorRegistry.Register(err2); e != nil {
		t.Fatal(e)
	}

	// When the app is created
	// Then it will fail
	_, e := appfx.NewApp(
		appDesc,
		app.NewTimeouts(),
		nil,
		zerolog.InfoLevel,
		foo.FxOptions(),
		bar.FxOptions(),
	)

	if e == nil {
		t.Fatal("the app should have failed to be created because the comp error registration should have failed")
	}
	t.Log(e)
	errInstance := e.(*err.Instance)
	if errInstance.SrcID != err.RegistryConflictErr.SrcID {
		t.Errorf("unexpected error: %v : %v", errInstance.SrcID, errInstance)
	}

}

func checkCompRegisteredEvents(t *testing.T, comps []*comp.Comp, events []*apptest.LogEvent) {
CompLoop:
	for _, c := range comps {
		for _, event := range events {
			if c.ID.String() == event.Comp.ID {
				t.Logf("checking %v against %v", c, event.Comp)
				if event.Comp.Version != c.Version.String() {
					t.Error("comp version did not match")
				}
				if len(c.Options) != len(event.Comp.Options) {
					t.Error("number of options does not match")
				} else {
					for _, opt := range c.Options {
						for _, eventOpt := range event.Comp.Options {
							if !strings.Contains(eventOpt, opt.FuncType.String()) {
								t.Error("option.Desc.FuncType did not match")
							}
							if !strings.Contains(eventOpt, opt.Type.String()) {
								t.Error("option.Desc.Type did not match")
							}
						}
					}
				}
				continue CompLoop
			}
		}
		t.Errorf("event was not logged for: %v", c.ID)
	}
}
