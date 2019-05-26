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

package app_test

import (
	"context"
	"crypto/rand"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/apptest"
	"go.uber.org/fx"
	"testing"
	"time"
)

const APP_NAME = app.Name("foobar")

var APP_VER = semver.MustParse("0.0.1")

func initEnvForDesc() app.Desc {
	// Given all of the required environment variables are set
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)

	apptest.Setenv(apptest.ID, id.String())
	apptest.Setenv(apptest.NAME, string(APP_NAME))
	apptest.Setenv(apptest.RELEASE_ID, releaseID.String())
	apptest.Setenv(apptest.VERSION, APP_VER.String())

	ver := app.Version(*APP_VER)
	return app.Desc{
		ID:        app.ID(id),
		Name:      APP_NAME,
		Version:   &ver,
		ReleaseID: app.ReleaseID(releaseID),
	}
}

func checkDescsAreEqual(t *testing.T, desc, expected app.Desc) {
	// And its properties match what was specified in the env
	if desc.ID != expected.ID {
		t.Errorf("ID did not match: %s != %s", desc.ID, expected.ID)
	}
	if desc.Name != expected.Name {
		t.Errorf("Name did not match: %s != %s", desc.Name, expected.Name)
	}
	if !(*semver.Version)(desc.Version).Equal((*semver.Version)(expected.Version)) {
		t.Errorf("Version did not match: %s != %s", (*semver.Version)(desc.Version), (*semver.Version)(expected.Version))
	}
	if desc.ReleaseID != expected.ReleaseID {
		t.Errorf("ReleaseID did not match: %s != %s", desc.ReleaseID, expected.ReleaseID)
	}
}

func TestDescConstruction(t *testing.T) {
	t.Parallel()
	v := app.Version(*semver.MustParse("0.0.1"))
	desc := &app.Desc{
		ID:        app.ID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)),
		Name:      app.Name("foo"),
		Version:   &v,
		ReleaseID: app.ReleaseID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)),
	}
	t.Logf("%s", desc)
}

func TestLoadDescFromEnv(t *testing.T) {
	// Given all of the required environment variables are set
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	version := semver.MustParse("0.0.1")
	name := app.Name("foobar")

	apptest.Setenv(apptest.ID, id.String())
	apptest.Setenv(apptest.NAME, string(name))
	apptest.Setenv(apptest.RELEASE_ID, releaseID.String())
	apptest.Setenv(apptest.VERSION, version.String())

	// When the Desc is loaded from the env
	desc, err := app.LoadDescFromEnv()

	// Then it is loaded successfully
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", &desc)

	// And its properties match what was specified in the env
	if desc.ID != app.ID(id) {
		t.Errorf("ID did not match: %s != %s", desc.ID, id)
	}
	if desc.Name != name {
		t.Errorf("Name did not match: %s != %s", desc.Name, name)
	}
	if !(*semver.Version)(desc.Version).Equal(version) {
		t.Errorf("Version did not match: %s != %s", (*semver.Version)(desc.Version), version)
	}
	if desc.ReleaseID != app.ReleaseID(releaseID) {
		t.Errorf("ReleaseID did not match: %s != %s", desc.ReleaseID, releaseID)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	defaultTimeout := 15 * time.Second
	t.Run("when env not set, then defaults are used", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		config := app.LoadConfigFromEnv()
		t.Logf("%s", config)
		if config.StartTimeout != defaultTimeout {
			t.Errorf("StartTimeout did not match the default: %v", config.StartTimeout)
		} else if config.StopTimeout != defaultTimeout {
			t.Errorf("StopTimeout did not match the default: %v", config.StartTimeout)
		}
	})

	t.Run("start timeout is set to 30 secs", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		apptest.Setenv(apptest.START_TIMEOUT, "30s")
		config := app.LoadConfigFromEnv()
		t.Logf("%s", config)
		if config.StartTimeout != 30*time.Second {
			t.Errorf("StartTimeout did not match 30s: %v", config.StartTimeout)
		} else if config.StopTimeout != defaultTimeout {
			t.Errorf("StopTimeout did not match the default: %v", config.StartTimeout)
		}
	})

	t.Run("stop timeout is set to 30 secs", func(t *testing.T) {
		apptest.ClearAppEnvSettings()
		apptest.Setenv(apptest.STOP_TIMEOUT, "30s")
		config := app.LoadConfigFromEnv()
		t.Logf("%s", config)
		if config.StartTimeout != defaultTimeout {
			t.Errorf("StartTimeout did not match the default: %v", config.StartTimeout)
		} else if config.StopTimeout != 30*time.Second {
			t.Errorf("StopTimeout did not match 30s: %v", config.StartTimeout)
		}
	})
}

func TestNewApp(t *testing.T) {
	// Given the env is initialized
	apptest.ClearAppEnvSettings()
	expectedDesc := initEnvForDesc()

	t.Run("using default app start and stop timeouts", func(t *testing.T) {
		// When the fx.App is created
		var desc app.Desc
		var instanceID app.InstanceID
		fxapp := app.New(
			fx.Populate(&desc),
			fx.Populate(&instanceID),
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
		checkDescsAreEqual(t, desc, expectedDesc)

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
		fxapp := app.New()
		if fxapp.StartTimeout() != 30*time.Second {
			t.Error("StartTimeout did not match the default")
		}
		if fxapp.StopTimeout() != 60*time.Second {
			t.Error("StopTimeout did not match the default")
		}
		if err := fxapp.Start(context.Background()); err != nil {
			panic(err)
		}
	})

}
