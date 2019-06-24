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

package fxapp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"strings"
	"testing"
	"time"
)

type HealthCheckFailure struct {
	healthCheckID string
	err           error
}

func (event *HealthCheckFailure) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", event.healthCheckID).
		Err(event.err)
}

type LogHealthCheckFailure func(event HealthCheckFailure, tags ...string)

func NewHealthCheckFailure(logger *zerolog.Logger) LogHealthCheckFailure {
	logEvent := fxapp.NewLogEventFunc(logger, zerolog.ErrorLevel, "01DE2Z4E07E4T0GJJXCG8NN6A0")
	return func(event HealthCheckFailure, tags ...string) {
		logEvent(&event, "healthcheck failed", tags...)
	}
}

func TestDomainEvent(t *testing.T) {
	buf := new(bytes.Buffer)
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		Provide(NewHealthCheckFailure).
		Invoke(func(lc fx.Lifecycle, logHealthCheckFailure LogHealthCheckFailure) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					healthCheckFailure := HealthCheckFailure{
						ulidgen.MustNew().String(),
						errors.New("failure to connect"),
					}
					logHealthCheckFailure(healthCheckFailure, "tag-a", "tag-b")
					return nil
				},
			})
		}).
		Build()

	switch {
	case err != nil:
		t.Errorf("*** app build failure: %v", err)
	default:
		go app.Run()
		<-app.Started()
		app.Shutdown()
		<-app.Done()

		t.Logf("\n%s", buf)
	}
}

func TestAppInitializedEventLogged(t *testing.T) {
	type Foo struct{}

	type Run func()

	buf := new(bytes.Buffer)
	_, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		SetStopTimeout(time.Minute).
		Provide(func() Foo { return Foo{} }).
		Invoke(func() {}).
		Build()

	switch {
	case err != nil:
		t.Errorf("** app build failed: %v", err)
	default:
		t.Logf("\n%v", buf)

		type Data struct {
			StartTimeout uint `json:"start_timeout"`
			StopTimeout  uint `json:"stop_timeout"`
			Provides     []string
			Invokes      []string
		}

		type LogEvent struct {
			Name    string `json:"n"`
			Message string `json:"m"`
			Data    Data   `json:"01DE4STZ0S24RG7R08PAY1RQX3"`
		}

		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				continue
			}
			if logEvent.Name == fxapp.AppInitializedEventID {
				break
			}
		}
		switch {
		case logEvent.Name == fxapp.AppInitializedEventID:
			if logEvent.Message != "app initialized" {
				t.Errorf("*** event message did not match: %v", logEvent.Message)
			}

			if logEvent.Data.StartTimeout*uint(time.Millisecond) != uint(fx.DefaultTimeout) {
				t.Errorf("*** start timeout did not match: %v", logEvent.Data.StartTimeout)
			}

			if logEvent.Data.StopTimeout*uint(time.Millisecond) != uint(time.Minute) {
				t.Errorf("*** stop timeout did not match: %v", logEvent.Data.StartTimeout)
			}
			if len(logEvent.Data.Provides) != 1 {
				t.Errorf("*** provides does not match: %v", logEvent.Data.Provides)
			}
			if len(logEvent.Data.Invokes) != 1 {
				t.Errorf("*** inokes does not match: %v", logEvent.Data.Invokes)
			}

		default:
			t.Error("*** app initialization event was not logged")
		}
	}
}

func TestAppStartingEventLogged(t *testing.T) {
	type Foo struct{}

	type Run func()

	buf := new(bytes.Buffer)
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		SetStopTimeout(time.Minute).
		Provide(func() Foo { return Foo{} }).
		Invoke(func() {}).
		Build()

	switch {
	case err != nil:
		t.Errorf("** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Started()
		app.Shutdown()
		<-app.Done()

		t.Logf("\n%v", buf)

		type LogEvent struct {
			Name    string `json:"n"`
			Message string `json:"m"`
		}

		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" {
				break
			}
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				continue
			}
			if logEvent.Name == fxapp.AppStartingEventID {
				break
			}
		}
		switch {
		case logEvent.Name == fxapp.AppStartingEventID:
			if logEvent.Message != "app starting" {
				t.Errorf("*** event message did not match: %v", logEvent.Message)
			}
		default:
			t.Error("*** app starting event was not logged")
		}

	}
}

