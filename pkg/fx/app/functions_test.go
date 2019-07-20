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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/oklog/ulid"
	"github.com/oysterpack/andiamo/pkg/fx/app"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"os"
	"testing"
)

func TestAppIDs(t *testing.T) {
	t.Parallel()

	t.Run("IDs are explicity set", func(t *testing.T) {
		t.Parallel()
		opts := app.Opts{ID: ulids.MustNew(), ReleaseID: ulids.MustNew()}
		a := fx.New(
			app.Module(opts),
			fx.Invoke(
				func(id app.ID, releaseID app.ReleaseID, instanceID app.InstanceID) {
					t.Logf("%s %s %s", id(), releaseID(), instanceID())
					assert.Equal(t, id(), opts.ID)
					assert.Equal(t, releaseID(), opts.ReleaseID)
					assert.NotEqual(t, instanceID(), ulid.ULID{}, "instanceID must not be a zero value")
				},
			),
		)

		assert.NoError(t, a.Err(), "app failed to initialize")
	})

	t.Run("load from env", func(t *testing.T) {
		prefix := ulids.MustNew().String()
		appID := ulids.MustNew()
		appReleaseID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())
		os.Setenv(prefix+"_RELEASE_ID", appReleaseID.String())

		a := fx.New(
			app.Module(app.Opts{EnvPrefix: prefix}),
			fx.Invoke(
				func(id app.ID, releaseID app.ReleaseID, instanceID app.InstanceID) {
					t.Logf("%s %s %s", id(), releaseID(), instanceID())
					assert.Equal(t, id(), appID)
					assert.Equal(t, releaseID(), appReleaseID)
					assert.NotEqual(t, instanceID(), ulid.ULID{}, "instanceID must not be a zero value")
				},
			),
		)

		assert.NoError(t, a.Err(), "app failed to initialize")
	})

	t.Run("load from env", func(t *testing.T) {
		prefix := app.EnvPrefix
		appID := ulids.MustNew()
		appReleaseID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())
		os.Setenv(prefix+"_RELEASE_ID", appReleaseID.String())

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(
				func(id app.ID, releaseID app.ReleaseID, instanceID app.InstanceID) {
					t.Logf("%s %s %s", id(), releaseID(), instanceID())
					assert.Equal(t, id(), appID)
					assert.Equal(t, releaseID(), appReleaseID)
					assert.NotEqual(t, instanceID(), ulid.ULID{}, "instanceID must not be a zero value")
				},
			),
		)

		assert.NoError(t, a.Err(), "app failed to initialize")
	})

	t.Run("load from env - missing ID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		appReleaseID := ulids.MustNew()
		os.Setenv(prefix+"_RELEASE_ID", appReleaseID.String())

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(func(app.ID) {}),
		)

		assert.Error(t, a.Err(), a.Err().Error())
	})

	t.Run("load from env - missing RELEASE_ID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		appID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(
				func(app.ReleaseID) {},
			),
		)

		assert.Error(t, a.Err(), a.Err().Error())
	})

	t.Run("load from env - invalid ID ULID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		os.Setenv(prefix+"_RELEASE_ID", ulids.MustNew().String())
		os.Setenv(prefix+"_ID", "INVALID")

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(func(app.ID) {}),
		)

		assert.Error(t, a.Err(), a.Err().Error())
	})

	t.Run("load from env - invalid RELEASE_ID ULID env var", func(t *testing.T) {
		prefix := app.EnvPrefix
		appID := ulids.MustNew()
		os.Setenv(prefix+"_ID", appID.String())
		os.Setenv(prefix+"_RELEASE_ID", "INVALID")

		defer func() {
			os.Unsetenv(prefix + "_ID")
			os.Unsetenv(prefix + "_RELEASE_ID")
		}()

		a := fx.New(
			app.Module(app.Opts{}),
			fx.Invoke(func(app.ReleaseID) {}),
		)

		assert.Error(t, a.Err(), a.Err().Error())

	})
}

