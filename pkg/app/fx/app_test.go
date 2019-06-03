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

package fx

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/app/ulidgen"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"io"
	"log"
	"os"
	"path"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	// reset the std logger when the test is done
	flags := log.Flags()
	defer func() {
		log.SetFlags(flags)
		log.SetOutput(os.Stderr)
	}()

	// Given the env is initialized
	expectedDesc := apptest.InitEnvForDesc()

	t.Run("using default settings", func(t *testing.T) {
		// When the fx.App is created
		var desc app.Desc
		var instanceID app.InstanceID
		fxapp := New(
			fx.Populate(&desc),
			fx.Populate(&instanceID),
			fx.Invoke(LogTestEvents),
		)
		if fxapp.StartTimeout() != 15*time.Second {
			t.Error("StartTimeout did not match the default")
		}
		if fxapp.StopTimeout() != 15*time.Second {
			t.Error("StopTimeout did not match the default")
		}

		// Then it starts with no errors
		if err := fxapp.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := fxapp.Stop(context.Background()); err != nil {
				t.Errorf("fxapp.Stop error: %v", err)
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

	})

	t.Run("using overidden app start and stop timeouts", func(t *testing.T) {
		apptest.Setenv(apptest.StartTimeout, "30s")
		apptest.Setenv(apptest.StopTimeout, "60s")
		fxapp := New()
		if fxapp.StartTimeout() != 30*time.Second {
			t.Error("StartTimeout did not match the default")
		}
		if fxapp.StopTimeout() != 60*time.Second {
			t.Error("StopTimeout did not match the default")
		}
		if err := fxapp.Start(context.Background()); err != nil {
			panic(err)
		}
		defer func() {
			if err := fxapp.Stop(context.Background()); err != nil {
				t.Errorf("fxapp.Stop error: %v", err)
			}
		}()
	})

	t.Run("using invalid app start/stop timeouts", func(t *testing.T) {
		apptest.InitEnvForDesc()
		apptest.Setenv(apptest.StartTimeout, "--")
		defer func() {
			if err := recover(); err == nil {
				t.Error("fx.New() should have because the app start timeout was misconfigured")
			} else {
				t.Logf("as expected, fx.New() failed because of: %v", err)
			}
		}()
		New()
	})

	t.Run("using invalid log config", func(t *testing.T) {
		apptest.InitEnvForDesc()
		apptest.Setenv(apptest.LogGlobalLevel, "--")
		defer func() {
			if err := recover(); err == nil {
				t.Error("fx.New() should have because the app global log level was misconfigured")
			} else {
				t.Logf("as expected, fx.New() failed because of: %v", err)
			}
		}()
		New()
	})
}

type empty struct{}

const (
	LogTestEventLogEventName = "LogTestEvents"
	LogTestEventOnStartMsg   = "OnStart"
	LogTestEventOnStopMsg    = "OnStop"
)

func LogTestEvents(logger *zerolog.Logger, lc fx.Lifecycle) {
	logger = logging.PackageLogger(logger, app.GetPackage(empty{}))
	foo := logging.NewEvent(LogTestEventLogEventName, zerolog.InfoLevel)

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

func TestLoadDesc(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("loading Desc should have failed and triggered a panic")
		} else {
			t.Logf("panic is expected: %v", err)
		}
	}()
	apptest.ClearAppEnvSettings()
	loadDesc()
}

func TestAppLifecycleEvents(t *testing.T) {
	checkLifecycleEvents := func(logFile io.Reader) {
		events := make([]string, 0, 6)
		scanner := bufio.NewScanner(logFile)
		for scanner.Scan() {
			logEventJSON := scanner.Text()
			t.Log(logEventJSON)

			var logEvent apptest.LogEvent
			err := json.Unmarshal([]byte(logEventJSON), &logEvent)
			if err != nil {
				t.Fatal(err)
			}

			switch logEvent.Event {
			case Start.Name, Running.Name, Stop.Name, Stopped.Name:
				events = append(events, logEvent.Event)
			case LogTestEventLogEventName:
				events = append(events, logEvent.Message)
			}
		}

		expectedEvents := []string{
			Start.Name,
			LogTestEventOnStartMsg,
			Running.Name,
			Stop.Name,
			LogTestEventOnStopMsg,
			Stopped.Name,
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
	logFile, err := os.Create(logFilePath)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}
	os.Stderr = logFile
	defer func() {
		os.Stderr = stderrBackup
		if err := logFile.Close(); err != nil {
			t.Fatalf("failed to close log file: %v", err)
		}
		logFile, err := os.Open(logFilePath)
		if err != nil {
			t.Fatalf("failed to open the log file for reading: %v", err)
		}
		checkLifecycleEvents(logFile)
		if err := logFile.Close(); err != nil {
			t.Fatalf("failed to close log file: %v", err)
		}

		if err := os.Remove(logFilePath); err != nil {
			t.Logf("failed to delete log file: %v", err)
		}
	}()

	// Given that the app log is captured
	apptest.InitEnvForDesc()
	fxapp := New(fx.Invoke(LogTestEvents))

	// When the app is started
	if err := fxapp.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Then it logs the Start event as the first lofecycle OnStart hook

	// And then after all other OnStart hooks are run, the Running event is logged

	// When the app is stopped
	if err := fxapp.Stop(context.Background()); err != nil {
		t.Errorf("fxapp.Stop error: %v", err)
	}

	// Then the Stop event is logged as the first OnStop hook

	// Then the Stopped event is logged as the jast OnStop hook

}