func TestAppStartedEventLogged(t *testing.T) {
	type Foo struct{}

	type Run func()

	buf := new(bytes.Buffer)
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		SetStopTimeout(time.Minute).
		Provide(func() Foo { return Foo{} }).
		Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					// simulate some startup work that consumes some time
					time.Sleep(time.Millisecond)
					return nil
				},
			})
		}).
		Build()

	switch {
	case err != nil:
		t.Errorf("** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Started()
		app.Shutdown()
		<-app.Done()

		t.Logf("\n%v", buf)

		type Data struct {
			Duration uint
		}

		type LogEvent struct {
			Name    string `json:"n"`
			Message string `json:"m"`
			Data    Data   `json:"01DE4X10QCV1M8TKRNXDK6AK7C"`
		}

		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" {
				break
			}
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				continue
			}
			if logEvent.Name == fxapp.AppStartedEventID {
				break
			}
		}
		switch {
		case logEvent.Name == fxapp.AppStartedEventID:
			if logEvent.Message != "app started" {
				t.Errorf("*** event message did not match: %v", logEvent.Message)
			}

			if logEvent.Data.Duration == 0 {
				t.Error("*** duration was not logged")
			}
		default:
			t.Error("*** app started event was not logged")
		}

	}
}

func TestAppStoppingEventLogged(t *testing.T) {
	type Foo struct{}

	type Run func()

	buf := new(bytes.Buffer)
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		SetStopTimeout(time.Minute).
		Provide(func() Foo { return Foo{} }).
		Invoke(func() {}).
		Build()

	switch {
	case err != nil:
		t.Errorf("** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Started()
		app.Shutdown()
		<-app.Done()

		t.Logf("\n%v", buf)

		type LogEvent struct {
			Name    string `json:"n"`
			Message string `json:"m"`
		}

		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" {
				break
			}
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				continue
			}
			if logEvent.Name == fxapp.AppStoppingEventID {
				break
			}
		}
		switch {
		case logEvent.Name == fxapp.AppStoppingEventID:
			if logEvent.Message != "app stopping" {
				t.Errorf("*** event message did not match: %v", logEvent.Message)
			}
		default:
			t.Error("*** app stopping event was not logged")
		}

	}
}

func TestAppStoppedEventLogged(t *testing.T) {
	type Foo struct{}

	type Run func()

	buf := new(bytes.Buffer)
	app, err := fxapp.NewAppBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		SetStopTimeout(time.Minute).
		Provide(func() Foo { return Foo{} }).
		Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{
				OnStop: func(context.Context) error {
					// simulate some work that consumes some time
					time.Sleep(time.Millisecond)
					return nil
				},
			})
		}).
		Build()

	switch {
	case err != nil:
		t.Errorf("** app build failed: %v", err)
	default:
		go app.Run()
		<-app.Started()
		app.Shutdown()
		<-app.Done()

		t.Logf("\n%v", buf)

		type Data struct {
			Duration uint
		}

		type LogEvent struct {
			Name    string `json:"n"`
			Message string `json:"m"`
			Data    Data   `json:"01DE4T1V9N50BB67V424S6MG5C"`
		}

		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" {
				break
			}
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				continue
			}
			if logEvent.Name == fxapp.AppStoppedEventID {
				break
			}
		}
		switch {
		case logEvent.Name == fxapp.AppStoppedEventID:
			if logEvent.Message != "app stopped" {
				t.Errorf("*** event message did not match: %v", logEvent.Message)
			}

			if logEvent.Data.Duration == 0 {
				t.Error("*** duration was not logged")
			}
		default:
			t.Error("*** app stopped event was not logged")
		}

	}
}