func TestLogger(t *testing.T) {
	t.Run("using defaults", func(t *testing.T) {
		a := fx.New(
			// Given the app module is plugged in
			app.Module(app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
			}),
			// When a Logger dependency is injected
			fx.Invoke(func(logger app.Logger) {
				// Then it available for used within the app
				event := logger("foo", zerolog.InfoLevel)
				event(nil, "bar")
			}),
		)

		// Then the app is initialized successfully
		assert.NoError(t, a.Err())
		assert.NoError(t, a.Start(context.Background()))
		assert.NoError(t, a.Stop(context.Background()))

		// And the default global log level is set to Info by default
		assert.Equal(t, zerolog.GlobalLevel(), zerolog.InfoLevel)
	})

	t.Run("with log writer", func(t *testing.T) {
		buf := new(bytes.Buffer)
		var ID app.ID
		var ReleaseID app.ReleaseID
		var InstanceID app.InstanceID
		a := fx.New(
			app.Module(app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			}),
			fx.Invoke(func(logger app.Logger) {
				event := logger("foo", zerolog.NoLevel)
				event(nil, "bar")
			}),
			fx.Populate(&ID, &ReleaseID, &InstanceID),
		)

		assert.NoError(t, a.Err())
		assert.NoError(t, a.Start(context.Background()))
		assert.NoError(t, a.Stop(context.Background()))

		assert.Equal(t, zerolog.GlobalLevel(), zerolog.InfoLevel)

		// {"a":"01DG138TTVDX5JH5F4GMNC3V67","r":"01DG138TTVK4MVW3B5TJGDSKHR","x":"01DG138TTVYGSN7QWBFT9660SS","n":"foo","z":"01DG138TTVBHCXQW29QTQAWPNM","t":1563405085,"m":"bar"}
		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"i"`
		}

		r := bufio.NewReader(buf)
		var logEvent LogEvent
		for {
			line, err := r.ReadString('\n')
			t.Log(line)
			if err != nil {
				break
			}
			assert.NoError(t, json.Unmarshal([]byte(line), &logEvent), "failed to parse line: %s", line)
			// Then the app ID is set on the log event
			assert.Equal(t, ID().String(), logEvent.AppID)
			// And the app release ID is set on the log event
			assert.Equal(t, ReleaseID().String(), logEvent.AppReleaseID)
			// And the app instance ID is set on the log event
			assert.Equal(t, InstanceID().String(), logEvent.InstanceID)
		}
	})

	t.Run("global log level loaded from env", func(t *testing.T) {
		// Given that APP12X_LOG_LEVEL env var is set to debug
		prefix := app.EnvPrefix
		os.Setenv(prefix+"_LOG_LEVEL", zerolog.DebugLevel.String())

		defer func() {
			os.Unsetenv(prefix + "_LOG_LEVEL")
		}()

		buf := new(bytes.Buffer)
		var ID app.ID
		var ReleaseID app.ReleaseID
		var InstanceID app.InstanceID
		const FooEvent = "01DG3P8XM1PAMV6B8XMSDX0QSH"
		a := fx.New(
			app.Module(app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			}),
			fx.Invoke(func(logger app.Logger) {
				// And a debug event is logged
				event := logger(FooEvent, zerolog.DebugLevel)
				event(nil, "bar")
			}),
			fx.Populate(&ID, &ReleaseID, &InstanceID),
		)

		assert.NoError(t, a.Err())
		// When the app is initialized
		// Then the zerolog global log level should be set to debug
		assert.Equal(t, zerolog.GlobalLevel(), zerolog.DebugLevel)

		// {"a":"01DG138TTVDX5JH5F4GMNC3V67","r":"01DG138TTVK4MVW3B5TJGDSKHR","x":"01DG138TTVYGSN7QWBFT9660SS","n":"foo","z":"01DG138TTVBHCXQW29QTQAWPNM","t":1563405085,"m":"bar"}
		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"i"`
		}

		r := bufio.NewReader(buf)
		var logEvent LogEvent
		// Then the debug event was logged
		fooEventLogged := false
		for {
			line, err := r.ReadString('\n')
			t.Log(line)
			if err != nil {
				break
			}
			assert.NoError(t, json.Unmarshal([]byte(line), &logEvent), "failed to parse line: %s", line)
			if logEvent.Name == FooEvent {
				fooEventLogged = true
				break
			}
		}
		assert.True(t, fooEventLogged, "event should have been logged at debug level")
	})

	t.Run("invalid global log level loaded from env", func(t *testing.T) {
		// Given that APP12X_LOG_LEVEL env var is set to an invalod value
		prefix := app.EnvPrefix
		os.Setenv(prefix+"_LOG_LEVEL", "invalid value")

		defer func() {
			os.Unsetenv(prefix + "_LOG_LEVEL")
		}()

		buf := new(bytes.Buffer)
		var ID app.ID
		var ReleaseID app.ReleaseID
		var InstanceID app.InstanceID
		const FooEvent = "01DG3P8XM1PAMV6B8XMSDX0QSH"
		a := fx.New(
			app.Module(app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			}),
			fx.Invoke(func(logger app.Logger) {
				// And a debug event is logged
				event := logger(FooEvent, zerolog.DebugLevel)
				event(nil, "bar")
			}),
			fx.Populate(&ID, &ReleaseID, &InstanceID),
		)

		assert.Error(t, a.Err())
	})

	t.Run("global log level is specified explicitly on Opts", func(t *testing.T) {
		buf := new(bytes.Buffer)
		var ID app.ID
		var ReleaseID app.ReleaseID
		var InstanceID app.InstanceID
		const FooEvent = "01DG3P8XM1PAMV6B8XMSDX0QSH"
		LogLevel := zerolog.DebugLevel
		a := fx.New(
			app.Module(app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),

				LogWriter: buf,
				// Given that the global log level is set
				GlobalLogLevel: &LogLevel,
			}),
			fx.Invoke(func(logger app.Logger) {
				// And a debug event is logged
				event := logger(FooEvent, zerolog.DebugLevel)
				event(nil, "bar")
			}),
			fx.Populate(&ID, &ReleaseID, &InstanceID),
		)

		assert.NoError(t, a.Err())
		// When the app is initialized
		// Then the zerolog global log level should be set to debug
		assert.Equal(t, zerolog.GlobalLevel(), zerolog.DebugLevel)

		// {"a":"01DG138TTVDX5JH5F4GMNC3V67","r":"01DG138TTVK4MVW3B5TJGDSKHR","x":"01DG138TTVYGSN7QWBFT9660SS","n":"foo","z":"01DG138TTVBHCXQW29QTQAWPNM","t":1563405085,"m":"bar"}
		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"i"`
		}

		r := bufio.NewReader(buf)
		var logEvent LogEvent
		// Then the debug event was logged
		fooEventLogged := false
		for {
			line, err := r.ReadString('\n')
			t.Log(line)
			if err != nil {
				break
			}
			assert.NoError(t, json.Unmarshal([]byte(line), &logEvent), "failed to parse line: %s", line)
			if logEvent.Name == FooEvent {
				fooEventLogged = true
				break
			}
		}
		assert.True(t, fooEventLogged, "event should have been logged at debug level")
	})

}
