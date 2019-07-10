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
	"context"
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/fxapptest"
	"github.com/oysterpack/partire-k8s/pkg/health"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"strings"
	"testing"
	"time"
)

// The app automatically provides health.Registry and health.Scheduler.
func TestAppHealthCheckRegistry(t *testing.T) {
	t.Parallel()
	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()

	var healthCheckRegistry health.Registry
	var healthCheckScheduler health.Scheduler
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		Invoke(func(registry health.Registry, logger *zerolog.Logger) {
			FooHealth := health.CheckOpts{
				Desc:        FooHealthDesc,
				ID:          ulidgen.MustNew().String(),
				Description: "Foo",
				RedImpact:   "fatal",
				Checker: func(ctx context.Context) health.Failure {
					return nil
				},
			}.MustNew()

			logger.Info().Msg(FooHealth.String())

			registry.Register(FooHealth)
		}).
		Populate(&healthCheckRegistry, &healthCheckScheduler).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** app failed to build: %v", err)
	}

	// health checks are scheduled to run as they are registered

	healthChecks := healthCheckRegistry.HealthChecks(nil)
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

	// Then the health.Scheduler instance is shutdown
	select {
	case <-healthCheckScheduler.Done():
	case <-time.After(time.Second): // health check scheduler is stopped async - thus, it may not have yet completed shutdown
		t.Errorf("*** health check scheduler has not been shutdown: stopping = %v", healthCheckScheduler.Stopping())
	}

}

func TestRegisteredHealthChecksAreLogged(t *testing.T) {
	t.Parallel()
	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()
	healthCheckID := ulidgen.MustNew()

	var healthCheckRegistry health.Registry
	var healthCheckRegistered <-chan health.Check
	buf := fxapptest.NewSyncLog()
	_, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		Invoke(func(registry health.Registry) {
			healthCheckRegistered = registry.Subscribe()
			FooHealth := health.CheckOpts{
				Desc:         FooHealthDesc,
				ID:           healthCheckID.String(),
				Description:  "Foo",
				RedImpact:    "fatal",
				YellowImpact: "yellow",
				Checker: func(ctx context.Context) health.Failure {
					return nil
				},
			}.MustNew()

			registry.Register(FooHealth)
		}).
		Populate(&healthCheckRegistry).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** failed to build app: %v", err)
	}

	t.Log(<-healthCheckRegistered)

	type Data struct {
		ID           string
		DescID       string `json:"desc_id"`
		Description  []string
		YellowImpact []string `json:"yellow_impact"`
		RedImpact    []string `json:"red_impact"`
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
		if logEvent.Data.ID != healthCheckID.String() {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, healthCheckID)
		}

		if logEvent.Data.DescID != FooHealthDesc.ID().String() {
			t.Errorf("*** health check desc ID did not match: %v != %v", logEvent.Data.DescID, FooHealthDesc.ID())
		}

		if len(logEvent.Data.Description) != 2 &&
			logEvent.Data.Description[0] != FooHealthDesc.Description() &&
			logEvent.Data.Description[1] != "Foo" {
			t.Error("*** health check description did not match")
		}
		if len(logEvent.Data.YellowImpact) != 1 &&
			logEvent.Data.YellowImpact[0] != FooHealthDesc.YellowImpact() {
			t.Error("*** health check yellow impact did not match")
		}
		if len(logEvent.Data.RedImpact) != 2 &&
			logEvent.Data.RedImpact[0] != FooHealthDesc.RedImpact() &&
			logEvent.Data.RedImpact[1] != "fatal" {
			t.Error("*** health check red impact did not match")
		}
	default:
		t.Error("*** health check registration event was not logged")
	}
}

func TestHealthCheckResultsAreLogged(t *testing.T) {
	t.Parallel()
	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()
	healthCheckID := ulidgen.MustNew()

	var healthCheckRegistry health.Registry
	var healthCheckResults <-chan health.Result
	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		Invoke(func(registry health.Registry, scheduler health.Scheduler) {
			healthCheckResults = scheduler.Subscribe(nil)
			FooHealth := health.CheckOpts{
				Desc:         FooHealthDesc,
				ID:           healthCheckID.String(),
				Description:  "Foo",
				RedImpact:    "fatal",
				YellowImpact: "yellow",
				Checker: func(ctx context.Context) health.Failure {
					time.Sleep(time.Millisecond)
					return nil
				},
			}.MustNew()

			registry.Register(FooHealth)
		}).
		Populate(&healthCheckRegistry).
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
		if logEvent.Data.ID != healthCheckID.String() {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, healthCheckID)
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
	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()
	healthCheckID := ulidgen.MustNew()

	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
		LogWriter(buf).
		Invoke(func(registry health.Registry, scheduler health.Scheduler) {
			FooHealth := health.CheckOpts{
				Desc:         FooHealthDesc,
				ID:           healthCheckID.String(),
				Description:  "Foo",
				RedImpact:    "fatal",
				YellowImpact: "yellow",
				Checker: func(ctx context.Context) health.Failure {
					return health.YellowFailure(errors.New("yellow"))
				},
			}.MustNew()

			registry.Register(FooHealth)
		}).
		DisableHTTPServer().
		Build()

	if err != nil {
		t.Errorf("*** failed to build app: %v", err)
	}

	err = app.Run()
	if err == nil {
		t.Error("*** app should have failed to startup because of health check failure")
		return
	}
	t.Log(err)

}
