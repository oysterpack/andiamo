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
	"context"
	"encoding/json"
	"errors"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/fxapptest"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"strings"
	"testing"
)

type HealthCheckFailure struct {
	healthCheckID string
	err           error
}

func (event *HealthCheckFailure) MarshalZerologObject(e *zerolog.Event) {
	e.Str("id", event.healthCheckID).
		Err(event.err)
}

func TestDomainEvent(t *testing.T) {
	type LogHealthCheckFailure func(event HealthCheckFailure, tags ...string)

	const HealthCheckEventID fxapp.EventTypeID = "01DE2Z4E07E4T0GJJXCG8NN6A0"

	NewHealthCheckFailure := func(logger *zerolog.Logger) LogHealthCheckFailure {
		logEvent := HealthCheckEventID.NewLogEventer(logger, zerolog.ErrorLevel)
		return func(event HealthCheckFailure, tags ...string) {
			logEvent(&event, "healthcheck failed", tags...)
		}
	}

	buf := fxapptest.NewSyncLog()
	app, err := fxapp.NewBuilder(newDesc("foo", "0.1.0")).
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

		type Data struct {
			Err string `json:"e"`
		}

		type LogEvent struct {
			Level   string   `json:"l"`
			Name    string   `json:"n"`
			Message string   `json:"m"`
			Tags    []string `json:"g"`
			Data    Data     `json:"01DE2Z4E07E4T0GJJXCG8NN6A0"`
		}
		var logEvent LogEvent
		for _, line := range strings.Split(buf.String(), "\n") {
			if line == "" {
				break
			}
			err := json.Unmarshal([]byte(line), &logEvent)
			if err != nil {
				t.Errorf("*** failed to parse log event: %v", err)
				break
			}
			if logEvent.Name == HealthCheckEventID.String() {
				t.Log(line)
				break
			}
		}
		switch {
		case logEvent.Name == HealthCheckEventID.String():
			if logEvent.Level != "error" {
				t.Errorf("*** level did not match: %v", logEvent.Level)
			}
			if logEvent.Data.Err != "failure to connect" {
				t.Error("*** event data was not logged")
			}
			if len(logEvent.Tags) != 2 {
				t.Errorf("*** tags were not logged: %v", logEvent.Tags)
			} else {
				if logEvent.Tags[0] != "tag-a" && logEvent.Tags[1] != "tag-b" {
					t.Errorf("*** tags don't match: %v", logEvent.Tags)
				}
			}

		default:
			t.Error("*** event was not logged")
		}
	}
}
