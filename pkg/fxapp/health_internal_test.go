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

package fxapp

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/oysterpack/andiamo/pkg/fx/health"
	"github.com/oysterpack/andiamo/pkg/fxapptest"
	"github.com/oysterpack/andiamo/pkg/ulids"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestLogYellowHealthCheckResult(t *testing.T) {
	t.Parallel()

	FooHealth := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo",
		YellowImpact: "app response times are slow",
		RedImpact:    "app is unavailable",
	}

	var shutdowner fx.Shutdowner
	var subscription health.CheckResultsSubscription
	var healthCheckResults health.CheckResults
	app := fx.New(
		health.Module(health.DefaultOpts()),
		fx.Invoke(
			func(subscribe health.SubscribeForCheckResults) {
				subscription = subscribe(func(result health.Result) bool {
					return result.ID == FooHealth.ID
				})
			},
			func(register health.Register) error {
				return register(FooHealth, health.CheckerOpts{}, func() (health.Status, error) {
					time.Sleep(time.Millisecond)
					return health.Yellow, errors.New("warning")
				})
			}),
		fx.Populate(&shutdowner, &healthCheckResults),
	)

	defer func() {
		go app.Run()
		shutdowner.Shutdown()
	}()

	buf := fxapptest.NewSyncLog()
	logger := zerolog.New(zerolog.SyncWriter(buf))
	done := make(chan struct{})
	defer close(done)

	f := startHealthCheckLoggerFunc(subscription, &logger, done)

	// wait until the health check logger routine is running
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		f()
	}()
	wg.Wait()

	t.Log(<-healthCheckResults(nil))

	type Data struct {
		ID     string
		Status uint8
		Start  uint
		Dur    uint
		Err    string `json:"e"`
	}

	type LogEvent struct {
		Level   string `json:"l"`
		Name    string `json:"n"`
		Message string `json:"m"`
		Data    Data   `json:"d"`
	}
	var logEvent LogEvent

	// Then the health check is logged with a warn log level
FoundEvent:
	for {
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
		runtime.Gosched()
	}

	switch {
	case logEvent.Name == "01DF3X60Z7XFYVVXGE9TFFQ7Z1":
		t.Logf("%#v", logEvent)
		if logEvent.Data.ID != FooHealth.ID {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, FooHealth.ID)
		}
		if logEvent.Data.Status != uint8(health.Yellow) {
			t.Error("*** status did not match")
		}
		if logEvent.Data.Start == 0 {
			t.Error("*** start should be set")
		}
		if logEvent.Data.Dur == 0 {
			t.Error("*** duration should be set")
		}
		if logEvent.Data.Err == "" {
			t.Error("*** error should be set")
		}
		if logEvent.Level != "warn" {
			t.Error("*** level should be warning")
		}

	default:
		t.Error("*** health check registration event was not logged")
	}
}

func TestLogRedHealthCheckResult(t *testing.T) {
	t.Parallel()

	FooHealth := health.Check{
		ID:           ulids.MustNew().String(),
		Description:  "Foo",
		YellowImpact: "app response times are slow",
		RedImpact:    "app is unavailable",
	}

	var shutdowner fx.Shutdowner
	var subscription health.CheckResultsSubscription
	var healthCheckResults health.CheckResults
	app := fx.New(
		health.Module(health.DefaultOpts()),
		fx.Invoke(
			func(subscribe health.SubscribeForCheckResults) {
				subscription = subscribe(func(result health.Result) bool {
					return result.ID == FooHealth.ID
				})
			},
			func(register health.Register) error {
				return register(FooHealth, health.CheckerOpts{}, func() (health.Status, error) {
					time.Sleep(time.Millisecond)
					return health.Red, errors.New("error")
				})
			}),
		fx.Populate(&shutdowner, &healthCheckResults),
	)

	defer func() {
		go app.Run()
		shutdowner.Shutdown()
	}()

	buf := fxapptest.NewSyncLog()
	logger := zerolog.New(zerolog.SyncWriter(buf))
	done := make(chan struct{})
	defer close(done)

	f := startHealthCheckLoggerFunc(subscription, &logger, done)

	// wait until the health check logger routine is running
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		f()
	}()
	wg.Wait()

	// When the health check is registered, it is run
	t.Log(<-healthCheckResults(nil))

	type Data struct {
		ID     string
		Status uint8
		Start  uint
		Dur    uint
		Err    string `json:"e"`
	}

	type LogEvent struct {
		Level   string `json:"l"`
		Name    string `json:"n"`
		Message string `json:"m"`
		Data    Data   `json:"d"`
	}
	var logEvent LogEvent

	// Then the health check is logged with a warn log level
FoundEvent:
	for {
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
		runtime.Gosched()
	}

	switch {
	case logEvent.Name == "01DF3X60Z7XFYVVXGE9TFFQ7Z1":
		t.Logf("%#v", logEvent)
		if logEvent.Data.ID != FooHealth.ID {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, FooHealth.ID)
		}
		if logEvent.Data.Status != uint8(health.Red) {
			t.Error("*** status did not match")
		}
		if logEvent.Data.Start == 0 {
			t.Error("*** start should be set")
		}
		if logEvent.Data.Dur == 0 {
			t.Error("*** duration should be set")
		}
		if logEvent.Data.Err == "" {
			t.Error("*** error should be set")
		}
		if logEvent.Level != "error" {
			t.Error("*** level should be error")
		}

	default:
		t.Error("*** health check registration event was not logged")
	}
}
