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
	"github.com/oysterpack/andiamo/pkg/fx/app"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"runtime"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	buf := new(bytes.Buffer)
	a := app.New(
		app.Opts{
			ID:        ulids.MustNew(),
			ReleaseID: ulids.MustNew(),
			LogWriter: buf,
		},
		fx.Invoke(
			func(logger app.Logger) {
				event := logger("TestNewApp", zerolog.InfoLevel)
				event(nil, "CIAO MUNDO!!!")
			},
		),
	)

	assert.NoError(t, a.Err())
	assert.NoError(t, a.Start(context.Background()))
	assert.NoError(t, a.Stop(context.Background()))

	type Data struct {
		Duration uint64
	}

	// {"a":"01DG138TTVDX5JH5F4GMNC3V67","r":"01DG138TTVK4MVW3B5TJGDSKHR","x":"01DG138TTVYGSN7QWBFT9660SS","n":"foo","z":"01DG138TTVBHCXQW29QTQAWPNM","t":1563405085,"m":"bar"}
	type LogEvent struct {
		Level   string `json:"l"`
		Name    string `json:"n"`
		Message string `json:"m"`

		AppID        string `json:"a"`
		AppReleaseID string `json:"r"`
		InstanceID   string `json:"x"`

		Data `json:"d"`
	}

	expectedEvents := map[string]struct{}{
		app.InitializedEvent: struct{}{},
		app.StartingEvent:    struct{}{},
		app.StartedEvent:     struct{}{},
		app.StoppingEvent:    struct{}{},
		app.StoppedEvent:     struct{}{},
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

		delete(expectedEvents, logEvent.Name)
	}

	assert.Empty(t, expectedEvents)
}

func TestApp_Go(t *testing.T) {
	t.Parallel()

	shutdown := func(t *testing.T, shutdowner fx.Shutdowner, done <-chan error, expectingStopError bool) {
	Shutdown:
		for i := 0; i < 10; i++ {
			shutdowner.Shutdown()
			select {
			case <-time.After(time.Millisecond):
				// if shutdown is called before the app has completed start up, then shutdown is not triggered
				// thus keep retrying until the app has started
				runtime.Gosched()
			case err := <-done:
				if !expectingStopError {
					assert.NoError(t, err)
				}

				break Shutdown
			}
		}
	}

	t.Run("run a simple app", func(t *testing.T) {
		t.Parallel()
		buf := new(bytes.Buffer)
		shutdowner, done, err := app.Go(
			app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			},
			fx.Invoke(
				func(logger app.Logger) {
					event := logger("TestNewApp", zerolog.InfoLevel)
					event(nil, "CIAO MUNDO!!!")
				},
			),
		)

		assert.NoError(t, err)
		shutdown(t, shutdowner, done, false)

		type Data struct {
			Duration uint64
		}

		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"x"`

			Data `json:"d"`
		}

		expectedEvents := map[string]struct{}{
			app.InitializedEvent: struct{}{},
			app.StartingEvent:    struct{}{},
			app.StartedEvent:     struct{}{},
			app.StoppingEvent:    struct{}{},
			app.StoppedEvent:     struct{}{},
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

			delete(expectedEvents, logEvent.Name)
		}

		assert.Empty(t, expectedEvents)
	})

	t.Run("run an app that fails to initialize", func(t *testing.T) {
		t.Parallel()
		buf := new(bytes.Buffer)
		shutdowner, done, err := app.Go(
			app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			},
			fx.Invoke(
				func() error {
					return errors.New("init failed")
				},
			),
		)

		assert.Error(t, err)
		t.Log(err)
		assert.Nil(t, shutdowner, "shutdowner should be nil")
		select {
		case err := <-done:
			assert.Error(t, err)
		default:
			assert.Fail(t, "done chan should have returned an error")
		}

		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"x"`
		}

		expectedEvents := map[string]struct{}{
			app.InitFailedEvent: struct{}{},
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

			delete(expectedEvents, logEvent.Name)
		}
		assert.Empty(t, expectedEvents)
	})

	t.Run("run an app that fails on start up", func(t *testing.T) {
		t.Parallel()
		buf := new(bytes.Buffer)
		_, done, err := app.Go(
			app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			},
			fx.Invoke(
				func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: func(context.Context) error {
							return errors.New("start failure")
						},
					})
				},
			),
		)

		assert.NoError(t, err)
		<-done // app should fail to start

		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"x"`
		}

		expectedEvents := map[string]struct{}{
			app.InitializedEvent: struct{}{},
			app.StartingEvent:    struct{}{},
			app.StartFailedEvent: struct{}{},
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

			delete(expectedEvents, logEvent.Name)
		}
		assert.Empty(t, expectedEvents)
	})

	t.Run("run an app that errors on shutdown", func(t *testing.T) {
		t.Parallel()
		buf := new(bytes.Buffer)
		shutdowner, done, err := app.Go(
			app.Opts{
				ID:        ulids.MustNew(),
				ReleaseID: ulids.MustNew(),
				LogWriter: buf,
			},
			fx.Invoke(
				func(lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStop: func(context.Context) error {
							return errors.New("stop failure")
						},
					})
				},
			),
		)

		assert.NoError(t, err)
		shutdown(t, shutdowner, done, true)

		type LogEvent struct {
			Level   string `json:"l"`
			Name    string `json:"n"`
			Message string `json:"m"`

			AppID        string `json:"a"`
			AppReleaseID string `json:"r"`
			InstanceID   string `json:"x"`
		}

		expectedEvents := map[string]struct{}{
			app.InitializedEvent: struct{}{},
			app.StartingEvent:    struct{}{},
			app.StartedEvent:     struct{}{},
			app.StoppingEvent:    struct{}{},
			app.StopFailedEvent:  struct{}{},
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

			delete(expectedEvents, logEvent.Name)
		}
		assert.Empty(t, expectedEvents)
	})

}
