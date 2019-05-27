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

// apptest package is used to support testing
package apptest

import (
	"crypto/rand"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"os"
	"strings"
	"testing"
	"time"
)

type Key string

const (
	ID         = Key("ID")
	NAME       = Key("NAME")
	VERSION    = Key("VERSION")
	RELEASE_ID = Key("RELEASE_ID")

	START_TIMEOUT = Key("START_TIMEOUT")
	STOP_TIMEOUT  = Key("STOP_TIMEOUT")

	LOG_GLOBAL_LEVEL     = Key("LOG_GLOBAL_LEVEL")
	LOG_DISABLE_SAMPLING = Key("LOG_DISABLE_SAMPLING")
)

// Setenv prefixes the key with "APPX12" and then sets the value of the environment variable named by the prefixed key.
func Setenv(key Key, value string) {
	if err := os.Setenv(prefix(key), value); err != nil {
		panic(err)
	}
}

// Unsetenv prefixes the key with "APPX12" and then tries to unset the env var
func Unsetenv(key Key) {
	if err := os.Unsetenv(prefix(key)); err != nil {
		panic(err)
	}
}

func ClearAppEnvSettings() {
	Unsetenv(ID)
	Unsetenv(VERSION)
	Unsetenv(NAME)
	Unsetenv(RELEASE_ID)

	Unsetenv(START_TIMEOUT)
	Unsetenv(STOP_TIMEOUT)

	Unsetenv(LOG_GLOBAL_LEVEL)
	Unsetenv(LOG_DISABLE_SAMPLING)
}

// Getenv prefixes the key with "APP12X" and then retrieves the value of the environment variable named by the prefixed key.
// If the variable is present in the environment the value (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will be false.
func LookupEnv(key Key) (string, bool) {
	return os.LookupEnv(prefix(key))
}

// prefixes the specified key
func prefix(key Key) string {
	return fmt.Sprintf("%s_%s", app.ENV_PREFIX, strings.ToUpper(string(key)))
}

func CheckDescsAreEqual(t *testing.T, desc, expected app.Desc) {
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

func InitEnvForDesc() app.Desc {
	const APP_NAME = app.Name("foobar")
	var appVer = semver.MustParse("0.0.1")

	ClearAppEnvSettings()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)

	Setenv(ID, id.String())
	Setenv(NAME, string(APP_NAME))
	Setenv(RELEASE_ID, releaseID.String())
	Setenv(VERSION, appVer.String())

	ver := app.Version(*appVer)
	return app.Desc{
		ID:        app.ID(id),
		Name:      APP_NAME,
		Version:   &ver,
		ReleaseID: app.ReleaseID(releaseID),
	}
}
