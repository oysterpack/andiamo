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
	"context"
	"encoding/json"
	"errors"
	"github.com/oysterpack/partire-k8s/pkg/fxapp/health"
	"github.com/oysterpack/partire-k8s/pkg/fxapptest"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"sync"
	"testing"
	"time"
)

func TestLogYellowHealthCheckResult(t *testing.T) {
	t.Parallel()

	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()
	FooHealth := health.NewBuilder(FooHealthDesc, ulidgen.MustNew()).
		Description("Foo").
		RedImpact("fatal").
		Checker(func(ctx context.Context) health.Failure {
			time.Sleep(time.Millisecond)
			return health.YellowFailure(errors.New("warning"))
		}).
		MustBuild()

	registry := health.NewRegistry()
	scheduler := health.StartScheduler(registry)
	healthCheckResults := scheduler.Subscribe(nil)
	defer scheduler.StopAsync()

	buf := fxapptest.NewSyncLog()
	logger := zerolog.New(zerolog.SyncWriter(buf))
	done := make(chan struct{})

	f := startHealthCheckLoggerFunc(scheduler.Subscribe(nil), &logger, done)

	// wait until the health check logger routine is running
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		f()
	}()
	wg.Wait()

	// When the health check is registered, it is run
	registry.Register(FooHealth)
	t.Log(<-healthCheckResults)

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
		Data    Data   `json:"01DF3X60Z7XFYVVXGE9TFFQ7Z1"`
	}
	var logEvent LogEvent

	// Then the health check is logged with a warn log level
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
		if logEvent.Data.ID != FooHealth.ID().String() {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, FooHealth.ID())
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

	FooHealthDesc := health.NewDescBuilder(ulidgen.MustNew()).
		Description("Foo").
		YellowImpact("app response times are slow").
		RedImpact("app is unavailable").
		MustBuild()
	FooHealth := health.NewBuilder(FooHealthDesc, ulidgen.MustNew()).
		Description("Foo").
		RedImpact("fatal").
		Checker(func(ctx context.Context) health.Failure {
			time.Sleep(time.Millisecond)
			return health.RedFailure(errors.New("error"))
		}).
		MustBuild()

	registry := health.NewRegistry()
	scheduler := health.StartScheduler(registry)
	healthCheckResults := scheduler.Subscribe(nil)
	defer scheduler.StopAsync()

	buf := fxapptest.NewSyncLog()
	logger := zerolog.New(zerolog.SyncWriter(buf))
	done := make(chan struct{})

	f := startHealthCheckLoggerFunc(scheduler.Subscribe(nil), &logger, done)

	// wait until the health check logger routine is running
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		f()
	}()
	wg.Wait()

	// When the health check is registered, it is run
	registry.Register(FooHealth)
	t.Log(<-healthCheckResults)

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
		Data    Data   `json:"01DF3X60Z7XFYVVXGE9TFFQ7Z1"`
	}
	var logEvent LogEvent

	// Then the health check is logged with a warn log level
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
		if logEvent.Data.ID != FooHealth.ID().String() {
			t.Errorf("*** health check ID did not match: %v != %v", logEvent.Data.ID, FooHealth.ID())
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
