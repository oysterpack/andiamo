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
	"context"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"log"
	"os"
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
			fx.Invoke(UseLogger),
		)
		if fxapp.StartTimeout() != 15*time.Second {
			t.Error("StartTimeout did not match the default")
		}
		if fxapp.StopTimeout() != 15*time.Second {
			t.Error("StopTimeout did not match the default")
		}

		// Then it starts with no errors
		if err := fxapp.Start(context.Background()); err != nil {
			panic(err)
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
		apptest.Setenv(apptest.START_TIMEOUT, "30s")
		apptest.Setenv(apptest.STOP_TIMEOUT, "60s")
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
		apptest.Setenv(apptest.START_TIMEOUT, "--")
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
		apptest.Setenv(apptest.LOG_GLOBAL_LEVEL, "--")
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

func UseLogger(logger *zerolog.Logger, desc app.Desc, instanceID app.InstanceID) {
	logger.Info().
		Str("instance_id", instanceID.String()).
		Str("desc", desc.String()).
		Msg("app is running")

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
