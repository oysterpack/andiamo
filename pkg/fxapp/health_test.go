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
	"bufio"
	"encoding/json"
	"github.com/oysterpack/andiamo/pkg/fx/health"
	"github.com/oysterpack/andiamo/pkg/fxapp"
	"github.com/oysterpack/andiamo/pkg/fxapptest"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strings"
	"testing"
	"time"
)

// The app automatically provides health.registry and health.Scheduler.
func TestAppHealthCheckRegistry(t *testing.T) {
	t.Parallel()

	var registeredChecks health.RegisteredChecks
	app, err := fxapp.NewBuilder(fxapp.ID(ulids.MustNew()), fxapp.ReleaseID(ulids.MustNew())).
		Invoke(func(register health.Register, logger *zerolog.Logger) error {
			return register(health.Check{
				ID:          ulids.MustNew().String(),
				Description: "Foo",
				RedImpact:   "Red",
			}, health.CheckerOpts{}, func() error {
				return nil
			})
		}).
		Populate(&registeredChecks).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** app failed to build: %v", err)
	}

	// health checks are scheduled to run as they are registered

	healthChecks := <-registeredChecks()
	if len(healthChecks) == 0 {
		t.Error("*** health check registry is empty")
		return
	}
	t.Log(healthChecks)

	go app.Run()
	<-app.Ready()

	// When the app is shutdown
	app.Shutdown()
	<-app.Done()
}

func TestRegisteredHealthChecksAreLogged(t *testing.T) {
	t.Parallel()
	Foo := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo",
		RedImpact:    "Red",
		YellowImpact: "Yellow",
	}

	var healthCheckRegistered <-chan health.RegisteredCheck
	buf := fxapptest.NewSyncLog()
	_, err := fxapp.NewBuilder(fxapp.ID(ulids.MustNew()), fxapp.ReleaseID(ulids.MustNew())).
		LogWriter(buf).
		Invoke(func(register health.Register, subscribe health.SubscribeForRegisteredChecks) {
			healthCheckRegistered = subscribe().Chan()
			register(Foo, health.CheckerOpts{}, func() error {
				return nil
			})
		}).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** failed to build app: %v", err)
	}

	t.Log(<-healthCheckRegistered)

	type Data struct {
		ID           string
		DescID       string `json:"desc_id"`
		Description  string
		YellowImpact string `json:"yellow_impact"`
		RedImpact    string `json:"red_impact"`
		Timeout      uint
		RunInterval  uint `json:"run_interval"`
	}

	type LogEvent struct {
		Name    string `json:"n"`
		Message string `json:"m"`
		Data    Data   `json:"d"`
	}
	var logEvent LogEvent

FoundEvent:
	for i := 0; i < 3; i++ {
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" {
				break
			}
			t.Log(line)
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				break
			}
			if logEvent.Name == "01DF3FV60A2J1WKX5NQHP47H61" {
				break FoundEvent
			}
		}
		time.Sleep(time.Millisecond)
	}

	switch {
	case logEvent.Name == "01DF3FV60A2J1WKX5NQHP47H61":
		t.Logf("%#v", logEvent)
		if logEvent.Data.ID != Foo.ID {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, Foo.ID)
		}
		if logEvent.Data.Description != Foo.Description {
			t.Error("*** health check description did not match")
		}
		if logEvent.Data.YellowImpact != Foo.YellowImpact {
			t.Error("*** health check yellow impact did not match")
		}
		if logEvent.Data.RedImpact != Foo.RedImpact {
			t.Error("*** health check red impact did not match")
		}
	default:
		t.Error("*** health check registration event was not logged")
	}
}

func TestHealthCheckResultsAreLogged(t *testing.T) {
	t.Parallel()

	Foo := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo",
		RedImpact:    "Red",
		YellowImpact: "Yellow",
	}

	var healthCheckResults <-chan health.Result
	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(fxapp.ID(ulids.MustNew()), fxapp.ReleaseID(ulids.MustNew())).
		LogWriter(buf).
		Invoke(func(register health.Register, subscribe health.SubscribeForCheckResults) error {
			healthCheckResults = subscribe(nil).Chan()
			return register(Foo, health.CheckerOpts{}, func() error {
				time.Sleep(time.Millisecond)
				return nil
			})
		}).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** failed to build app: %v", err)
	}

	go app.Run()
	defer func() {
		app.Shutdown()
		<-app.Done()
	}()
	<-app.Ready()

	t.Log(<-healthCheckResults)

	type Data struct {
		ID     string
		Status uint8
		Start  uint
		Dur    uint
	}

	type LogEvent struct {
		Name    string `json:"n"`
		Message string `json:"m"`
		Data    Data   `json:"d"`
	}
	var logEvent LogEvent

FoundEvent:
	for i := 0; i < 3; i++ {
		reader := bufio.NewReader(buf)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			t.Log(line)
			err = json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v : %v", err, line)
				break
			}
			if logEvent.Name == "01DF3X60Z7XFYVVXGE9TFFQ7Z1" {
				break FoundEvent
			}
		}
		time.Sleep(time.Millisecond)
	}

	switch {
	case logEvent.Name == "01DF3X60Z7XFYVVXGE9TFFQ7Z1":
		t.Logf("%#v", logEvent)
		if logEvent.Data.ID != Foo.ID {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, Foo.ID)
		}
		if logEvent.Data.Status != uint8(health.Green) {
			t.Error("*** status did not match")
		}
		if logEvent.Data.Start == 0 {
			t.Error("*** start should be set")
		}
		if logEvent.Data.Dur == 0 {
			t.Error("*** duration should be set")
		}

	default:
		t.Error("*** health check result event was not logged")
	}

}

func TestHealthCheckFailureCausesAppStartupFailure(t *testing.T) {
	t.Parallel()
	Foo := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo",
		RedImpact:    "Red",
		YellowImpact: "Yellow",
	}

	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(fxapp.ID(ulids.MustNew()), fxapp.ReleaseID(ulids.MustNew())).
		LogWriter(buf).
		Invoke(func(register health.Register) error {
			return register(Foo, health.CheckerOpts{}, func() error {
				return errors.New("BOOM!!!")
			})
		}).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** failed to build app: %v", err)
	}

	time.Sleep(time.Millisecond)

	err = app.Run()
	if err == nil {
		t.Error("*** app should have failed to startup because of health check failure")
		return
	}
	t.Log(err)

}
