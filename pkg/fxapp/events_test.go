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
	"errors"
	"github.com/oysterpack/partire-k8s/pkg/fxapp"
	"github.com/oysterpack/partire-k8s/pkg/ulidgen"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
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

type LogHealthCheckFailure func(event HealthCheckFailure, tags ...string)

func NewHealthCheckFailure(logger *zerolog.Logger) LogHealthCheckFailure {
	const HealthCheckEventID fxapp.EventTypeID = "01DE2Z4E07E4T0GJJXCG8NN6A0"
	logEvent := HealthCheckEventID.NewLogEventer(logger, zerolog.ErrorLevel)
	return func(event HealthCheckFailure, tags ...string) {
		logEvent(&event, "healthcheck failed", tags...)
	}
}

func TestDomainEvent(t *testing.T) {
	buf := new(bytes.Buffer)
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

		t.Logf("\n%s", buf)
	}
}
