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

package logging_test

import (
	"encoding/json"
	"github.com/oysterpack/partire-k8s/pkg/app"
	"github.com/oysterpack/partire-k8s/pkg/app/logging"
	"github.com/oysterpack/partire-k8s/pkg/apptest"
	"github.com/pkg/errors"
	"testing"
)

const PACKAGE app.Package = "github.com/oysterpack/partire-k8s/pkg/log_test"

func TestPackageLogger(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)

	t.Run("log error with stacktrace", func(t *testing.T) {
		// in order for the stack to be logged, Stack() needs to be called before Err()
		logger.Error().Stack().Err(errors.New("BOOM!!!")).Msg("")
		t.Logf("error log event: %s", logger.Buf.String())

		var logEvent apptest.LogEvent
		e := json.Unmarshal([]byte(logger.Buf.String()), &logEvent)
		if e != nil {
			t.Fatalf("Failed to unmarshal log event as JSON: %v", e)
		}
		if len(logEvent.Stack) == 0 {
			t.Error("The stacktrace should have been logged")
		}
	})
}

func TestComponentLogger(t *testing.T) {
	logger := apptest.NewTestLogger(PACKAGE)
	const Foo = "foo"
	compLogger := logging.ComponentLogger(logger.Logger, Foo)

	compLogger.Info().Msg("")
	t.Log(logger.Buf.String())
	var logEvent apptest.LogEvent
	e := json.Unmarshal([]byte(logger.Buf.String()), &logEvent)
	if e != nil {
		t.Fatalf("Failed to unmarshal log event as JSON: %v", e)
	}
	if logEvent.Component != Foo {
		t.Errorf("log event component name did not match: %v", logEvent.Component)
	}
}

func TestField_String(t *testing.T) {
	if logging.ID.String() != string(logging.ID) {
		t.Errorf("Field.String() should simply return the Field as a string: %s", logging.ID.String())
	}
}
