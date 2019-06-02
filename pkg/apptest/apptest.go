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

// Package apptest is used to support testing
package apptest

import (
	"crypto/rand"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/oklog/ulid"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logcfg"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/rs/zerolog"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// Key represents env var config property names - with out the envconfig name prefix
type Key string

// envconfig keys
const (
	ID        = Key("ID")
	Name      = Key("NAME")
	Version   = Key("VERSION")
	ReleaseID = Key("RELEASE_ID")

	StartTimeout = Key("START_TIMEOUT")
	StopTimeout  = Key("STOP_TIMEOUT")

	LogGlobalLevel     = Key("LOG_GLOBAL_LEVEL")
	LogDisableSampling = Key("LOG_DISABLE_SAMPLING")
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

// ClearAppEnvSettings clears the app specific env vars
func ClearAppEnvSettings() {
	Unsetenv(ID)
	Unsetenv(Version)
	Unsetenv(Name)
	Unsetenv(ReleaseID)

	Unsetenv(StartTimeout)
	Unsetenv(StopTimeout)

	Unsetenv(LogGlobalLevel)
	Unsetenv(LogDisableSampling)
}

// LookupEnv prefixes the key with "APP12X" and then retrieves the value of the environment variable named by the prefixed key.
// If the variable is present in the environment the value (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will be false.
func LookupEnv(key Key) (string, bool) {
	return os.LookupEnv(prefix(key))
}

// prefixes the specified key
func prefix(key Key) string {
	return fmt.Sprintf("%s_%s", app.EnvPrefix, strings.ToUpper(string(key)))
}

// CheckDescsAreEqual checks that the descs match - logging errors for any mis-matches
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

// InitEnvForDesc initializes the env for testing and returns an app.Desc that has been loaded from the env.
// - app name = "foobar"
// - app version = 0.0.1
// - sets the env vars to be able to load the app.Desc from it
func InitEnvForDesc() app.Desc {
	const AppName = app.Name("foobar")
	var appVer = semver.MustParse("0.0.1")

	ClearAppEnvSettings()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	releaseID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)

	Setenv(ID, id.String())
	Setenv(Name, string(AppName))
	Setenv(ReleaseID, releaseID.String())
	Setenv(Version, appVer.String())

	ver := app.Version(*appVer)
	return app.Desc{
		ID:        app.ID(id),
		Name:      AppName,
		Version:   &ver,
		ReleaseID: app.ReleaseID(releaseID),
	}
}

// TestLogger writes to a string.Builder, which can then be inspected while testing.
type TestLogger struct {
	*zerolog.Logger
	Buf *strings.Builder
	app.Desc
	app.InstanceID
}

// NewTestLogger constructs a new TestLogger instance.
func NewTestLogger(p app.Package) *TestLogger {
	// Given an app.Desc and app.InstanceID
	desc := InitEnvForDesc()
	instanceID := app.InstanceID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
	// And zerolog is configured
	if err := logcfg.ConfigureZerolog(); err != nil {
		log.Fatalf("app.ConfigureZerolog() failed: %v", err)
	}
	// When a new zerolog.Logger is created
	logger := logcfg.NewLogger(instanceID, desc)
	logger = logging.PackageLogger(logger, p)
	// And the log output is captured in a strings.Builder
	buf := new(strings.Builder)
	logger2 := logger.Output(buf)
	logger = &logger2
	return &TestLogger{logger, buf, desc, instanceID}
}

// LogEvent is used to unmarshal zerolog JSON log events
type LogEvent struct {
	Level        string       `json:"l"`
	Timestamp    int64        `json:"t"`
	Message      string       `json:"m"`
	App          AppDesc      `json:"a"`
	Event        string       `json:"n"`
	ErrorMessage string       `json:"e"`
	Error        *Error       `json:"f"`
	Tags         []string     `json:"g"`
	Stack        []Stackframe `json:"s"`
}

// Time converts the log event event UNIX time into a time.Time
func (e *LogEvent) Time() time.Time {
	return time.Unix(e.Timestamp, 0)
}

// MatchesDesc returns true if the specified Desc matches.
func (e *LogEvent) MatchesDesc(desc *app.Desc) bool {
	return e.App.ID == desc.ID.String() &&
		e.App.Name == string(desc.Name) &&
		e.App.Version == desc.Version.String() &&
		e.App.ReleaseID == desc.ReleaseID.String()
}

// AppDesc is used to unmarshal zerolog JSON log events.
type AppDesc struct {
	ID         string `json:"i"`
	ReleaseID  string `json:"r"`
	Name       string `json:"n"`
	Version    string `json:"v"`
	InstanceID string `json:"x"`
}

// Error represents the error details that were logged.
type Error struct {
	ID         string   `json:"i"`
	Name       string   `json:"n"`
	SrcID      string   `json:"s"`
	InstanceID string   `json:"x"`
	Tags       []string `json:"g"`
}

// Stackframe represents a stack frame that is logged.
type Stackframe struct {
	Func   string `json:"func"`
	Line   string `json:"line"`
	Source string `json:"source"`
}
